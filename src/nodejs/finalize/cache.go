package finalize

import (
	"bytes"
	"io/ioutil"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

type Cache struct {
	Stager      *libbuildpack.Stager
	NodeVersion string
	NPMVersion  string
	YarnVersion string
}

func NewCache(stager *libbuildpack.Stager) (*Cache, error) {
	var err error
	c := &Cache{Stager: stager}

	if c.NodeVersion, err = c.findVersion("node"); err != nil {
		return nil, err
	}

	if c.NPMVersion, err = c.findVersion("npm"); err != nil {
		return nil, err
	}

	if c.YarnVersion, err = c.findVersion("yarn"); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Cache) findVersion(binary string) (string, error) {
	buffer := new(bytes.Buffer)
	if err := c.Stager.Command.Execute("", buffer, ioutil.Discard, binary, "--version"); err != nil {
		return "", err
	}
	return strings.TrimSpace(buffer.String()), nil
}
