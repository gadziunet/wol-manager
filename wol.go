package main

import (
	"bytes"
	"net"
	"time"
)

const magicPacketLen = 102

func magicPacket(mac net.HardwareAddr) []byte {
	packet := bytes.Repeat([]byte{0xFF}, 6)
	for i := 0; i < 16; i++ {
		packet = append(packet, mac...)
	}
	return packet
}

func sendPacket(local, remote net.IP, port int, mac net.HardwareAddr) error {
	var laddr *net.UDPAddr
	if local != nil {
		laddr = &net.UDPAddr{IP: local}
	}
	conn, err := net.DialUDP("udp4", laddr, &net.UDPAddr{IP: remote, Port: port})
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Write(magicPacket(mac))
	return err
}

func sendTo(addr string, mac net.HardwareAddr) error {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(magicPacket(mac))
	return err
}
