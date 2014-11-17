package main

import (
	"fmt"
	"strings"
)

type Fund int

const (
	DIN Fund = iota
	DBS
)

var fundStrings = []string{
	DIN: "DIN",
	DBS: "DBS",
}

const BaseUrl = "https://www.dovreforvaltning.com/sites/default/files/"

var fundUrlFormats = []string{
	DIN: BaseUrl + "din_a_nok_%s.csv",
	DBS: BaseUrl + "dbs_nok_%s.csv",
}

var fundColHeaders = map[Fund][]string{
	DIN: {"Date", "DIN A (NOK)", "BI (NOK)"},
	DBS: {"Date", "DBS (NOK)",   "BI (NOK)"},
}

func (f Fund) String() string {
	return fundStrings[f]
}

func (f Fund) UrlFormat() string {
	return fundUrlFormats[f]
}

func (f Fund) ColHeaderLen() int {
	return len(fundColHeaders[f])
}

func (f Fund) ColHeader(col int) string {
	return fundColHeaders[f][col]
}

func (f *Fund) Set(value string) error {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case DIN.String():
		*f = DIN
	case DBS.String():
		*f = DBS
	default:
		return fmt.Errorf("Invalid fund code. Got \"%s\", expected one of %s", value, fundList())
	}

	return nil
}

func fundList() string {
	return fmt.Sprintf("\"%s\"", strings.Join(fundStrings, "\", \""))
}
