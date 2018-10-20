package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
)

var verbose bool
var debug bool

func main() {
	verbose = *flag.Bool("v", false, "Verbose mode.")
	debug = *flag.Bool("d", false, "Debug mode.")
	flag.Parse()
	if debug {
		verbose = true
	}

	basePath, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("failed to get absolute path: %s", err)
	}

	err = do(basePath)
	if err != nil {
		log.Fatal(err)
	}
}

func do(basePath string) {
	fmt.Println(basePath)
}
