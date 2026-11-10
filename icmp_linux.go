//go:build linux

package main

import (
	"errors"
	"net"
	"syscall"
	"time"
)

func pingICMP(ip net.IP, timeout time.Duration) bool {
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_ICMP)
	if err != nil {
		return pingRaw(ip4, timeout)
	}
	defer syscall.Close(fd)

	tv := syscall.NsecToTimeval(int64(timeout))
	if err := syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv); err != nil {
		return false
	}
	if err := syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_SNDTIMEO, &tv); err != nil {
		return false
	}

	var to syscall.SockaddrInet4
	copy(to.Addr[:], ip4)
	if err := syscall.Sendto(fd, echoRequest(1, 1), 0, &to); err != nil {
		return false
	}

	deadline := time.Now().Add(timeout)
	buf := make([]byte, 1500)
	for time.Now().Before(deadline) {
		n, from, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			return false
		}
		sa, ok := from.(*syscall.SockaddrInet4)
		if !ok || !net.IP(sa.Addr[:]).Equal(ip4) {
			continue
		}
		if n >= 1 && buf[0] == 0 {
			return true
		}
	}
	return false
}
