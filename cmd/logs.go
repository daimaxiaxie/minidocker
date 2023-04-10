package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"minidocker/container"
)

var logsCommand = &cobra.Command{
	Use:     "logs",
	Short:   "print logs",
	Long:    "print logs of a container",
	Example: "minidocker logs [CONTAINER]",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return LogContainer(args[0])
	},
}

func init() {

}

func LogContainer(containerName string) error {
	if content, err := container.GetContainerLog(containerName); err != nil {
		return err
	} else {
		fmt.Println(string(content))
	}

	return nil
}
