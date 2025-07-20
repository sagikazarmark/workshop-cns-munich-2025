package main

import "testing"

func TestFoo(t *testing.T) {
	if false {
		t.Fatal("The universe is broken forever")
	}
}
