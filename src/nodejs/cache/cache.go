package cache

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

type Cache struct {
	Stager      *libbuildpack.Stager
	NodeVersion string
	NPMVersion  string
	YarnVersion string
}

func New(stager *libbuildpack.Stager) (*Cache, error) {
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

func (c *Cache) Save() error {
	return nil
}

func (c *Cache) Restore() error {
	c.Stager.Log.BeginStep("Restoring cache")

	signature, err := ioutil.ReadFile(filepath.Join(c.Stager.CacheDir, "node", "signature"))
	if err != nil {
		if os.IsNotExist(err) {
			c.Stager.Log.Info("Skipping cache restore (no previous cache)")
			return nil
		}

		return err
	}

	if strings.TrimSpace(string(signature)) != c.signature() {
		c.Stager.Log.Info("Skipping cache restore (new runtime signature)")
		return nil
	}

	if os.Getenv("NODE_MODULES_CACHE") != "" {
		c.Stager.Log.Info("Skipping cache restore (disabled by config)")
		return nil
	}

	dirsToRestore, err := c.selectCacheDirs()
	if err != nil {
		return err
	}

	return c.restoreCacheDirs(dirsToRestore)
}

func (c *Cache) selectCacheDirs() ([]string, error) {
	dirs := []string{".npm", ".yarn/cache", "bower_components"}

	var p struct {
		Dirs1 []string `json:"cacheDirectories"`
		Dirs2 []string `json:"cache_directories"`
	}

	if err := libbuildpack.NewJSON().Load(filepath.Join(c.Stager.BuildDir, "package.json"), &p); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	if len(p.Dirs1) > 0 {
		dirs = p.Dirs1
		c.Stager.Log.Info("Loading %d from cacheDirectories (package.json):", len(dirs))
	} else if len(p.Dirs2) > 0 {
		dirs = p.Dirs2
		c.Stager.Log.Info("Loading %d from cacheDirectories (package.json):", len(dirs))
	} else {
		c.Stager.Log.Info("Loading 3 from cacheDirectories (default):")
	}

	return dirs, nil
}

func (c *Cache) restoreCacheDirs(dirsToRestore []string) error {
	for _, dir := range dirsToRestore {
		dest := filepath.Join(c.Stager.BuildDir, dir)

		exists, err := libbuildpack.FileExists(dest)
		if err != nil {
			return err
		}

		if exists {
			c.Stager.Log.Info("- %s (exists - skipping)", dir)
			continue
		}

		source := filepath.Join(c.Stager.CacheDir, "node", dir)
		exists, err = libbuildpack.FileExists(source)
		if err != nil {
			return err
		}

		if !exists {
			c.Stager.Log.Info("- %s (not cached - skipping)", dir)
			continue
		}

		c.Stager.Log.Info("- %s", dir)

		if err = os.MkdirAll(path.Dir(dest), 0755); err != nil {
			return err
		}

		if err := os.Rename(source, dest); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cache) findVersion(binary string) (string, error) {
	buffer := new(bytes.Buffer)
	if err := c.Stager.Command.Execute("", buffer, ioutil.Discard, binary, "--version"); err != nil {
		return "", err
	}
	return strings.TrimSpace(buffer.String()), nil
}

func (c *Cache) signature() string {
	return fmt.Sprintf("%s; %s; %s", c.NodeVersion, c.NPMVersion, c.YarnVersion)
}
