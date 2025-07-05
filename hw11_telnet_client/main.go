package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: go-telnet [--timeout=10s] host port")
		os.Exit(2)
	}
	host, port := flag.Arg(0), flag.Arg(1)
	address := net.JoinHostPort(host, port)

	client := NewTelnetClient(address, *timeout, os.Stdin, os.Stdout)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "...Connection error: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()
	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	// Обработка SIGINT (Ctrl+C)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})
	var once sync.Once

	go func() {
		_ = client.Receive()
		once.Do(func() { close(done) })
	}()
	go func() {
		_ = client.Send()
		once.Do(func() { close(done) })
	}()

	select {
	case <-done:
		fmt.Fprintln(os.Stderr, "...Connection was closed by peer or EOF")
	case <-sigCh:
		fmt.Fprintln(os.Stderr, "...Interrupted")
	}
}
