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
	fileDatas map[string]*fileData
	// lineImageScanner
	basePath string
	fset     *token.FileSet
}

func (g *global) fileData(path string) *fileData {
	if val, ok := g.fileDatas[path]; ok {
		return val
	}

	relpath, _ := filepath.Rel(g.basePath, path)
	relpath = "./" + relpath
	if verbose {
		log.Println(relpath)
	}

	new := &fileData{
		fileID:      len(g.fileDatas) + 1,
		absFilePath: relpath,
		gtagsData:   make([]standard, 0),
		grtagsData:  make(map[string]*compact),
	}

	g.fileDatas[path] = new
	return new
}

type fileData struct {
	fileID      int
	absFilePath string
	gtagsData   []standard
	grtagsData  map[string]*compact
}

func newGlobal(fset *token.FileSet, basePath string) (*global, error) {
	g := &global{
		fileDatas: make(map[string]*fileData),
		basePath:  basePath,
		fset:      fset,
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
	dbfiles := []tagType{
		gtags,
		grtags,
		gpath,
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

	insertEntry(transaction[gtags], " __.COMPRESS", " __.COMPRESS ddefine ttypedef", nil)
	insertEntry(transaction[gtags], " __.COMPNAME", " __.COMPNAME", nil)
	insertEntry(transaction[gtags], " __.VERSION", " __.VERSION 6", nil)

	insertEntry(transaction[grtags], " __.COMPACT", " __.COMPACT", nil)
	insertEntry(transaction[grtags], " __.COMPLINE", " __.COMPLINE", nil)
	insertEntry(transaction[grtags], " __.COMPNAME", " __.COMPNAME", nil)
	insertEntry(transaction[grtags], " __.VERSION", " __.VERSION 6", nil)

	insertEntry(transaction[gpath], " __.VERSION", " __.VERSION 2", nil)
	insertEntry(transaction[gpath], " __.NEXTKEY", "1", nil)

	for _, fd := range g.fileDatas {
		for _, s := range fd.gtagsData {
			insertEntry(transaction[gtags], s.tagName, s.String(), strconv.Itoa(s.fileID))
		}
		for tagName, compact := range fd.grtagsData {
			insertEntry(transaction[grtags], tagName, compact.String(), strconv.Itoa(compact.fileID))
		}

		insertEntry(transaction[gpath], fd.absFilePath, fd.fileID, nil)
		insertEntry(transaction[gpath], fd.fileID, fd.absFilePath, nil)
		insertEntry(transaction[gpath], " __.NEXTKEY", strconv.Itoa(fd.fileID+1), nil)
	}

	for _, file := range dbfiles {
		err = transaction[file].Commit()
		if err != nil {
			return err
		}
		db[file].Close()
	}

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
	path := pos.Filename
	li, err := getLineImage(pos.Filename, pos.Line)
	if err != nil {
		return
	}
	lineImage := strings.Replace(strings.TrimSpace(li), node.Name.Name, "@n", -1)

	g.fileData(path).gtagsData = append(g.fileData(path).gtagsData, standard{
		tagName:    node.Name.Name,
		fileID:     g.fileData(path).fileID,
		lineNumber: pos.Line,
		lineImage:  lineImage,
	})
}

func (g *global) addIdent(ident *ast.Ident) {
	pos := g.fset.Position(ident.Pos())
	path := pos.Filename
	r, found := g.fileData(path).grtagsData[ident.Name]
	if found {
		r.lineNumbers = append(r.lineNumbers, pos.Line)
	} else {
		g.fileData(path).grtagsData[ident.Name] = &compact{
			fileID:      g.fileData(path).fileID,
			lineNumbers: []int{pos.Line},
		}
	}
}

func (g *global) addTypeSpec(typeSpec *ast.TypeSpec) {
	pos := g.fset.Position(typeSpec.Pos())
	path := pos.Filename

	li, err := getLineImage(pos.Filename, pos.Line)
	if err != nil {
		return
	}
	lineImage := strings.Replace(strings.TrimSpace(li), typeSpec.Name.Name, "@n", -1)
	g.fileData(path).gtagsData = append(g.fileData(path).gtagsData, standard{
		tagName:    typeSpec.Name.Name,
		fileID:     g.fileData(path).fileID,
		lineNumber: pos.Line,
		lineImage:  lineImage,
	})
}


func (g *global) parse(node ast.Node) bool {
	if node == nil {
		return false
	}
	if _, ok := node.(*ast.Package); ok {
		return true
	}

	switch node := node.(type) {
	case *ast.FuncDecl:
		g.addFuncDecl(node)
	case *ast.Ident:
		g.addIdent(node)
	case *ast.TypeSpec:
		g.addTypeSpec(node)
	}
	return true
}
