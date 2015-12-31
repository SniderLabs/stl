// Copyright 2013 Silas Snider. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package stl

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	ident       = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_0123456789."
	spaces      = " \n\t\r"
	floatDigits = "0123456789."
)

//Vertex represents a point in 3d space.
type Vertex struct {
	X, Y, Z float32
}

//Facet represents a triangle from the STL file.
type Facet struct {
	Vertices [3]Vertex
	Valid    bool
	Normal   Vertex
}

//LexerError represents an error generated by the lexer.
type LexerError struct {
	e string
}
type stateFn func(*lexer) stateFn
type lexer struct {
	name     string
	input    []byte
	state    stateFn
	start    int
	pos      int
	items    chan *Facet
	error    string
	meshName string

	pendingFacet *Facet
}

func (l LexerError) Error() string {
	return l.e
}

func lex(name string, input []byte, initialState stateFn) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		state: initialState,
		items: make(chan *Facet, 2), // Two items sufficient.
	}
	return l
}

func (l *lexer) nextFacet() *Facet {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			l.state = l.state(l)
			if l.state == nil {
				return nil
			}
		}
	}
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		return -1
	}
	r = rune(l.input[l.pos])
	l.pos++
	return r
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos--
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) bool {
	result := false
	for strings.IndexRune(valid, l.next()) >= 0 {
		result = true
	}
	l.backup()
	return result
}

func (l *lexer) acceptKeyword(keyword string) bool {
	if bytes.Equal(l.input[l.pos:l.pos+len(keyword)], []byte(keyword)) {
		l.pos += len(keyword)
		return true
	}
	return false
}

func (l *lexer) emit(f *Facet) {
	l.items <- f
	l.start = l.pos
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.error = fmt.Sprintf(format, args...)
	return nil
}

func (l *lexer) acceptNumber() bool {
	l.accept("+-")
	if !l.acceptRun(floatDigits) {
		return false
	}
	if l.accept("eE") {
		l.accept("+-")
		if !l.acceptRun(floatDigits) {
			return false
		}
	}
	return true
}

func asciiVertex(l *lexer) stateFn {
	var vs [3]Vertex
	for i := 0; i < 3; i++ {
		l.acceptRun(spaces)
		if !l.acceptKeyword("vertex") {
			return l.errorf("Expected keyword 'vertex' at %v. Saw %q instead", l.pos, l.input[l.pos:l.pos+6])
		}
		var xyz [3]float32
		for j := 0; j < 3; j++ {
			l.acceptRun(spaces)
			c := l.pos
			if !l.acceptNumber() {
				return l.errorf("Expected a number at %v. Saw %q instead", l.pos, l.input[l.pos:l.pos+1])
			}
			t, err := strconv.ParseFloat(string(l.input[c:l.pos]), 32)
			if err != nil {
				return l.errorf("Unable to parse float from %q", l.input[c:l.pos])
			}
			xyz[j] = float32(t)
		}
		vs[i] = Vertex{X: xyz[0], Y: xyz[1], Z: xyz[2]}
	}
	l.acceptRun(spaces)
	if !l.acceptKeyword("endloop") {
		return l.errorf("Expected keyword 'endloop' at %v. Saw %q instead", l.pos, l.input[l.pos:l.pos+7])
	}
	l.acceptRun(spaces)
	if !l.acceptKeyword("endfacet") {
		return l.errorf("Expected keyword 'endfacet' at %v. Saw %q instead", l.pos, l.input[l.pos:l.pos+8])
	}
	l.acceptRun(spaces)
	l.pendingFacet.Vertices = vs
	l.pendingFacet.Valid = true
	l.emit(l.pendingFacet)
	l.pendingFacet = nil
	return asciiFacet
}

