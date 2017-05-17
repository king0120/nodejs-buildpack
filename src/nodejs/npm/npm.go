package npm

import (
	"io"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Command interface {
	Execute(dir string, stdout io.Writer, stderr io.Writer, program string, args ...string) error
}

type Logger interface {
	Info(format string, args ...interface{})
	Warning(format string, args ...interface{})
}

type NPM struct {
	BuildDir string
	Command  Command
	Logger   Logger
}

func (n *NPM) Build() error {
	pkgExists, err := libbuildpack.FileExists(filepath.Join(n.BuildDir, "package.json"))
	if err != nil {
		return err
	}

	if !pkgExists {
		n.Logger.Info("Skipping (no package.json)")
		return nil
	}

	shrinkwrapExists, err := libbuildpack.FileExists(filepath.Join(n.BuildDir, "npm-shrinkwrap.json"))
	if err != nil {
		return err
	}

	if shrinkwrapExists {
		n.Logger.Info("Installing node modules (package.json + shrinkwrap)")
	} else {
		n.Logger.Info("Installing node modules (package.json)")
	}

	return nil
}

func (n *NPM) Rebuild() error {
	return nil
}
