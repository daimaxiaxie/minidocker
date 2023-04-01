package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"minidocker/cgroups"
	"minidocker/cgroups/subsystems"
	"os"
	"os/exec"
	"strings"
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
		res := &subsystems.ResourceConfig{}
		if res.MemoryLimit, err = cmd.Flags().GetString("memory"); err != nil {
			return
		}
		if res.CpuShare, err = cmd.Flags().GetString("cpushare"); err != nil {
			return
		}
		if res.CpuSet, err = cmd.Flags().GetString("cpuset"); err != nil {
			return
		}
		Run(tty, cmd.Flags().Args(), res)
	},
}

func init() {
	runCommand.Flags().BoolP("terminal", "t", false, "enable tty")
	runCommand.Flags().StringP("memory", "m", "1024m", "memory limit")
	runCommand.Flags().StringP("cpushare", "", "1024", "cpushare limit")
	runCommand.Flags().StringP("cpuset", "", "", "cpuset limit")
}

func Run(tty bool, commands []string, res *subsystems.ResourceConfig) {
	fmt.Println(tty, commands, res)
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}

	if err = cmd.Start(); err != nil {
		fmt.Println(err)
		return
	}
	cgroupManager := cgroups.NewCgroupManager("minidocker-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Set(res)
	cgroupManager.Apply(cmd.Process.Pid)

	writePipe.WriteString(strings.Join(commands, " "))
	writePipe.Close()

	cmd.Wait()
}
