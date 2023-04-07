package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"minidocker/network"
)

var networkCommand = &cobra.Command{
	Use:     "network",
	Short:   "manager network",
	Long:    "manager container network",
	Example: "minidocker network [COMMAND] [FLAGS]",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var createCommand = &cobra.Command{
	Use:   "create",
	Short: "create network",
	Long:  "create container network",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		network.Init()
		driver, _ := cmd.Flags().GetString("driver")
		subnet, _ := cmd.Flags().GetString("subnet")
		if err := network.CreateNetwork(driver, subnet, args[0]); err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	},
}

var listCommand = &cobra.Command{
	Use:   "list",
	Short: "list network",
	Long:  "list all container network",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = network.Init()
		network.ListNetwork()
		return nil
	},
}

var rmCommand = &cobra.Command{
	Use:   "remove",
	Short: "remove network",
	Long:  "remove container network",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		network.Init()
		if err := network.DeleteNetwork(args[0]); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	createCommand.Flags().StringP("driver", "d", "", "network driver")
	createCommand.Flags().StringP("subnet", "s", "", "subnet cider")

	networkCommand.AddCommand(createCommand)
	networkCommand.AddCommand(listCommand)
	networkCommand.AddCommand(rmCommand)
}
