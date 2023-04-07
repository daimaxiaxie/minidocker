package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"math/rand"
	"minidocker/cgroups"
	"minidocker/cgroups/subsystems"
	"minidocker/container"
	"minidocker/network"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var runCommand = &cobra.Command{
	Use:     "run [OPTIONS] IMAGE [COMMAND] [ARG...]",
	Short:   "Fast start a container",
	Long:    "Create a container with namespace and cgroups limit",
	Example: "minidocker run -t IMAGE [command]",
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
		volume, err := cmd.Flags().GetString("volume")
		if err != nil {
			return
		}
		detach, err := cmd.Flags().GetBool("detach")
		if err != nil {
			return
		}
		if tty && detach {
			fmt.Println("tty and detach can not both provided")
			return
		} else if !(tty || detach) {
			tty = true
		}
		containerName, err := cmd.Flags().GetString("name")
		if err != nil {
			return
		}
		env, err := cmd.Flags().GetStringSlice("env")
		if err != nil {
			return
		}
		net, err := cmd.Flags().GetString("net")
		portMapping, err := cmd.Flags().GetStringSlice("port")
		imageName := args[0]
		Run(tty, args[1:], volume, res, containerName, imageName, env, net, portMapping)
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

func Run(tty bool, commands []string, volume string, res *subsystems.ResourceConfig, containerName string, imageName string, env []string, net string, portMapping []string) {
	fmt.Println(tty, commands, res)
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	id := RandID(10)
	if len(containerName) == 0 {
		containerName = id
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		pathUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(pathUrl, 0622); err != nil {
			fmt.Println(err)
			return
		}
		logFilePath := pathUrl + ContainerLogFile
		logFile, err := os.Create(logFilePath)
		if err != nil {
			fmt.Println(err)
			return
		}
		cmd.Stdout = logFile
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Env = append(os.Environ(), env...)

	NewWorkSpace(volume, containerName, imageName)
	cmd.Dir = fmt.Sprintf(MntURL, containerName)

	if err = cmd.Start(); err != nil {
		fmt.Println(err)
		return
	}

	containerName, err = recordContainerInfo(cmd.Process.Pid, commands, containerName, id, volume)
	if err != nil {
		fmt.Println("Record container info error ", err)
		return
	}

	cgroupManager := cgroups.NewCgroupManager("minidocker-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Set(res)
	cgroupManager.Apply(cmd.Process.Pid)

	if net != "" {
		network.Init()
		info := &container.Info{
			Pid:         strconv.Itoa(cmd.Process.Pid),
			Id:          id,
			Name:        containerName,
			PortMapping: portMapping,
		}
		if err := network.Connect(net, info); err != nil {
			fmt.Println(err)
			return
		}
	}

	writePipe.WriteString(strings.Join(commands, " "))
	writePipe.Close()

	if tty {
		cmd.Wait()
		deleteContainerInfo(containerName)
		DeleteWorkSpace(volume, containerName)
	}
}

// NewWorkSpace Create a AUFS
func NewWorkSpace(volume string, containerName string, imageName string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	_ = CreateMountPoint(containerName, imageName)

	if len(volume) > 0 {
		volumeURLs := volumeUrlExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(volumeURLs, containerName)
		} else {
			fmt.Println("mount volume error")
		}
	}
}

func CreateReadOnlyLayer(imageName string) {
	unTarFolderURL := RootURL + "/" + imageName + "/"
	imageURL := RootURL + "/" + imageName + ".tar"
	exist, err := PathExist(unTarFolderURL)
	if err != nil {
		fmt.Println("fail to judge dir ", unTarFolderURL, " exists ", err)
	}
	if !exist {
		if err = os.Mkdir(unTarFolderURL, 0777); err != nil {
			fmt.Println(err)
		}
		if _, err := exec.Command("tar", "-xvf", imageURL, "-C", unTarFolderURL).CombinedOutput(); err != nil {
			fmt.Println(err)
		}
	}
}

func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerURL, containerName)
	if err := os.Mkdir(writeURL, 0777); err != nil {
		fmt.Println(err)
	}
}

func MountVolume(volumeURLs []string, containerName string) {
	parentURL := volumeURLs[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		fmt.Println(err)
	}
	containerURL := volumeURLs[1]
	containerVolumeURL := fmt.Sprintf(MntURL, containerName) + "/" + containerURL
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		fmt.Println(err)
	}
	dirs := "dirs=" + parentURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

func CreateMountPoint(containerName string, imageName string) error {
	mntURL := fmt.Sprintf(MntURL, containerName)
	if err := os.Mkdir(mntURL, 0777); err != nil {
		fmt.Println(err)
	}

	writeLayer := fmt.Sprintf(WriteLayerURL, containerName)
	imageLayer := RootURL + "/" + imageName
	dirs := "dirs=" + writeLayer + ":" + imageLayer
	if _, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL).CombinedOutput(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func DeleteWorkSpace(volume string, containerName string) {
	if len(volume) > 0 {
		volumeURLs := volumeUrlExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(volumeURLs, containerName)
		} else {
			DeleteMountPoint(containerName)
		}
	} else {
		DeleteMountPoint(containerName)
	}

	writeURL := fmt.Sprintf(WriteLayerURL, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		fmt.Println(err)
	}
}

func DeleteMountPoint(containerName string) {
	mntURL := fmt.Sprintf(MntURL, containerName)
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		fmt.Println(err)
	}
}

func DeleteMountPointWithVolume(volumeURLs []string, containerName string) {
	mntURL := fmt.Sprintf(MntURL, containerName)
	containerURL := mntURL + "/" + volumeURLs[1]
	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}

	DeleteMountPoint(containerName)
}

func recordContainerInfo(pid int, commands []string, containerName string, id string, volume string) (string, error) {

	containerInfo := &container.Info{
		Pid:        strconv.Itoa(pid),
		Id:         id,
		Name:       containerName,
		Command:    strings.Join(commands, " "),
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Status:     RUNNING,
		Volume:     volume,
	}
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return "", err
	}

	pathUrl := fmt.Sprintf(DefaultInfoLocation, containerInfo.Name)
	if err := os.MkdirAll(pathUrl, 0622); err != nil {
		return "", err
	}
	fileName := pathUrl + "/" + ConfigName
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err = file.Write(jsonBytes); err != nil {
		return "", err
	}
	return containerName, nil
}

func deleteContainerInfo(containerId string) {
	pathUrl := fmt.Sprintf(DefaultInfoLocation, containerId)
	if err := os.RemoveAll(pathUrl); err != nil {
		fmt.Println("Remove dir ", pathUrl, " error ", err)
	}
}

func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func volumeUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
}

func RandID(length int) string {
	rand.Seed(time.Now().UnixNano())
	res := make([]byte, length)
	candidate := "0123456789"
	for i := range res {
		res[i] = candidate[rand.Intn(len(candidate))]
	}
	return string(res)
}

const (
	RUNNING string = "running"
	STOP    string = "stopped"
	EXIT    string = "exited"

	DefaultInfoLocation string = "/var/run/minidocker/%s/"
	ConfigName          string = "config.json"
	ContainerLogFile    string = "container.log"
	RootURL             string = "/root"
	MntURL              string = "/root/mnt/%s"
	WriteLayerURL       string = "/root/writeLayer/%s"
)
