package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var initCommand = &cobra.Command{
	Use:     "init",
	Short:   "Init a container",
	Long:    "start a container process",
	Example: "minidocker init [command]",
	Args:    cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := ContainerInit(); err != nil {
			fmt.Println("container error : ", err)
			return
		}

	},
	Hidden: true,
}

func init() {

}

func ContainerInit() error {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		return err
	}
	commands := strings.Split(string(msg), " ")
	if commands == nil || len(commands) == 0 {
		return fmt.Errorf("run container get command error, args is nil")
	}

	/*
		https://github.com/xianlubird/mydocker/issues/41
		mount namespace default shared
	*/
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	//defer syscall.Unmount("/proc", defaultMountFlags)
	path, err := exec.LookPath(commands[0])
	if err != nil {
		return err
	}
	if err := syscall.Exec(path, commands, os.Environ()); err != nil {
		return err
	}
	return nil
}
