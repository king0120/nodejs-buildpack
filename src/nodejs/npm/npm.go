package npm

import "io"

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
	return nil
}

func (n *NPM) Rebuild() error {
	return nil
}
