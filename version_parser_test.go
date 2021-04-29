package roadrunner_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/laraboot-io/laraboot-buildpacks-roadrunner"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testVersionParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		versionParser roadrunner.VersionParser
	)

	it.Before(func() {
		versionParser = roadrunner.NewVersionParser()
	})

	context("ParseVersion", func() {
		context("when there is no buildpack.yml", func() {
			it("returns a * for the version and empty version source", func() {
				version, versionSource, err := versionParser.ParseVersion("some-path")
				Expect(err).NotTo(HaveOccurred())
				Expect(version).To(Equal("*"))
				Expect(versionSource).To(Equal(""))
			})
		})

		context("when there is a buildpack.yml", func() {
			var path string

			it.Before(func() {
				file, err := ioutil.TempFile("", "buildpack.yml")
				Expect(err).NotTo(HaveOccurred())

				_, err = file.WriteString(`{"roadrunner": {"version": "some-version"}}`)
				Expect(err).NotTo(HaveOccurred())

				path = file.Name()

				Expect(file.Close()).To(Succeed())
			})

			it.After(func() {
				Expect(os.Remove(path)).To(Succeed())
			})

			it("parses the version", func() {
				version, versionSource, err := versionParser.ParseVersion(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(version).To(Equal("some-version"))
				Expect(versionSource).To(Equal("buildpack.yml"))
			})

			context("when there is not roadrunner version in the buildpack.yml", func() {
				it.Before(func() {
					err := ioutil.WriteFile(path, []byte(`{"some-thing": {"version": "some-version"}}`), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns a * for the version and empty version source", func() {
					version, versionSource, err := versionParser.ParseVersion(path)
					Expect(err).NotTo(HaveOccurred())
					Expect(version).To(Equal("*"))
					Expect(versionSource).To(Equal(""))
				})
			})

			context("failure cases", func() {
				context("when the file cannot be opened", func() {
					it.Before(func() {
						Expect(os.Chmod(path, 0000)).To(Succeed())
					})

					it("returns an error", func() {
						_, _, err := versionParser.ParseVersion(path)
						Expect(err).To(MatchError(ContainSubstring("failed to parse buildpack.yml")))
						Expect(err).To(MatchError(ContainSubstring("permission denied")))
					})
				})

				context("when the file contains malformed yaml", func() {
					it.Before(func() {
						err := ioutil.WriteFile(path, []byte("%%%"), 0644)
						Expect(err).NotTo(HaveOccurred())
					})

					it("returns an error", func() {
						_, _, err := versionParser.ParseVersion(path)
						Expect(err).To(MatchError(ContainSubstring("failed to parse buildpack.yml")))
						Expect(err).To(MatchError(ContainSubstring("could not find expected directive name")))
					})
				})
			})
		})
	})
}
