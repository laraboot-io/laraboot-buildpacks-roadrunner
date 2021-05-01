package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSimpleApp(t *testing.T, when spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		//Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker

		name      string
		source    string
		image     occam.Image
		container occam.Container
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()

		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
		Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		Expect(os.RemoveAll(source)).To(Succeed())
	})

	it("serves up staticfile", func() {
		var err error

		source, err = occam.Source(filepath.Join("testdata", "simple_app"))
		Expect(err).NotTo(HaveOccurred())

		image, _, err = pack.WithVerbose().Build.
			WithBuildpacks(phpBuildpack, goBuildpack, roadRunnerBuildpack).
			Execute(name, source)
		Expect(err).NotTo(HaveOccurred())

		//container, err = docker.Container.Run.
		//	WithEnv(map[string]string{"PORT": "8080"}).
		//	WithPublish("8080").
		//	WithPublishAll().
		//	Execute(image.ID)
		//Expect(err).NotTo(HaveOccurred())
		//
		//Eventually(container).Should(BeAvailable())
		//
		//response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
		//Expect(err).NotTo(HaveOccurred())
		//Expect(response.StatusCode).To(Equal(http.StatusOK))
	})
}
