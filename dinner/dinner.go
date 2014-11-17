package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

const (
	DateLayout = "2006-01-02"
	UserAgent  = "dinner/1.0"
	FilePerm   = 0660
)

var (
	fund        Fund
	formatting  Formatting
	currentFile string
	verbose     bool
)

func formatPct(cur, prev float64) string {
	var color string

	pct := (cur / prev - 1) * 100

	if pct < 0 {
		color = formatting.Code(Red)
	} else {
		color = formatting.Code(Green)
	}

	return fmt.Sprintf("%s%+.2f%%%s", color, pct, formatting.Code(Reset))
}

func prevWeekday(date time.Time) time.Time {
	if date.Weekday() == time.Monday {
		date = date.AddDate(0, 0, -3) // Friday's update on monday
	} else if date.Weekday() == time.Sunday {
		date = date.AddDate(0, 0, -3) // Last update before weekend is thursday
	} else if date.Weekday() == time.Saturday {
		date = date.AddDate(0, 0, -2)
	} else {
		date = date.AddDate(0, 0, -1) // Regular tue-fri update for last day's figures
	}

	return date
}

func updateProcessedDate(newdate time.Time) bool {
	if currentFile == "" {
		return true
	} else if isAlreadyProcessed(newdate) {
		return false
	}

	file, err := os.OpenFile(currentFile, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, FilePerm)

	if err != nil {
		log.Fatal("Error opening file with current date for writing: ", err)
	}

	defer file.Close()

	data := []byte(newdate.Format(DateLayout))
	_, err = file.Write(data)

	if err != nil {
		log.Fatal("Error writing file with current date: ", err)
	}

	return true
}

func isAlreadyProcessed(newDate time.Time) bool {
	if currentFile == "" {
		return false
	}

	file, err := os.Open(currentFile)

	if err != nil {
		log.Print("Error opening file with current date for reading: ", err)
		return false
	}

	defer file.Close()

	data := make([]byte, len(DateLayout))
	_, err = file.Read(data)

	if err != nil {
		log.Print("Error reading file with current date: ", err)
		return false
	}

	prevDate, err := time.Parse(DateLayout, string(data))

	if err != nil {
		log.Print("Error parsing file with current date: ", err)
		return false
	}

	return !prevDate.Before(newDate)
}

