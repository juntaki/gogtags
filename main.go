package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var verbose bool
var debug bool

func main() {
	verbosep := flag.Bool("v", false, "Verbose mode.")
	debugp := flag.Bool("d", false, "Debug mode.")
	flag.Parse()

	verbose = *verbosep
	debug = *debugp
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

func do(basePath string) error {
	fset := token.NewFileSet() // positions are relative to fset
	g, err := newGlobal(fset, basePath)
	if err != nil {
		return err
	}

	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// if hidden directory - skip the entire dir
			if info.Name()[0] == '.' {
				if verbose {
					fmt.Println("Hidden folder, skipping: ", path)
				}
				return filepath.SkipDir
			}
			pkgs, err := parser.ParseDir(fset, path, nil, 0)
			if err != nil {
				log.Println("Error in parsing directory, skipping: ", path, err)
				return nil
			}
			for _, p := range pkgs {
				ast.Inspect(p, g.parse)
			}
		}
		return nil
	})

	err = g.finalize()
	if err != nil {
		return err
	}

	return nil
}
