package main

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func buildPonzuServer() error {
	err := copyContentAndAddons()
	if err != nil {
		return err
	}

	// execute go build -o ponzu-cms cmd/ponzu/*.go
	cmdPackageName := strings.Join([]string{".", "cmd", "ponzu"}, "/")
	buildOptions := []string{"build", "-o", buildOutputName(), cmdPackageName}
	return execAndWait(gocmd, buildOptions...)
}

// copyContentAndAddons copies all ./content and addon files to internal vendor directory
func copyContentAndAddons() error {
	src := "content"
	dst := filepath.Join("cmd", "ponzu", "vendor", "github.com", "ponzu-cms", "ponzu", "content")
	err := emptyDir(dst)
	if err != nil {
		return err
	}
	err = copyFilesWarnConflicts(src, dst, []string{"doc.go"})
	if err != nil {
		return err
	}

	// copy all ./addons files & dirs to internal vendor directory
	src = "addons"
	dst = filepath.Join("cmd", "ponzu", "vendor")
	err = copyFilesWarnConflicts(src, dst, nil)
	return err
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build will build/compile the project to then be run.",
	Long: `From within your Ponzu project directory, running build will copy and move
the necessary files from your workspace into the vendored directory, and
will build/compile the project to then be run.

By providing the 'gocmd' flag, you can specify which Go command to build the
project, if testing a different release of Go.

Errors will be reported, but successful build commands return nothing.`,
	Example: `$ ponzu build
(or)
$ ponzu -gocmd=go1.8rc1 build`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return buildPonzuServer()
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
