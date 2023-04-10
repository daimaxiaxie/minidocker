package cgroups

import "minidocker/cgroups/subsystems"

type CgroupManager struct {
	Path     string
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

func (c *CgroupManager) Apply(pid int) error {
	for _, subSys := range subsystems.Subsystems {
		_ = subSys.Apply(c.Path, pid)
	}
	return nil
}

func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSys := range subsystems.Subsystems {
		_ = subSys.Set(c.Path, res)
	}
	return nil
}
func (c *CgroupManager) Destroy() error {
	for _, subSys := range subsystems.Subsystems {
		if err := subSys.Remove(c.Path); err != nil {
			return err
		}
	}
	return nil
}
