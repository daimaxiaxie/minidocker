package cmd

import (
	"github.com/spf13/cobra"
	"minidocker/container"
)

var stopCommand = &cobra.Command{
	Use:     "stop",
	Short:   "stop a container",
	Long:    "stop a container by SIGTERM",
	Example: "minidocker stop [CONTAINER]",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return container.StopContainer(args[0])
	},
}
