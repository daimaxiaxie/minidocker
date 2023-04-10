package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"minidocker/container"
)

var removeCommand = &cobra.Command{
	Use:     "remove",
	Short:   "remove a container",
	Long:    "remove a container",
	Example: "minidocker remove [CONTAINER]",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return RemoveContainer(args[0])
	},
}

func init() {

}

func RemoveContainer(containerName string) error {
	containerInfo, err := container.GetContainerInfoByName(containerName)
	if err != nil {
		return fmt.Errorf("get container info by name error %s", err)
	}
	if containerInfo.Status != container.STOP {
		return fmt.Errorf("container is running")
	}

	container.DestroyContainer(containerInfo.Name, containerInfo.Volume)
	return nil
}
