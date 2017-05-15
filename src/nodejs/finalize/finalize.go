package finalize

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Finalizer struct {
	Stager *libbuildpack.Stager
}

func Run(f *Finalizer) error {
	if err := f.TipVendorDependencies(); err != nil {
		f.Stager.Log.Error(err.Error())
		return err
	}

	if err := f.WarnMissingPackageJSON(); err != nil {
		f.Stager.Log.Error(err.Error())
		return err
	}

	return nil
}

func (f *Finalizer) TipVendorDependencies() error {
	subdirs, err := hasSubdirs(filepath.Join(f.Stager.BuildDir, "node_modules"))
	if err != nil {
		return err
	}
	if !subdirs {
		f.Stager.Log.Protip("It is recommended to vendor the application's Node.js dependencies",
			"http://docs.cloudfoundry.org/buildpacks/node/index.html#vendoring")
	}

	return nil
}

func (f *Finalizer) WarnMissingPackageJSON() error {
	exists, err := libbuildpack.FileExists(filepath.Join(f.Stager.BuildDir, "package.json"))
	if err != nil {
		return err
	}

	if !exists {
		f.Stager.Log.Warning("No package.json found")
	}
	return nil
}

func hasSubdirs(path string) (bool, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	for _, file := range files {
		if file.IsDir() {
			return true, nil
		}
	}

	return false, nil
}
