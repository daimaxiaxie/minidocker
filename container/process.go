package container

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"minidocker/cgroups"
	"minidocker/cgroups/subsystems"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

const (
	ENV_EXEC_PID = "minidocker_pid"
	ENV_EXEC_CMD = "minidocker_cmd"
)

var logger = zap.NewExample().Sugar()

type Config struct {
	Resource      *subsystems.ResourceConfig
	Volume        string
	ContainerName string
	ImageName     string
	Net           string
	Env           []string
	PortMapping   []string
	Commands      []string
}

func NewContainer(tty bool, config *Config) (*exec.Cmd, Info, error) {
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		return nil, Info{}, err
	}

	id := RandID(10)
	if len(config.ContainerName) == 0 {
		config.ContainerName = id
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		pathUrl := fmt.Sprintf(DefaultInfoLocation, config.ContainerName)
		if err := os.MkdirAll(pathUrl, 0622); err != nil {
			return nil, Info{}, err
		}
		logFilePath := pathUrl + LogFile
		logFile, err := os.Create(logFilePath)
		if err != nil {
			return nil, Info{}, err
		}
		cmd.Stdout = logFile
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Env = append(os.Environ(), config.Env...)

	NewWorkSpace(config.Volume, config.ContainerName, config.ImageName)
	cmd.Dir = fmt.Sprintf(MntURL, config.ContainerName)

	if err = cmd.Start(); err != nil {
		return nil, Info{}, err
	}

	info, err := recordContainerInfo(cmd.Process.Pid, config.Commands, config.ContainerName, id, config.Volume)
	if err != nil {
		return nil, Info{}, fmt.Errorf("record container info error %s", err)
	}

	cgroupManager := cgroups.NewCgroupManager("minidocker-cgroup")
	defer cgroupManager.Destroy()
	_ = cgroupManager.Set(config.Resource)
	_ = cgroupManager.Apply(cmd.Process.Pid)

	_, _ = writePipe.WriteString(strings.Join(config.Commands, " "))
	_ = writePipe.Close()

	return cmd, *info, nil
}

func StopContainer(containerName string) error {
	containerInfo, err := GetContainerInfoByName(containerName)
	if err != nil {
		return fmt.Errorf("get container info by name error %s", err)
	}
	pid, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		return fmt.Errorf("get container pid error %s", err)
	}

	if err = syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("stop container error %s", err)
	}

	containerInfo.Status = STOP
	containerInfo.Pid = ""

	newContent, err := json.Marshal(containerInfo)
	if err != nil {
		return fmt.Errorf("marshal container info error %s", err)
	}

	configFile := fmt.Sprintf(DefaultInfoLocation, containerName) + ConfigName
	if err = ioutil.WriteFile(configFile, newContent, 0622); err != nil {
		logger.Warnf("write container config file error %s", err)
	}
	return nil
}

func DestroyContainer(containerName, volume string) {
	deleteContainerInfo(containerName)
	DeleteWorkSpace(volume, containerName)
}

func ExecContainer(containerName, command string) error {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		return fmt.Errorf("get container pid by name error %s", err)
	}

	logger.Infof("process %s run %s", pid, command)
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_ = os.Setenv(ENV_EXEC_PID, pid)
	defer os.Unsetenv(ENV_EXEC_PID)
	_ = os.Setenv(ENV_EXEC_CMD, command)
	defer os.Unsetenv(ENV_EXEC_CMD)

	containerEnv := getEnvByPid(pid)
	cmd.Env = append(os.Environ(), containerEnv...)

	if err = cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getContainerPidByName(containerName string) (string, error) {
	containerInfo, err := GetContainerInfoByName(containerName)
	if err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

func getEnvByPid(pid string) []string {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Errorf("get process env error %s", err)
		return nil
	}
	env := strings.Split(string(content), "\u0000")
	return env
}
