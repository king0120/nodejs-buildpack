package supply

import (
	"errors"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Supplier struct {
	Stager      *libbuildpack.Stager
	PackageJSON PackageJSON
}

type PackageJSON struct {
	Engines Engines `json:"engines"`
}

type Engines struct {
	Node string `json:"node"`
	Yarn string `json:"yarn"`
	NPM  string `json:"npm"`
	Iojs string `json:"iojs"`
}

func Run(ss *Supplier) error {
	if err := ss.LoadPackageJSON(); err != nil {
		ss.Stager.Log.Error("Unable to load package.json: %s", err.Error())
		return err
	}

	return nil
}

func (ss *Supplier) LoadPackageJSON() error {
	err := libbuildpack.NewJSON().Load(filepath.Join(ss.Stager.BuildDir, "package.json"), &ss.PackageJSON)
	if err != nil {
		return err
	}

	if ss.PackageJSON.Engines.Iojs != "" {
		return errors.New("io.js not supported by this buildpack")
	}

	return nil
}
