package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strconv"
	"syscall"
)

var stopCommand = &cobra.Command{
	Use:     "stop",
	Short:   "stop a container",
	Long:    "stop a container by SIGTERM",
	Example: "minidocker stop [CONTAINER]",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		StopContainer(args[0])
	},
}

func init() {

}

func StopContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		fmt.Println("get container info by name error ", err)
		return
	}
	pid, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		fmt.Println("get container pid error ", err)
		return
	}

	if err = syscall.Kill(pid, syscall.SIGTERM); err != nil {
		fmt.Println("stop container error ", err)
		return
	}

	containerInfo.Status = STOP
	containerInfo.Pid = ""

	newContent, err := json.Marshal(containerInfo)
	if err != nil {
		fmt.Println("marshal container info error ", err)
		return
	}

	configFile := fmt.Sprintf(DefaultInfoLocation, containerName) + ConfigName
	if err = ioutil.WriteFile(configFile, newContent, 0622); err != nil {
		fmt.Println("write container config file error ", err)
	}
}

func getContainerInfoByName(containerName string) (*ContainerInfo, error) {
	pathUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	content, err := ioutil.ReadFile(pathUrl + ConfigName)
	if err != nil {
		return nil, err
	}
	var containerInfo = new(ContainerInfo)
	if err = json.Unmarshal(content, containerInfo); err != nil {
		return nil, err
	}
	return containerInfo, nil
}
