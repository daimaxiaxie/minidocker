package network

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io/fs"
	"minidocker/container"
	"net"
	"os"
	"path"
	"path/filepath"
	"text/tabwriter"

	"github.com/vishvananda/netlink"
)

var (
	defaultNetworkPath = "/var/run/minidocker/network/network"
	drivers            = map[string]Driver{}
	networks           = map[string]*Network{}
	logger             = zap.NewExample().Sugar()
)

type Network struct {
	Name    string
	IPRange *net.IPNet
	Driver  string
}

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string         `json:"portMapping"`
	Network     *Network
}

type Driver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network Network, endpoint *Endpoint) error
}

func (n *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}

	dumpPath = path.Join(dumpPath, n.Name)
	file, err := os.OpenFile(dumpPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(n)
	if err != nil {
		return err
	}

	if _, err = file.Write(data); err != nil {
		return err
	}
	return nil
}

func (n *Network) load(dumpPath string) error {
	configFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	data := make([]byte, 2048)
	length, err := configFile.Read(data)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data[:length], n); err != nil {
		return err
	}
	return nil
}

func (n *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, n.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	return os.Remove(path.Join(dumpPath, n.Name))
}

func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
		}
	}

	_ = filepath.Walk(defaultNetworkPath, func(filepath string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		_, name := path.Split(filepath)
		network := &Network{
			Name: name,
		}

		if err := network.load(filepath); err != nil {
			logger.Errorf("load network error %s", err)
		}

		networks[name] = network
		return nil
	})
	return nil
}

func CreateNetwork(driver, subnet, name string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	gatewayIP, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = gatewayIP

	network, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}

	return network.dump(defaultNetworkPath)
}

func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprintf(w, "NAME\tIPRange\tDriver\n")
	for _, network := range networks {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", network.Name, network.IPRange.String(), network.Driver)
	}
	if err := w.Flush(); err != nil {
		return
	}
}

func DeleteNetwork(name string) error {
	network, ok := networks[name]
	if !ok {
		return fmt.Errorf("no suach network %s", name)
	}

	if err := ipAllocator.Release(network.IPRange, &network.IPRange.IP); err != nil {
		return err
	}

	if err := drivers[network.Driver].Delete(*network); err != nil {
		return err
	}
	return network.remove(defaultNetworkPath)
}

func Connect(name string, info *container.Info) error {
	network, ok := networks[name]
	if !ok {
		return fmt.Errorf("no such network %s", name)
	}

	ip, err := ipAllocator.Allocate(network.IPRange)
	if err != nil {
		return err
	}

	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", info.Id, name),
		IPAddress:   ip,
		Network:     network,
		PortMapping: info.PortMapping,
	}

	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	if err = configEndpointIPAddressAndRouter(ep, info); err != nil {
		return err
	}
	return configPortMapping(ep, info)
}

func Disconnect(name string, info *container.Info) error {
	return nil
}
