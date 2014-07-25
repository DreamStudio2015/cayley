// Copyright 2014 The Cayley Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package nquads implements the RDF 1.1 N-Quads line-based syntax for RDF
// datasets as defined by http://www.w3.org/TR/n-quads/.
package nquads

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/google/cayley/graph"
)

// Parse returns a valid graph.Triple or a non-nil error. Parse does
// handle comments except where the comment placement does not prevent
// a complete valid graph.Triple from being defined.
func Parse(str string) (*graph.Triple, error) {
	t, err := parse([]rune(str))
	return &t, err
}

// Decoder implements N-Quad document parsing according to the RDF
// 1.1 N-Quads specification.
type Decoder struct {
	r    *bufio.Reader
	line []byte
}

// NewDecoder returns an N-Quad decoder that takes its input from the
// provided io.Reader.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: bufio.NewReader(r)}
}

// Unmarshal returns the next valid N-Quad as a graph.Triple, or an error.
func (dec *Decoder) Unmarshal() (*graph.Triple, error) {
	dec.line = dec.line[:0]
	var line []byte
	for {
		for {
			l, pre, err := dec.r.ReadLine()
			if err != nil {
				return nil, err
			}
			dec.line = append(dec.line, l...)
			if !pre {
				break
			}
		}
		if line = bytes.TrimSpace(dec.line); len(line) != 0 && line[0] != '#' {
			break
		}
		dec.line = dec.line[:0]
	}
	triple, err := Parse(string(line))
	if err != nil {
		return nil, fmt.Errorf("failed to parse %q: %v", dec.line, err)
	}
	if triple == nil {
		return dec.Unmarshal()
	}
	return triple, nil
}

func unEscape(r []rune, isEscaped bool) string {
	if !isEscaped {
		return string(r)
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(r)))

	for i := 0; i < len(r); {
		switch r[i] {
		case '\\':
			i++
			var c byte
			switch r[i] {
			case 't':
				c = '\t'
			case 'b':
				c = '\b'
			case 'n':
				c = '\n'
			case 'r':
				c = '\r'
			case 'f':
				c = '\f'
			case '"':
				c = '"'
			case '\'':
				c = '\''
			case '\\':
				c = '\\'
			case 'u':
				rc, err := strconv.ParseInt(string(r[i+1:i+5]), 16, 32)
				if err != nil {
					panic(fmt.Errorf("internal parser error: %v", err))
				}
				buf.WriteRune(rune(rc))
				i += 5
				continue
			case 'U':
				rc, err := strconv.ParseInt(string(r[i+1:i+9]), 16, 32)
				if err != nil {
					panic(fmt.Errorf("internal parser error: %v", err))
				}
				buf.WriteRune(rune(rc))
				i += 9
				continue
			}
			buf.WriteByte(c)
		default:
			buf.WriteRune(r[i])
		}
		i++
	}

	return buf.String()
}