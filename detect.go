package roadrunner

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

const PlanDependencyRoadRunner = "road-runner"
const PlanDependencyGolang = "go"
const PlanDependencyPhp = "php"

//go:generate faux --interface Parser --output fakes/parser.go
type Parser interface {
	ParseVersion(path string) (version, versionSource string, err error)
}

type BuildPlanMetadata struct {
	Version       string `toml:"version,omitempty"`
	VersionSource string `toml:"version-source,omitempty"`
	Launch        bool   `toml:"launch"`
	Build         bool   `toml:"build"`
}

type RRConfig struct {
	Rpc    string `toml:"rpc,omitempty"`
	Server string `toml:"server,omitempty"`
	Httpd  string `toml:"httpd,omitempty"`
	Logs   string `toml:"logs,omitempty"`
}

func Detect(parser Parser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		plan := packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: PlanDependencyRoadRunner},
				},
			},
		}

		_, err := os.Stat(filepath.Join(context.WorkingDir, ".rr.yaml"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return plan, nil
			}
			return packit.DetectResult{}, err
		}

		var requirements []packit.BuildPlanRequirement

		// Require Golang
		requirements = append(requirements, packit.BuildPlanRequirement{
			Name: PlanDependencyGolang,
			Metadata: BuildPlanMetadata{
				Version: "1.*",
				Build:   true,
				Launch:  false,
			},
		})

		// Require Php
		requirements = append(requirements, packit.BuildPlanRequirement{
			Name: PlanDependencyPhp,
			Metadata: BuildPlanMetadata{
				Version: "7.*",
				Build:   false,
				Launch:  true,
			},
		})

		if version, ok := os.LookupEnv("BP_ROADRUNNER_VERSION"); ok {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: PlanDependencyRoadRunner,
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: "BP_ROADRUNNER_VERSION",
					Launch:        true,
				},
			})
		}

		version, versionSource, err := parser.ParseVersion(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: PlanDependencyRoadRunner,
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: versionSource,
					Launch:        true,
				},
			})
			plan.Plan.Requires = requirements
		}

		return plan, nil
	}
}
