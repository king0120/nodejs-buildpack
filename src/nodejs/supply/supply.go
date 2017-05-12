package supply

import (
	"errors"
	"path/filepath"
	"strings"

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

	ss.WarnNodeEngine()

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

func (ss *Supplier) WarnNodeEngine() {
	docsLink := "http://docs.cloudfoundry.org/buildpacks/node/node-tips.html"

	if ss.PackageJSON.Engines.Node == "" {
		ss.Stager.Log.Warning("Node version not specified in package.json. See: %s", docsLink)
	}
	if ss.PackageJSON.Engines.Node == "*" {
		ss.Stager.Log.Warning("Dangerous semver range (*) in engines.node. See: %s", docsLink)
	}
	if strings.HasPrefix(ss.PackageJSON.Engines.Node, ">") {
		ss.Stager.Log.Warning("Dangerous semver range (>) in engines.node. See: %s", docsLink)
	}
	return
}
