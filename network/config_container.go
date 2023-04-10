package network

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"minidocker/container"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func configEndpointIPAddressAndRouter(ep *Endpoint, info *container.Info) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return err
	}
	defer enterContainerNetns(&peerLink, info)()

	interfaceIP := *ep.Network.IPRange
	interfaceIP.IP = ep.IPAddress
	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return err
	}
	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}
	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRouter := &netlink.Route{LinkIndex: peerLink.Attrs().Index, Gw: ep.Network.IPRange.IP, Dst: cidr}
	if err = netlink.RouteAdd(defaultRouter); err != nil {
		return err
	}
	return nil
}

func configPortMapping(ep *Endpoint, info *container.Info) error {
	for _, port := range ep.PortMapping {
		mapping := strings.Split(port, ":")
		if len(mapping) != 2 {
			logger.Errorln("port mapping format error")
			continue
		}

		command := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s", mapping[0], ep.IPAddress.String(), mapping[1])
		if _, err := exec.Command("iptables", strings.Split(command, " ")...).CombinedOutput(); err != nil {
			logger.Errorf("exec iptables error %s", err)
			continue
		}
	}
	return nil
}

func enterContainerNetns(link *netlink.Link, info *container.Info) func() {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", info.Pid), os.O_RDONLY, 0)
	if err != nil {
		logger.Errorf("get container net namespace error %s", err)
	}

	fd := f.Fd()
	runtime.LockOSThread()
	if err = netlink.LinkSetNsFd(*link, int(fd)); err != nil {
		logger.Errorf("set link netns error %s", err)
	}

	origins, err := netns.Get()
	if err != nil {
		logger.Errorf("get current netns error %s", err)
	}

	if err = netns.Set(netns.NsHandle(fd)); err != nil {
		logger.Errorf("set netns error %s", err)
	}

	return func() {
		_ = netns.Set(origins)
		_ = origins.Close()
		runtime.UnlockOSThread()
		_ = f.Close()
	}
}
