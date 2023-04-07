package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"minidocker/container"
	"os"
	"text/tabwriter"
)

var SkipList = map[string]bool{
	"network": true,
}

var psCommand = &cobra.Command{
	Use:     "ps",
	Short:   "list all containers",
	Long:    "list all containers",
	Example: "minidocker ps",
	Args:    cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ListContainers()
	},
}

func init() {

}

func ListContainers() {
	pathUrl := fmt.Sprintf(DefaultInfoLocation, "")
	pathUrl = pathUrl[:len(pathUrl)-1]
	files, err := ioutil.ReadDir(pathUrl)
	if err != nil {
		fmt.Println(err)
		return
	}

	var containers []*container.Info
	for _, file := range files {
		if _, ok := SkipList[file.Name()]; ok {
			continue
		}
		info, err := getContainerInfo(file)
		if err != nil {
			fmt.Println(err)
			continue
		}
		containers = append(containers, info)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATE\n")
	for _, item := range containers {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", item.Id, item.Name, item.Pid, item.Status, item.Command, item.CreateTime)
	}
	if err = w.Flush(); err != nil {
		fmt.Println(err)
	}
}

func getContainerInfo(file os.FileInfo) (*container.Info, error) {
	containerName := file.Name()
	configFile := fmt.Sprintf(DefaultInfoLocation, containerName) + ConfigName
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var containerInfo = new(container.Info)
	if err = json.Unmarshal(content, containerInfo); err != nil {
		return nil, err
	}
	return containerInfo, nil
}