func asciiFacet(l *lexer) stateFn {
	l.acceptRun(spaces)
	if !l.acceptKeyword("facet") {
		if l.acceptKeyword("endsolid") {
			return nil
		}
		return l.errorf("Expected keyword 'facet' at %v. Saw %q instead", l.pos, l.input[l.pos:l.pos+5])
	}
	l.acceptRun(spaces)
	if !l.acceptKeyword("normal") {
		return l.errorf("Expected keyword 'normal' at %v. Saw %q instead", l.pos, l.input[l.pos:l.pos+6])
	}

	// Parse the normal
	l.acceptRun(spaces)
	l.ignore()
	if !l.acceptNumber() {
		return l.errorf("Expected a number at %v. Saw %q instead", l.pos, l.input[l.pos:l.pos+1])
	}
	ni, err := strconv.ParseFloat(string(l.input[l.start:l.pos]), 32)
	if err != nil {
		return l.errorf("Unable to parse float from %q", l.input[l.start:l.pos])
	}

	l.acceptRun(spaces)
	l.ignore()
	if !l.acceptNumber() {
		return l.errorf("Expected a number at %v. Saw %q instead", l.pos, l.input[l.pos:l.pos+1])
	}
	nj, err := strconv.ParseFloat(string(l.input[l.start:l.pos]), 32)
	if err != nil {
		return l.errorf("Unable to parse float from %q", l.input[l.start:l.pos])
	}

	l.acceptRun(spaces)
	l.ignore()
	if !l.acceptNumber() {
		return l.errorf("Expected a number at %v. Saw %q instead", l.pos, l.input[l.pos:l.pos+1])
	}
	nk, err := strconv.ParseFloat(string(l.input[l.start:l.pos]), 32)
	if err != nil {
		return l.errorf("Unable to parse float from %q", l.input[l.start:l.pos])
	}

	l.acceptRun(spaces)

	if !l.acceptKeyword("outer") {
		return l.errorf("Expected keyword 'outer' at %v. Saw %q instead", l.pos, l.input[l.pos:l.pos+5])
	}
	l.acceptRun(spaces)
	if !l.acceptKeyword("loop") {
		return l.errorf("Expected keyword 'loop' at %v. Saw %q instead", l.pos, l.input[l.pos:l.pos+4])
	}

	l.pendingFacet = &Facet{
		Normal: Vertex{
			X: float32(ni),
			Y: float32(nj),
			Z: float32(nk),
		},
	}
	return asciiVertex
}

func asciiFileHeader(l *lexer) stateFn {
	l.acceptRun(spaces)
	l.acceptRun(ident)
	c := string(l.input[l.start:l.pos])
	if c == "facet" {
		l.ignore()
		return asciiFacet
	}
	l.meshName = c
	return asciiFacet
}

// parseASCII knows how to parse ASCII STL files.
func parseASCII(stl []byte) ([]*Facet, error) {
	var facets []*Facet
	l := lex("ASCII STL", stl, asciiFileHeader)
	f := l.nextFacet()
	for f != nil && f.Valid {
		facets = append(facets, f)
		f = l.nextFacet()
	}
	if l.error != "" {
		return nil, LexerError{e: l.error}
	}
	return facets, nil
}

// parseBinary knows how to parse binary STL files.
func parseBinary(stl []byte) ([]*Facet, error) {
	//Skip 80 byte header
	if len(stl) < 80 {
		return nil, errors.New("Incomplete header on binary STL file.")
	}
	stl = stl[80:]
	if len(stl) < 4 {
		return nil, errors.New("Binary STL file contains no data.")
	}
	l := bytes.NewBuffer(stl[:4])
	triangles := bytes.NewBuffer(stl[4:])

	var numTriangles uint32
	var t float32
	var ut uint16
	var ps [3]float32
	err := binary.Read(l, binary.LittleEndian, &numTriangles)
	if err != nil {
		return nil, err
	}
	facets := make([]*Facet, numTriangles)
	for i := uint32(0); i < numTriangles; i++ {
		var vs [3]Vertex
		var normal Vertex
		//Read the normal
		for j := 0; j < 3; j++ {
			err = binary.Read(triangles, binary.LittleEndian, &t)
			if err != nil {
				return nil, err
			}
			normal = Vertex{X: ps[0], Y: ps[1], Z: ps[2]}
		}
		//Read the vertices
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				err = binary.Read(triangles, binary.LittleEndian, &ps[k])
				if err != nil {
					return nil, err
				}
			}
			vs[j] = Vertex{X: ps[0], Y: ps[1], Z: ps[2]}
		}
		facets[i] = &Facet{
			Vertices: vs,
			Normal:   normal,
			Valid:    true,
		}
		err = binary.Read(triangles, binary.LittleEndian, &ut)
		if err != nil {
			return nil, err
		}
	}

	return facets, nil
}

// ParseSTL returns the list of Facets that corresponds to the STL file passed in.
func ParseSTL(stlPath string) ([]*Facet, error) {
	stl, err := ioutil.ReadFile(stlPath)
	if err != nil {
		return nil, err
	}
	return ParseSTLBytes(stl)
}

// ParseSTLBytes returns the list of Facets that corresponds to the STL bytes passed in.
func ParseSTLBytes(stl []byte) ([]*Facet, error) {
	if len(stl) < 5 {
		return nil, errors.New("STL file too small (<5 bytes)")
	}
	if string(stl[0:5]) == "solid" {
		facets, err := parseASCII(stl[5:])
		if err != nil {
			facets, err2 := parseBinary(stl)
			if err2 != nil {
				return nil, err
			}
			return facets, nil
		}
		return facets, nil
	}
	return parseBinary(stl)
}
