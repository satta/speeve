package util

import (
	"math/rand"
	"net"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestIPSampler(t *testing.T) {
	val, err := MakeIPSampler("192.168.1.2/24")
	if err != nil {
		t.Fatal(err)
	}
	_, ipNet, _ := net.ParseCIDR("192.168.1.2/24")
	for i := 0; i < 1000; i++ {
		ip := val.GetIP()
		if !ipNet.Contains(ip) {
			t.Fatalf("net %s does not contain IP %s", ipNet, ip)
		}
	}
}

func TestIPSamplerSingleIP(t *testing.T) {
	val, err := MakeIPSampler("192.168.1.43/32")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		ip := val.GetIP()
		if ip.String() != "192.168.1.43" {
			t.Fatalf("IP %s does not contain expected value", ip)
		}
	}
}

func TestIPSamplerBadRange(t *testing.T) {
	val, err := MakeIPSampler("fobar")
	if err == nil {
		t.Fatal("expected error")
	}
	if val != nil {
		t.Fatal("expected nil")
	}
}
