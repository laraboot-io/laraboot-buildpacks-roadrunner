package integration_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	phpBuildpack               string
	goBuildpack                string
	roadRunnerBuildpack        string
	offlineRoadRunnerBuildpack string
	buildpackInfo              struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}
)

func TestIntegration(t *testing.T) {
	var Expect = NewWithT(t).Expect

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()

	_, err = toml.DecodeReader(file, &buildpackInfo)
	Expect(err).NotTo(HaveOccurred())

	phpBuildpack = "paketo-buildpacks/php-dist"
	goBuildpack = "paketo-buildpacks/go-dist"

	buildpackStore := occam.NewBuildpackStore()

	roadRunnerBuildpack, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	offlineRoadRunnerBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	//suite("Caching", testCaching)
	//suite("Logging", testLogging)
	//suite("Offline", testOffline)
	suite("SimpleApp", testSimpleApp)
	suite.Run(t)
}
