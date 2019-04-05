package action

import (
	"fmt"
	"os"

	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/cobra"
)

var globalOptions = struct {
	Box *packr.Box
}{}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gmucli",
	Short: "gmucli is a command line tool to manage gRPC microservices creation",
	Long: `gmucli is a heavy opinionated command line utility than allows for easy creation of gRPC microservers.

It allows to include REST gateway (via grpc-gateway), monitoring (prometheus), middleware (logging, status, ...) in a hopefully easy way.

Developed with ♡ in Barcelona by the GMU Team`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		globalOptions.Box = packr.New("Templates", "../templates")
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
