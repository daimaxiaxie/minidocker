package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"minidocker/cgroups/subsystems"
	"minidocker/container"
	"minidocker/network"
)

var runCommand = &cobra.Command{
	Use:     "run [OPTIONS] IMAGE [COMMAND] [ARG...]",
	Short:   "Fast start a container",
	Long:    "Create a container with namespace and cgroups limit",
	Example: "minidocker run -t IMAGE [command]",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tty, err := cmd.Flags().GetBool("terminal")
		res := &subsystems.ResourceConfig{}
		res.MemoryLimit, err = cmd.Flags().GetString("memory")
		res.CpuShare, err = cmd.Flags().GetString("cpushare")
		res.CpuSet, err = cmd.Flags().GetString("cpuset")
		volume, err := cmd.Flags().GetString("volume")
		detach, err := cmd.Flags().GetBool("detach")
		if tty && detach {
			return fmt.Errorf("tty and detach can not both provided")
		} else if !(tty || detach) {
			tty = true
		}
		containerName, err := cmd.Flags().GetString("name")
		env, err := cmd.Flags().GetStringSlice("env")
		net, err := cmd.Flags().GetString("net")
		portMapping, err := cmd.Flags().GetStringSlice("port")
		imageName := args[0]
		if err != nil {
			return err
		}
		config := &container.Config{
			Resource:      res,
			Volume:        volume,
			ContainerName: containerName,
			ImageName:     imageName,
			Net:           net,
			Env:           env,
			PortMapping:   portMapping,
			Commands:      args[1:],
		}
		return Run(tty, config)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	runCommand.Flags().BoolP("terminal", "t", false, "enable tty")
	runCommand.Flags().BoolP("detach", "d", false, "detach container")
	runCommand.Flags().StringP("memory", "m", "1024m", "memory limit")
	runCommand.Flags().StringP("cpushare", "", "1024", "cpushare limit")
	runCommand.Flags().StringP("cpuset", "", "", "cpuset limit")
	runCommand.Flags().StringP("volume", "v", "", "volume")
	runCommand.Flags().StringP("name", "n", "", "container name")
	runCommand.Flags().StringSliceP("env", "e", []string{}, "set environment")
	runCommand.Flags().StringP("net", "", "", "join network")
	runCommand.Flags().StringSliceP("port", "p", []string{}, "port mapping")
	runCommand.Flags().SetInterspersed(false)
}

func Run(tty bool, config *container.Config) error {
	logger.Infof("use args : %v, %+v", tty, config)

	cmd, info, err := container.NewContainer(tty, config)
	if err != nil {
		return err
	}

	if config.Net != "" {
		_ = network.Init()
		if err := network.Connect(config.Net, &info); err != nil {
			return err
		}
	}

	if tty {
		_ = cmd.Wait()
		container.DestroyContainer(config.ContainerName, config.Volume)
	}
	return nil
}
