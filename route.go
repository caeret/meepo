package meepo

import "net"

type Routes struct {
	nets map[byte][]*net.IPNet
}

func NewRoutes() *Routes {
	return &Routes{
		nets: make(map[byte][]*net.IPNet),
	}
}

func (r *Routes) Add(cidr string) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	r.nets[ipNet.IP.To4()[0]] = append(r.nets[ipNet.IP.To4()[0]], ipNet)
	return nil
}

func (r *Routes) Test(ip net.IP) bool {
	nets := r.nets[ip.To4()[0]]
	if len(nets) > 0 {
		for _, ipNet := range nets {
			if ipNet.Contains(ip.To4()) {
				return true
			}
		}
	}
	return false
}
