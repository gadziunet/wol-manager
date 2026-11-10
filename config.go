package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr         string
	DataDir      string
	CookieSecure bool
	TrustProxy   bool
	Ports        []int
	StatusEvery  time.Duration
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envBool(key string) (bool, error) {
	v := strings.ToLower(envOr(key, "false"))
	switch v {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	}
	return false, fmt.Errorf("%s: expected true or false, got %q", key, v)
}

func parsePorts(s string) ([]int, error) {
	var out []int
	seen := map[int]bool{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 || n > 65535 {
			return nil, fmt.Errorf("invalid port %q", part)
		}
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	if len(out) == 0 || len(out) > 8 {
		return nil, fmt.Errorf("expected 1 to 8 ports")
	}
	return out, nil
}

func loadConfig() (Config, error) {
	c := Config{
		Addr:    envOr("WOL_ADDR", ":8080"),
		DataDir: envOr("WOL_DATA", "./data"),
	}
	var err error
	if c.CookieSecure, err = envBool("WOL_COOKIE_SECURE"); err != nil {
		return c, err
	}
	if c.TrustProxy, err = envBool("WOL_TRUST_PROXY"); err != nil {
		return c, err
	}
	if c.Ports, err = parsePorts(envOr("WOL_PORTS", "9,7")); err != nil {
		return c, fmt.Errorf("WOL_PORTS: %w", err)
	}
	c.StatusEvery, err = time.ParseDuration(envOr("WOL_STATUS_INTERVAL", "15s"))
	if err != nil {
		return c, fmt.Errorf("WOL_STATUS_INTERVAL: %w", err)
	}
	if c.StatusEvery < 5*time.Second {
		return c, fmt.Errorf("WOL_STATUS_INTERVAL: minimum is 5s")
	}
	return c, nil
}
