package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os/exec"
)

var commitCommand = &cobra.Command{
	Use:     "commit",
	Short:   "commit a container into image",
	Long:    "commit a container into image",
	Example: "minidocker commit [CONTAINER] [IMAGE]",
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		Commit(args[0], args[1])
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {

}

func Commit(containerName string, imageName string) {
	mntURL := fmt.Sprintf(MntURL, containerName) + "/"
	imageTar := RootURL + "/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		fmt.Println(err)
	}
}