func downloadData(url string) *http.Response {
	if verbose {
		log.Print("Fetching ", url)
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	return resp
}

func checkColumnParsing(line int, column int, expected string, cols []string, err error) {
	if err != nil {
		log.Fatalf("Error parsing column %d on line %d: Expected %s, got \"%s\" (%s)", column + 1, line, expected, cols[column], err)
	}
}

func parseData(file io.Reader) []Day {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = fund.ColHeaderLen()

	hdr, err := reader.Read()

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < reader.FieldsPerRecord; i++ {
		if hdr[i] != fund.ColHeader(i) {
			log.Fatal("Unexpected column headers. Got \"%s\", expected \"%s\"", hdr[i], fund.ColHeader(i))
		}
	}

	days := make([]Day, 0, (time.Now().Year() - 2012 + 1) * 365)

	for line := 2; ; line++ {
		cols, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		date, err := time.Parse(DateLayout, cols[0])

		checkColumnParsing(line, 0, "date", cols, err)

		rate, err := strconv.ParseFloat(cols[1], 64)

		checkColumnParsing(line, 1, "decimal", cols, err)

		index, err := strconv.ParseFloat(cols[2], 64)

		// Skip missing index values.
		if cols[2] != "" && err != nil {
			checkColumnParsing(line, 2, "decimal", cols, err)
		}

		days = append(days, Day{Date: date, Rate: rate, Index: index})
	}

	if len(days) < 100 {
		log.Fatalf("Insufficient data (got %d rows)", len(days))
	}

	return days
}

func generateMessage(days []Day) string {
	// Sort by descending dates.
	sort.Sort(sort.Reverse(ByDate(days)))

	lastDay := &days[0]
	prevDay := &days[1]

	if !updateProcessedDate(lastDay.Date) {
		if verbose {
			log.Print("OK: Current date already processed")
		}
		return ""
	}

	var weekDay, mnt1Day, mnt3Day, mnt6Day, yearDay, athDay *Day

	weekDiff := lastDay.Date.AddDate(0,  0, -7)
	mnt1Diff := lastDay.Date.AddDate(0, -1,  0)
	mnt3Diff := lastDay.Date.AddDate(0, -3,  0)
	mnt6Diff := lastDay.Date.AddDate(0, -6,  0)
	yearDiff := time.Date(lastDay.Date.Year(), 1, 1, 0, 0, 0, 0, time.UTC)

	var ath float64 = 0

	for i := 0; i < len(days); i++ {
		if weekDay == nil && (days[i].Date.Equal(weekDiff) || days[i].Date.Before(weekDiff)) {
			weekDay = &days[i]
		} else if mnt1Day == nil && (days[i].Date.Equal(mnt1Diff) || days[i].Date.Before(mnt1Diff)) {
			mnt1Day = &days[i]
		} else if mnt3Day == nil && (days[i].Date.Equal(mnt3Diff) || days[i].Date.Before(mnt3Diff)) {
			mnt3Day = &days[i]
		} else if mnt6Day == nil && (days[i].Date.Equal(mnt6Diff) || days[i].Date.Before(mnt6Diff)) {
			mnt6Day = &days[i]
		}

		if yearDay == nil && days[i].Date.Before(yearDiff) {
			yearDay = &days[i]
		}

		if days[i].Rate > ath {
			ath = days[i].Rate
			athDay = &days[i]
		}
	}

	if lastDay == nil || prevDay == nil || weekDay == nil || mnt1Day == nil || mnt3Day == nil || mnt6Day == nil || yearDay == nil || athDay == nil {
		log.Fatal("Insufficient historic data")
	}

	return fmt.Sprintf("%s %s  1D: %s  1W: %s  1M: %s  3M: %s  6M: %s  YTD: %s  ATH: %s\n",
		fund,
		lastDay.Date.Format(DateLayout),
		formatPct(lastDay.Rate, prevDay.Rate),
		formatPct(lastDay.Rate, weekDay.Rate),
		formatPct(lastDay.Rate, mnt1Day.Rate),
		formatPct(lastDay.Rate, mnt3Day.Rate),
		formatPct(lastDay.Rate, mnt6Day.Rate),
		formatPct(lastDay.Rate, yearDay.Rate),
		formatPct(lastDay.Rate, athDay.Rate))
}

func serveDinner() {
	today := time.Now().Truncate(time.Hour * 24)

	if isAlreadyProcessed(prevWeekday(today)) {
		if verbose {
			log.Print("OK: Previous weekday already processed")
		}
		return
	}

	url := fmt.Sprintf(fund.UrlFormat(), today.Format("01-02"))

	resp := downloadData(url)

	defer resp.Body.Close()

	days := parseData(resp.Body)

	msg := generateMessage(days)

	if len(msg) == 0 {
		if verbose {
			log.Print("OK: No message generated")
		}

		return
	}

	fmt.Print(msg)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-f FUND] [-t TYPE] [-c CurrentFile] [-v]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Var(&fund,               "f", fmt.Sprintf("fund code, accepted values: %s", fundList()))
	flag.Var(&formatting,         "t", fmt.Sprintf("message formatting type, accepted values: %s", formattingList()))
	flag.StringVar(&currentFile,  "c", "",    "file for caching last processed data")
	flag.BoolVar(&verbose,        "v", false, "verbose")
	flag.Usage = usage
}

func main() {
	flag.Parse()

	if flag.NArg() > 0 {
		usage()
	}

	if verbose {
		if currentFile == "" {
			log.Print("Ignoring current date checks")
		} else {
			log.Print("Current: ", currentFile)
		}

		log.Print("Formatting: ", formatting)
		log.Print("Fund: ", fund)
	}

	serveDinner()
}
