package roadrunner

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type VersionParser struct{}

func NewVersionParser() VersionParser {
	return VersionParser{}
}

func (v VersionParser) ParseVersion(path string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "*", "", nil
		}
		return "", "", fmt.Errorf("failed to parse buildpack.yml: %w", err)
	}

	defer file.Close()

	var buildpack struct {
		RoadRunner struct {
			Version string `yaml:"version"`
		} `yaml:"roadrunner"`
	}
	err = yaml.NewDecoder(file).Decode(&buildpack)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse buildpack.yml: %w", err)
	}

	if buildpack.RoadRunner.Version == "" {
		return "*", "", nil
	}

	return buildpack.RoadRunner.Version, "buildpack.yml", nil
}
