//go:generate go run generate.go "../.."

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/docopt/docopt-go"
)

// Convert Dockerfile.template into version specific Dockerfile
// along with resources

type Edition string

const (
	EditionEnterprise = Edition("enterprise")
	EditionCommunity  = Edition("community")
)

type Product string

const (
	ProductServer  = Product("couchbase-server")
	ProductSyncGw  = Product("sync-gateway")
	ProductSandbox = Product("server-sandbox")
)

// These are Docker's idea of architecture names, eg. amd64, arm64.
// "Archgeneric" is for a filename with @@ARCH@@ in place of the
// actual architecture, which will be substituted at build time in
// the Dockerfile.
type Arch string

const (
	Archamd64   = Arch("amd64")
	Archarm64   = Arch("arm64")
	Archgeneric = Arch("@@ARCH@@")
)

// A map of "overrides" which specify custom package download urls and package names
// for unreleased or otherwise special version.
// Key format: $product_$edition_$version (eg, sync-gateway_community_2.0.0-latestbuild)
// Note: currently only implemented for sync gateway
type VersionCustomizations map[string]VersionCustomization

// Parameters that can be customized
type VersionCustomization struct {
	PackageUrl      string `json:"package_url"`
	PackageFilename string `json:"package_filename"`
}

// ProductVersionFilter is a map of Product to a regular expression that should match versions
// This can be used to exclude older versions from being updated.
// For an example of usage, see the Sync Gateway entries in init()
type ProductVersionFilter map[Product]*regexp.Regexp

// Matches returns true if the given product/version matched by the given filter
func (filter ProductVersionFilter) Matches(product Product, version string) bool {
	if r := filter[product]; r != nil && r.MatchString(version) {
		return true
	}
	return false
}

var (
	default_editions      []Edition
	default_products      []Product
	versionCustomizations VersionCustomizations
	baseDir               string
	skipGeneration        ProductVersionFilter
)

func init() {
	default_editions = []Edition{
		EditionCommunity,
		EditionEnterprise,
	}

	default_products = []Product{
		ProductServer,
		ProductSyncGw,
		ProductSandbox,
	}

	// TODO: Read the version_customizations.json file into map
	versionCustomizations = map[string]VersionCustomization{}
	versionCustomizations["sync-gateway_community_2.0.0-devbuild"] = VersionCustomization{
		PackageUrl:      "http://cbmobile-packages.s3.amazonaws.com/couchbase-sync-gateway-community_2.0.0-827_x86_64.rpm",
		PackageFilename: "couchbase-sync-gateway-community_2.0.0-827_x86_64.rpm",
	}
	versionCustomizations["sync-gateway_enterprise_2.0.0-devbuild"] = VersionCustomization{
		PackageUrl:      "http://cbmobile-packages.s3.amazonaws.com/couchbase-sync-gateway-enterprise_2.0.0-827_x86_64.rpm",
		PackageFilename: "couchbase-sync-gateway-enterprise_2.0.0-827_x86_64.rpm",
	}

	skipGeneration = ProductVersionFilter{
		ProductSyncGw: regexp.MustCompile(`^(1\.|2\.0\.).+$`), // 1.x and 2.0.x
	}
}

func main() {
	usage := `Dockerfile Generator

Usage:
  generate BASE_DIRECTORY -p PRODUCT -v VERSION -e EDITION -o DIR [ -t TEMPLATE_ARG ]...
  generate BASE_DIRECTORY

The first form generates a single Dockerfile and its associated resources
in the specified directory (which must exist). The second form will
search for directories under the specified directory with the form

    EDITION/PRODUCT/VERSION

and for each such directory that does not contain a Dockerfile, will
create the corresponding Dockerfile with its associated resources.

Arguments:
  BASE_DIRECTORY                  Root of "docker" repository

Options:
  -p PRODUCT, --product PRODUCT   Product name
  -v VERSION, --version VERSION   Product version
  -e EDITION, --edition EDITION   Product edition (community/enterprise)
  -o OUTPUT_DIRECTORY             Directory to write Dockerfile to
  -t TEMPLATE_ARG                 KEY=VALUE to provide to the template
  -h, --help                      Print this usage message
`

	args, _ := docopt.ParseDoc(usage)
	baseDir = args["BASE_DIRECTORY"].(string)

	if args["--product"] != nil {
		log.Println("Generating single product")
		generateOneDockerfile(
			Edition(args["--edition"].(string)),
			Product(args["--product"].(string)),
			args["--version"].(string),
			args["-o"].(string),
			generateOverrides(args["-t"].([]string)),
		)
	} else {
		log.Println("Generating multiple products")
		generateAllDockerfiles()
	}

	log.Printf("Successfully finished!")
}

