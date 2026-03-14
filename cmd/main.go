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
	"github.com/fibrchat/worker/pkg/event"
	"github.com/fibrchat/worker/pkg/message"
	"github.com/fibrchat/worker/pkg/request"
)

type handler struct{}

func (h handler) OnMessage(msg message.Message) {
	fmt.Printf("\r\033[K[%s] %s: %s\n> ", msg.Timestamp.Format("15:04:05"), msg.Src, msg.Content)
}

func (h handler) OnConnect(evt event.Event) {
	fmt.Printf("\r\033[K* %s connected\n> ", evt.User)
}

func (h handler) OnDisconnect(evt event.Event) {
	fmt.Printf("\r\033[K* %s disconnected\n> ", evt.User)
}

// Temp CLI client
func main() {
	server := flag.String("server", "ws://localhost:4222", "NATS WebSocket server URL")
	username := flag.String("user", "bob", "username (required)")
	password := flag.String("pass", "password", "password")
	domain := flag.String("domain", "localhost", "server domain")
	flag.Parse()

	c, err := client.New(client.Options{
		ServerURL: *server,
		Username:  *username,
		Password:  *password,
		Domain:    *domain,
		Handler:   handler{},
	})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	fmt.Printf("Connected as %s\n", c.Address())
	fmt.Println("Usage: /msg <user@domain> <message>")
	fmt.Println("       /list - list online users")

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

		if line == "/list" {
			users, err := c.ListOnline()
			if err != nil {
				fmt.Printf("  Error: %v\n", err)
				continue
			}
			if len(users) == 0 {
				fmt.Println("  No users online")
			} else {
				fmt.Println("  Online:")
				for _, u := range users {
					fmt.Printf("    %s\n", u)
				}
			}
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
		dst, err := address.Parse(to)
		if err != nil {
			fmt.Printf("  Invalid address: %v\n", err)
			continue
		}

		resp, err := c.SendMessage(dst, body)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		if resp.Code != request.CodeSuccess {
			fmt.Printf("  Server error: %s\n", resp.Error)
		}
	}
}
