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
	"encoding/hex"
	"errors"
	"strings"
)

type Target struct {
	b       []byte
	special []bool
}

var ErrNoInput = errors.New("no input")
var ErrInputIsShort = errors.New("input buffer is short")
var ErrContinue = errors.New("continue")

const (
	magicWildcard = 0xff
)

func (t Target) IsSpecial(i uint64) bool {
	if uint64(len(t.special)) <= i {
		return false
	}
	return t.special[i]
}

func parseTargetNoSpecial(t string) ([]byte, error) {
	t = strings.TrimPrefix(t, "0x")
	t = strings.TrimPrefix(t, "0X")

	lenT := len(t)
	if lenT == 0 {
		return []byte{}, ErrNoInput
	}
	if lenT%2 == 1 {
		t = "0" + t
	}
	return hex.DecodeString(t)
}

func parseTarget(t string) (*Target, error) {
	lenT := len(t)
	if lenT == 0 {
		return nil, ErrNoInput
	}
	wildCount := strings.Count(t, ".")

	if wildCount == 0 {
		b, err := parseTargetNoSpecial(t)
		if err != nil {
			return &Target{}, nil
		}
		return &Target{b: b, special: make([]bool, len(b))}, nil
	}

	ret := &Target{}
	strs := strings.Split(t, ".")

	// if t == "...", strs = []string{"", "", "", ""}

	wildIndex := []int{}
	index := 0

	for i, v := range strs {
		b, err := parseTargetNoSpecial(v)
		if err != nil && err != ErrNoInput {
			return &Target{}, err
		}
		ret.b = append(ret.b, b...)
		index += len(b)
		if i < len(strs)-1 {
			wildIndex = append(wildIndex, index)
			ret.b = append(ret.b, magicWildcard)
			index += 1
		}
	}
	ret.special = make([]bool, len(ret.b))
	for _, v := range wildIndex {
		ret.special[v] = true
	}

	return ret, nil
}

func match(startAddr uint64, r *result, input []byte, t *Target) (uint64, error) {
	tIndex := uint64(0)
	icache := uint64(0) // start index cache
	var i uint64

	if len(t.b) > len(input) {
		return startAddr, ErrInputIsShort
	}
	for i = 0; i < uint64(len(input)); i++ {
		if input[i] == t.b[tIndex] || t.IsSpecial(tIndex) {
			if tIndex == uint64(len(t.b)-1) {
				/* matched */
				r.count += 1

				/////////////////////////
				m := &matchedBytes{}
				tailIndex := icache + uint64(len(t.b))
				m.startAddr = startAddr + icache
				m.b[1] = input[icache:tailIndex]

				startindex := icache
				start := (m.startAddr & 0xf)
				last := 0x10 - ((m.startAddr + tIndex) & 0xf)
				if icache >= start {
					startindex -= start
					m.startAddr -= start
				}
				m.b[0] = input[startindex:icache]

				lastindex := icache + tIndex
				if (last + lastindex) >= uint64(len(input)) {
					m.b[2] = input[tailIndex:]
				} else {
					m.b[2] = input[tailIndex : last+lastindex]
				}
				////////////////////////

				r.print(m)
				tIndex = 0
			} else {
				if tIndex == 0 {
					icache = i
				}
				tIndex++
			}
		} else {
			if tIndex > 0 {
				tIndex = 0
				i = icache
			}
		}
	}
	if tIndex > 0 {
		return startAddr + i - tIndex, ErrContinue
	}

	return startAddr + i, nil
}
