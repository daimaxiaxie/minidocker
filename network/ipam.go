package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath = "/var/run/minidocker/network/ipam/subnet.json"

var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

type IPAM struct {
	SubnetAllocatorPath string
	Subnets             *map[string]string
}

func (i *IPAM) load() error {
	if _, err := os.Stat(i.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	configFile, err := os.Open(i.SubnetAllocatorPath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	data := make([]byte, 2048)
	length, err := configFile.Read(data)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data[:length], i.Subnets); err != nil {
		return err
	}
	return nil
}

func (i *IPAM) dump() error {
	configDir, _ := path.Split(i.SubnetAllocatorPath)
	if _, err := os.Stat(configDir); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(configDir, 0644)
		} else {
			return err
		}
	}

	configFile, err := os.OpenFile(i.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer configFile.Close()

	data, err := json.Marshal(i.Subnets)
	if err != nil {
		return err
	}

	if _, err = configFile.Write(data); err != nil {
		return err
	}
	return nil
}

func (i *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	i.Subnets = &map[string]string{}

	if err := i.load(); err != nil {
		fmt.Println("load allocation info error ", err)
	}

	_, subnet, _ = net.ParseCIDR(subnet.String())
	ones, bits := subnet.Mask.Size()
	if _, exist := (*i.Subnets)[subnet.String()]; !exist {
		(*i.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(bits-ones))
	}

	for index := range (*i.Subnets)[subnet.String()] {
		if (*i.Subnets)[subnet.String()][index] == '0' {
			ipAlloc := []byte((*i.Subnets)[subnet.String()])
			ipAlloc[index] = '1'
			(*i.Subnets)[subnet.String()] = string(ipAlloc)
			ip = subnet.IP

			for offset := uint(4); offset > 0; offset -= 1 {
				[]byte(ip)[4-offset] += uint8(index >> ((offset - 1) * 8))
			}
			ip[3] += 1
			break
		}
	}
	_ = i.dump()
	return
}

func (i *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	i.Subnets = &map[string]string{}

	if err := i.load(); err != nil {
		fmt.Println("load allocation info error ", err)
	}

	_, subnet, _ = net.ParseCIDR(subnet.String())
	index := 0
	ip := ipaddr.To4()
	ip[3] -= 1
	for Offset := uint(4); Offset > 0; Offset -= 1 {
		index += int(ip[Offset-1]-subnet.IP[Offset-1]) << ((4 - Offset) * 8)
	}

	ipAlloc := []byte((*i.Subnets)[subnet.String()])
	ipAlloc[index] = '0'
	(*i.Subnets)[subnet.String()] = string(ipAlloc)
	_ = i.dump()
	return nil
}