func generateOverrides(args []string) (retval map[string]any) {

	retval = map[string]any{}
	for _, mapping := range args {
		vals := strings.Split(mapping, "=")
		if len(vals) != 2 {
			log.Fatalf("-t '%s' not of form KEY=VALUE", mapping)
		}
		retval[vals[0]] = vals[1]
	}

	return
}

func generateAllDockerfiles() {
	for _, edition := range default_editions {
		for _, product := range default_products {
			// find corresponding directory for this edition/product combo
			dir := path.Join(baseDir, string(edition), string(product))

			// find all version subdirectories (must match regex)
			versions := versionSubdirectories(dir)

			// for each version
			for _, ver := range versions {

				if skipGeneration.Matches(product, ver) {
					log.Printf("Skipping generation for %v %v %v", product, edition, ver)
					continue
				}
				generateOneDockerfile(edition, product, ver, "", nil)
			}
		}
	}
}

func generateOneDockerfile(
	edition Edition, product Product, ver string, outputDir string,
	overrides map[string]any,
) error {
	// Start with a basic DockerfileVariant, then tweak if necessary
	variant := DockerfileVariant{
		Edition:           edition,
		Product:           product,
		Version:           strings.TrimSuffix(ver, "-staging"),
		TargetVersion:     strings.TrimSuffix(ver, "-staging"),
		Arches:            []Arch{Archamd64},
		IsStaging:         strings.HasSuffix(ver, "-staging"),
		TemplateFilename:  "Dockerfile.template",
		OutputDir:         outputDir,
		TemplateOverrides: overrides,
	}

	productVer, _ := intVer(variant.Version)

	// Update according to special cases based on Product and Version.
	if product == ProductServer {
		if productVer == 70003 {
			// CBD-4603: 7.0.3 actually builds from 7.0.3-MP1 for complete
			// Log4Shell remediation
			variant.Version = "7.0.3-MP1"
		}

		if productVer >= 70100 {
			// 7.1.0 and higher also support arm64
			variant.Arches = append(variant.Arches, Archarm64)
		}
	} else if product == ProductSyncGw {
		if productVer <= 30003 {
			variant.TemplateFilename = "Dockerfile.centos.template"
		} else {
			variant.TemplateFilename = "Dockerfile.ubuntu.template"
			variant.Arches = append(variant.Arches, Archarm64)
		}
	} else if product == ProductSandbox {
		if productVer >= 71000 {
			// 7.1.0 and higher also support arm64
			variant.Arches = append(variant.Arches, Archarm64)
		}
	}

	// Now generate the Dockerfile(s) based on the constructed variant
	if err := generateVariant(variant); err != nil {
		log.Fatalf("Failed (%v/%v/%v): %v", edition, product, ver, err)
	}

	return nil
}

func generateVariant(variant DockerfileVariant) error {
	if _, err := os.Stat(variant.dockerfile()); !os.IsNotExist(err) {
		log.Printf("%s exists, not regenerating...", variant.dockerfile())
	} else {
		if err := generateDockerfile(variant); err != nil {
			return err
		}

		if err := deployScriptResources(variant); err != nil {
			return err
		}

		if err := deployConfigResources(variant); err != nil {
			return err
		}
	}

	// We always want to ensure the readme is updated, to avoid the current
	// description on docker hub being overwritten by legacy documentation.
	if err := deployReadme(variant); err != nil {
		return err
	}

	return nil
}

