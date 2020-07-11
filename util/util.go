package util

import (
	"encoding/binary"
	"net"
)

func ParseIPRange(rng string) ([]net.IP, error) {
	_, ipv4Net, err := net.ParseCIDR(rng)
	if err != nil {
		return nil, err
	}

	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)
	finish := (start & mask) | (mask ^ 0xffffffff)

	ips := make([]net.IP, 0)
	for i := start; i <= finish; i++ {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		ips = append(ips, ip)
	}
	return ips, nil
}
