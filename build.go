package roadrunner

import (
	"fmt"
	"github.com/paketo-buildpacks/packit/pexec"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
//go:generate faux --interface DependencyService --output fakes/dependency_service.go

type EntryResolver interface {
	Resolve(string, []packit.BuildpackPlanEntry, []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(string, []packit.BuildpackPlanEntry) (launch, build bool)
}

type DependencyService interface {
	Resolve(path, name, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
}

func Build(entries EntryResolver, dependencies DependencyService, clock chronos.Clock, logger LogEmitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title(context.BuildpackInfo)
		logger.Process("Resolving RoadRunner Server version")

		priorities := []interface{}{
			"BP_ROADRUNNER_VERSION",
			"buildpack.yml",
		}
		entry, sortedEntries := entries.Resolve("road-runner", context.Plan.Entries, priorities)
		logger.Candidates(sortedEntries)

		roadRunnerLayer, err := context.Layers.Get("road-runner")
		if err != nil {
			return packit.BuildResult{}, err
		}

		version, ok := entry.Metadata["version"].(string)
		if !ok {
			version = ""
		}

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), "road-runner", version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		source, _ := entry.Metadata["version-source"].(string)
		if source == "buildpack.yml" {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logger.Subprocess("WARNING: Setting the server version through buildpack.yml will be deprecated soon in Apache HTTP Server Buildpack v%s.", nextMajorVersion.String())
			logger.Subprocess("Please specify the version through the $BP_ROADRUNNER_VERSION environment variable instead. See docs for more information.")
			logger.Break()
		}

		if sha, ok := roadRunnerLayer.Metadata["cache_sha"].(string); !ok || sha != dependency.SHA256 {
			logger.Process("Executing build process")

			roadRunnerLayer, err = roadRunnerLayer.Reset()
			if err != nil {
				return packit.BuildResult{}, err
			}
			roadRunnerLayer.Launch, _ = entries.MergeLayerTypes("road-runner", context.Plan.Entries)

			logger.Subprocess("Installing RoadRunner Server %s", dependency.Version)
			duration, err := clock.Measure(func() error {
				platformPath, _ := os.MkdirTemp("", "platform")
				return dependencies.Deliver(dependency, context.CNBPath, roadRunnerLayer.Path, platformPath)

			})

			if err != nil {
				return packit.BuildResult{}, err
			}

			logger.Break()
			logger.Action("Completed in %s", duration.Round(time.Millisecond))

			if strings.HasPrefix(dependency.URI, "https://") {

				//// --------------
				logger.Subprocess("Downloading RoadRunner Server %s", dependency.URI)
				duration, err = clock.Measure(func() error {

					curl := pexec.NewExecutable("curl")
					tar := pexec.NewExecutable("tar")
					ls := pexec.NewExecutable("ls")

					err := curl.Execute(pexec.Execution{
						Args: []string{dependency.URI,
							"-o",
							filepath.Join(roadRunnerLayer.Path, "roadrunner.tar.gz"),
						},
						Stdout: os.Stdout,
					})

					err = tar.Execute(pexec.Execution{
						Args: []string{"-xvf",
							filepath.Join(roadRunnerLayer.Path, "roadrunner.tar.gz")},
						Stdout: os.Stdout,
					})

					if err != nil {
						panic(err)
					}

					err = ls.Execute(pexec.Execution{
						Args:   []string{filepath.Join(roadRunnerLayer.Path, "roadrunner.tar.gz")},
						Stdout: os.Stdout,
					})

					if err != nil {
						return err
					}

					return nil
				})

				if err != nil {
					fmt.Printf("Error: %s\n", err)
					return packit.BuildResult{}, err
				}
				logger.Break()
				logger.Action("Completed in %s", duration.Round(time.Millisecond))
				//// --------------

				//logger.Subprocess("Building RoadRunner Server %s", dependency.Version)
				//
				//dir := fmt.Sprintf("%s", filepath.Join(roadRunnerLayer.Path))
				//
				////Check if install succeeded and source path is available for build
				//if _, derr := os.Stat(dir); os.IsNotExist(derr) {
				//	log.Println(derr)
				//	return packit.BuildResult{}, derr
				//} else {
				//	buildDuration, prerr := clock.Measure(func() error {
				//
				//		// Run make to build RoadRunner specifying the directory (-C)
				//		return RunProcs(procmgr.Procs{
				//			Processes: map[string]procmgr.Proc{
				//				"buildRoadRunner": {
				//					Command: "make",
				//					Args:    []string{"-C", dir},
				//				},
				//			},
				//		})
				//
				//	})
				//
				//	if prerr != nil {
				//		log.Println(derr)
				//		return packit.BuildResult{}, derr
				//	}
				//
				//	logger.Break()
				//	logger.Action("Built in %s", buildDuration.Round(time.Millisecond))
				//}
			}

			roadRunnerLayer.Metadata = map[string]interface{}{
				"built_at":  clock.Now().Format(time.RFC3339Nano),
				"cache_sha": dependency.SHA256,
			}

			logger.Process("Configuring environment")
			roadRunnerLayer.LaunchEnv.Override("APP_ROOT", context.WorkingDir)
			roadRunnerLayer.LaunchEnv.Override("SERVER_ROOT", roadRunnerLayer.Path)

			logger.Environment(roadRunnerLayer.LaunchEnv)
		}

		return packit.BuildResult{
			Layers: []packit.Layer{roadRunnerLayer},
			Launch: packit.LaunchMetadata{
				Processes: []packit.Process{
					{
						Type:    "web",
						Command: "tail -f /dev/null",
					},
				},
			},
		}, nil
	}
}
