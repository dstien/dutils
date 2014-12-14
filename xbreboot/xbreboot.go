package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	DebugBiosPort  = 731
	ResponseOk     = "200- OK"
	ResponseQuit   = "200- bye"
	ResponseBanner = "201- connected"
	CommandReboot  = "reboot"
	ArgumentWarm   = " warm"
	CommandQuit    = "bye"
	MessageSuffix  = "\r\n"
)

var (
	verbose bool
	cold    bool
)

func connect(host string) (conn net.Conn, reader *bufio.Reader, writer *bufio.Writer, err error) {
	socket := fmt.Sprintf("%s:%d", host, DebugBiosPort)

	if verbose {
		log.Printf("Connecting to %s", socket)
	}

	conn, err = net.Dial("tcp4", socket)
	if err != nil {
		return nil, nil, nil, err
	}

	reader = bufio.NewReader(conn)
	writer = bufio.NewWriter(conn)

	_, err = readResponse(reader, ResponseBanner)
	if err != nil {
		defer conn.Close()
		return nil, nil, nil, fmt.Errorf("Error reading protocol banner: %s", err)
	}

	return conn, reader, writer, err
}

func readResponse(reader *bufio.Reader, expected string) (response string, err error) {
	response, err = reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if strings.HasSuffix(response, MessageSuffix) {
		response = response[:len(response)-len(MessageSuffix)]
	}

	if verbose {
		log.Printf("Received response \"%s\"", response)
	}

	if expected != "" && response != expected {
		err = fmt.Errorf("Got \"%s\", expected \"%s\".", response, expected)
	}

	return response, err
}

func sendCommand(writer *bufio.Writer, reader *bufio.Reader, command, expected string) (response string, err error) {
	if verbose {
		log.Printf("Sending command \"%s\"", command)
	}

	_, err = writer.WriteString(command + MessageSuffix)
	if err != nil {
		return "", err
	}

	err = writer.Flush()
	if err != nil {
		return "", err
	}

	return readResponse(reader, expected)
}

func reboot(host string) {
	conn, reader, writer, err := connect(host)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	command := CommandReboot

	if !cold {
		command += ArgumentWarm
	}

	_, err = sendCommand(writer, reader, command, ResponseOk)
	if err != nil {
		log.Fatalf("Command \"%s\" failed: %s", command, err)
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
