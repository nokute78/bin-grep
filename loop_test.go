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
	"testing"
)

func TestScan(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	r := NewResult("", out, printSimple)

	b, err := hex.DecodeString("aabbccddeeff")
	if err != nil {
		t.Fatalf("hex.DecodeString:%s", err)
	}
	in := bytes.NewReader(b)
	targ, err := parseTarget("aabb")
	if err != nil {
		t.Fatalf("parseTarget:%s", err)
	}

	err = scan(in, r, targ, 0)
	if err != nil {
		t.Fatalf("scan:%s", err)
	}

	if out.Len() == 0 {
		t.Fatalf("no output")
	}
}
