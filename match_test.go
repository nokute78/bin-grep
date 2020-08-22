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
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

func TestParseTarget(t *testing.T) {
	type testcase struct {
		name   string
		input  string
		expect *Target
	}

	cases := []testcase{
		{"simple", "aabbccdd", &Target{b: []byte{0xaa, 0xbb, 0xcc, 0xdd}, special: []bool{false, false, false, false}}},
		{"0xsimple", "0xaabbccdd", &Target{b: []byte{0xaa, 0xbb, 0xcc, 0xdd}, special: []bool{false, false, false, false}}},
		{"0Xsimple", "0Xaabbccdd", &Target{b: []byte{0xaa, 0xbb, 0xcc, 0xdd}, special: []bool{false, false, false, false}}},
		{"lack0", "abbccdd", &Target{b: []byte{0x0a, 0xbb, 0xcc, 0xdd}, special: []bool{false, false, false, false}}},
		{"wild0", ".aabbccdd", &Target{b: []byte{magicWildcard, 0xaa, 0xbb, 0xcc, 0xdd}, special: []bool{true, false, false, false, false}}},
		{"wild1", "...", &Target{b: []byte{magicWildcard, magicWildcard, magicWildcard}, special: []bool{true, true, true}}},
		{"wild2", "ab.", &Target{b: []byte{0xab, magicWildcard}, special: []bool{false, true}}},
		{"wild3", "aa.bb", &Target{b: []byte{0xaa, magicWildcard, 0xbb}, special: []bool{false, true, false}}},
	}

	for _, v := range cases {
		ret, err := parseTarget(v.input)
		if err != nil {
			t.Errorf("%s:err %s\n", v.name, err)
			continue
		}
		if len(ret.b) != len(v.expect.b) {
			t.Errorf("%s:target len mistmach given=%d expect=%d\n", v.name, len(ret.b), len(v.expect.b))
			continue
		}
		for i := 0; i < len(ret.b); i++ {
			if ret.b[i] != v.expect.b[i] {
				t.Errorf("%s:target mistmach given=%v expect=%v\n", v.name, ret.b, v.expect.b)
			}
		}

		if bytes.Compare(ret.b, v.expect.b) != 0 {
			t.Errorf("%s:mistmach given=%x expect=%x\n", v.name, ret.b, v.expect.b)
			continue
		}
		if len(ret.special) != len(v.expect.special) {
			t.Errorf("%s:wildcard len mistmach given=%d expect=%d\n", v.name, len(ret.special), len(v.expect.special))
			continue
		}

		for i := 0; i < len(ret.special); i++ {
			if ret.special[i] != v.expect.special[i] {
				t.Errorf("%s:wildcard mistmach given=%v expect=%v\n", v.name, ret.special, v.expect.special)
			}
		}
	}

	_, err := parseTarget("")
	if err == nil {
		t.Errorf("err is nil !?")
	}
}

func TestIsSpecial(t *testing.T) {
	targ, err := parseTarget("0xaa.bb")
	if err != nil {
		t.Fatalf("parseTarget error:%s", err)
	}

	if targ.IsSpecial(0) {
		t.Errorf("index 0 is not special")
	}
	if !targ.IsSpecial(1) {
		t.Errorf("index 1 is special")
	}
	if targ.IsSpecial(2) {
		t.Errorf("index 2 is not special")
	}

	if targ.IsSpecial(100) {
		t.Errorf("out of range and should be false")
	}

}

func TestMatchSimple(t *testing.T) {
	targetStr := "aabbccdd"
	hexTarget, err := hex.DecodeString(targetStr)
	if err != nil {
		t.Fatalf("DecodeString err=%s", err)
	}

	buf := bytes.NewBuffer([]byte{})
	targ, err := parseTarget(targetStr)
	if err != nil {
		t.Fatalf("parseTarget err=%s", err)
	}

	r := NewResult("", buf, nil)
	input := make([]byte, len(targetStr)/2-2)

	_, err = match(0, r, input, targ)
	if err != ErrInputIsShort {
		t.Errorf("given %s expect %s", err, ErrInputIsShort)
	}

	input = make([]byte, 16)
	input[len(input)-1] = hexTarget[0]
	_, err = match(0, r, input, targ)
	if err != ErrContinue {
		t.Errorf("given %s expect %s", err, ErrContinue)
	}

	buf.Reset()
	inputBuf := bytes.NewBuffer([]byte{})

	_, err = inputBuf.WriteString(strings.Repeat("aaaaaaaa", 256))
	if err != nil {
		t.Fatalf("WriteString:%s", err)
	}

	_, err = inputBuf.Write(hexTarget)
	if err != nil {
		t.Fatalf("WriteString:%s", err)
	}

	_, err = inputBuf.WriteString(strings.Repeat("aaaaaaaa", 256))
	if err != nil {
		t.Fatalf("WriteString:%s", err)
	}

	_, err = match(0, r, inputBuf.Bytes(), targ)
	if err != nil {
		t.Errorf("match err=%s", err)
	}

	if buf.Len() == 0 {
		t.Errorf("output is blank")
	}
}
