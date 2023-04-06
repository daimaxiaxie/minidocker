package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var logsCommand = &cobra.Command{
	Use:     "logs",
	Short:   "print logs",
	Long:    "print logs of a container",
	Example: "minidocker logs [CONTAINER]",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		LogContainer(args[0])
	},
}

func init() {

}

func LogContainer(containerName string) {
	pathUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	file, err := os.Open(pathUrl + ContainerLogFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(content))
}
