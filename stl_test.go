// Copyright 2013 Silas Snider. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package stl

import (
	"fmt"
	"testing"
)

func TestEmptyParse(t *testing.T) {
	const input = ""
	sf, err := ParseSTLBytes([]byte(input))
	if err == nil {
		t.Errorf("Expecting an error with 0-length STL.")
	}
	if sf != nil {
		t.Errorf("Expecting nil SurfaceMesh if STL file is 0-length.")
	}
}

func TestASCIIParse(t *testing.T) {
	facets, err := ParseSTL("testdata/cube.ascii.stl")
	if err != nil {
		t.Fatalf("Unable to parse test file \"cube.ascii.stl\": %v", err)
	}

	if len(facets) != 12 {
		t.Error(fmt.Sprintf("Expected %v facets, got %v", 12, len(facets)))
	}
}

func TestBinaryParse(t *testing.T) {
	facets, err := ParseSTL("testdata/cube.binary.stl")
	if err != nil {
		t.Fatalf("Unable to parse test file \"cube.ascii.stl\": %v", err)
	}

	if len(facets) != 12 {
		t.Error(fmt.Sprintf("Expected %v facets, got %v", 12, len(facets)))
	}
}

func TestSimpleASCIIParse(t *testing.T) {
	const input = `solid OpenSCAD_Model
  facet normal 1 0 0
    outer loop
      vertex -2 0.5 -0.5
      vertex -2 1.5 -0.5
      vertex -2 1.5 0.5
    endloop
  endfacet
  facet normal 1 0 -0
    outer loop
      vertex -2 0.5 0.5
      vertex -2 0.5 -0.5
      vertex -2 1.5 0.5
    endloop
  endfacet
endsolid OpenSCAD_Model`
	facets, err := ParseSTLBytes([]byte(input))
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(facets) != 2 {
		t.Fatalf("Wrong number of facets: got %d, expected 2", len(facets))
	}
	tv := Vertex{X: -2, Y: 0.5, Z: -0.5}
	if facets[0].Vertices[0] != tv {
		t.Fatalf("Incorrect parse. Got: %+v, expected %+v", facets[0].Vertices[0], tv)
	}
	tv = Vertex{1, 0, 0}
	if facets[0].Normal != tv {
		t.Fatalf("Incorrect parse. Got %+v, expected %+v", facets[0].Normal, tv)
	}
	tv = Vertex{1, 0, -0}
	if facets[1].Normal != tv {
		t.Fatalf("Incorrect parse. Got %+v, expected %+v", facets[1].Normal, tv)
	}
	tv = Vertex{-2, 0.5, 0.5}
	if facets[1].Vertices[0] != tv {
		t.Fatalf("Incorrect parse. Got %+v, expected %+v", facets[1].Vertices[0], tv)
	}
}

func TestSimpleBinaryParse(t *testing.T) {
	const input = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x00\x00\x00\x00\x80?\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xc0\x00\x00\x00?\x00\x00\x00\xbf\x00\x00\x00\xc0\x00\x00\xc0?\x00\x00\x00\xbf\x00\x00\x00\xc0\x00\x00\xc0?\x00\x00\x00?\x00\x00\x00\x00\x80?\x00\x00\x00\x00\x00\x00\x00\x80\x00\x00\x00\xc0\x00\x00\x00?\x00\x00\x00?\x00\x00\x00\xc0\x00\x00\x00?\x00\x00\x00\xbf\x00\x00\x00\xc0\x00\x00\xc0?\x00\x00\x00?\x00\x00"
	facets, err := ParseSTLBytes([]byte(input))
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(facets) != 2 {
		t.Fatalf("Wrong number of facets: got %d, expected 2", len(facets))
	}
	tv := Vertex{X: -2, Y: 0.5, Z: -0.5}
	if facets[0].Vertices[0] != tv {
		t.Fatalf("Incorrect parse. Got: %+v, expected %+v", facets[0].Vertices[0], tv)
	}
}

func TestNonStandardBinaryParse(t *testing.T) {
	const input = "solid\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x00\x00\x00\x00\x80?\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xc0\x00\x00\x00?\x00\x00\x00\xbf\x00\x00\x00\xc0\x00\x00\xc0?\x00\x00\x00\xbf\x00\x00\x00\xc0\x00\x00\xc0?\x00\x00\x00?\x00\x00\x00\x00\x80?\x00\x00\x00\x00\x00\x00\x00\x80\x00\x00\x00\xc0\x00\x00\x00?\x00\x00\x00?\x00\x00\x00\xc0\x00\x00\x00?\x00\x00\x00\xbf\x00\x00\x00\xc0\x00\x00\xc0?\x00\x00\x00?\x00\x00"
	facets, err := ParseSTLBytes([]byte(input))
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(facets) != 2 {
		t.Fatalf("Wrong number of facets: got %d, expected 2", len(facets))
	}
	tv := Vertex{X: -2, Y: 0.5, Z: -0.5}
	if facets[0].Vertices[0] != tv {
		t.Fatalf("Incorrect parse. Got: %+v, expected %+v", facets[0].Vertices[0], tv)
	}
}
