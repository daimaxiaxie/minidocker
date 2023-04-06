package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var removeCommand = &cobra.Command{
	Use:     "remove",
	Short:   "remove a container",
	Long:    "remove a container",
	Example: "minidocker remove [CONTAINER]",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RemoveContainer(args[0])
	},
}

func init() {

}

func RemoveContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		fmt.Println("get container info by name error ", err)
		return
	}

	if containerInfo.Status != STOP {
		fmt.Println("container is running")
		return
	}

	pathUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err = os.RemoveAll(pathUrl); err != nil {
		fmt.Println("remove dir ", pathUrl, " error ", err)
		return
	}

	DeleteWorkSpace(containerInfo.Volume, containerInfo.Name)
}
