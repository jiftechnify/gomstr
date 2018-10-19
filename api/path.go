package main

import (
	"strings"
)

// PathSeparator is the separeator of path.
const PathSeparator = "/"

// Path represents a path of resource(s) and a resource ID(optional).
type Path struct {
	Path string
	ID   string
}

// NewPath parses raw path string and constructs a Path object.
func NewPath(p string) *Path {
	var id string
	p = strings.Trim(p, PathSeparator)
	s := strings.Split(p, PathSeparator)
	if len(s) > 1 {
		id = s[len(s)-1]
		p = strings.Join(s[:len(s)-1], PathSeparator)
	}
	return &Path{Path: p, ID: id}
}

// HasID returns if the Path object contains resource ID.
func (p *Path) HasID() bool {
	return len(p.ID) > 0
}
