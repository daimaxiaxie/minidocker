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
		Run(tty, args, volume, res)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	runCommand.Flags().BoolP("terminal", "t", false, "enable tty")
	runCommand.Flags().StringP("memory", "m", "1024m", "memory limit")
	runCommand.Flags().StringP("cpushare", "", "1024", "cpushare limit")
	runCommand.Flags().StringP("cpuset", "", "", "cpuset limit")
	runCommand.Flags().StringP("volume", "v", "", "volume")
	runCommand.Flags().SetInterspersed(false)
}

func Run(tty bool, commands []string, volume string, res *subsystems.ResourceConfig) {
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
	mntURL := "/root/mnt/"
	rootURL := "/root/"
	NewWorkSpace(rootURL, mntURL, volume)
	cmd.Dir = mntURL

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
	DeleteWorkSpace(rootURL, mntURL, volume)
}

// NewWorkSpace Create a AUFS
func NewWorkSpace(rootURL string, mntURL string, volume string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)

	if len(volume) > 0 {
		volumeURLs := volumeUrlExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(rootURL, mntURL, volumeURLs)
		} else {
			fmt.Println("mount volume error")
		}
	}
}

func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExist(busyboxURL)
	if err != nil {
		fmt.Println("fail to judge dir ", busyboxURL, " exists ", err)
	}
	if !exist {
		if err = os.Mkdir(busyboxURL, 0777); err != nil {
			fmt.Println(err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			fmt.Println(err)
		}
	}
}

func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL, 0777); err != nil {
		fmt.Println(err)
	}
}

func MountVolume(rootURL string, mntURL string, volumeURLs []string) {
	parentURL := volumeURLs[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		fmt.Println(err)
	}
	containerURL := volumeURLs[1]
	containerVolumeURL := mntURL + containerURL
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

func CreateMountPoint(rootURL string, mntURL string) {
	if err := os.Mkdir(mntURL, 0777); err != nil {
		fmt.Println(err)
	}
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	if len(volume) > 0 {
		volumeURLs := volumeUrlExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(rootURL, mntURL, volumeURLs)
		} else {
			DeleteMountPoint(rootURL, mntURL)
		}
	} else {
		DeleteMountPoint(rootURL, mntURL)
	}

	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		fmt.Println(err)
	}
}

func DeleteMountPoint(rootURL string, mntURL string) {
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

func DeleteMountPointWithVolume(rootURL string, mntURL string, volumeURLs []string) {
	containerURL := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}

	DeleteMountPoint(rootURL, mntURL)
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
