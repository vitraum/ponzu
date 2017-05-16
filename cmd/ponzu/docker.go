package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	dockerCmd = &cobra.Command{
		Use:   "docker",
		Short: "build a docker container of the project.",
		Long: `ponzu docker builds a docker container from the current project.
Database files will be added to the container.`,
		Example: `$ ponzu docker`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return buildDocker(dockerTags)
		},
	}
	dockerTags []string
	depcmd     string
)

// buildDocker builds a docker container using the docker command
func buildDocker(tags []string) error {
	err := buildPonzuServerForDocker(depcmd)
	if err != nil {
		return err
	}

	err = buildDockerImage(tags)
	if err != nil {
		return err
	}

	return nil
}

func buildPonzuServerForDocker(depcmd string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = copyContentAndAddons()
	if err != nil {
		return err
	}

	packageName, err := getPackageName()
	if err != nil {
		return err
	}

	if depcmd != "" {
		dep := strings.Split(depcmd, " ")

		fmt.Printf("ponzu docker: %s\n", strings.Join(dep, " "))
		cmd := exec.Command(dep[0], dep[1:]...)
		cmd.Dir = filepath.Join(pwd, "cmd", "ponzu")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		err := cmd.Start()
		if err != nil {
			return err

		}
		return cmd.Wait()
	}

	dockerDir := "/go/src/" + packageName
	volumeMapping := pwd + ":" + dockerDir

	dockerArgs := []string{
		"run",
		"--rm",
		"-w", dockerDir + "/cmd/ponzu",
		"-v", volumeMapping,
		"golang:1.8.1",
		"sh", "-c",
		`go get ./...` +
			` && CGO_ENABLED=0 go build -tags netgo -ldflags "-s -w -extldflags '-static'" -o ../../ponzu-server-docker .`,
	}

	fmt.Println("docker", strings.Join(dockerArgs, " "))
	return execAndWait("docker", dockerArgs...)
}

func buildDockerImage(tags []string) error {
	dockerArgs := []string{
		"build",
		"-f", "deployment/docker/Dockerfile.prebuilt",
	}

	for _, tag := range tags {
		dockerArgs = append(dockerArgs, "-t", tag)
	}

	dockerArgs = append(dockerArgs, ".")

	return execAndWait("docker", dockerArgs...)
}

// getPackageName returns the name of the project
func getPackageName() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	gopath, err := getGOPATH()
	if err != nil {
		return "", err
	}

	relpath, err := filepath.Rel(gopath, pwd)
	if err != nil {
		return "", err
	}
	if relpath[:3] == "src" {
		return relpath[4:], nil
	}
	return "", errors.New("current working directory is not in GOPATH")
}

func init() {
	dockerCmd.Flags().StringSliceVarP(&dockerTags, "tag", "t", []string{},
		"tag to add to this container, can be given multiple times")
	dockerCmd.Flags().StringVar(&depcmd, "depcmd", "",
		"dependency management command to be run before building the container")

	rootCmd.AddCommand(dockerCmd)
}
