package roadrunner_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/laraboot-io/laraboot-buildpacks-roadrunner"
	"github.com/laraboot-io/laraboot-buildpacks-roadrunner/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		parser *fakes.Parser

		workingDir string
		detect     packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		parser = &fakes.Parser{}
		parser.ParseVersionCall.Returns.Version = "some-version"
		parser.ParseVersionCall.Returns.VersionSource = "some-version-source"

		detect = roadrunner.Detect(parser)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a DetectResult that provides httpd", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: workingDir,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: roadrunner.PlanDependencyRoadRunner},
				},
			},
		}))

		Expect(parser.ParseVersionCall.CallCount).To(Equal(0))
	})

	context("when there is an roadrunner.conf file in the workspace", func() {
		it.Before(func() {
			_, err := os.Create(filepath.Join(workingDir, "roadrunner.conf"))
			Expect(err).NotTo(HaveOccurred())
		})

		it("returns a DetectResult that provides and required httpd", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: roadrunner.PlanDependencyRoadRunner},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: roadrunner.PlanDependencyRoadRunner,
							Metadata: roadrunner.BuildPlanMetadata{
								Version:       "some-version",
								VersionSource: "some-version-source",
								Launch:        true,
							},
						},
					},
				},
			}))

			Expect(parser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
		})
	})

	context("BP_ROADRUNNER_VERSION is set", func() {
		it.Before(func() {
			_, err := os.Create(filepath.Join(workingDir, "roadrunner.conf"))
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Setenv("BP_ROADRUNNER_VERSION", "env-var-version")).To(Succeed())
		})

		it.After(func() {
			err := os.Remove(filepath.Join(workingDir, "roadrunner.conf"))
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Unsetenv("BP_ROADRUNNER_VERSION")).To(Succeed())
		})

		it("returns a DetectResult that required specified version of httpd", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: roadrunner.PlanDependencyRoadRunner},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: roadrunner.PlanDependencyRoadRunner,
							Metadata: roadrunner.BuildPlanMetadata{
								Version:       "env-var-version",
								VersionSource: "BP_ROADRUNNER_VERSION",
								Launch:        true,
							},
						},
						{
							Name: roadrunner.PlanDependencyRoadRunner,
							Metadata: roadrunner.BuildPlanMetadata{
								Version:       "some-version",
								VersionSource: "some-version-source",
								Launch:        true,
							},
						},
					},
				},
			}))
		})
	})

	context("failure cases", func() {
		context("when ParseVersion fails", func() {
			it.Before(func() {
				_, err := os.Create(filepath.Join(workingDir, "roadrunner.conf"))
				Expect(err).NotTo(HaveOccurred())

				parser.ParseVersionCall.Returns.Err = errors.New("failed to parse version")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{WorkingDir: workingDir})
				Expect(err).To(MatchError("failed to parse version"))
			})
		})
	})
}
