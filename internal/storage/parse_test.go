package storage

import (
	"testing"
)

func TestParseConnStr(t *testing.T) {
	_, err := ToConnData("postgres://foo:bar@localhost:5432/baz")
	if err != nil {
		t.Fatal(err)	
	}
	_, err = ToConnData("/foo/bar")
	if err != nil {
		t.Fatal(err)	
	}
	_, err = ToConnData("/foo/bar/")
	if err != nil {
		t.Fatal(err)	
	}
	_, err = ToConnData("foo/bar")
	if err == nil {
		t.Fatalf("expected error")
	}
	_, err = ToConnData("http://foo/bar")
	if err == nil {
		t.Fatalf("expected error")
	}
}
