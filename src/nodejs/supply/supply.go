package supply

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

type Supplier struct {
	Stager *libbuildpack.Stager
	Node   string
	Yarn   string
	NPM    string
}

type packageJSON struct {
	Engines engines `json:"engines"`
}

type engines struct {
	Node string `json:"node"`
	Yarn string `json:"yarn"`
	NPM  string `json:"npm"`
	Iojs string `json:"iojs"`
}

func Run(ss *Supplier) error {
	ss.Stager.Log.BeginStep("Installing binaries")
	if err := ss.LoadPackageJSON(); err != nil {
		ss.Stager.Log.Error("Unable to load package.json: %s", err.Error())
		return err
	}

	ss.WarnNodeEngine()

	if err := ss.InstallNode(); err != nil {
		ss.Stager.Log.Error("Unable to install node: %s", err.Error())
		return err
	}

	if err := ss.InstallNPM(); err != nil {
		ss.Stager.Log.Error("Unable to install node: %s", err.Error())
		return err
	}

	if err := ss.InstallYarn(); err != nil {
		ss.Stager.Log.Error("Unable to install node: %s", err.Error())
		return err
	}

	return nil
}

func (ss *Supplier) LoadPackageJSON() error {
	var p packageJSON

	err := libbuildpack.NewJSON().Load(filepath.Join(ss.Stager.BuildDir, "package.json"), &p)
	if err != nil {
		return err
	}

	if p.Engines.Iojs != "" {
		return errors.New("io.js not supported by this buildpack")
	}

	if p.Engines.Node != "" {
		ss.Stager.Log.Info("engines.node (package.json): %s", p.Engines.Node)
	} else {
		ss.Stager.Log.Info("engines.node (package.json): unspecified")
	}

	if p.Engines.NPM != "" {
		ss.Stager.Log.Info("engines.npm (package.json): %s", p.Engines.NPM)
	} else {
		ss.Stager.Log.Info("engines.npm (package.json): unspecified (use default)")
	}

	ss.Node = p.Engines.Node
	ss.NPM = p.Engines.NPM
	ss.Yarn = p.Engines.Yarn

	return nil
}

func (ss *Supplier) WarnNodeEngine() {
	docsLink := "http://docs.cloudfoundry.org/buildpacks/node/node-tips.html"

	if ss.Node == "" {
		ss.Stager.Log.Warning("Node version not specified in package.json. See: %s", docsLink)
	}
	if ss.Node == "*" {
		ss.Stager.Log.Warning("Dangerous semver range (*) in engines.node. See: %s", docsLink)
	}
	if strings.HasPrefix(ss.Node, ">") {
		ss.Stager.Log.Warning("Dangerous semver range (>) in engines.node. See: %s", docsLink)
	}
	return
}

func (ss *Supplier) InstallNode() error {
	return nil
}

func (ss *Supplier) InstallNPM() error {
	return nil
}
func (ss *Supplier) InstallYarn() error {
	return nil
}
