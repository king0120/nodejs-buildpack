package finalize

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

type Finalizer struct {
	Stager      *libbuildpack.Stager
	NodeVersion string
	NPMVersion  string
	YarnVersion string
}

func Run(f *Finalizer) error {
	if err := f.Init(); err != nil {
		f.Stager.Log.Error("unable to find binaries: %s", err.Error())
		return err
	}

	if err := f.TipVendorDependencies(); err != nil {
		f.Stager.Log.Error(err.Error())
		return err
	}

	if err := f.WarnMissingPackageJSON(); err != nil {
		f.Stager.Log.Error(err.Error())
		return err
	}

	f.ListNodeConfig(os.Environ())

	return nil
}

func (f *Finalizer) Init() error {
	var err error
	if f.NodeVersion, err = f.findVersion("node"); err != nil {
		return err
	}

	if f.NPMVersion, err = f.findVersion("npm"); err != nil {
		return err
	}

	if f.YarnVersion, err = f.findVersion("yarn"); err != nil {
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

func (f *Finalizer) ListNodeConfig(environment []string) {
	npmConfigProductionTrue := false
	nodeEnv := "production"

	for _, env := range environment {
		if strings.HasPrefix(env, "NPM_CONFIG_") || strings.HasPrefix(env, "YARN_") || strings.HasPrefix(env, "NODE_") {
			f.Stager.Log.Info(env)
		}

		if env == "NPM_CONFIG_PRODUCTION=true" {
			npmConfigProductionTrue = true
		}

		if strings.HasPrefix(env, "NODE_ENV=") {
			nodeEnv = env[9:]
		}
	}

	if npmConfigProductionTrue && nodeEnv != "production" {
		f.Stager.Log.Info("npm scripts will see NODE_ENV=production (not '%s')\nhttps://docs.npmjs.com/misc/config#production", nodeEnv)
	}
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

func (f *Finalizer) findVersion(binary string) (string, error) {
	buffer := new(bytes.Buffer)
	if err := f.Stager.Command.Execute("", buffer, ioutil.Discard, binary, "--version"); err != nil {
		return "", err
	}
	return strings.TrimSpace(buffer.String()), nil
}
