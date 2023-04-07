package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:     "minidocker",
	Short:   "A simple container runtime",
	Long:    "A simple container runtimr",
	Example: "minidocker run [command]",
	Args:    cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		fmt.Println("pre")
	},
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCommand.AddCommand(runCommand)
	rootCommand.AddCommand(initCommand)
	rootCommand.AddCommand(commitCommand)
	rootCommand.AddCommand(psCommand)
	rootCommand.AddCommand(logsCommand)
	rootCommand.AddCommand(execCommand)
	rootCommand.AddCommand(stopCommand)
	rootCommand.AddCommand(removeCommand)
	rootCommand.AddCommand(networkCommand)
}

func Execute() error {
	return rootCommand.Execute()
}
