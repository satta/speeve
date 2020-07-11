package util

import (
	"encoding/binary"
	"math/big"
	"math/rand"
	"net"
)

// IPSampler supports only IPv4 so far
type IPSampler struct {
	baseIP   uint32
	hostSize uint32
}

func MakeIPSampler(rng string) (*IPSampler, error) {
	ip, ipnet, err := net.ParseCIDR(rng)
	if err != nil {
		return nil, err
	}
	maskOnes, _ := ipnet.Mask.Size()
	is := &IPSampler{
		hostSize: uint32(maskOnes),
		baseIP:   uint32(0),
	}
	if len(ip) == 16 {
		is.baseIP = binary.BigEndian.Uint32(ip[12:16])
	} else {
		is.baseIP = binary.BigEndian.Uint32(ip)
	}
	return is, nil
}

func ipv6ToInt(IPv6Addr net.IP) *big.Int {
	IPv6Int := big.NewInt(0)
	IPv6Int.SetBytes(IPv6Addr)
	return IPv6Int
}

func intToIpv6(intipv6 *big.Int) net.IP {
	ip := intipv6.Bytes()
	var a net.IP = ip
	return a
}

func toIPv4(val uint32) net.IP {
	return net.IPv4(byte(val>>24), byte(val>>16&0xFF), byte(val>>8)&0xFF, byte(val&0xFF))
}

func (is *IPSampler) GetIP() net.IP {
	randIP := rand.Uint32() >> is.hostSize
	randIP = randIP ^ is.baseIP
	return toIPv4(randIP)
}
