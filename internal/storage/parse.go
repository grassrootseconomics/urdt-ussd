package storage

import (
	"fmt"
	"net/url"
	"path"
)

const (
	DBTYPE_MEM = iota
	DBTYPE_GDBM
	DBTYPE_POSTGRES
)

type ConnData struct {
	typ int
	str string
	domain string
}

func (cd *ConnData) DbType() int {
	return cd.typ
}

func (cd *ConnData) String() string {
	return cd.str
}

func (cd *ConnData) Domain() string {
	return cd.domain
}

func (cd *ConnData) Path() string {
	v, _ := url.Parse(cd.str)
	v.RawQuery = ""
	return v.String()
}

func probePostgres(s string) (string, string, bool) {
	domain := "public"
	v, err := url.Parse(s)
	if err != nil {
		return "", "", false
	}
	if v.Scheme != "postgres" {
		return "", "", false
	}
	vv := v.Query()
	if vv.Has("search_path") {
		domain = vv.Get("search_path")
	}
	return s, domain, true
}

func probeGdbm(s string) (string, string, bool) {
	if !path.IsAbs(s) {
		return "", "", false
	}
	s = path.Clean(s)
	return s, "", true
}

func ToConnData(connStr string) (ConnData, error) {
	var o ConnData

	if connStr == "" {
		return o, nil
	}

	v, domain, ok := probePostgres(connStr)
	if ok {
		o.typ = DBTYPE_POSTGRES
		o.str = v
		o.domain = domain
		return o, nil
	}

	v, _, ok = probeGdbm(connStr)
	if ok {
		o.typ = DBTYPE_GDBM
		o.str = v
		return o, nil
	}

	return o, fmt.Errorf("invalid connection string: %s", connStr)
}
