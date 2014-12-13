package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	DebugBiosPort      = 731
	ResponseOk         = "200- OK"
	ResponseQuit       = "200- bye"
	ResponseBanner     = "201- connected"
	ResponseSendBinary = "204- send binary data"
	CommandSendFile    = "sendfile name=\"%s\" length=0x%x"
	CommandQuit        = "bye"
	MessageSuffix      = "\r\n"
	XboxPathSeparator  = '\\'
)

var (
	verbose bool
)

func openLocal(name string) (file *os.File, length int64, err error) {
	if verbose {
		log.Printf("Opening local file \"%s\"", name)
	}

	stat, err := os.Stat(name)
	if err != nil {
		return nil, 0, err
	} else if !stat.Mode().IsRegular() {
		return nil, 0, fmt.Errorf("Not a regular file: \"%s\"", name)
	}

	length = stat.Size()

	file, err = os.Open(name)

	return file, length, err
}

func parseRemote(name, localfile string) (host, path string, err error) {
	dest := strings.SplitN(name, ":", 2)

	if len(dest) != 2 || dest[0] == "" || dest[1] == "" {
		return "", "", fmt.Errorf("Destination filename must be on the format \"host:X:\\full\\path\\file\"")
	}

	host = dest[0]
	path = dest[1]

	// Append local filename if remote path is a directory.
	if strings.HasSuffix(path, string(XboxPathSeparator)) {
		path += localfile
	}

	if verbose {
		log.Printf("Destination host: \"%s\"", host)
		log.Printf("Destination file: \"%s\"", path)
	}

	return host, path, nil
}

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

func copyFile(sourcefilename, destfilename string) {
	sourcefile, sourcelength, err := openLocal(sourcefilename)
	if err != nil {
		log.Fatal(err)
	}
	defer sourcefile.Close()

	desthost, destpath, err := parseRemote(destfilename, filepath.Base(sourcefilename))
	if err != nil {
		log.Fatal(err)
	}

	destconn, destreader, destwriter, err := connect(desthost)
	if err != nil {
		log.Fatal(err)
	}

	defer destconn.Close()

	command := fmt.Sprintf(CommandSendFile, destpath, sourcelength)
	_, err = sendCommand(destwriter, destreader, command, ResponseSendBinary)
	if err != nil {
		log.Fatalf("Command \"%s\" failed: %s", command, err)
	}

	if verbose {
		log.Printf("Sending %d bytes of binary data", sourcelength)
	} else {
		fmt.Printf("Copying \"%s\" (%d bytes) to %s:\"%s\"... ", sourcefilename, sourcelength, desthost, destpath)
	}

	_, err = io.Copy(destwriter, sourcefile)
	if err != nil {
		log.Fatal("Copying file data failed: ", err)
	}

	_, err = readResponse(destreader, ResponseOk)
	if err != nil {
		log.Fatal("Copying file data failed: ", err)
	} else if !verbose {
		fmt.Println("Success")
	}

	_, err = sendCommand(destwriter, destreader, CommandQuit, ResponseQuit)
	if err != nil {
		log.Fatal("Farewell failed: ", err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-v] [sourcefile] [host:destfile]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Usage = usage
}

func main() {
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
	}

	copyFile(flag.Args()[0], flag.Args()[1])
}
