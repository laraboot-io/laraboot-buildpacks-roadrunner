package roadrunner_test

import (
	"bytes"
	"errors"
	roadrunner "github.com/laraboot-io/laraboot-buildpacks-roadrunner"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/laraboot-io/laraboot-buildpacks-roadrunner/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		layersDir  string
		cnbPath    string
		timestamp  string

		entryResolver     *fakes.EntryResolver
		dependencyService *fakes.DependencyService
		clock             chronos.Clock
		buffer            *bytes.Buffer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbPath, err = ioutil.TempDir("", "cnb-path")
		Expect(err).NotTo(HaveOccurred())

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "http",
			Metadata: map[string]interface{}{
				"version-source": "BP_ROADRUNNER_VERSION",
				"version":        "2.1.1",
				"launch":         true,
			},
		}
		entryResolver.MergeLayerTypesCall.Returns.Launch = true

		dependencyService = &fakes.DependencyService{}
		dependencyService.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:           "road-runner",
			SHA256:       "some-sha",
			Source:       "some-source",
			SourceSHA256: "some-source-sha",
			Stacks:       []string{"some-stack"},
			URI:          "some-uri",
			Version:      "2.1.1",
		}

		now := time.Now()
		clock = chronos.NewClock(func() time.Time { return now })
		timestamp = now.Format(time.RFC3339Nano)

		buffer = bytes.NewBuffer(nil)
		logEmitter := roadrunner.NewLogEmitter(buffer)

		build = roadrunner.Build(entryResolver, dependencyService, clock, logEmitter)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbPath)).To(Succeed())
	})

	it("builds roadrunner", func() {
		result, err := build(packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "1.2.3",
			},
			WorkingDir: workingDir,
			Layers:     packit.Layers{Path: layersDir},
			CNBPath:    cnbPath,
			Stack:      "some-stack",
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "road-runner",
						Metadata: map[string]interface{}{
							"version-source": "BP_ROADRUNNER_VERSION",
							"version":        "2.1.1",
							"launch":         true,
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.BuildResult{
			Layers: []packit.Layer{
				{
					Name:   "road-runner",
					Path:   filepath.Join(layersDir, "road-runner"),
					Launch: true,
					SharedEnv: packit.Environment{
						"PATH.append": filepath.Join(layersDir, "road-runner", "sbin"),
						"PATH.delim":  ":",
					},
					BuildEnv: packit.Environment{},
					LaunchEnv: packit.Environment{
						"APP_ROOT.override":    workingDir,
						"SERVER_ROOT.override": filepath.Join(layersDir, "road-runner"),
					},
					ProcessLaunchEnv: map[string]packit.Environment{},
					Metadata: map[string]interface{}{
						"built_at":  timestamp,
						"cache_sha": "some-sha",
					},
				},
			},
			Launch: packit.LaunchMetadata{
				Processes: []packit.Process{
					{
						Type:    "web",
						Command: "rr serve -v -d",
					},
				},
			},
		}))

		Expect(dependencyService.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbPath, "buildpack.toml")))
		Expect(dependencyService.ResolveCall.Receives.Name).To(Equal("road-runner"))
		Expect(dependencyService.ResolveCall.Receives.Version).To(Equal("2.1.1"))
		Expect(dependencyService.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyService.DeliverCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:           "road-runner",
			SHA256:       "some-sha",
			Source:       "some-source",
			SourceSHA256: "some-source-sha",
			Stacks:       []string{"some-stack"},
			URI:          "some-uri",
			Version:      "2.1.1",
		}))
		Expect(dependencyService.DeliverCall.Receives.CnbPath).To(Equal(cnbPath))
		Expect(dependencyService.DeliverCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "road-runner")))
	})

	context("when the entry contains a version constraint", func() {
		it.Before(func() {
			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "http",
				Metadata: map[string]interface{}{
					"version-source": "BP_ROADRUNNER_VERSION",
					"version":        "2.1.*",
					"launch":         true,
				},
			}

			dependencyService.ResolveCall.Returns.Dependency = postal.Dependency{
				ID:           "road-runner",
				SHA256:       "some-sha",
				Source:       "some-source",
				SourceSHA256: "some-source-sha",
				Stacks:       []string{"some-stack"},
				URI:          "some-uri",
				Version:      "2.1.1",
			}
		})
		it("builds roadrunner with that version", func() {
			result, err := build(packit.BuildContext{
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "1.2.3",
				},
				WorkingDir: workingDir,
				Layers:     packit.Layers{Path: layersDir},
				CNBPath:    cnbPath,
				Stack:      "some-stack",
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "road-runner",
							Metadata: map[string]interface{}{
								"version-source": "BP_ROADRUNNER_VERSION",
								"version":        "2.1.*",
								"launch":         true,
							},
						},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.BuildResult{
				Layers: []packit.Layer{
					{
						Name:   "road-runner",
						Path:   filepath.Join(layersDir, "road-runner"),
						Launch: true,
						SharedEnv: packit.Environment{
							"PATH.append": filepath.Join(layersDir, "road-runner", "sbin"),
							"PATH.delim":  ":",
						},
						BuildEnv: packit.Environment{},
						LaunchEnv: packit.Environment{
							"APP_ROOT.override":    workingDir,
							"SERVER_ROOT.override": filepath.Join(layersDir, "road-runner"),
						},
						ProcessLaunchEnv: map[string]packit.Environment{},
						Metadata: map[string]interface{}{
							"built_at":  timestamp,
							"cache_sha": "some-sha",
						},
					},
				},
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: "rr serve -v -d",
						},
					},
				},
			}))

			Expect(dependencyService.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbPath, "buildpack.toml")))
			Expect(dependencyService.ResolveCall.Receives.Name).To(Equal("road-runner"))
			Expect(dependencyService.ResolveCall.Receives.Version).To(Equal("2.1.*"))
			Expect(dependencyService.ResolveCall.Receives.Stack).To(Equal("some-stack"))

			Expect(dependencyService.DeliverCall.Receives.Dependency).To(Equal(postal.Dependency{
				ID:           "road-runner",
				SHA256:       "some-sha",
				Source:       "some-source",
				SourceSHA256: "some-source-sha",
				Stacks:       []string{"some-stack"},
				URI:          "some-uri",
				Version:      "2.1.1",
			}))
			Expect(dependencyService.DeliverCall.Receives.CnbPath).To(Equal(cnbPath))
			Expect(dependencyService.DeliverCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "road-runner")))
		})
	})

	context("when the version source is buildpack.yml", func() {
		it.Before(func() {
			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "http",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
					"version":        "2.1.1",
					"launch":         true,
				},
			}
		})

		it("builds roadrunner with that version", func() {
			_, err := build(packit.BuildContext{
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "1.2.3",
				},
				WorkingDir: workingDir,
				Layers:     packit.Layers{Path: layersDir},
				CNBPath:    cnbPath,
				Stack:      "some-stack",
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "road-runner",
							Metadata: map[string]interface{}{
								"version-source": "buildpack.yml",
								"version":        "2.1.1",
								"launch":         true,
							},
						},
					},
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(dependencyService.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbPath, "buildpack.toml")))
			Expect(dependencyService.ResolveCall.Receives.Name).To(Equal("road-runner"))
			Expect(dependencyService.ResolveCall.Receives.Version).To(Equal("2.1.1"))
			Expect(dependencyService.ResolveCall.Receives.Stack).To(Equal("some-stack"))

			Expect(buffer.String()).To(ContainSubstring("WARNING: Setting the server version through buildpack.yml will be deprecated soon in Apache HTTP Server Buildpack v2.0.0"))
			Expect(buffer.String()).To(ContainSubstring("Please specify the version through the $BP_ROADRUNNER_VERSION environment variable instead. See docs for more information."))
		})
	})

	context("failure cases", func() {
		//context("when the httpd layer cannot be retrieved", func() {
		//	it.Before(func() {
		//		Expect(ioutil.WriteFile(filepath.Join(layersDir, "roadrunner.toml"), nil, 0000)).To(Succeed())
		//	})
		//
		//	it("returns an error", func() {
		//		_, err := build(packit.BuildContext{
		//			BuildpackInfo: packit.BuildpackInfo{
		//				Name:    "Some Buildpack",
		//				Version: "1.2.3",
		//			},
		//			WorkingDir: workingDir,
		//			Layers:     packit.Layers{Path: layersDir},
		//			CNBPath:    cnbPath,
		//			Stack:      "some-stack",
		//			Plan: packit.BuildpackPlan{
		//				Entries: []packit.BuildpackPlanEntry{
		//					{Name: "road-runner", Metadata: map[string]interface{}{"launch": true}},
		//				},
		//			},
		//		})
		//		Expect(err).To(MatchError(ContainSubstring("permission denied")))
		//	})
		//})

		context("when the dependency cannot be resolved", func() {
			it.Before(func() {
				dependencyService.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "1.2.3",
					},
					WorkingDir: workingDir,
					Layers:     packit.Layers{Path: layersDir},
					CNBPath:    cnbPath,
					Stack:      "some-stack",
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: "road-runner", Metadata: map[string]interface{}{"launch": true}},
						},
					},
				})
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})

		context("when the dependency cannot be installed", func() {
			it.Before(func() {
				dependencyService.DeliverCall.Returns.Error = errors.New("failed to install dependency")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "1.2.3",
					},
					WorkingDir: workingDir,
					Layers:     packit.Layers{Path: layersDir},
					CNBPath:    cnbPath,
					Stack:      "some-stack",
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: "road-runner", Metadata: map[string]interface{}{"launch": true}},
						},
					},
				})
				Expect(err).To(MatchError("failed to install dependency"))
			})
		})
	})
}
