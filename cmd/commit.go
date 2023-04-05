package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os/exec"
)

var commitCommand = &cobra.Command{
	Use:     "commit IMAGE",
	Short:   "commit a container into image",
	Long:    "commit a container into image",
	Example: "minidocker commit IMAGE",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		Commit(args[0])
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {

}

func Commit(imageName string) {
	mntURL := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		fmt.Println(err)
	}
}
