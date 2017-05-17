package npm_test

import (
	"bytes"
	"io/ioutil"
	n "nodejs/npm"
	"os"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=npm.go --destination=mocks_test.go --package=npm_test

var _ = Describe("Yarn", func() {
	var (
		err         error
		buildDir    string
		npm         *n.NPM
		logger      libbuildpack.Logger
		buffer      *bytes.Buffer
		mockCtrl    *gomock.Controller
		mockCommand *MockCommand
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "nodejs-buildpack.build.")
		Expect(err).To(BeNil())

		buffer = new(bytes.Buffer)

		logger = libbuildpack.NewLogger()
		logger.SetOutput(ansicleaner.New(buffer))

		mockCtrl = gomock.NewController(GinkgoT())
		mockCommand = NewMockCommand(mockCtrl)
	})

	JustBeforeEach(func() {
		npm = &n.NPM{
			BuildDir: buildDir,
			Logger:   logger,
			Command:  mockCommand,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()

		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())
	})

	Describe("Build", func() {
	})

	Describe("Rebuild", func() {
	})
})
