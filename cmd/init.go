package cmd

import "C"
import (
	"fmt"
	"github.com/spf13/cobra"
	"minidocker/container"
)

var initCommand = &cobra.Command{
	Use:     "init",
	Short:   "Init a container",
	Long:    "start a container process",
	Example: "minidocker init [command]",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := container.Init(); err != nil {
			return fmt.Errorf("container error : %s", err)
		}
		return nil
	},
	Hidden: true,
}
