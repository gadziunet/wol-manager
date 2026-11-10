package main

import (
	"net"
	"time"
)

func icmpChecksum(b []byte) uint16 {
	var sum uint32
	for i := 0; i+1 < len(b); i += 2 {
		sum += uint32(b[i])<<8 | uint32(b[i+1])
	}
	if len(b)%2 == 1 {
		sum += uint32(b[len(b)-1]) << 8
	}
	for sum>>16 != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}

func echoRequest(id, seq uint16) []byte {
	b := []byte{8, 0, 0, 0, byte(id >> 8), byte(id), byte(seq >> 8), byte(seq), 'w', 'o', 'l', '!'}
	c := icmpChecksum(b)
	b[2], b[3] = byte(c>>8), byte(c)
	return b
}

func pingRaw(ip net.IP, timeout time.Duration) bool {
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return false
	}
	defer conn.Close()
	const id = 0x574c
	if _, err := conn.WriteTo(echoRequest(id, 1), &net.IPAddr{IP: ip4}); err != nil {
		return false
	}
	deadline := time.Now().Add(timeout)
	_ = conn.SetReadDeadline(deadline)
	buf := make([]byte, 1500)
	for time.Now().Before(deadline) {
		n, from, err := conn.ReadFrom(buf)
		if err != nil {
			return false
		}
		fromIP, ok := from.(*net.IPAddr)
		if !ok || !fromIP.IP.Equal(ip4) {
			continue
		}
		if n >= 6 && buf[0] == 0 && uint16(buf[4])<<8|uint16(buf[5]) == id {
			return true
		}
	}
	return false
}
