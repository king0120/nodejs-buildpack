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
	Stager               *libbuildpack.Stager
	NodeVersion          string
	NPMVersion           string
	YarnVersion          string
	PackageJSONCacheDirs []string
}

var defaultCacheDirs = []string{".npm", ".cache/yarn", "bower_components"}

func New(stager *libbuildpack.Stager, cacheDirs []string) (*Cache, error) {
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

	c.PackageJSONCacheDirs = cacheDirs

	return c, nil
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

	if os.Getenv("NODE_MODULES_CACHE") == "false" {
		c.Stager.Log.Info("Skipping cache restore (disabled by config)")
		return nil
	}

	source, dirsToRestore := c.selectCacheDirs()
	c.Stager.Log.Info("Loading %d from cacheDirectories (%s):", len(dirsToRestore), source)

	return c.restoreCacheDirs(dirsToRestore)
}

func (c *Cache) Save() error {
	c.Stager.Log.BeginStep("Caching build")
	c.Stager.Log.Info("Clearing previous node cache")

	if err := c.Stager.ClearCache(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(c.Stager.CacheDir, "node"), 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(c.Stager.CacheDir, "node", "signature"), []byte(c.signature()+"\n"), 0644); err != nil {
		return err
	}

	if os.Getenv("NODE_MODULES_CACHE") == "false" {
		c.Stager.Log.Info("Skipping cache save (disabled by config)")
		return nil
	}

	source, dirsToSave := c.selectCacheDirs()
	c.Stager.Log.Info("Saving %d cacheDirectories (%s):", len(dirsToSave), source)

	if err := c.saveCacheDirs(dirsToSave); err != nil {
		return err

	}

	dirsToRemove := []string{".npm", ".cache/yarn"}
	for _, dir := range dirsToRemove {
		if err := os.RemoveAll(filepath.Join(c.Stager.BuildDir, dir)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cache) saveCacheDirs(dirsToSave []string) error {
	for _, dir := range dirsToSave {
		dest := filepath.Join(c.Stager.CacheDir, "node", dir)
		source := filepath.Join(c.Stager.BuildDir, dir)

		sourceExists, err := libbuildpack.FileExists(source)
		if err != nil {
			return err
		}

		if sourceExists {
			c.Stager.Log.Info("- %s", dir)

			if err := os.MkdirAll(dest, 0755); err != nil {
				return err
			}

			if err := libbuildpack.CopyDirectory(source, dest); err != nil {
				return err
			}
		} else {
			c.Stager.Log.Info("- %s (nothing to cache)", dir)
		}
	}

	return nil
}

func (c *Cache) selectCacheDirs() (string, []string) {
	if len(c.PackageJSONCacheDirs) > 0 {
		return "package.json", c.PackageJSONCacheDirs
	}

	return "default", defaultCacheDirs
}

func (c *Cache) restoreCacheDirs(dirsToRestore []string) error {
	for _, dir := range dirsToRestore {
		dest := filepath.Join(c.Stager.BuildDir, dir)
		source := filepath.Join(c.Stager.CacheDir, "node", dir)

		destExists, err := libbuildpack.FileExists(dest)
		if err != nil {
			return err
		}

		sourceExists, err := libbuildpack.FileExists(source)
		if err != nil {
			return err
		}

		if destExists {
			c.Stager.Log.Info("- %s (exists - skipping)", dir)
		} else if !sourceExists {
			c.Stager.Log.Info("- %s (not cached - skipping)", dir)
		} else {
			c.Stager.Log.Info("- %s", dir)

			if err := os.MkdirAll(path.Dir(dest), 0755); err != nil {
				return err
			}

			if err := os.Rename(source, dest); err != nil {
				return err
			}
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
