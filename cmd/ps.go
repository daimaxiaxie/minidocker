package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"minidocker/container"
	"os"
	"text/tabwriter"
)

var psCommand = &cobra.Command{
	Use:     "ps",
	Short:   "list all containers",
	Long:    "list all containers",
	Example: "minidocker ps",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return ListContainers()
	},
}

func init() {

}

func ListContainers() error {
	containers, err := container.GetAllContainer()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATE\n")
	for _, item := range containers {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", item.Id, item.Name, item.Pid, item.Status, item.Command, item.CreateTime)
	}
	if err = w.Flush(); err != nil {
		return err
	}
	return nil
}
