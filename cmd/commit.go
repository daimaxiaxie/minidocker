package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"minidocker/container"
	"os/exec"
)

var commitCommand = &cobra.Command{
	Use:     "commit",
	Short:   "commit a container into image",
	Long:    "commit a container into image",
	Example: "minidocker commit [CONTAINER] [IMAGE]",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return Commit(args[0], args[1])
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func Commit(containerName string, imageName string) error {
	mntURL := fmt.Sprintf(container.MntURL, containerName) + "/"
	imageTar := container.RootURL + "/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		return err
	}
	return nil
}
