package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
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
	ProductServer = Product("couchbase-server")
	ProductSyncGw = Product("sync-gateway")
)

var (
	editions       []Edition
	products       []Product
	processingRoot string
)

func init() {

	editions = []Edition{
		EditionCommunity,
		EditionEnterprise,
	}

	products = []Product{
		ProductServer,
		ProductSyncGw,
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
	for _, version := range versions {

		variant := DockerfileVariant{
			Edition: edition,
			Product: product,
			Version: version,
		}

		if err := generateVariant(variant); err != nil {
			return err
		}

	}

	return nil

}

func generateVariant(variant DockerfileVariant) error {

	if err := generateDockerfile(variant); err != nil {
		return err
	}

	if err := deployResources(variant); err != nil {
		return err
	}

	if err := deployReadme(variant); err != nil {
		return err
	}

	return nil

}

func generateDockerfile(variant DockerfileVariant) error {

	log.Printf("generateDockerfile called with: %v", variant)

	versionDir := variant.versionDir()

	// figure out output filename
	targetDockerfile := path.Join(versionDir, "Dockerfile")

	// open a file at destPath
	out, err := os.Create(targetDockerfile)
	if err != nil {
		return err
	}
	defer out.Close()

	// find the path to the source template
	sourceTemplate := path.Join(
		processingRoot,
		"generate",
		"templates",
		string(variant.Product),
		"Dockerfile.template",
	)

	// template parameters
	params := struct {
		CB_VERSION string
		CB_PACKAGE string
		CB_EXTRA_DEPS string
	}{
		CB_VERSION: variant.VersionWithSubstitutions(),
		CB_PACKAGE: variant.debPackageName(),
		CB_EXTRA_DEPS: variant.extraDependencies(),
	}

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

func deployResources(variant DockerfileVariant) error {

	srcDir := path.Join(
		processingRoot,
		"generate",
		"resources",
		string(variant.Product),
		"scripts",
	)

	exists, err := exists(srcDir)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	versionDir := variant.versionDir()

	destDir := path.Join(versionDir, "scripts")

	return CopyDir(srcDir, destDir)

}

func deployReadme(variant DockerfileVariant) error {

	srcDir := path.Join(
		processingRoot,
		"generate",
		"resources",
		string(variant.Product),
	)

	srcFile := path.Join(srcDir, "README.md")
	versionDir := variant.versionDir()
	destFile := path.Join(versionDir, "README.md")

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
		if err != nil {
			// this isn't working as expected.
			//
			// source file:
			// -rwxr-xr-x  1 tleyden  staff   1.3K Apr 26 19:05 sync-gw-start
			// dest file:
			// -rw-r--r--  1 tleyden  staff   1.3K Apr 27 09:20 sync-gw-start
			// ^^ why isn't dest file +x?
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
	Version string
}

func (variant DockerfileVariant) isVersion2() bool {
	return strings.HasPrefix(variant.Version, "2")
}

// Get the version for this variant, possibly doing substitutions
func (variant DockerfileVariant) VersionWithSubstitutions() string {
	if variant.Product == "sync-gateway" {
		// if version is 0.0.0-xxx, replace with feature/xxx.
		// (example: 0.0.0-forestdb -> feature/forestdb)
		branchName := extraStuffAfterVersion(variant.Version)
		if branchName != "" {
			// do something with the branch name ..
			return fmt.Sprintf("feature/%v", branchName)
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

// Generate the rpm package name for this variant:
// eg: couchbase-server-enterprise-3.0.2-centos6.x86_64.rpm
func (variant DockerfileVariant) rpmPackageName() string {
	// for 2.x, leave centos out of the rpm name
	if variant.isVersion2() {
		return fmt.Sprintf(
			"%v-%v_%v_x86_64.rpm",
			variant.Product,
			variant.Edition,
			variant.Version,
		)
	} else {
		return fmt.Sprintf(
			"%v-%v-%v-centos6.x86_64.rpm",
			variant.Product,
			variant.Edition,
			variant.Version,
		)
	}
}

// Generate the rpm package name for this variant:
// eg: couchbase-server-enterprise-3.0.2-ubuntu12.04_amd64.deb
func (variant DockerfileVariant) debPackageName() string {
	// for 2.x, leave ubuntu12.04 out of the deb name
	if variant.isVersion2() {
		return fmt.Sprintf(
			"%v-%v_%v_x86_64.deb",
			variant.Product,
			variant.Edition,
			variant.Version,
		)
	} else {
		return fmt.Sprintf(
			"%v-%v_%v-ubuntu12.04_amd64.deb",
			variant.Product,
			variant.Edition,
			variant.Version,
		)
	}
}

// Specify any extra dependencies, based on variant
func (variant DockerfileVariant) extraDependencies() string {
	if variant.Product == "couchbase-server" &&
	   variant.isVersion2() {
		return "librtmp0"
	}
	return ""
}

func (variant DockerfileVariant) versionDir() string {
	versionDir := path.Join(
		processingRoot,
		string(variant.Edition),
		string(variant.Product),
		string(variant.Version),
	)
	return versionDir
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
