package meepo

import (
	"net"
)

type Routes struct {
	nets map[byte][]*net.IPNet
}

func NewRoutes() *Routes {
	return &Routes{
		nets: make(map[byte][]*net.IPNet),
	}
}

func (r *Routes) Add(cidr string) error {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	if x := ip.To4(); x != nil {
		ip = x
	} else {
		ip = x.To16()
	}
	r.nets[ip[0]] = append(r.nets[ip[0]], ipNet)
	return nil
}

func (r *Routes) Test(ip net.IP) bool {
	if x := ip.To4(); x != nil {
		ip = x
	} else {
		ip = x.To16()
	}
	nets := r.nets[ip[0]]
	if len(nets) > 0 {
		for _, ipNet := range nets {
			if ipNet.Contains(ip) {
				return true
			}
		}
	}
	return false
}
