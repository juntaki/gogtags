package main

import "testing"

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
