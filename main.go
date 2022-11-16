package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

func main() {
	cfg, printHelp, err := parseConfig(os.Args[1:])
	if err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}

	if printHelp {
		fmt.Print("tcpfw - forwards tcp connections from one port to another\n\n")
		fmt.Print("\t--host    host to listen to, defaults to localhost\n")
		fmt.Print("\t--inport  port to listen to\n")
		fmt.Print("\t--outport port to forward the connection to\n")
		os.Exit(0)
	}

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Host, cfg.InPort))
	if err != nil {
		log.Fatalf("%s: failed to create listener", err)
	}

	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("%s: failed to accept connection", err)
		}
		go handleRequest(cfg, conn)
	}
}

type Config struct {
	Host    string
	InPort  string
	OutPort string
}

var ErrUnknownArgument = errors.New("unknown argument")
var ErrInvalidConfig = errors.New("invalid config")

func parseConfig(args []string) (Config, bool, error) {
	cfg := Config{
		Host:    "localhost",
		InPort:  "",
		OutPort: "",
	}

	printHelp := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--host":
			if i < len(args) {
				i++
				cfg.Host = args[i]
			}
		case "--inport":
			if i < len(args) {
				i++
				cfg.InPort = args[i]
			}
		case "--outport":
			if i < len(args) {
				i++
				cfg.OutPort = args[i]
			}
		case "--help":
			return Config{}, true, nil
		default:
			return Config{}, false, fmt.Errorf("%w %s", ErrUnknownArgument, arg)
		}
	}

	if cfg.Host == "" {
		return Config{}, false, fmt.Errorf("%w: host must not be empty", ErrInvalidConfig)
	}

	if cfg.InPort == "" {
		return Config{}, false, fmt.Errorf("%w: inport must not be empty", ErrInvalidConfig)
	}

	if cfg.OutPort == "" {
		return Config{}, false, fmt.Errorf("%w: outport must not be empty", ErrInvalidConfig)
	}

	return cfg, printHelp, nil
}

func handleRequest(cfg Config, connIn net.Conn) {
	defer connIn.Close()

	tcpTarget, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", cfg.Host, cfg.OutPort))
	if err != nil {
		log.Printf("%s: resolve failed", err)
		return
	}

	connOut, err := net.DialTCP("tcp", nil, tcpTarget)
	if err != nil {
		log.Printf("%s: dial failed", err)
		return
	}
	defer connOut.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// connIn -> connOut
	go func() {
		defer wg.Done()
		if _, err := io.Copy(connIn, connOut); err != nil {
			log.Printf("error while copying to target tcp: %s", err)
		}
	}()

	// connOut -> connIn
	go func() {
		defer wg.Done()
		if _, err := io.Copy(connOut, connIn); err != nil {
			log.Printf("error while copying to receiver tcp: %s", err)
		}
	}()

	wg.Wait()
}