func generateDockerfile(variant DockerfileVariant) error {
	log.Printf("generateDockerfile called with: %v", variant)

	targetDir := variant.targetDir()
	log.Printf("targetDir: %v", targetDir)

	// figure out output filename
	targetDockerfile := variant.dockerfile()
	log.Printf("targetDockerfile: %v", targetDockerfile)

	// find the path to the source template
	sourceTemplate := path.Join(
		baseDir,
		"generate",
		"templates",
		string(variant.Product),
		string(variant.TemplateFilename),
	)

	log.Printf("template: %v", sourceTemplate)
	log.Printf("product: %v", variant.Product)
	var params map[string]any

	if variant.Product == ProductServer {
		// template parameters
		params = map[string]any{
			"CB_VERSION":         variant.VersionWithSubstitutions(),
			"CB_PACKAGE":         variant.serverPackageFile(Archgeneric),
			"CB_PACKAGE_NAME":    variant.serverPackageName(),
			"CB_EXTRA_DEPS":      variant.extraDependencies(),
			"CB_SHA256_arm64":    variant.getSHA256(Archarm64),
			"CB_SHA256_amd64":    variant.getSHA256(Archamd64),
			"CB_RELEASE_URL":     variant.releaseURL(),
			"DOCKER_BASE_IMAGE":  variant.dockerBaseImage(),
			"PKG_COMMAND":        variant.serverPkgCommand(),
			"SYSTEMD_WORKAROUND": variant.systemdWorkaround(),
			"CB_MULTIARCH":       len(variant.Arches) > 1,
			"CB_SKIP_CHECKSUM":   "false",
		}

	} else if variant.Product == ProductSyncGw {
		// template parameters
		params = map[string]any{
			"SYNC_GATEWAY_PACKAGE_URL":      variant.sgPackageUrl(),
			"SYNC_GATEWAY_PACKAGE_FILENAME": variant.sgPackageFilename(),
			"DOCKER_BASE_IMAGE":             variant.dockerBaseImage(),
		}

	} else if variant.Product == ProductSandbox {
		// template parameters
		params = map[string]any{
			"CB_VERSION":        variant.VersionWithSubstitutions(),
			"DOCKER_BASE_IMAGE": variant.dockerBaseImage(),
			"CB_MULTIARCH":      len(variant.Arches) > 1,
		}
	}

	// Apply any user-requested template overrides
	for key, value := range variant.TemplateOverrides {
		params[key] = value
	}

	// open a file at destPath
	out, err := os.Create(targetDockerfile)
	if err != nil {
		return err
	}
	defer out.Close()

	templateBytes, err := ioutil.ReadFile(sourceTemplate)
	if err != nil {
		return err
	}

	tmpl, err := template.New("docker").Parse(string(templateBytes))
	if err != nil {
		return err
	}
	err = tmpl.Execute(out, params)
	if err != nil {
		return err
	}

	return nil
}

func deployResourcesSubdir(variant DockerfileVariant, subdir string) error {
	srcDir := path.Join(
		baseDir,
		"generate",
		"resources",
		string(variant.Product),
		subdir,
	)

	exists, err := exists(srcDir)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	targetDir := variant.targetDir()

	destDir := path.Join(targetDir, subdir)

	return CopyDir(srcDir, destDir)
}

func deployScriptResources(variant DockerfileVariant) error {
	return deployResourcesSubdir(variant, "scripts")
}

func deployConfigResources(variant DockerfileVariant) error {
	return deployResourcesSubdir(variant, "config")
}

func deployReadme(variant DockerfileVariant) error {
	srcDir := path.Join(
		baseDir,
		"generate",
		"resources",
		string(variant.Product),
	)

	srcFile := path.Join(srcDir, "README.md")
	targetDir := variant.targetDir()
	destFile := path.Join(targetDir, "README.md")

	if err := CopyFile(srcFile, destFile); err != nil {
		return err
	}

	return nil
}

func versionSubdirectories(dir string) []string {
	// eg, 3.0.25
	versionDirGlobPattern := "[0-9]*.[0-9]*.[0-9]*"

	versions := []string{}

	files, _ := filepath.Glob(fmt.Sprintf("%v/%v", dir, versionDirGlobPattern))
	for _, file := range files {
		versions = append(versions, filepath.Base(file))
	}

	return versions
}

func CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err == nil {
			err = os.Chmod(dest, sourceinfo.Mode())
			if err != nil {
				log.Printf("Error chmod %v", dest)
			}
		}

	}

	return
}

func CopyDir(source string, dest string) (err error) {
	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir

	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourcefilepointer := source + "/" + obj.Name()

		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			err = CopyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = CopyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}

