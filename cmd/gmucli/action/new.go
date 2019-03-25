package action

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	newCmd = &cobra.Command{
		Use:   "new",
		Short: "create a new gRPC project",
		Run:   newCmdRun,
	}

	path string
)

func init() {
	RootCmd.AddCommand(newCmd)

	newCmd.Flags().StringVarP(&path, "path", "p", "", "path of the project. Defaults to current working directory")
}

func newCmdRun(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Printf("You must supply a path for the service, e.g gmu new my-project\n")
		return
	}

	projectPath, err := getProjectPath(path, args[0])
	if err != nil {
		fmt.Printf("can not set the project path. Error: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Creating project %v\n", args[0])
	fmt.Printf("Path: %v\n", projectPath)
	checkDirMustNotExist(projectPath)

	fmt.Println("Done!")
}

func getProjectPath(path, pname string) (string, error) {
	var projectPath, cwd string
	var err error
	if cwd, err = os.Getwd(); err != nil {
		return "", err
	}

	if &path == nil || path == "" {
		projectPath = cwd + "/" + pname
	} else {
		if !strings.HasPrefix(path, "/") {
			fmt.Printf("found relative path %v\n", path)
			projectPath = cwd + "/" + path + "/" + pname
			fmt.Printf("creating project at %s\n", projectPath)
		}
	}
	return projectPath, nil
}

func checkDirMustNotExist(path string) {
	if _, err := os.Stat(path); err == nil {
		// dir exists
		fmt.Printf("Error: directory %s already exists. Can not override", path)
		os.Exit(1)
	}
}
