package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "serve":
		case "healthcheck":
			os.Exit(runHealthcheck())
		case "send":
			os.Exit(runSend(os.Args[2:]))
		case "version":
			fmt.Println(version)
			return
		default:
			fmt.Fprintf(os.Stderr, "Usage: wol [serve | send [-addr host:port] MAC | healthcheck | version]\n")
			os.Exit(2)
		}
	}
	os.Exit(runServer())
}

func runSend(args []string) int {
	fs := flag.NewFlagSet("send", flag.ContinueOnError)
	addr := fs.String("addr", "255.255.255.255:9", "destination host:port")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: wol send [-addr host:port] AA:BB:CC:DD:EE:FF\n")
		return 2
	}
	mac, err := net.ParseMAC(fs.Arg(0))
	if err != nil || len(mac) != 6 {
		fmt.Fprintf(os.Stderr, "Invalid MAC address: %q\n", fs.Arg(0))
		return 2
	}
	if err := sendTo(*addr, mac); err != nil {
		fmt.Fprintf(os.Stderr, "Sending failed: %v\n", err)
		return 1
	}
	fmt.Printf("Sent magic packet to %s for %s\n", *addr, mac)
	return 0
}

func runHealthcheck() int {
	host, port, err := net.SplitHostPort(envOr("WOL_ADDR", ":8080"))
	if err != nil {
		return 1
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://" + net.JoinHostPort(host, port) + "/healthz")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func runServer() int {
	log.SetFlags(log.LstdFlags | log.LUTC)
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("configuration error: %v", err)
		return 2
	}
	st, err := openStore(cfg.DataDir)
	if err != nil {
		log.Printf("opening data directory: %v", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tracker := newTracker()
	prober := &Prober{store: st}
	wakes := newManager(ctx, cfg.Ports, prober, tracker)
	mon := &Monitor{store: st, prober: prober, tracker: tracker, every: cfg.StatusEvery}
	srv := newServer(cfg, st, wakes, tracker)

	go mon.Run(ctx)
	go srv.janitor(ctx)

	httpSrv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}()

	log.Printf("wol %s listening on %s (data: %s, wake ports: %v)", version, cfg.Addr, cfg.DataDir, cfg.Ports)
	if !pingICMP(net.IPv4(127, 0, 0, 1), 500*time.Millisecond) {
		log.Printf("note: ICMP ping is not available here, online status uses TCP ports only (see README)")
	}
	if tok := srv.SetupToken(); tok != "" {
		log.Printf("FIRST RUN: no account exists yet. Open the web page and enter this setup token: %s", tok)
	}

	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("server: %v", err)
		return 1
	}
	wakes.Wait()
	return 0
}
