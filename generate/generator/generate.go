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
	ProductServer   = Product("couchbase-server")
	ProductSyncGw   = Product("sync-gateway")
	ProductOperator = Product("couchbase-operator")
)

type Arch string

const (
	Archamd64   = Arch("x86_64")
	Archaarch64 = Arch("aarch64")
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
	editions              []Edition
	products              []Product
	versionCustomizations VersionCustomizations
	processingRoot        string
	skipGeneration        ProductVersionFilter
)

func init() {

	editions = []Edition{
		EditionCommunity,
		EditionEnterprise,
	}

	products = []Product{
		ProductServer,
		ProductSyncGw,
		ProductOperator,
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

	// get args with this binary stripped off
	args := os.Args[1:]

	if len(args) == 0 {
		log.Fatalf("Usage: ./generate <path> where <path> is the directory where you checked out couchbase-docker, eg: /home/you/dev/couchbase-docker")
	}

	processingRoot = args[0]

	for _, edition := range editions {
		for _, product := range products {
			if err := generateVersions(edition, product); err != nil {
				log.Fatalf("Failed (%v/%v): %v", edition, product, err)
			}
		}
	}

	log.Printf("Successfully finished!")

}

func generateVersions(edition Edition, product Product) error {

	// find corresponding directory for this edition/product combo
	dir := path.Join(processingRoot, string(edition), string(product))

	// find all version subdirectories (must match regex)
	versions := versionSubdirectories(dir)

	// for each version
	for _, ver := range versions {

		if skipGeneration.Matches(product, ver) {
			log.Printf("Skipping generation for %v %v %v", product, edition, ver)
			continue
		}

		// Start with a basic DockerfileVariant, then tweak if necessary
		variant := DockerfileVariant{
			Edition:       edition,
			Product:       product,
			Multiplatform: false,
			Version:       strings.TrimSuffix(ver, "-staging"),
			TargetVersion: strings.TrimSuffix(ver, "-staging"),
			IsStaging:     strings.HasSuffix(ver, "-staging"),
		}
		variantArches := []Arch{Archamd64}

		// Couchbase Server has some special cases based on Version.
		if product == ProductServer {
			serverVer, _ := intVer(variant.Version)
			if serverVer == 70003 {
				// CBD-4603: 7.0.3 actually builds from 7.0.3-MP1 for complete
				// Log4Shell remediation
				variant.Version = "7.0.3-MP1"
			}
			if serverVer >= 70100 {
				// From Neo onwards, we have both amd64 and aarch64 images
				variant.Multiplatform = true
				variantArches = []Arch{Archamd64, Archaarch64}
			}
		}

		// Now generate the Dockerfile(s) based on the constructed variant
		for _, arch := range variantArches {
			variant.Arch = arch
			if err := generateVariant(variant); err != nil {
				return err

			}
		}
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
	targetDockerfile := path.Join(targetDir, "Dockerfile")
	log.Printf("targetDockerfile: %v", targetDockerfile)

	// find the path to the source template
	sourceTemplate := path.Join(
		processingRoot,
		"generate",
		"templates",
		string(variant.Product),
		"Dockerfile.template",
	)

	log.Printf("template: %v", sourceTemplate)
	log.Printf("product: %v", variant.Product)
	var params interface{}

	if variant.Product == ProductServer {
		// template parameters
		params = struct {
			CB_VERSION         string
			CB_PACKAGE         string
			CB_PACKAGE_NAME    string
			CB_EXTRA_DEPS      string
			CB_SHA256          string
			CB_RELEASE_URL     string
			DOCKER_BASE_IMAGE  string
			PKG_COMMAND        string
			SYSTEMD_WORKAROUND bool
			ARCH               string
		}{
			CB_VERSION:         variant.VersionWithSubstitutions(),
			CB_PACKAGE:         variant.serverPackageFile(),
			CB_PACKAGE_NAME:    variant.serverPackageName(),
			CB_EXTRA_DEPS:      variant.extraDependencies(),
			CB_SHA256:          variant.getSHA256(),
			CB_RELEASE_URL:     variant.releaseURL(),
			DOCKER_BASE_IMAGE:  variant.dockerBaseImage(),
			PKG_COMMAND:        variant.pkgCommand(),
			SYSTEMD_WORKAROUND: variant.systemdWorkaround(),
			ARCH:               string(variant.Arch),
		}

	} else if variant.Product == ProductSyncGw {
		// template parameters
		params = struct {
			SYNC_GATEWAY_PACKAGE_URL      string
			SYNC_GATEWAY_PACKAGE_FILENAME string
			DOCKER_BASE_IMAGE             string
		}{
			SYNC_GATEWAY_PACKAGE_URL:      variant.sgPackageUrl(),
			SYNC_GATEWAY_PACKAGE_FILENAME: variant.sgPackageFilename(),
			DOCKER_BASE_IMAGE:             variant.dockerBaseImage(),
		}
	} else if variant.Product == ProductOperator {
		// template parameters
		params = struct {
			CO_VERSION     string
			CO_RELEASE_URL string
			CO_PACKAGE     string
			CO_SHA256      string
		}{
			CO_VERSION:     variant.VersionWithSubstitutions(),
			CO_RELEASE_URL: variant.releaseURL(),
			CO_PACKAGE:     variant.operatorPackageName(),
			CO_SHA256:      variant.getSHA256(),
		}
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
		processingRoot,
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
		processingRoot,
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
	Edition       Edition
	Arch          Arch
	Product       Product
	Multiplatform bool
	// Version is the real version of the product as seen in the outside
	// world - eg., in download URLs, package filenames, etc.
	Version string
	// TargetVersion is the version of the Docker image (which in turn
	// is the directory name in this repository). 99.99% of the time
	// this will be the same as Version, but very occasionally we need
	// to translate a bit here
	TargetVersion string
	IsStaging     bool
}

func (variant DockerfileVariant) getSHA256() string {
	var sha256url string
	if variant.Product == "couchbase-server" {
		sha256url = variant.releaseURL() + "/" +
			variant.Version + "/" + variant.serverPackageFile() + ".sha256"
	} else if variant.Product == "couchbase-operator" {
		sha256url = variant.releaseURL() + "/" +
			variant.Version + "/" + variant.operatorPackageName() + ".sha256"
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
		if strings.Contains(variant.Version, "forestdb") {
			return "tleyden5iwx/forestdb"
		}
		return "centos:centos7"
	case ProductServer:
		if variant.Arch == Archaarch64 {
			return "amazonlinux:2"
		}
		return fmt.Sprintf("ubuntu:%s", variant.ubuntuVersion())
	default:
		log.Printf("Failed %v", variant.Product)
		panic("Unexpected product")
	}
}

func (variant DockerfileVariant) pkgCommand() string {

	log.Printf("Arch: %v", variant.Arch)
	if variant.Arch == Archaarch64 {
		return "yum"
	}
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
// eg: couchbase-server-enterprise-7.1.0-ubuntu20.04_amd64.deb
// eg: couchbase-server-enterprise-7.1.0-amzn2.aarch64.rpm
func (variant DockerfileVariant) serverPackageFile() string {
	if variant.Arch == Archaarch64 {
		return fmt.Sprintf(
			"%v-%v-%v-amzn2.aarch64.rpm",
			variant.Product,
			variant.Edition,
			variant.Version,
		)
	}
	return fmt.Sprintf(
		"%v-%v_%v-ubuntu%v_amd64.deb",
		variant.Product,
		variant.Edition,
		variant.Version,
		variant.ubuntuVersion(),
	)
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

// Generate the bits package name for this couchbase-operator variant
func (variant DockerfileVariant) operatorPackageName() string {
	return fmt.Sprintf(
		"couchbase-autonomous-operator-dist_%v.tar.gz",
		variant.Version,
	)
}

// Specify any extra dependencies, based on variant
func (variant DockerfileVariant) extraDependencies() string {
	if variant.Product == "couchbase-server" {
		if variant.Arch == Archamd64 {
			if variant.isMadHatterOrNewer() {
				return "bzip2 runit"
			} else {
				return "python-httplib2 runit"
			}
		} else {
			if variant.isMadHatterOrNewer() {
				return "bzip2"
			} else {
				return "python-httplib2"
			}
		}
	}
	return ""
}

func (variant DockerfileVariant) targetDir() string {
	// Here we use TargetVersion rather than Version
	version := string(variant.TargetVersion)
	if variant.IsStaging {
		version = fmt.Sprintf("%s-staging", version)
	}
	targetDir := path.Join(
		processingRoot,
		string(variant.Edition),
		string(variant.Product),
		version,
	)
	if variant.Multiplatform {
		targetDir = path.Join(
			targetDir,
			string(variant.Arch),
		)

		err := os.MkdirAll(targetDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create %v: %v", targetDir, err)
		}
	}
	return targetDir
}

func (variant DockerfileVariant) dockerfile() string {
	return path.Join(variant.targetDir(), "Dockerfile")
}

func (variant DockerfileVariant) releaseURL() string {
	if variant.Product == "couchbase-operator" {
		return "https://packages.couchbase.com/kubernetes"
	}
	if variant.IsStaging {
		return "http://packages-staging.couchbase.com/releases"
	} else {
		return "https://packages.couchbase.com/releases"
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
		return fmt.Sprintf(
			"couchbase-sync-gateway-%s_%s_x86_64.rpm",
			strings.ToLower(string(variant.Edition)),
			variant.Version,
		)

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
