package supply

import (
	"bytes"
	"errors"
	"io/ioutil"
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
		ss.Stager.Log.Error("Unable to install npm: %s", err.Error())
		return err
	}

	if err := ss.InstallYarn(); err != nil {
		ss.Stager.Log.Error("Unable to install yarn: %s", err.Error())
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
	var dep libbuildpack.Dependency

	if ss.Node != "" {
		versions := ss.Stager.Manifest.AllDependencyVersions("node")
		ver, err := libbuildpack.FindMatchingVersion(ss.Node, versions)
		if err != nil {
			return err
		}
		dep.Name = "node"
		dep.Version = ver
	} else {
		var err error

		dep, err = ss.Stager.Manifest.DefaultVersion("node")
		if err != nil {
			return err
		}
	}

	if err := ss.Stager.Manifest.InstallDependency(dep, filepath.Join(ss.Stager.DepDir(), "node")); err != nil {
		return err
	}
	return ss.Stager.LinkDirectoryInDepDir(filepath.Join(ss.Stager.DepDir(), "node", "bin"), "bin")
}

func (ss *Supplier) InstallNPM() error {
	buffer := new(bytes.Buffer)
	if err := ss.Stager.Command.Execute(ss.Stager.BuildDir, buffer, buffer, "npm", "--version"); err != nil {
		return err
	}

	npmVersion := strings.TrimSpace(buffer.String())

	if ss.NPM == "" {
		ss.Stager.Log.Info("Using default npm version: %s", npmVersion)
		return nil
	}
	if ss.NPM == npmVersion {
		ss.Stager.Log.Info("npm %s already installed with node", npmVersion)
		return nil
	}

	ss.Stager.Log.Info("Downloading and installing npm %s (replacing version %s)...", ss.NPM, npmVersion)

	if err := ss.Stager.Command.Execute(ss.Stager.BuildDir, ioutil.Discard, ioutil.Discard, "npm", "install", "--unsafe-perm", "--quiet", "-g", "npm@"+ss.NPM); err != nil {
		ss.Stager.Log.Error("We're unable to download the version of npm you've provided (%s).\nPlease remove the npm version specification in package.json", ss.NPM)
		return err
	}
	return nil
}

func (ss *Supplier) InstallYarn() error {
	if ss.Yarn != "" {
		versions := ss.Stager.Manifest.AllDependencyVersions("yarn")
		_, err := libbuildpack.FindMatchingVersion(ss.Yarn, versions)
		if err != nil {
			ss.Stager.Log.Warning("package.json requested yarn version %s, but buildpack only includes yarn version %s", ss.Yarn, versions[0])
		}
	}

	yarnInstallDir := filepath.Join(ss.Stager.DepDir(), "yarn")

	if err := ss.Stager.Manifest.InstallOnlyVersion("yarn", yarnInstallDir); err != nil {
		return err
	}

	if err := ss.Stager.LinkDirectoryInDepDir(filepath.Join(yarnInstallDir, "dist", "bin"), "bin"); err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	if err := ss.Stager.Command.Execute(ss.Stager.BuildDir, buffer, buffer, "yarn", "--version"); err != nil {
		return err
	}

	yarnVersion := strings.TrimSpace(buffer.String())
	ss.Stager.Log.Info("Installed yarn %s", yarnVersion)

	return nil
}
