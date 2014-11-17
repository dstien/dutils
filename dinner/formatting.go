package main

import (
	"fmt"
	"strings"
)

type Formatting int

const (
	None Formatting = iota
	Terminal
	IRC
)

var formattingStrings = []string{
	None:     "None",
	Terminal: "Terminal",
	IRC:      "IRC",
}

type Code int

const (
	Reset Code = iota
	Red
	Green
)

var formattingCodes = map[Formatting][]string{
	None:     {Reset: "",       Red: "",           Green: ""          },
	Terminal: {Reset: "\x1b[m", Red: "\x1b[31;1m", Green: "\x1b[32;1m"},
	IRC:      {Reset: "\x0f",   Red: "\x02\x035",  Green: "\x02\x033" },
}

func (f Formatting) Code(code Code) string {
	return formattingCodes[f][code]
}

func (f Formatting) String() string {
	return formattingStrings[f]
}

func (f *Formatting) Set(value string) error {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case strings.ToUpper(None.String()):
		*f = None
	case strings.ToUpper(Terminal.String()):
		*f = Terminal
	case strings.ToUpper(IRC.String()):
		*f = IRC
	default:
		return fmt.Errorf("Invalid formatting type. Got \"%s\", expected one of %s", value, formattingList())
	}

	return nil
}

func formattingList() string {
	return fmt.Sprintf("\"%s\"", strings.Join(formattingStrings, "\", \""))
}
