package finalize

import (
	"bytes"
	"io/ioutil"
	"nodejs/cache"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

type Finalizer struct {
	Stager    *libbuildpack.Stager
	CacheDirs []string
	PreBuild  string
	PostBuild string
}

func Run(f *Finalizer) error {
	if err := f.ReadPackageJson(); err != nil {
		f.Stager.Log.Error("Failed parsing package.json: %s", err.Error())
		return err
	}

	if err := f.TipVendorDependencies(); err != nil {
		f.Stager.Log.Error(err.Error())
		return err
	}

	f.ListNodeConfig(os.Environ())

	cacher, err := cache.New(f.Stager, f.CacheDirs)
	if err != nil {
		f.Stager.Log.Error("Unable to initialize cache: %s", err.Error())
		return err
	}

	if err := cacher.Restore(); err != nil {
		f.Stager.Log.Error("Unable to restore cache: %s", err.Error())
		return err
	}

	if err := cacher.Save(); err != nil {
		f.Stager.Log.Error("Unable to save cache: %s", err.Error())
		return err
	}

	return nil
}

func (f *Finalizer) ReadPackageJson() error {
	var p struct {
		CacheDirs1 []string `json:"cacheDirectories"`
		CacheDirs2 []string `json:"cache_directories"`
		Scripts    struct {
			PreBuild  string `json:"heroku-prebuild"`
			PostBuild string `json:"heroku-postbuild"`
		} `json:"scripts"`
	}
	f.CacheDirs = []string{}

	if err := libbuildpack.NewJSON().Load(filepath.Join(f.Stager.BuildDir, "package.json"), &p); err != nil {
		if os.IsNotExist(err) {
			f.Stager.Log.Warning("No package.json found")
			return nil
		} else {
			return err
		}
	}

	if len(p.CacheDirs1) > 0 {
		f.CacheDirs = p.CacheDirs1
	} else if len(p.CacheDirs2) > 0 {
		f.CacheDirs = p.CacheDirs2
	}
	f.PreBuild = p.Scripts.PreBuild
	f.PostBuild = p.Scripts.PostBuild

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
