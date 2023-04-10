package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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
			logger.Errorln("mount volume error")
		}
	}
}

func CreateReadOnlyLayer(imageName string) {
	unTarFolderURL := RootURL + "/" + imageName + "/"
	imageURL := RootURL + "/" + imageName + ".tar"
	exist, err := PathExist(unTarFolderURL)
	if err != nil {
		logger.Errorf("fail to judge dir %s exists %s", unTarFolderURL, err)
	}
	if !exist {
		if err = os.Mkdir(unTarFolderURL, 0777); err != nil {
			logger.Error(err)
		}
		if _, err := exec.Command("tar", "-xvf", imageURL, "-C", unTarFolderURL).CombinedOutput(); err != nil {
			logger.Error(err)
		}
	}
}

func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerURL, containerName)
	if err := os.Mkdir(writeURL, 0777); err != nil {
		logger.Error(err)
	}
}

func MountVolume(volumeURLs []string, containerName string) {
	parentURL := volumeURLs[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		logger.Error(err)
	}
	containerURL := volumeURLs[1]
	containerVolumeURL := fmt.Sprintf(MntURL, containerName) + "/" + containerURL
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		logger.Error(err)
	}
	dirs := "dirs=" + parentURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Error(err)
	}
}

func CreateMountPoint(containerName string, imageName string) error {
	mntURL := fmt.Sprintf(MntURL, containerName)
	if err := os.Mkdir(mntURL, 0777); err != nil {
		logger.Error(err)
	}

	writeLayer := fmt.Sprintf(WriteLayerURL, containerName)
	imageLayer := RootURL + "/" + imageName
	dirs := "dirs=" + writeLayer + ":" + imageLayer
	if _, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL).CombinedOutput(); err != nil {
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
		logger.Error(err)
	}
}

func DeleteMountPoint(containerName string) {
	mntURL := fmt.Sprintf(MntURL, containerName)
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Error(err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		logger.Error(err)
	}
}

func DeleteMountPointWithVolume(volumeURLs []string, containerName string) {
	mntURL := fmt.Sprintf(MntURL, containerName)
	containerURL := mntURL + "/" + volumeURLs[1]
	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Error(err)
	}

	DeleteMountPoint(containerName)
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
