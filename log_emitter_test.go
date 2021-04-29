package roadrunner_test

import (
	"bytes"
	"testing"

	"github.com/laraboot-io/laraboot-buildpacks-roadrunner"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLogEmitter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer  *bytes.Buffer
		emitter roadrunner.LogEmitter
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		emitter = roadrunner.NewLogEmitter(buffer)
	})

	context("Title", func() {
		it("logs the buildpack title", func() {
			emitter.Title(packit.BuildpackInfo{
				Name:    "some-name",
				Version: "some-version",
			})
			Expect(buffer.String()).To(Equal("some-name some-version\n"))
		})
	})

	context("Environment", func() {
		it("logs the environment variables", func() {
			emitter.Environment(packit.Environment{
				"SOME_VAR.override": "some-value",
			})
			Expect(buffer.String()).To(Equal("    SOME_VAR -> \"some-value\"\n"))
		})
	})
}
