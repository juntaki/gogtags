package main

import (
	"bufio"
	"database/sql"
	"go/ast"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type global struct {
	fileDatas []*fileData
	// lineImageScanner
	basePath    string
	currentFile *os.File
	currentLine int
	scanner     *bufio.Scanner
	fset        *token.FileSet
}

func (g *global) appendFileData(path string) {
	new := &fileData{
		fileID:      len(g.fileDatas) + 1,
		absFilePath: path,
		gtagsData:   make([]standard, 0),
		grtagsData:  make(map[string]*compact),
	}

	g.fileDatas = append(g.fileDatas, new)
}

func (g *global) latestFileData() *fileData {
	if len(g.fileDatas) == 0 {
		return nil
	}
	return g.fileDatas[len(g.fileDatas)-1]
}

type fileData struct {
	fileID      int
	absFilePath string
	gtagsData   []standard
	grtagsData  map[string]*compact
}

func newGlobal(fset *token.FileSet, basePath string) (*global, error) {
	g := &global{
		fileDatas:   []*fileData{},
		basePath:    basePath,
		currentFile: nil,
		currentLine: 0,
		scanner:     nil,
		fset:        fset,
	}

	return g, nil
}

func insertEntry(tx *sql.Tx, key, dat, extra interface{}) {
	_, err := tx.Exec(`insert into db (key, dat, extra) values (?, ?, ?)`, key, dat, extra)
	if err != nil {
		log.Panicln("failed to exec", err, "|key:", key, "|dat:", dat, "|extra:", extra)
	}
}

func (g *global) finalize() error {
	if g.currentFile != nil {
		err := g.currentFile.Close()
		if err != nil {
			return err
		}
	}

	dbfiles := []tagType{
		GTAGS,
		GRTAGS,
		GPATH,
	}
	db := make(map[tagType]*sql.DB)
	transaction := make(map[tagType]*sql.Tx)
	var err error
	for _, file := range dbfiles {
		os.Remove("./" + file.String())
		db[file], err = sql.Open("sqlite3", file.String())
		if err != nil {
			return err
		}
		_, err = db[file].Exec(`create table db (key text, dat text, extra text)`)
		if err != nil {
			return err
		}
		transaction[file], err = db[file].Begin()
		if err != nil {
			return err
		}
	}

	insertEntry(transaction[GTAGS], " __.COMPRESS", " __.COMPRESS ddefine ttypedef", nil)
	insertEntry(transaction[GTAGS], " __.COMPNAME", " __.COMPNAME", nil)
	insertEntry(transaction[GTAGS], " __.VERSION", " __.VERSION 6", nil)

	insertEntry(transaction[GRTAGS], " __.COMPACT", " __.COMPACT", nil)
	insertEntry(transaction[GRTAGS], " __.COMPLINE", " __.COMPLINE", nil)
	insertEntry(transaction[GRTAGS], " __.COMPNAME", " __.COMPNAME", nil)
	insertEntry(transaction[GRTAGS], " __.VERSION", " __.VERSION 6", nil)

	insertEntry(transaction[GPATH], " __.VERSION", " __.VERSION 2", nil)
	insertEntry(transaction[GPATH], " __.NEXTKEY", "1", nil)

	for _, fd := range g.fileDatas {
		for _, s := range fd.gtagsData {
			insertEntry(transaction[GTAGS], s.tagName, s.String(), strconv.Itoa(s.fileID))
		}
		for tagName, compact := range fd.grtagsData {
			insertEntry(transaction[GRTAGS], tagName, compact.String(), strconv.Itoa(compact.fileID))
		}

		filepath, _ := filepath.Rel(g.basePath, fd.absFilePath)
		filepath = "./" + filepath
		if verbose {
			log.Println(filepath)
		}

		insertEntry(transaction[GPATH], filepath, fd.fileID, nil)
		insertEntry(transaction[GPATH], fd.fileID, filepath, nil)
		insertEntry(transaction[GPATH], " __.NEXTKEY", strconv.Itoa(fd.fileID+1), nil)
	}

	for _, file := range dbfiles {
		transaction[file].Commit()
		db[file].Close()
	}

	return nil
}

func (g *global) switchFile(node ast.Node) (err error) {
	pos := g.fset.Position(node.Pos())
	abspath, err := filepath.Abs(pos.Filename)
	if err != nil {
		log.Fatal("failed to get absolute path: ", err)
	}

	if g.currentFile != nil && g.currentFile.Name() == abspath {
		return nil
	}

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
	g.appendFileData(abspath)

	return nil
}

func getLineImage(filename string, line int) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", errors.Wrap(err, "failed to open next file ")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	for l := 0; l < line; l++ {
		scanner.Scan()
	}

	return scanner.Text(), nil
}

func (g *global) addFuncDecl(node *ast.FuncDecl) {
	pos := g.fset.Position(node.Pos())
	li, err := getLineImage(pos.Filename, pos.Line)
	if err != nil {
		return
	}
	lineImage := strings.Replace(strings.TrimSpace(li), node.Name.Name, "@n", -1)

	g.latestFileData().gtagsData = append(g.latestFileData().gtagsData, standard{
		tagName:    node.Name.Name,
		fileID:     g.latestFileData().fileID,
		lineNumber: pos.Line,
		lineImage:  lineImage,
	})
}

func (g *global) addIdent(ident *ast.Ident) {
	pos := g.fset.Position(ident.Pos())
	r, found := g.latestFileData().grtagsData[ident.Name]
	if found {
		r.lineNumbers = append(r.lineNumbers, pos.Line)
	} else {
		g.latestFileData().grtagsData[ident.Name] = &compact{
			fileID:      g.latestFileData().fileID,
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

	err := g.switchFile(node)
	if err != nil {
		log.Print("failed to switch file: ", err)
		return false
	}

	switch node.(type) {
	case *ast.FuncDecl:
		g.addFuncDecl(node.(*ast.FuncDecl))
	case *ast.Ident:
		g.addIdent(node.(*ast.Ident))
	}
	return true
}
