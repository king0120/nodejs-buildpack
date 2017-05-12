package supply_test

import (
	"bytes"
	"io/ioutil"
	"nodejs/supply"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Supply", func() {
	var (
		err      error
		buildDir string
		depsDir  string
		depsIdx  string
		depDir   string
		supplier *supply.Supplier
		logger   libbuildpack.Logger
		buffer   *bytes.Buffer
	)

	BeforeEach(func() {
		depsDir, err = ioutil.TempDir("", "nodejs-buildpack.deps.")
		Expect(err).To(BeNil())

		buildDir, err = ioutil.TempDir("", "nodejs-buildpack.build.")
		Expect(err).To(BeNil())

		depsIdx = "14"
		depDir = filepath.Join(depsDir, depsIdx)

		err = os.MkdirAll(depDir, 0755)
		Expect(err).To(BeNil())

		buffer = new(bytes.Buffer)

		logger = libbuildpack.NewLogger()
		logger.SetOutput(buffer)
	})

	JustBeforeEach(func() {
		bps := &libbuildpack.Stager{
			BuildDir: buildDir,
			DepsDir:  depsDir,
			DepsIdx:  depsIdx,
			Log:      logger,
		}

		supplier = &supply.Supplier{
			Stager: bps,
		}
	})

	AfterEach(func() {
		err = os.RemoveAll(depsDir)
		Expect(err).To(BeNil())
	})

	Describe("LoadPackageJSON", func() {
		var packageJSON string

		JustBeforeEach(func() {
			ioutil.WriteFile(filepath.Join(buildDir, "package.json"), []byte(packageJSON), 0644)
		})

		Context("File is invalid JSON", func() {
			BeforeEach(func() {
				packageJSON = `not actually JSON`
			})

			It("returns an error", func() {
				err = supplier.LoadPackageJSON()
				Expect(err).NotTo(BeNil())
			})
		})

		Context("File is valid JSON", func() {
			Context("has an engines section", func() {
				BeforeEach(func() {
					packageJSON = `
{
  "name": "node",
  "version": "1.0.0",
  "main": "server.js",
  "author": "CF Buildpacks Team",
  "dependencies": {
    "logfmt": "~1.1.2",
    "express": "~4.0.0"
  },
  "engines" : {
		"yarn" : "*",
		"npm"  : "npm-x",
		"node" : "node-y",
		"something" : "3.2.1"
	}
}
`
				})

				It("loads the engines into the supplier", func() {
					err = supplier.LoadPackageJSON()
					Expect(err).To(BeNil())

					Expect(supplier.PackageJSON.Engines.Node).To(Equal("node-y"))
					Expect(supplier.PackageJSON.Engines.Yarn).To(Equal("*"))
					Expect(supplier.PackageJSON.Engines.NPM).To(Equal("npm-x"))
					Expect(supplier.PackageJSON.Engines.Iojs).To(Equal(""))
				})

				Context("the engines section contains iojs", func() {
					BeforeEach(func() {
						packageJSON = `
{
  "engines" : {
		"iojs" : "*"
	}
}
`
					})

					It("returns an error", func() {
						err = supplier.LoadPackageJSON()
						Expect(err).NotTo(BeNil())

						Expect(err.Error()).To(ContainSubstring("io.js not supported by this buildpack"))
					})
				})
			})

			Context("does not have an engines section", func() {
				BeforeEach(func() {
					packageJSON = `
{
  "name": "node",
  "version": "1.0.0",
  "main": "server.js",
  "author": "CF Buildpacks Team",
  "dependencies": {
    "logfmt": "~1.1.2",
    "express": "~4.0.0"
  }
}
`
				})

				It("loads the engine struct with empty strings", func() {
					err = supplier.LoadPackageJSON()
					Expect(err).To(BeNil())

					Expect(supplier.PackageJSON.Engines.Node).To(Equal(""))
					Expect(supplier.PackageJSON.Engines.Yarn).To(Equal(""))
					Expect(supplier.PackageJSON.Engines.NPM).To(Equal(""))
					Expect(supplier.PackageJSON.Engines.Iojs).To(Equal(""))
				})
			})
		})
	})
})
