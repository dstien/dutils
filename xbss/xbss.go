package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"net"
	"os"
	"time"
)

const (
	FilenameFormat    = "xbss-2006-01-02_15-04-05.000.png"
	DebugBiosPort     = 731
	ResponseOK        = "200- OK\r\n"
	ResponseQuit      = "200- bye\r\n"
	ResponseBanner    = "201- connected\r\n"
	ResponseBinary    = "203- binary response follows\r\n"
	CommandScreenshot = "screenshot\r\n"
	CommandQuit       = "bye\r\n"
	HeaderScreenshot  = "pitch=0x%x width=0x%x height=0x%x format=0x%x, framebuffersize=0x%x\r\n"
	FormatBGRA        = 18
)

var (
	filename string
	verbose  bool
)

func bgra2rgba(data []byte, pitch, width, height int) {
	for i := 0; i < pitch * height; i += pitch / width {
		data[i], data[i + 2] = data[i + 2], data[i]
	}
}

func readImage(reader *bufio.Reader, length int) []byte {
	data := make([]byte, length)
	totalread := 0

	if verbose {
		log.Print("Reading image data")
	}

	for totalread < length {
		tmp := make([]byte, length)
		read, err := reader.Read(tmp)
		if read > 0 && err != nil {
			log.Fatal("Reading image data failed: ", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, ".")
		}

		copy(data[totalread:], tmp)

		totalread += read
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "\n")
	}

	return data
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

func screenshot(host string) {
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

	if verbose {
		log.Print("Connected. Sending screenshot command.")
	}

	_, err = writer.WriteString(CommandScreenshot)
	writer.Flush()

	if err != nil {
		log.Fatal("Couldn't send screenshot command: ", err)
	}

	status, err := reader.ReadString('\n')

	if err != nil {
		log.Fatal("Couldn't read screenshot status: ", err)
	} else if status != ResponseBinary {
		log.Fatal("Got unexpected screenshot response: ", banner)
	}

	header, err := reader.ReadString('\n')

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

	data := readImage(reader, fbsize)

	writeImage(data, pitch, width, height)

	_, err = writer.WriteString(CommandQuit)
	writer.Flush()

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
