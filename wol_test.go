package main

import (
	"bytes"
	"net"
	"testing"
)

func TestMagicPacket(t *testing.T) {
	mac, _ := net.ParseMAC("AA:BB:CC:DD:EE:FF")
	p := magicPacket(mac)

	if len(p) != magicPacketLen {
		t.Fatalf("packet length = %d, want %d", len(p), magicPacketLen)
	}
	if !bytes.Equal(p[:6], bytes.Repeat([]byte{0xFF}, 6)) {
		t.Errorf("bad header: % x", p[:6])
	}
	for i := 0; i < 16; i++ {
		if got := p[6+i*6 : 12+i*6]; !bytes.Equal(got, mac) {
			t.Errorf("repetition %d = % x, want % x", i, got, []byte(mac))
		}
	}
}

func TestSendPacketUnicastLoopback(t *testing.T) {
	pc, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("no UDP socket: %v", err)
	}
	defer pc.Close()

	mac, _ := net.ParseMAC("AA:BB:CC:DD:EE:FF")
	port := pc.LocalAddr().(*net.UDPAddr).Port
	if err := sendPacket(nil, net.IPv4(127, 0, 0, 1), port, mac); err != nil {
		t.Fatalf("sendPacket: %v", err)
	}

	buf := make([]byte, 256)
	n, _, err := pc.ReadFrom(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(buf[:n], magicPacket(mac)) {
		t.Errorf("received packet differs from the expected one")
	}
}

func TestPingLoopback(t *testing.T) {
	if !pingICMP(net.IPv4(127, 0, 0, 1), 1500*1000*1000) {
		t.Skip("ICMP ping is not permitted in this environment")
	}
}
