package roadrunner

import (
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/postal"
	"log"
	"os"
	"path/filepath"
)

type SourceDep struct {
	dependency postal.Dependency
	localPath  string
	filePath   string
}

func (sourcedep *SourceDep) Download(cwd string) error {

	curl := pexec.NewExecutable("curl")
	tarFile := filepath.Join(cwd, "roadrunner.tar.gz")

	err := curl.Execute(pexec.Execution{
		Args: []string{
			"-L",
			sourcedep.dependency.URI,
			"-o",
			tarFile,
		},
		Stderr: os.Stderr,
	})

	if err != nil {
		return err
	}

	sourcedep.localPath = cwd
	sourcedep.filePath = tarFile

	return nil
}

func (sourcedep SourceDep) Untar() error {

	tar := pexec.NewExecutable("tar")

	err := tar.Execute(pexec.Execution{
		Args: []string{
			"-zxvf",
			sourcedep.filePath,
		},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	if err != nil {
		return err
	}

	return nil
}

func (sourcedep SourceDep) ModDownload(cwd string) error {

	goBin := pexec.NewExecutable("go")

	err := goBin.Execute(pexec.Execution{
		Dir: cwd,
		Args: []string{
			"mod",
			"download",
		},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	if err != nil {
		return err
	}

	return nil
}

func (sourcedep SourceDep) MakeCmd(args []string) error {

	makeBin := pexec.NewExecutable("make")

	err := makeBin.Execute(pexec.Execution{
		Dir:    sourcedep.localPath,
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	if err != nil {
		return err
	}

	return nil
}

func (sourcedep *SourceDep) WholeEnchilada(path string) error {

	err := sourcedep.Download(path)
	if err != nil {
		log.Printf("An error occurred downloading from source: %s\n", err)
		return err
	}

	err = sourcedep.Untar()
	if err != nil {
		log.Printf("An error occurred unpacking source tarball: %s\n", err)
		return err
	}

	err = sourcedep.ModDownload(path)
	if err != nil {
		log.Printf("An error occurred downloading go modules: %s\n", err)
		return err
	}

	err = sourcedep.MakeCmd([]string{"build"})
	if err != nil {
		log.Printf("An error occurred while building: %s\n", err)
		return err
	}

	return nil
}
