package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
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
	}{
		CB_VERSION: variant.Version,
		CB_PACKAGE: variant.rpmPackageName(),
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

	versionDir := variant.versionDir()

	destDir := path.Join(versionDir, "scripts")

	return CopyDir(srcDir, destDir)

}

func deployReadme(variant DockerfileVariant) error {
	srcFile := path.Join(processingRoot, "README.md")
	versionDir := variant.versionDir()
	destFile := path.Join(versionDir, "README.md")
	return CopyFile(srcFile, destFile)
}

func generateSyncGw(edition, version string) error {
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
			err = os.Chmod(dest, sourceinfo.Mode())
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

func (variant DockerfileVariant) versionDir() string {
	versionDir := path.Join(
		processingRoot,
		string(variant.Edition),
		string(variant.Product),
		string(variant.Version),
	)
	return versionDir
}
