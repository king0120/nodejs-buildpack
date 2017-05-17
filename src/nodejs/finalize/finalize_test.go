package finalize_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"nodejs/finalize"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=../vendor/github.com/cloudfoundry/libbuildpack/command_runner.go --destination=mocks_command_runner_test.go --package=finalize_test
//go:generate mockgen -source=finalize.go --destination=mocks_test.go --package=finalize_test

var _ = Describe("Finalize", func() {
	var (
		err               error
		buildDir          string
		finalizer         *finalize.Finalizer
		logger            libbuildpack.Logger
		buffer            *bytes.Buffer
		mockCtrl          *gomock.Controller
		mockCommandRunner *MockCommandRunner
		mockYarn          *MockYarn
		mockNPM           *MockNPM
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "nodejs-buildpack.build.")
		Expect(err).To(BeNil())

		buffer = new(bytes.Buffer)

		logger = libbuildpack.NewLogger()
		logger.SetOutput(ansicleaner.New(buffer))

		mockCtrl = gomock.NewController(GinkgoT())
		mockCommandRunner = NewMockCommandRunner(mockCtrl)
		mockYarn = NewMockYarn(mockCtrl)
		mockNPM = NewMockNPM(mockCtrl)

		bps := &libbuildpack.Stager{
			BuildDir: buildDir,
			Log:      logger,
			Command:  mockCommandRunner,
		}

		finalizer = &finalize.Finalizer{
			Stager: bps,
			Yarn:   mockYarn,
			NPM:    mockNPM,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()

		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())
	})

	Describe("TipVendorDependencies", func() {
		Context("node_modules exists and has subdirectories", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(buildDir, "node_modules", "exciting_module"), 0755)).To(BeNil())
			})

			It("does not log anything", func() {
				Expect(finalizer.TipVendorDependencies()).To(BeNil())
				Expect(buffer.String()).To(Equal(""))
			})
		})

		Context("node_modules exists and has NO subdirectories", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(buildDir, "node_modules"), 0755)).To(BeNil())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "node_modules", "a_file"), []byte("content"), 0644)).To(BeNil())
			})

			It("logs a pro tip", func() {
				Expect(finalizer.TipVendorDependencies()).To(BeNil())
				Expect(buffer.String()).To(ContainSubstring("PRO TIP: It is recommended to vendor the application's Node.js dependencies"))
				Expect(buffer.String()).To(ContainSubstring("http://docs.cloudfoundry.org/buildpacks/node/index.html#vendoring"))
			})
		})

		Context("node_modules does not exist", func() {
			It("logs a pro tip", func() {
				Expect(finalizer.TipVendorDependencies()).To(BeNil())
				Expect(buffer.String()).To(ContainSubstring("PRO TIP: It is recommended to vendor the application's Node.js dependencies"))
				Expect(buffer.String()).To(ContainSubstring("http://docs.cloudfoundry.org/buildpacks/node/index.html#vendoring"))
			})
		})
	})

	Describe("ReadPackageJSON", func() {
		Context("package.json has cacheDirectories", func() {
			BeforeEach(func() {
				packageJSON := `
{
  "cacheDirectories" : [
		"first",
		"second"
	]
}
`
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "package.json"), []byte(packageJSON), 0644)).To(Succeed())
			})

			It("sets CacheDirs", func() {
				Expect(finalizer.ReadPackageJSON()).To(Succeed())
				Expect(finalizer.CacheDirs).To(Equal([]string{"first", "second"}))
			})
		})

		Context("package.json has cache_directories", func() {
			BeforeEach(func() {
				packageJSON := `
{
  "cache_directories" : [
		"third",
		"fourth"
	]
}
`
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "package.json"), []byte(packageJSON), 0644)).To(Succeed())
			})

			It("sets CacheDirs", func() {
				Expect(finalizer.ReadPackageJSON()).To(Succeed())
				Expect(finalizer.CacheDirs).To(Equal([]string{"third", "fourth"}))
			})
		})

		Context("package.json has prebuild script", func() {
			BeforeEach(func() {
				packageJSON := `
{
  "scripts" : {
		"script": "script",
		"heroku-prebuild": "makestuff",
		"thing": "thing"
	}
}
`
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "package.json"), []byte(packageJSON), 0644)).To(Succeed())
			})

			It("sets PreBuild", func() {
				Expect(finalizer.ReadPackageJSON()).To(Succeed())
				Expect(finalizer.PreBuild).To(Equal("makestuff"))
			})
		})

		Context("package.json has postbuild script", func() {
			BeforeEach(func() {
				packageJSON := `
{
  "scripts" : {
		"script": "script",
		"heroku-postbuild": "logstuff",
		"thing": "thing"
	}
}
`
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "package.json"), []byte(packageJSON), 0644)).To(Succeed())
			})

			It("sets PostBuild", func() {
				Expect(finalizer.ReadPackageJSON()).To(Succeed())
				Expect(finalizer.PostBuild).To(Equal("logstuff"))
			})
		})

		Context("package.json does not exist", func() {
			It("warns user", func() {
				Expect(finalizer.ReadPackageJSON()).To(Succeed())
				Expect(buffer.String()).To(ContainSubstring("**WARNING** No package.json found"))
			})
			It("initializes config based values", func() {
				Expect(finalizer.ReadPackageJSON()).To(Succeed())
				Expect(finalizer.CacheDirs).To(Equal([]string{}))
			})
		})

		Context("yarn.lock exists", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "yarn.lock"), []byte("{}"), 0644)).To(Succeed())
			})
			It("sets UseYarn", func() {
				Expect(finalizer.ReadPackageJSON()).To(Succeed())
				Expect(finalizer.UseYarn).To(BeTrue())
			})
		})

		Context("yarn.lock does not exist", func() {
			It("sets UseYarn", func() {
				Expect(finalizer.ReadPackageJSON()).To(Succeed())
				Expect(finalizer.UseYarn).To(BeFalse())
			})
		})
	})

	Describe("ListNodeConfig", func() {
		DescribeTable("outputs relevant env vars",
			func(key string, value string, expected string) {
				finalizer.ListNodeConfig([]string{fmt.Sprintf("%s=%s", key, value)})
				Expect(buffer.String()).To(Equal(expected))
			},

			Entry("NPM_CONFIG_", "NPM_CONFIG_THING", "someval", "       NPM_CONFIG_THING=someval\n"),
			Entry("YARN_", "YARN_KEY", "aval", "       YARN_KEY=aval\n"),
			Entry("NODE_", "NODE_EXCITING", "newval", "       NODE_EXCITING=newval\n"),
			Entry("NOT_RELEVANT", "NOT_RELEVANT", "anything", ""),
		)

		It("warns about NODE_ENV override", func() {
			finalizer.ListNodeConfig([]string{"NPM_CONFIG_PRODUCTION=true", "NODE_ENV=development"})
			Expect(buffer.String()).To(ContainSubstring("npm scripts will see NODE_ENV=production (not 'development')"))
			Expect(buffer.String()).To(ContainSubstring("https://docs.npmjs.com/misc/config#production"))
		})
	})

	Describe("BuildDependencies", func() {
		Context("yarn.lock exists", func() {
			BeforeEach(func() {
				finalizer.UseYarn = true
				mockYarn.EXPECT().Build().Return(nil)
			})

			It("runs yarn install", func() {
				Expect(finalizer.BuildDependencies()).To(Succeed())
			})

			Context("prebuild is specified", func() {
				BeforeEach(func() {
					finalizer.PreBuild = "prescriptive"
				})

				It("runs the prebuild script", func() {
					mockCommandRunner.EXPECT().Execute(buildDir, gomock.Any(), gomock.Any(), "yarn", "run", "prescriptive")
					Expect(finalizer.BuildDependencies()).To(Succeed())
					Expect(buffer.String()).To(ContainSubstring("Running prescriptive (yarn)"))
				})
			})

			Context("postbuild is specified", func() {
				BeforeEach(func() {
					finalizer.PostBuild = "descriptive"
				})

				It("runs the prebuild script", func() {
					mockCommandRunner.EXPECT().Execute(buildDir, gomock.Any(), gomock.Any(), "yarn", "run", "descriptive")
					Expect(finalizer.BuildDependencies()).To(Succeed())
					Expect(buffer.String()).To(ContainSubstring("Running descriptive (yarn)"))
				})
			})
		})
	})
})
