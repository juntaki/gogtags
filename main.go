package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

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

type global struct {
	fileID     int
	gtagsData  []standard
	grtagsData map[string]*compact
	db         map[string]*sql.DB
	// lineImageScanner
	basePath       string
	currentFile    *os.File
	currentRelPath string
	currentLine    int
	scanner        *bufio.Scanner
	fset           *token.FileSet
}

func newGlobal(fset *token.FileSet, basePath string) (*global, error) {
	g := &global{
		fileID:      0,
		gtagsData:   make([]standard, 0),
		grtagsData:  make(map[string]*compact),
		db:          make(map[string]*sql.DB),
		basePath:    basePath,
		currentFile: nil,
		currentLine: 0,
		scanner:     nil,
		fset:        fset,
	}

	dbfiles := []string{
		"GTAGS",
		"GRTAGS",
		"GPATH",
	}

	for _, file := range dbfiles {
		var err error
		os.Remove("./" + file)
		g.db[file], err = sql.Open("sqlite3", file)
		if err != nil {
			return nil, err
		}
		_, err = g.db[file].Exec(`create table db (key text primary key, dat text, extra text)`)
		if err != nil {
			return nil, err
		}
	}

	stmt, err := g.db["GTAGS"].Prepare(`insert into db (key, dat) values (?, ?)`)
	if err != nil {
		return nil, err
	}
	stmt.Exec(" __.COMPRESS", " __.COMPRESS ddefine ttypedef")
	stmt.Exec(" __.COMPNAME", " __.COMPNAME")
	stmt.Exec(" __.VERSION", " __.VERSION 6")

	stmt, err = g.db["GRTAGS"].Prepare(`insert into db (key, dat) values (?, ?)`)
	if err != nil {
		return nil, err
	}
	stmt.Exec(" __.COMPACT", " __.COMPACT")
	stmt.Exec(" __.COMPLINE", " __.COMPLINE")
	stmt.Exec(" __.COMPNAME", " __.COMPNAME")
	stmt.Exec(" __.VERSION", " __.VERSION 6")

	stmt, err = g.db["GPATH"].Prepare(`insert into db (key, dat) values (?, ?)`)
	if err != nil {
		return nil, err
	}
	stmt.Exec(" __.VERSION", " __.VERSION 2")
	stmt.Exec(" __.NEXTKEY", "1")

	return g, nil
}

func (g *global) dump() {
	if g.fileID == 0 {
		return
	}
	stmt, err := g.db["GTAGS"].Prepare(`insert into db (key, dat, extra) values (?, ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range g.gtagsData {
		stmt.Exec(s.tagName, s.String(), strconv.Itoa(s.fileID))
	}
	stmt, err = g.db["GRTAGS"].Prepare(`insert into db (key, dat, extra) values (?, ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	for tagName, compact := range g.grtagsData {
		stmt.Exec(tagName, compact.String(), strconv.Itoa(compact.fileID))
	}
	stmt, err = g.db["GPATH"].Prepare(`insert into db (key, dat) values (?, ?)`)
	if err != nil {
		log.Fatal(err)
	}

	filepath, _ := filepath.Rel(g.basePath, g.currentFile.Name())
	filepath = "./" + filepath
	log.Println(filepath)
	stmt.Exec(filepath, g.fileID)
	stmt.Exec(g.fileID, filepath)
	stmt.Exec(" __.NEXTKEY", strconv.Itoa(g.fileID+1))
}

func (g *global) finalize() error {
	if g.currentFile != nil {
		err := g.currentFile.Close()
		if err != nil {
			return err
		}
	}
	g.dump()
	return nil
}

func (g *global) switchFile(abspath string) (err error) {
	// Close and Setup Scanner
	if g.currentFile != nil {
		err := g.currentFile.Close()
		if err != nil {
			return errors.Wrapf(err, "failed to close current file, current: %s abspath: %s", g.currentFile.Name(), abspath)
		}
	}
	g.currentFile, err = os.Open(abspath)
	if err != nil {
		return errors.Wrap(err, "failed to open next file ")
	}
	g.scanner = bufio.NewScanner(g.currentFile)
	g.currentLine = 0

	// Reset parsed data
	g.gtagsData = make([]standard, 0)
	g.grtagsData = make(map[string]*compact)
	g.fileID++

	return nil
}

func (g *global) addFuncDecl(node *ast.FuncDecl) {
	ident := node.Name
	pos := g.fset.Position(node.Pos())
	for ; g.currentLine < pos.Line; g.currentLine++ {
		g.scanner.Scan()
	}
	lineImage := strings.Replace(strings.TrimSpace(g.scanner.Text()), ident.Name, "@n", -1)

	g.gtagsData = append(g.gtagsData, standard{
		tagName:    ident.Name,
		fileID:     g.fileID,
		lineNumber: pos.Line,
		lineImage:  lineImage,
	})
}

func (g *global) addIdent(ident *ast.Ident) {
	pos := g.fset.Position(ident.Pos())
	r, found := g.grtagsData[ident.Name]
	if found {
		r.lineNumbers = append(r.lineNumbers, pos.Line)
	} else {
		g.grtagsData[ident.Name] = &compact{
			fileID:      g.fileID,
			lineNumbers: []int{pos.Line},
		}
	}
}

func (g *global) parse(node ast.Node) bool {
	if node == nil {
		return false
	}
	if _, ok := node.(*ast.Package); ok {
		return true
	}
	pos := g.fset.Position(node.Pos())
	absPath, err := filepath.Abs(pos.Filename)
	if err != nil {
		log.Fatal("failed to get absolute path: ", err)
	}
	if g.currentFile == nil || g.currentFile.Name() != absPath {
		g.dump()

		err = g.switchFile(absPath)
		if err != nil {
			log.Fatal("failed to switch file: ", err)
		}
	}

	switch node.(type) {
	case *ast.FuncDecl:
		g.addFuncDecl(node.(*ast.FuncDecl))
	case *ast.Ident:
		g.addIdent(node.(*ast.Ident))
	}
	return true
}

func main() {
	basePath, err := filepath.Abs(".")
	if err != nil {
		log.Fatal("failed to get absolute path: ", err)
	}

	fset := token.NewFileSet() // positions are relative to fset
	g, err := newGlobal(fset, basePath)
	if err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			pkgs, err := parser.ParseDir(fset, path, nil, 0)
			if err != nil {
				log.Fatal(err)
			}
			for _, p := range pkgs {
				ast.Inspect(p, g.parse)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	err = g.finalize()
	if err != nil {
		log.Fatal(err)
	}
}
