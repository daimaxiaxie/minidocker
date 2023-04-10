package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func Init() error {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		return err
	}
	commands := strings.Split(string(msg), " ")
	if commands == nil || len(commands) == 0 {
		return fmt.Errorf("run container get command error, args is nil")
	}

	if err = setupMount(); err != nil {
		return err
	}

	path, err := exec.LookPath(commands[0])
	if err != nil {
		return err
	}

	if err := syscall.Exec(path, commands, os.Environ()); err != nil {
		return err
	}
	return nil
}

func setupMount() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	_ = syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err = pivotRoot(pwd); err != nil {
		return err
	}
	/*
		https://github.com/xianlubird/mydocker/issues/41
		mount namespace default shared
	*/
	//syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	_ = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	_ = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	return nil
}

func pivotRoot(root string) error {
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error : %v", err)
	}
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	return os.Remove(pivotDir)
}
