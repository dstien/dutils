package main

import (
	"time"
)

type Day struct {
	Date  time.Time
	Rate  float64
	Index float64
}

type ByDate []Day

func (d ByDate) Len() int {
	return len(d)
}

func (d ByDate) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d ByDate) Less(i, j int) bool {
	return d[i].Date.Before(d[j].Date)
}
