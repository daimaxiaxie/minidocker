package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"syscall"
)

var runCommand = &cobra.Command{
	Use:     "run",
	Short:   "Fast start a container",
	Long:    "Create a container with namespace and cgroups limit",
	Example: "minidocker run -t [command]",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tty, err := cmd.Flags().GetBool("terminal")
		if err != nil {
			return
		}
		Run(tty, cmd.Flags().Arg(0))
	},
}

func init() {
	runCommand.Flags().BoolP("terminal", "t", false, "enable tty")
}

func Run(tty bool, command string) {
	fmt.Println(tty, command)
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
}
