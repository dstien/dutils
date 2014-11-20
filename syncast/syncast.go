package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"time"
)

type RSSTime struct {
	time.Time
}

func (t *RSSTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw string
	d.DecodeElement(&raw, &start)
	parsed, err := time.Parse(time.RFC1123, raw)
	*t = RSSTime{parsed}
	return err
}

type EnclosureURL struct {
	Full     string
	Filename string
}

func (u *EnclosureURL) UnmarshalXMLAttr(attr xml.Attr) error {
	u.Full = attr.Value

	url, err := url.Parse(attr.Value)

	if err == nil {
		u.Filename = path.Base(url.Path)
	}

	return err
}

type Enclosure struct {
	URL       EnclosureURL `xml:"url,attr"`
	Length    uint64       `xml:"length,attr"`
	Type      string       `xml:"type,attr"`
}

type Link struct {
	URL  string `xml:"href,attr"`
	Type string `xml:"type,attr"`
}

type ItunesDuration struct {
	time.Duration
}

func (t *ItunesDuration) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw string
	d.DecodeElement(&raw, &start)
	parsed, err := time.Parse("15:04:05", raw)

	if err == nil {
		parsed = parsed.AddDate(1, 0, 0)
		*t = ItunesDuration{parsed.Sub(time.Time{})}
	}

	return err
}

type Item struct {
	Title       string         `xml:"title"`
	Description string         `xml:"description"`
	Enclosure   Enclosure      `xml:"enclosure"`
	PubDate     RSSTime        `xml:"pubDate"`
	GUID        string         `xml:"guid"`
	Summary     string         `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd summary"`
	Subtitle    string         `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd subtitle"`
	Author      string         `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd author"`
	Keywords    string         `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd keywords"`
	Duration    ItunesDuration `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd duration"`
}

type RSSChannel struct {
	Title       string `xml:"channel>title"`
	Description string `xml:"channel>description"`
	Link        Link   `xml:"http://www.w3.org/2005/Atom channel>link"`
	Items       []Item `xml:"channel>item"`
}

func parseFile(filename string) *RSSChannel {
	fmt.Fprintf(os.Stderr, "Reading \"%s\"\n", os.Args[1])

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	rss := RSSChannel{}

	err = xml.Unmarshal([]byte(data), &rss)
	if err != nil {
		log.Fatal(err)
	}

	return &rss
}

func printContent(rss *RSSChannel) {
	fmt.Printf("Feed\n")
	fmt.Printf("  Title:       \"%s\"\n", rss.Title)
	fmt.Printf("  Description: \"%s\"\n", rss.Description)
	fmt.Printf("  URL:         \"%s\"\n", rss.Link.URL)
	fmt.Println()

	for i, item := range rss.Items {
		fmt.Printf("Item %d\n", i)
		fmt.Printf("  Title:       \"%s\"\n", item.Title)
		fmt.Printf("  Description: \"%s\"\n", item.Description)
		fmt.Printf("  Summary:     \"%s\"\n", item.Summary)
		fmt.Printf("  Subtitle:    \"%s\"\n", item.Subtitle)
		fmt.Printf("  URL:         \"%s\"\n", item.Enclosure.URL.Full)
		fmt.Printf("  Filename:    \"%s\"\n", item.Enclosure.URL.Filename)
		fmt.Printf("  Length:      %d bytes\n", item.Enclosure.Length)
		fmt.Printf("  Type:        \"%s\"\n", item.Enclosure.Type)
		fmt.Printf("  Duration:    %s\n", item.Duration)
		fmt.Printf("  PubDate:     %s\n", item.PubDate)
		fmt.Printf("  GUID:        \"%s\"\n", item.GUID)
		fmt.Printf("  Author:      \"%s\"\n", item.Author)
		fmt.Printf("  Keywords:    \"%s\"\n", item.Keywords)
		fmt.Println()
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s feed.xml\n", os.Args[0])
		os.Exit(2)
	}

	rss := parseFile(os.Args[1])
	printContent(rss)
}
