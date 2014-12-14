package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	FilenameFormat    = "xbss-2006-01-02_15-04-05.000.png"
	DebugBiosPort     = 731
	ResponseOk        = "200- OK"
	ResponseQuit      = "200- bye"
	ResponseBanner    = "201- connected"
	ResponseBinary    = "203- binary response follows"
	CommandScreenshot = "screenshot"
	CommandQuit       = "bye"
	HeaderScreenshot  = "pitch=0x%x width=0x%x height=0x%x format=0x%x, framebuffersize=0x%x"
	MessageSuffix     = "\r\n"
	FormatBGRA        = 18
)

var (
	filename string
	verbose  bool
)

func bgra2rgba(data []byte, pitch, width, height int) {
	for i := 0; i < pitch*height; i += pitch / width {
		data[i], data[i+2] = data[i+2], data[i]
	}
}

func writeImage(data []byte, pitch, width, height int) {
	bgra2rgba(data, pitch, width, height)
	img := &image.RGBA{data, pitch, image.Rect(0, 0, width, height)}

	if filename == "" {
		filename = time.Now().Format(FilenameFormat)
	}

	file, err := os.Create(filename)

	if err != nil {
		log.Fatal("Couldn't create output file: ", err)
	}

	defer file.Close()

	filewriter := bufio.NewWriter(file)

	err = png.Encode(filewriter, img)

	if err != nil {
		log.Fatal("PNG encoding failed: ", err)
	}

	fmt.Println(filename)
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

func screenshot(host string) {
	conn, reader, writer, err := connect(host)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	_, err = sendCommand(writer, reader, CommandScreenshot, ResponseBinary)
	if err != nil {
		log.Fatalf("Command \"%s\" failed: %s", CommandScreenshot, err)
	}

	header, err := readResponse(reader, "")
	if err != nil {
		log.Fatal("Couldn't read screenshot header: ", err)
	}

	var pitch, width, height, format, fbsize int
	fmt.Sscanf(header, HeaderScreenshot, &pitch, &width, &height, &format, &fbsize)

	if verbose {
		fmt.Fprintf(os.Stderr, "\n%-16s: %7d\n%-16s: %7d\n%-16s: %7d\n%-16s: %7d\n%-16s: %7d\n\n", "Pitch", pitch, "Width", width, "Height", height, "Format", format, "Framebuffer size", fbsize)
	}

	if pitch == 0 || width == 0 || height == 0 || format != FormatBGRA || fbsize == 0 {
		log.Fatal("Invalid image format")
	}

	buf := bytes.NewBuffer(nil)
	buf.Grow(fbsize)

	_, err = io.CopyN(buf, reader, int64(fbsize))
	if err != nil {
		log.Fatal("Reading image data failed: ", err)
	}

	writeImage(buf.Bytes(), pitch, width, height)

	_, err = sendCommand(writer, reader, CommandQuit, ResponseQuit)
	if err != nil {
		log.Fatal("Farewell failed: ", err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-f filename.png] [-v] host\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.StringVar(&filename, "f", "", "output filename")
	flag.Usage = usage
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
	}

	screenshot(flag.Args()[0])
}
