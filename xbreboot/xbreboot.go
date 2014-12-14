package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	DebugBiosPort     = 731
	ResponseOK        = "200- OK\r\n"
	ResponseBanner    = "201- connected\r\n"
	CommandRebootCold = "reboot\r\n"
	CommandRebootWarm = "reboot warm\r\n"
	CommandQuit       = "bye\r\n"
)

var (
	verbose bool
	cold    bool
)

func reboot(host string) {
	socket := fmt.Sprintf("%s:%d", host, DebugBiosPort)

	if verbose {
		log.Print("Connecting to ", socket)
	}

	conn, err := net.Dial("tcp4", socket)

	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	banner, err := reader.ReadString('\n')

	if err != nil {
		log.Fatal("Couldn't read banner: ", err)
	}

	defer conn.Close()

	if banner != ResponseBanner {
		log.Fatal("Got unexpected protocol banner: ", banner)
	}

	var command string

	if cold {
		command = CommandRebootCold
	} else {
		command = CommandRebootWarm
	}

	if verbose {
		log.Print("Connected. Sending command: ", command)
	}

	_, err = writer.WriteString(command)
	writer.Flush()

	if err != nil {
		log.Fatal("Couldn't send reboot command: ", err)
	}

	status, err := reader.ReadString('\n')

	if err != nil {
		log.Fatal("Couldn't read reboot response: ", err)
	} else if status != ResponseOK {
		log.Fatal("Got unexpected reboot response: ", status)
	} else if verbose {
		log.Fatal("Reboot command accepted. Disconnecting.")
	}

	_, err = writer.WriteString(CommandQuit)
	writer.Flush()

	if err != nil {
		log.Fatal("Farewell failed: ", err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-cold] [-v] host\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.BoolVar(&cold, "cold", false, "reload BIOS")
	flag.Usage = usage
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
	}

	reboot(flag.Args()[0])
}
