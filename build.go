package roadrunner

import (
	"fmt"
	"os"
	"path/filepath"
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

			roadRunnerLayer.Metadata = map[string]interface{}{
				"built_at":  clock.Now().Format(time.RFC3339Nano),
				"cache_sha": dependency.SHA256,
			}

			logger.Process("Configuring environment")

			roadRunnerLayer.SharedEnv.Append("PATH", filepath.Join(roadRunnerLayer.Path,
				fmt.Sprintf("roadrunner-%s-linux-amd64", dependency.Version)), ":")

			logger.Environment(roadRunnerLayer.SharedEnv)

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
						Command: "rr serve -d",
					},
				},
			},
		}, nil
	}
}
