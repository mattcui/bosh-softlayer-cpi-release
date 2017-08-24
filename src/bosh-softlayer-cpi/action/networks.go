package action

import (
	"bosh-softlayer-cpi/registry"

	"bosh-softlayer-cpi/softlayer/virtual_guest_service"
	"fmt"
)

const (
	NetworkTypeManual string = "manual"
)

type Networks map[string]Network

type Network struct {
	Type            string                 `json:"type,omitempty"`
	IP              string                 `json:"ip,omitempty"`
	Gateway         string                 `json:"gateway,omitempty"`
	Netmask         string                 `json:"netmask,omitempty"`
	DNS             []string               `json:"dns,omitempty"`
	DHCP            bool                   `json:"use_dhcp,omitempty"`
	Default         []string               `json:"default,omitempty"`
	MAC             string                 `json:"mac,omitempty"`
	Alias           string                 `json:"alias,omitempty"`
	Routes          registry.Routes        `json:"routes,omitempty"`
	CloudProperties NetworkCloudProperties `json:"cloud_properties,omitempty"`
}

func (ns Networks) AsInstanceServiceNetworks() instance.Networks {
	networks := instance.Networks{}

	// Need to think deep
	for netName, network := range ns {
		netSlim := instance.Network{
			Type:    network.Type,
			IP:      network.IP,
			Gateway: network.Gateway,
			Netmask: network.Netmask,
			DNS:     network.DNS,
		}

		if len(network.CloudProperties.NetworkVlans) > 0 {
			parseCloudProperties(networks, netName, netSlim, network.CloudProperties.NetworkVlans, network.CloudProperties.SourcePolicyRouting)
		}
	}

	return networks
}

func (ns Networks) HasManualNetwork() bool {
	for _, network := range ns {
		if network.IsManual() {
			return true
		}
	}

	return false
}

func (n Network) IsManual() bool {
	return n.Type == NetworkTypeManual
}

func parseCloudProperties(networks instance.Networks, netName string, network instance.Network, networkVlans []NetworkVlan, sourcePolicyRouting bool) {
	for index, networkVlan := range networkVlans {
		var newNetName string
		var cloudProps instance.NetworkCloudProperties
		if networkVlan.SubnetId != 0 {
			cloudProps = instance.NetworkCloudProperties{
				VlanID:   networkVlan.VlanId,
				SubnetID: networkVlan.SubnetId,
			}
		} else {
			cloudProps = instance.NetworkCloudProperties{
				VlanID: networkVlan.VlanId,
			}
		}

		if index > 0 {
			newNetName = fmt.Sprintf("%s_%d", netName, index)
			network.CloudProperties = cloudProps
			networks[newNetName] = network
		} else {
			newNetName = netName
			network.CloudProperties = cloudProps
			networks[newNetName] = network
		}
	}
}