type DockerfileVariant struct {
	Edition Edition
	Product Product
	// Version is the real version of the product as seen in the outside
	// world - eg., in download URLs, package filenames, etc.
	Version string
	// TargetVersion is the version of the Docker image (which in turn
	// is the directory name in this repository). 99.99% of the time
	// this will be the same as Version, but very occasionally we need
	// to translate a bit here
	TargetVersion     string
	TemplateFilename  string
	Arches            []Arch
	IsStaging         bool
	OutputDir         string
	TemplateOverrides map[string]any
}

func (variant DockerfileVariant) getSHA256(arch Arch) string {
	var sha256url string
	if variant.Product == "couchbase-server" {
		sha256url = variant.releaseURL() + "/" +
			variant.serverPackageFile(arch) + ".sha256"
	}

	resp, err := http.Get(sha256url)
	log.Print(sha256url)

	if err != nil || resp.StatusCode != 200 {
		log.Printf("Error downloading SHA256 file")
		return "MISSING_SHA256_ERROR"
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error download content of SHA256 file")
			return "HTTP_ERROR"
		}
		return strings.Fields(fmt.Sprintf("%s", body))[0]
	}
}

func (variant DockerfileVariant) dockerBaseImage() string {
	switch variant.Product {
	case ProductSyncGw:
		productVer, _ := intVer(variant.Version)
		if strings.Contains(variant.Version, "forestdb") {
			return "tleyden5iwx/forestdb"
		}
		if productVer <= 30003 {
			return "centos:centos7"
		} else {
			return "ubuntu:22.04"
		}
	case ProductServer:
		return fmt.Sprintf("ubuntu:%s", variant.ubuntuVersion())
	case ProductSandbox:
		return fmt.Sprintf("couchbase/server:%s", variant.Version)
	default:
		log.Printf("Failed %v", variant.Product)
		panic("Unexpected product")
	}
}

func (variant DockerfileVariant) serverPkgCommand() string {
	// Currently all Server Dockerfiles are based on Ubuntu, so this is
	// always "apt-get". However we did some work in the Dockerfile
	// template to support "yum" as well. Leaving that in place for now
	// in case we work on integrating the RHCC Dockerfile in future.
	return "apt-get"
}

func (variant DockerfileVariant) systemdWorkaround() bool {
	if variant.Product == ProductServer {
		ver, _ := intVer(variant.Version)
		if ver < 70000 {
			return true
		}
	}
	return false
}

func intVer(v string) (int64, error) {
	baseVer := strings.Split(v, "-")[0]
	sections := strings.Split(baseVer, ".")
	intVerSection := func(n int) string {
		return fmt.Sprintf("%02s", sections[n])
	}
	s := ""
	for i := 0; i < 3; i++ {
		s += intVerSection(i)
	}
	return strconv.ParseInt(s, 10, 64)
}

func (variant DockerfileVariant) isMadHatterOrNewer() bool {
	ver, _ := intVer(variant.Version)
	return ver >= 60500
}

func (variant DockerfileVariant) ubuntuVersion() string {
	// Intended for use by Couchbase Server only
	if strings.HasPrefix(variant.Version, "4") {
		return "14.04"
	} else if strings.HasPrefix(variant.Version, "5") {
		return "16.04"
	} else if strings.HasPrefix(variant.Version, "6") {
		if variant.Version == "6.0.0" {
			return "16.04"
		} else if strings.HasPrefix(variant.Version, "6.0") || strings.HasPrefix(variant.Version, "6.5") || variant.Version == "6.6.0" || variant.Version == "6.6.1" {
			return "18.04"
		}
	}
	// Value for 7.x and 6.6.2+
	return "20.04"
}

// Get the version for this variant, possibly doing substitutions
func (variant DockerfileVariant) VersionWithSubstitutions() string {
	if variant.Product == "sync-gateway" {
		// if version is 0.0.0-xxx, replace with feature/xxx.
		// (example: 0.0.0-forestdb -> feature/forestdb)
		extraStuff := extraStuffAfterVersion(variant.Version)
		switch extraStuff {
		case "forestdb":
			return fmt.Sprintf("feature/%v", extraStuff)
		default:
			return variant.Version
		}
	}
	return variant.Version
}

// Given a version like "1.0.0" or "0.0.0-forestdb", return
// the extra stuff after the version, like "" or "forestdb" (respectively)
func extraStuffAfterVersion(version string) string {
	re := regexp.MustCompile(`[0-9]*.[0-9]*.[0-9]*-?(.*)`)
	result := re.FindStringSubmatch(version)
	if len(result) > 1 {
		group1 := result[1]
		return group1
	}
	return ""
}

