package finalize_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"nodejs/finalize"
	"os"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=../vendor/github.com/cloudfoundry/libbuildpack/command_runner.go --destination=mocks_command_runner_test.go --package=finalize_test

var _ = Describe("Cache", func() {
	var (
		err               error
		buildDir          string
		cache             *finalize.Cache
		stager            *libbuildpack.Stager
		logger            libbuildpack.Logger
		buffer            *bytes.Buffer
		mockCtrl          *gomock.Controller
		mockCommandRunner *MockCommandRunner
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "nodejs-buildpack.build.")
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
			Log:      logger,
			Command:  mockCommandRunner,
		}

		cache = &finalize.Cache{
			Stager: stager,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()

		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())
	})

	Describe("NewCache", func() {
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
			cache, err = finalize.NewCache(stager)
			Expect(err).To(BeNil())
			Expect(cache.NodeVersion).To(Equal("6.9.3"))
		})

		It("sets npm version", func() {
			cache, err = finalize.NewCache(stager)
			Expect(err).To(BeNil())
			Expect(cache.NPMVersion).To(Equal("4.5.6"))
		})

		It("sets yarn version", func() {
			cache, err = finalize.NewCache(stager)
			Expect(err).To(BeNil())
			Expect(cache.YarnVersion).To(Equal("9.8.7"))
		})
	})
})
