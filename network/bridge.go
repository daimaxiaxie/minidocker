package network

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
)

type BridgeNetworkDriver struct{}

func (b *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (b *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip

	n := &Network{
		Name:    name,
		IPRange: ipRange,
		Driver:  b.Name(),
	}

	err := b.init(n)
	if err != nil {
		logger.Errorf("init bridge error %s", err)
	}
	return n, err
}

func (b *BridgeNetworkDriver) Delete(network Network) error {
	br, err := netlink.LinkByName(network.Name)
	if err != nil {
		return err
	}
	return netlink.LinkDel(br)
}

func (b *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	br, err := netlink.LinkByName(network.Name)
	if err != nil {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	la.MasterIndex = br.Attrs().Index

	endpoint.Device = netlink.Veth{LinkAttrs: la, PeerName: "cif-" + endpoint.ID[:5]}

	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return err
	}

	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return err
	}
	return nil
}

func (b *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {
	return nil
}

func (b *BridgeNetworkDriver) init(n *Network) error {
	if err := createBridgeInterface(n.Name); err != nil {
		return err
	}

	gatewayIP := *n.IPRange
	gatewayIP.IP = n.IPRange.IP
	if err := setInterfaceIP(n.Name, gatewayIP.String()); err != nil {
		return err
	}

	if err := setInterfaceUP(n.Name); err != nil {
		return err
	}
	if err := setupIPTables(n.Name, n.IPRange); err != nil {
		return err
	}
	return nil
}

func createBridgeInterface(name string) error {
	_, err := net.InterfaceByName(name)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = name
	br := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge create failed %s : %v", name, err)
	}
	return nil
}

func setInterfaceIP(name string, ip string) error {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	ipNet, err := netlink.ParseIPNet(ip)
	if err != nil {
		return err
	}

	addr := &netlink.Addr{IPNet: ipNet}
	return netlink.AddrAdd(iface, addr)
}

func setInterfaceUP(name string) error {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	if err = netlink.LinkSetUp(iface); err != nil {
		return err
	}
	return nil
}

func setupIPTables(name string, subnet *net.IPNet) error {
	command := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), name)

	if _, err := exec.Command("iptables", strings.Split(command, " ")...).CombinedOutput(); err != nil {
		return err
	}
	return nil
}
