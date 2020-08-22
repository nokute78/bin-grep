/*
   Copyright 2020 Takahiro Yamashita

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

type result struct {
	header  string
	count   uint64
	output  io.Writer
	printer func(m *matchedBytes, o io.Writer, h string)
}

var colorP func(w io.Writer, s string, a ...interface{})

func init() {
	colorP = noRetFprintf
}

func noRetFprintf(w io.Writer, s string, a ...interface{}) {
	fmt.Fprintf(w, s, a...)
}

func SetColorOutput() {
	colorP = color.New(color.FgRed).FprintfFunc()
}

func NewResult(h string, o io.Writer, f func(m *matchedBytes, o io.Writer, h string)) *result {
	fPrinter := f
	if f == nil {
		fPrinter = printSimple
	}

	return &result{header: h, printer: fPrinter, output: o}
}

func (r result) print(m *matchedBytes) {
	r.printer(m, r.output, r.header)
}

func pHeader(sb *strings.Builder, startAddr uint64, header string) uint64 {
	count := uint64(startAddr & 0xf)
	index := uint64(0)
	if count > 0 {
		fmt.Fprintf(sb, "%s%016x ", header, startAddr-count)
		for i := 0; i < int(count); i++ {
			sb.WriteRune('-')
			if index%4 == 3 {
				sb.WriteRune(' ')
			}
			index++
		}
	}
	return index
}

func printDevNull(m *matchedBytes, o io.Writer, header string) {
	// nothing to do
}

func printSimple(m *matchedBytes, o io.Writer, header string) {
	fmt.Fprintf(o, "%s%016x\n", header, m.startAddr+uint64(len(m.b[0])))
}

func printV(m *matchedBytes, o io.Writer, header string) {
	sb := &strings.Builder{}
	count := pHeader(sb, m.startAddr, header)
	index := 0

	//	fmt.Printf("%s\n", m)

	for _, v := range m.b[0] {
		if count&0xf == 0 {
			sb.WriteString(header)
			fmt.Fprintf(sb, "%016x ", m.startAddr+count)
		}

		fmt.Fprintf(sb, "%02x", v)

		if count%2 == 1 {
			sb.WriteRune(' ')
		}
		if count&0xf == 15 {
			sb.WriteRune('\n')
		}
		count++
		index++
	}

	for _, v := range m.b[1] {
		if count&0xf == 0 {
			sb.WriteString(header)
			fmt.Fprintf(sb, "%016x ", m.startAddr+count)
		}

		colorP(sb, "%02x", v)

		if count%2 == 1 {
			sb.WriteRune(' ')
		}
		if count&0xf == 15 {
			sb.WriteRune('\n')
		}
		count++
		index++
	}

	for _, v := range m.b[2] {
		if count&0xf == 0 {
			sb.WriteString(header)
			fmt.Fprintf(sb, "%016x ", m.startAddr+count)
		}

		fmt.Fprintf(sb, "%02x", v)

		if count%2 == 1 {
			sb.WriteRune(' ')
		}
		if count&0xf == 15 {
			sb.WriteRune('\n')
		}
		count++
		index++
	}
	sb.WriteRune('\n')
	io.WriteString(o, sb.String())

}

type matchedBytes struct {
	startAddr uint64
	b         [3][]byte
}

func (m matchedBytes) String() string {
	return fmt.Sprintf("start=0x%x\nb(len=%d)=%x\n", m.startAddr, len(m.b), m.b)
}
