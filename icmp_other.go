//go:build !linux

package main

import (
	"net"
	"time"
)

func pingICMP(ip net.IP, timeout time.Duration) bool {
	return pingRaw(ip, timeout)
}
