package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"minidocker/container"
	"os"
	"strings"

	_ "minidocker/nsenter"
)

var execCommand = &cobra.Command{
	Use:     "exec",
	Short:   "exec a command",
	Long:    "exec a command in container",
	Example: "minidocker exec [CONTAINER] [command]",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv(container.ENV_EXEC_PID) != "" {
			logger.Infof("subProcess env %s, %s", os.Getenv(container.ENV_EXEC_PID), os.Getenv(container.ENV_EXEC_CMD))
			return nil
		}
		if len(args) < 2 {

			return fmt.Errorf("requires at least 2 args")
		}

		return ExecContainer(args[0], args[1:])
	},
}

func init() {
	execCommand.Flags().SetInterspersed(false)
}

func ExecContainer(containerName string, commands []string) error {
	command := strings.Join(commands, " ")
	return container.ExecContainer(containerName, command)
}
