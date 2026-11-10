package main

import (
	"net"
	"regexp"
	"slices"
	"strings"
)

const maxDevices = 200

var (
	hostRe  = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	ifaceRe = regexp.MustCompile(`^[a-zA-Z0-9_.:@-]{1,15}$`)
	userRe  = regexp.MustCompile(`^[a-z0-9_.-]{3,32}$`)
)

type Methods struct {
	Global   bool `json:"global"`
	Directed bool `json:"directed"`
	Unicast  bool `json:"unicast"`
}

type Device struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	MAC        string  `json:"mac"`
	Host       string  `json:"host"`
	Netmask    string  `json:"netmask"`
	Interface  string  `json:"interface"`
	Methods    Methods `json:"methods"`
	CheckPorts []int   `json:"checkPorts"`
	ICMP       bool    `json:"icmp"`
	TimeoutSec int     `json:"timeoutSec"`
	RetrySec   int     `json:"retrySec"`
	LastIP     string  `json:"lastIp,omitempty"`
}

type deviceInput struct {
	Name       string   `json:"name"`
	MAC        string   `json:"mac"`
	Host       string   `json:"host"`
	Netmask    string   `json:"netmask"`
	Interface  string   `json:"interface"`
	Methods    *Methods `json:"methods"`
	CheckPorts []int    `json:"checkPorts"`
	ICMP       *bool    `json:"icmp"`
	TimeoutSec int      `json:"timeoutSec"`
	RetrySec   int      `json:"retrySec"`
}

type fieldError struct {
	Code  string
	Field string
}

func (e *fieldError) Error() string { return e.Code }

func bad(code, field string) error { return &fieldError{Code: code, Field: field} }

func parseMask(s string) net.IPMask {
	ip := net.ParseIP(strings.TrimSpace(s)).To4()
	if ip == nil {
		return nil
	}
	m := net.IPMask(ip)
	if ones, bits := m.Size(); bits == 0 || ones < 1 {
		return nil
	}
	return m
}

func (in deviceInput) device() (Device, error) {
	d := Device{}

	d.Name = strings.TrimSpace(in.Name)
	if n := len([]rune(d.Name)); n < 1 || n > 80 {
		return d, bad("invalid_name", "name")
	}

	mac, err := net.ParseMAC(strings.TrimSpace(in.MAC))
	if err != nil || len(mac) != 6 {
		return d, bad("invalid_mac", "mac")
	}
	d.MAC = mac.String()

	host := strings.TrimSpace(in.Host)
	if host != "" {
		if ip := net.ParseIP(host); ip != nil {
			ip4 := ip.To4()
			if ip4 == nil {
				return d, bad("invalid_host", "host")
			}
			host = ip4.String()
		} else {
			labels := strings.Split(host, ".")
			last := labels[len(labels)-1]
			numeric := last != "" && strings.Trim(last, "0123456789") == ""
			if len(host) > 253 || !hostRe.MatchString(host) || numeric {
				return d, bad("invalid_host", "host")
			}
		}
	}
	d.Host = host

	if nm := strings.TrimSpace(in.Netmask); nm != "" {
		m := parseMask(nm)
		if m == nil {
			return d, bad("invalid_netmask", "netmask")
		}
		d.Netmask = net.IP(m).String()
	}

	d.Interface = strings.TrimSpace(in.Interface)
	if d.Interface != "" && !ifaceRe.MatchString(d.Interface) {
		return d, bad("invalid_interface", "interface")
	}

	d.Methods = Methods{Global: true, Directed: true, Unicast: true}
	if in.Methods != nil {
		d.Methods = *in.Methods
	}
	if !d.Methods.Global && !d.Methods.Directed && !d.Methods.Unicast {
		return d, bad("no_method", "methods")
	}

	if len(in.CheckPorts) > 5 {
		return d, bad("invalid_ports", "checkPorts")
	}
	for _, p := range in.CheckPorts {
		if p < 1 || p > 65535 {
			return d, bad("invalid_ports", "checkPorts")
		}
	}
	d.CheckPorts = slices.Compact(slices.Sorted(slices.Values(in.CheckPorts)))
	if d.CheckPorts == nil {
		d.CheckPorts = []int{}
	}

	d.ICMP = true
	if in.ICMP != nil {
		d.ICMP = *in.ICMP
	}

	d.TimeoutSec = in.TimeoutSec
	if d.TimeoutSec == 0 {
		d.TimeoutSec = 120
	}
	if d.TimeoutSec < 10 || d.TimeoutSec > 600 {
		return d, bad("invalid_timeout", "timeoutSec")
	}
	d.RetrySec = in.RetrySec
	if d.RetrySec == 0 {
		d.RetrySec = 10
	}
	if d.RetrySec < 2 || d.RetrySec > 60 {
		return d, bad("invalid_retry", "retrySec")
	}
	return d, nil
}
