package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	ENV_EXEC_PID = "minidocker_pid"
	ENV_EXEC_CMD = "minidocker_cmd"
)

var execCommand = &cobra.Command{
	Use:     "exec",
	Short:   "exec a command",
	Long:    "exec a command in container",
	Example: "minidocker exec [CONTAINER] [command]",
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv(ENV_EXEC_PID) != "" {
			return
		}

		ExecContainer(args[0], args[1:])
	},
}

func init() {
	execCommand.Flags().SetInterspersed(false)
}

func ExecContainer(containerName string, commands []string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		fmt.Println("get container pid by name error ", err)
		return
	}

	command := strings.Join(commands, " ")
	fmt.Println(pid, command)
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, pid)
	defer os.Unsetenv(ENV_EXEC_PID)
	os.Setenv(ENV_EXEC_CMD, command)
	defer os.Unsetenv(ENV_EXEC_CMD)

	containerEnv := getEnvByPid(pid)
	cmd.Env = append(os.Environ(), containerEnv...)

	if err = cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

func getContainerPidByName(containerName string) (string, error) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

func getEnvByPid(pid string) []string {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	env := strings.Split(string(content), "\u0000")
	return env
}