// Generate the package filename for this variant:
// eg: couchbase-server-enterprise-7.1.1-linux_amd64.deb
func (variant DockerfileVariant) serverPackageFile(arch Arch) string {
	serverVer, _ := intVer(variant.Version)
	if serverVer >= 70100 {
		// From Neo onwards, use "linux" package since it's all the same.
		return fmt.Sprintf(
			"%v-%v_%v-linux_%v.deb",
			variant.Product,
			variant.Edition,
			variant.Version,
			arch,
		)
	} else {
		// For earlier releases, no arm64 builds, so just hardcode amd64
		return fmt.Sprintf(
			"%v-%v_%v-ubuntu%v_amd64.deb",
			variant.Product,
			variant.Edition,
			variant.Version,
			variant.ubuntuVersion(),
		)
	}
}

// Generate the package name (couchbase-server or couchbase-server-community)
// for this variant
func (variant DockerfileVariant) serverPackageName() string {
	if variant.Edition == EditionCommunity {
		return "couchbase-server-community"
	} else {
		return "couchbase-server"
	}
}

// Specify any extra dependencies, based on variant
func (variant DockerfileVariant) extraDependencies() string {
	if variant.Product == "couchbase-server" {
		if variant.isMadHatterOrNewer() {
			return "bzip2 runit"
		} else {
			return "python-httplib2 runit"
		}
	}
	return ""
}

func (variant DockerfileVariant) targetDir() string {
	// If variant has an explicit output directory, use that
	if variant.OutputDir != "" {
		return variant.OutputDir
	}

	// Here we use TargetVersion rather than Version
	version := string(variant.TargetVersion)
	if variant.IsStaging {
		version = fmt.Sprintf("%s-staging", version)
	}
	targetDir := path.Join(
		baseDir,
		string(variant.Edition),
		string(variant.Product),
		version,
	)
	return targetDir
}

func (variant DockerfileVariant) dockerfile() string {
	return path.Join(variant.targetDir(), "Dockerfile")
}

func (variant DockerfileVariant) releaseURL() string {
	if variant.IsStaging {
		return "http://packages-staging.couchbase.com/releases/" + variant.Version
	} else {
		return "https://packages.couchbase.com/releases/" + variant.Version
	}
}

// Find the package URL for this Sync Gateway version
func (variant DockerfileVariant) sgPackageUrl() string {
	var packagesBaseUrl string
	if variant.IsStaging {
		packagesBaseUrl = "http://packages-staging.couchbase.com/releases/couchbase-sync-gateway"
	} else {
		packagesBaseUrl = "http://packages.couchbase.com/releases/couchbase-sync-gateway"
	}

	versionCustomization, hasCustomization := variant.versionCustomization()

	switch hasCustomization {
	case true:
		return fmt.Sprintf("%s", versionCustomization.PackageUrl)
	default:
		sgFileName := variant.sgPackageFilename()

		return fmt.Sprintf(
			"%s/%s/%s",
			packagesBaseUrl,
			variant.Version,
			sgFileName,
		)
	}
}

func (variant DockerfileVariant) sgPackageFilename() string {
	versionCustomization, hasCustomization := variant.versionCustomization()
	switch hasCustomization {
	case true:
		return fmt.Sprintf("%s", versionCustomization.PackageFilename)
	default:
		productVer, _ := intVer(variant.Version)
		// Containers for SGW versions <= 3.0.3 were only produced for x64
		if productVer <= 30003 {
			return fmt.Sprintf(
				"couchbase-sync-gateway-%s_%s_@@ARCH@@.rpm",
				strings.ToLower(string(variant.Edition)),
				variant.Version,
			)
		} else {
			return fmt.Sprintf(
				"couchbase-sync-gateway-%s_%s_@@ARCH@@.deb",
				strings.ToLower(string(variant.Edition)),
				variant.Version,
			)
		}
	}
}

func (variant DockerfileVariant) versionCustomization() (v VersionCustomization, exists bool) {
	// eg, "sync-gateway_community_2.0.0-build
	key := variant.versionCustomizationKey()

	v, exists = versionCustomizations[key]
	return v, exists
}

func (variant DockerfileVariant) versionCustomizationKey() string {
	return fmt.Sprintf("%s_%s_%s", variant.Product, variant.Edition, variant.Version)
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
