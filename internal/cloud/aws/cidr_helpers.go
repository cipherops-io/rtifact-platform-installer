package aws

import (
	"fmt"
	"net"
)

func cidrInAllowed(cidr string, allowed []*net.IPNet) bool {
	ip, n, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	for _, a := range allowed {
		if a.Contains(ip) {
			ones, _ := n.Mask.Size()
			aOnes, _ := a.Mask.Size()
			if ones >= aOnes {
				return true
			}
		}
	}
	return false
}

func checkCIDROverlaps(cidrs []string) error {
	var nets []*net.IPNet
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return fmt.Errorf("invalid CIDR %s: %w", c, err)
		}
		for _, existing := range nets {
			if n.Contains(existing.IP) || existing.Contains(n.IP) {
				return fmt.Errorf("CIDR overlap: %s and %s", c, existing.String())
			}
		}
		nets = append(nets, n)
	}
	return nil
}