package main

import (
	"os"

	"github.com/laraboot-io/laraboot-buildpacks-roadrunner"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/draft"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	transport := cargo.NewTransport()
	dependencyService := postal.NewService(transport)
	logEmitter := roadrunner.NewLogEmitter(os.Stdout)
	versionParser := roadrunner.NewVersionParser()
	entryResolver := draft.NewPlanner()

	packit.Run(
		roadrunner.Detect(versionParser),
		roadrunner.Build(
			entryResolver,
			dependencyService,
			chronos.DefaultClock,
			logEmitter,
		),
	)
}
