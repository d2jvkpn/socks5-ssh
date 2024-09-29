package utils

import (
	"fmt"
	"net"
	"regexp"
)

type DeivceIPs struct {
	Name string   `json:"name"`
	IPs  []string `json:"ips"`
}

func GetIPs() (items []DeivceIPs, err error) {
	var (
		re         *regexp.Regexp
		interfaces []net.Interface
	)

	// lo,virbr0,docker0,br-
	re = regexp.MustCompile(`^(lo|virbr|docker|br-)`)

	if interfaces, err = net.Interfaces(); err != nil {
		return nil, fmt.Errorf("fetching interfaces: %w", err)
	}

	for _, v := range interfaces {
		if re.MatchString(v.Name) {
			continue
		}
		var (
			item  DeivceIPs
			ip    net.IP
			addrs []net.Addr
		)
		item = DeivceIPs{
			Name: v.Name,
			IPs:  make([]string, 0),
		}

		if addrs, err = v.Addrs(); err != nil {
			continue
		}

		for _, addr := range addrs {
			switch t := addr.(type) {
			case *net.IPNet:
				ip = t.IP
			case *net.IPAddr:
				ip = t.IP
			}
			if ip != nil {
				item.IPs = append(item.IPs, ip.String())
			}
		}

		items = append(items, item)
	}

	return items, nil
}
