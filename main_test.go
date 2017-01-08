package main

import (
	"database/sql"
	"os"
	"testing"
)

func TestCompact(t *testing.T) {
	c := compact{
		lineNumbers: []int{10, 13, 15},
		fileID:      1,
	}
	if c.String() != "1 @n 10,3,2" {
		t.Error(c.String())
	}

	c = compact{
		lineNumbers: []int{10, 11, 12, 13, 15, 16},
		fileID:      1,
	}
	if c.String() != "1 @n 10-3,2-1" {
		t.Error(c.String())
	}
}

func BenchmarkInsert(b *testing.B) {
	file := GTAGS
	os.Remove("./" + file.String())
	g, _ := sql.Open("sqlite3", file.String())
	g.Exec(`create table db (key text, dat text, extra text)`)
	for i := 0; i < b.N; i++ {
		g.Exec(`insert into db (key, dat, extra) values (?, ?, ?)`, "key", "dat", "extra")
	}
}

func BenchmarkInsertCommit(b *testing.B) {
	file := GTAGS
	os.Remove("./" + file.String())
	g, _ := sql.Open("sqlite3", file.String())
	g.Exec(`create table db (key text, dat text, extra text)`)
	t, _ := g.Begin()
	for i := 0; i < b.N; i++ {
		t.Exec(`insert into db (key, dat, extra) values (?, ?, ?)`, "key", "dat", "extra")
	}
	t.Commit()
}
