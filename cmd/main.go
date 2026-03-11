package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fibrchat/client/pkg/client"
	"github.com/fibrchat/worker/pkg/address"
	"github.com/fibrchat/worker/pkg/message"
)

// Temp CLI client
func main() {
	server := flag.String("server", "ws://localhost:4222", "NATS WebSocket server URL")
	username := flag.String("user", "", "username (required)")
	password := flag.String("pass", "password", "password")
	domain := flag.String("domain", "server-a", "server domain")
	flag.Parse()

	if *username == "" {
		fmt.Fprintln(os.Stderr, "error: -user is required")
		flag.Usage()
		os.Exit(1)
	}

	c, err := client.New(client.Options{
		ServerURL: *server,
		Username:  *username,
		Password:  *password,
		Domain:    *domain,
		OnMessage: func(msg message.Message) {
			fmt.Printf("\r\033[K[%s] %s: %s\n> ", msg.Timestamp.Format("15:04:05"), msg.From, msg.Body)
		},
	})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	fmt.Printf("Connected as %s\n", c.Address())
	fmt.Println("Usage: /msg <user@domain> <message>")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nBye!")
		c.Close()
		os.Exit(0)
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		args := strings.TrimPrefix(line, "/msg ")
		if args == line {
			continue
		}

		parts := strings.SplitN(args, " ", 2)
		if len(parts) < 2 {
			continue
		}

		to, body := parts[0], parts[1]
		if _, _, err := address.Split(to); err != nil {
			fmt.Printf("  Invalid address: %v\n", err)
			continue
		}

		resp, err := c.SendMessage(to, body)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		if resp.Code != message.CodeSuccess {
			fmt.Printf("  Server error: %s\n", resp.Error)
		}
	}
}
