package cache_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"nodejs/cache"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=../vendor/github.com/cloudfoundry/libbuildpack/command_runner.go --destination=mocks_command_runner_test.go --package=cache_test

var _ = Describe("Cache", func() {
	var (
		err               error
		buildDir          string
		cacheDir          string
		cacher            *cache.Cache
		stager            *libbuildpack.Stager
		logger            libbuildpack.Logger
		buffer            *bytes.Buffer
		mockCtrl          *gomock.Controller
		mockCommandRunner *MockCommandRunner
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "nodejs-buildpack.build.")
		Expect(err).To(BeNil())

		cacheDir, err = ioutil.TempDir("", "nodejs-buildpack.cache.")
		Expect(err).To(BeNil())

		buffer = new(bytes.Buffer)

		logger = libbuildpack.NewLogger()
		logger.SetOutput(ansicleaner.New(buffer))

		mockCtrl = gomock.NewController(GinkgoT())
		mockCommandRunner = NewMockCommandRunner(mockCtrl)
	})

	JustBeforeEach(func() {
		stager = &libbuildpack.Stager{
			BuildDir: buildDir,
			CacheDir: cacheDir,
			Log:      logger,
			Command:  mockCommandRunner,
		}

		cacher = &cache.Cache{
			Stager:      stager,
			NodeVersion: "1.1.1",
			NPMVersion:  "2.2.2",
			YarnVersion: "3.3.3",
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()

		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(cacheDir)
		Expect(err).To(BeNil())
	})

	Describe("New", func() {
		BeforeEach(func() {
			mockCommandRunner.EXPECT().Execute("", gomock.Any(), gomock.Any(), "node", "--version").Do(func(_ string, buffer io.Writer, _ io.Writer, _ string, _ string) {
				buffer.Write([]byte("6.9.3\n"))
			}).Return(nil)

			mockCommandRunner.EXPECT().Execute("", gomock.Any(), gomock.Any(), "npm", "--version").Do(func(_ string, buffer io.Writer, _ io.Writer, _ string, _ string) {
				buffer.Write([]byte("4.5.6\n"))
			}).Return(nil)

			mockCommandRunner.EXPECT().Execute("", gomock.Any(), gomock.Any(), "yarn", "--version").Do(func(_ string, buffer io.Writer, _ io.Writer, _ string, _ string) {
				buffer.Write([]byte("9.8.7\n"))
			}).Return(nil)
		})

		It("sets node version", func() {
			cacher, err = cache.New(stager)
			Expect(err).To(BeNil())
			Expect(cacher.NodeVersion).To(Equal("6.9.3"))
		})

		It("sets npm version", func() {
			cacher, err = cache.New(stager)
			Expect(err).To(BeNil())
			Expect(cacher.NPMVersion).To(Equal("4.5.6"))
		})

		It("sets yarn version", func() {
			cacher, err = cache.New(stager)
			Expect(err).To(BeNil())
			Expect(cacher.YarnVersion).To(Equal("9.8.7"))
		})
	})

	Describe("Restore", func() {
		Context("there is a cache", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(cacheDir, "node", ".npm"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(cacheDir, "node", ".npm", "cached"), []byte("xxx"), 0644)).To(Succeed())

				Expect(os.MkdirAll(filepath.Join(cacheDir, "node", ".yarn", "cache"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(cacheDir, "node", ".yarn", "cache", "cached"), []byte("yyy"), 0644)).To(Succeed())

				Expect(os.MkdirAll(filepath.Join(cacheDir, "node", "bower_components"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(cacheDir, "node", "bower_components", "cached"), []byte("zzz"), 0644)).To(Succeed())

				Expect(os.MkdirAll(filepath.Join(cacheDir, "node", "other_dir"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(cacheDir, "node", "other_dir", "cached"), []byte("aaa"), 0644)).To(Succeed())
			})

			Context("signature changed", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(cacheDir, "node", "signature"), []byte("1; 2; 3\n"), 0644)).To(BeNil())
				})

				It("alerts user", func() {
					Expect(cacher.Restore()).To(Succeed())

					Expect(buffer.String()).To(ContainSubstring("Skipping cache restore (new runtime signature)"))
				})

				It("does not restore the cache", func() {
					Expect(cacher.Restore()).To(Succeed())
					files, err := ioutil.ReadDir(filepath.Join(buildDir))
					Expect(err).To(BeNil())
					Expect(len(files)).To(Equal(0))
				})
			})

			Context("signatures match", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(cacheDir, "node", "signature"), []byte("1.1.1; 2.2.2; 3.3.3\n"), 0644)).To(BeNil())
				})

				Context("cached directories are not in build dir", func() {
					It("alerts user", func() {
						Expect(cacher.Restore()).To(Succeed())

						Expect(buffer.String()).To(ContainSubstring("Loading 3 from cacheDirectories (default):"))
						Expect(buffer.String()).To(ContainSubstring("- .npm\n"))
						Expect(buffer.String()).To(ContainSubstring("- .yarn/cache\n"))
						Expect(buffer.String()).To(ContainSubstring("- bower_components\n"))
					})

					It("moves the requested cached directories", func() {
						Expect(cacher.Restore()).To(Succeed())
						files, err := ioutil.ReadDir(filepath.Join(buildDir))
						Expect(err).To(BeNil())

						Expect(len(files)).To(Equal(3))
						Expect(ioutil.ReadFile(filepath.Join(buildDir, ".npm", "cached"))).To(Equal([]byte("xxx")))
						Expect(ioutil.ReadFile(filepath.Join(buildDir, ".yarn", "cache", "cached"))).To(Equal([]byte("yyy")))
						Expect(ioutil.ReadFile(filepath.Join(buildDir, "bower_components", "cached"))).To(Equal([]byte("zzz")))
					})
				})

				Context("some cached directories are already in build dir", func() {
					BeforeEach(func() {
						Expect(os.MkdirAll(filepath.Join(buildDir, ".npm"), 0755)).To(Succeed())
						Expect(ioutil.WriteFile(filepath.Join(buildDir, ".npm", "cached"), []byte("from app"), 0644)).To(Succeed())
					})

					It("alerts user", func() {
						Expect(cacher.Restore()).To(Succeed())

						Expect(buffer.String()).To(ContainSubstring("Loading 3 from cacheDirectories (default):"))
						Expect(buffer.String()).To(ContainSubstring("- .npm (exists - skipping)\n"))
						Expect(buffer.String()).To(ContainSubstring("- .yarn/cache\n"))
						Expect(buffer.String()).To(ContainSubstring("- bower_components\n"))
					})

					It("moves the requested cached directories", func() {
						Expect(cacher.Restore()).To(Succeed())
						files, err := ioutil.ReadDir(filepath.Join(buildDir))
						Expect(err).To(BeNil())

						Expect(len(files)).To(Equal(3))
						Expect(ioutil.ReadFile(filepath.Join(buildDir, ".npm", "cached"))).To(Equal([]byte("from app")))
						Expect(ioutil.ReadFile(filepath.Join(buildDir, ".yarn", "cache", "cached"))).To(Equal([]byte("yyy")))
						Expect(ioutil.ReadFile(filepath.Join(buildDir, "bower_components", "cached"))).To(Equal([]byte("zzz")))
					})
				})

				Context("some cached directories are already in build dir", func() {
					BeforeEach(func() {
						Expect(os.RemoveAll(filepath.Join(cacheDir, "node", ".npm"))).To(Succeed())
					})

					It("alerts user", func() {
						Expect(cacher.Restore()).To(Succeed())

						Expect(buffer.String()).To(ContainSubstring("Loading 3 from cacheDirectories (default):"))
						Expect(buffer.String()).To(ContainSubstring("- .npm (not cached - skipping)\n"))
						Expect(buffer.String()).To(ContainSubstring("- .yarn/cache\n"))
						Expect(buffer.String()).To(ContainSubstring("- bower_components\n"))
					})

					It("moves the requested cached directories", func() {
						Expect(cacher.Restore()).To(Succeed())
						files, err := ioutil.ReadDir(filepath.Join(buildDir))
						Expect(err).To(BeNil())

						Expect(len(files)).To(Equal(2))
						Expect(ioutil.ReadFile(filepath.Join(buildDir, ".yarn", "cache", "cached"))).To(Equal([]byte("yyy")))
						Expect(ioutil.ReadFile(filepath.Join(buildDir, "bower_components", "cached"))).To(Equal([]byte("zzz")))
					})
				})
			})
		})

		Context("there is not a cache", func() {
			It("alerts user", func() {
				Expect(cacher.Restore()).To(Succeed())
				Expect(buffer.String()).To(ContainSubstring("Skipping cache restore (no previous cache)"))
			})
		})
	})
})
