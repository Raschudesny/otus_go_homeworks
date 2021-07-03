package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

var timeout time.Duration

func init() {
	pflag.DurationVarP(&timeout, "timeout", "t", time.Second*10, "client timeout")
	pflag.Parse()
}

func main() {
	args := pflag.Args()

	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "Host or port arguments must be provided.")
		fmt.Fprintln(os.Stderr, "Usage: go-telnet [--timeout] host port.")
		os.Exit(1)
	}

	host := args[0]
	port := args[1]

	address := net.JoinHostPort(host, port)

	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to establish connection : %s", err)
	}

	defer func() {
		if err := client.Close(); err != nil {
			log.Println(err)
		}
	}()

	notifyContext, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()

	ctx, cancelFunc := context.WithCancel(notifyContext)

	go func(telnetClient TelnetClient) {
		if err := telnetClient.Send(); err != nil {
			log.Println(err)
		}
		cancelFunc()
	}(client)

	go func(telnetClient TelnetClient) {
		if err := telnetClient.Receive(); err != nil {
			log.Println(err)
		}
		cancelFunc()
	}(client)
	<-ctx.Done()
}
