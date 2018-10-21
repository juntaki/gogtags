package main

import "fmt"

type tagType int

const (
	gtags tagType = iota
	grtags
	gpath
)

func (t tagType) String() string {
	switch t {
	case gtags:
		return "GTAGS"
	case grtags:
		return "GRTAGS"
	case gpath:
		return "GPATH"
	}
	panic("invalid tagType")
}

type standard struct {
	tagName    string
	fileID     int
	lineNumber int
	lineImage  string
}

func (s standard) String() string {
	return fmt.Sprintf("%d @n %d %s", s.fileID, s.lineNumber, s.lineImage)
}

type compact struct {
	fileID      int
	lineNumbers []int
}

func (c compact) String() string {
	continueCounter := 0
	output := fmt.Sprintf("%d", c.lineNumbers[0]) // [0] must be exist
	for l := 1; l < len(c.lineNumbers); l++ {
		diff := c.lineNumbers[l] - c.lineNumbers[l-1]
		if continueCounter == 0 {
			if diff == 1 {
				output += "-"
				continueCounter++
			} else {
				output += fmt.Sprintf(",%d", diff)
			}
		} else {
			if diff == 1 {
				continueCounter++
			} else {
				output += fmt.Sprintf("%d,%d", continueCounter, diff)
				continueCounter = 0
			}
		}
	}
	if continueCounter != 0 {
		output += fmt.Sprintf("%d", continueCounter)
	}
	return fmt.Sprintf("%d @n %s", c.fileID, output)
}
