package data

import (
	"fmt"
	"net"
	"os"

	"github.com/ghetzel/go-stockutil/netutil"
	"github.com/ghetzel/go-stockutil/stringutil"
)

type Network struct {
}

func (self Network) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	if hostname, err := os.Hostname(); err == nil {
		out[`hostname`] = hostname
	}

	if fqdn := shell(`hostname -f`).String(); fqdn != `` {
		_, domain := stringutil.SplitPair(fqdn, `.`)

		out[`fqdn`] = fqdn
		out[`domain`] = domain
	}

	if defaultIP := netutil.DefaultAddress(); defaultIP != nil {
		out[`network.default.ip`] = defaultIP.IP.String()
		out[`network.default.interface`] = defaultIP.Interface.Name
		out[`network.default.gateway`] = defaultIP.Gateway.String()

		// old key name
		out[`network.ip`] = out[`network.default.ip`]
		out[`network.gateway`] = out[`network.default.gateway`]
	}

	if ifaces, err := net.Interfaces(); err == nil {
		for i, iface := range ifaces {
			prefix := fmt.Sprintf("network.interfaces.%d", i)

			out[prefix+`.name`] = iface.Name
			out[prefix+`.mtu`] = iface.MTU
			out[prefix+`.hwaddr`] = iface.HardwareAddr.String()

			if iface.Flags&net.FlagUp != 0 {
				out[prefix+`.up`] = true
			} else {
				out[prefix+`.up`] = false
			}

			if iface.Flags&net.FlagLoopback != 0 {
				out[prefix+`.link_type`] = `loopback`
			} else if iface.Flags&net.FlagPointToPoint != 0 {
				out[prefix+`.link_type`] = `pointtopoint`
			} else {
				out[prefix+`.link_type`] = `ethernet`
			}

			if addrs, err := iface.Addrs(); err == nil {
				for j, addr := range addrs {
					if ipnet, ok := addr.(*net.IPNet); ok {
						if ip4 := ipnet.IP.To4(); ip4 != nil {
							out[fmt.Sprintf("%s.addresses.%d.ipversion", prefix, j)] = 4
							out[fmt.Sprintf("%s.addresses.%d.ip", prefix, j)] = ip4.String()

							if len(ipnet.Mask) > 0 {
								out[fmt.Sprintf("%s.addresses.%d.netmask", prefix, j)] = net.ParseIP(
									`255.255.255.255`,
								).Mask(ipnet.Mask).String()
							}
						} else if ip6 := ipnet.IP.To16(); ip6 != nil {
							out[fmt.Sprintf("%s.addresses.%d.ipversion", prefix, j)] = 6
							out[fmt.Sprintf("%s.addresses.%d.ip", prefix, j)] = ip6.String()
						}

						if len(ipnet.Mask) > 0 {
							out[fmt.Sprintf("%s.addresses.%d.cidr", prefix, j)], _ = ipnet.Mask.Size()
						}
					} else {
						out[fmt.Sprintf("%s.addresses.%d.address", prefix, j)] = addr.String()
					}
				}
			}
		}
	}

	return out
}
