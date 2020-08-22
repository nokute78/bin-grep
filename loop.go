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
	"fmt"
	"io"
	"os"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} { return make([]byte, 4096) },
}

func scan(r io.Reader, result *result, t *Target, addr uint64) error {
	b := bufPool.Get().([]byte)
	defer bufPool.Put(b)
	var err error
	prev := []byte{}
	buf := []byte{}

	for {
		_, rErr := r.Read(b)
		if rErr != nil && rErr != io.EOF {
			return rErr
		}

		// append tail of last bytes
		if len(prev) > 0 {
			buf = bytes.Join([][]byte{prev, b}, []byte{})
		} else {
			buf = b
		}
		prev = []byte{}

		// grep
		addr, err = match(addr, result, buf, t)
		if err == ErrContinue {
			prev = b[len(b)-len(t.b)-1:]
		} else if err == ErrInputIsShort {
			prev = buf
		} else if err != nil {
			return err
		}

		// finish
		if rErr == io.EOF {
			break
		}
	}

	return nil
}

func loopMain(cnf *Config, o io.Writer, errStream io.Writer) {
	lenPaths := len(cnf.files)

	for i := 0; i < lenPaths; i++ {
		// setup result
		var r *result
		if cnf.addrOnly {
			r = NewResult(cnf.files[i]+": ", o, nil)
		} else {
			r = NewResult(cnf.files[i]+": ", o, printV)
		}
		if cnf.countMode {
			r.printer = printDevNull
		}

		// main loop
		f, err := os.Open(cnf.files[i])
		if err != nil {
			fmt.Fprintf(errStream, "%s:err=%s\n", cnf.files[i], err)
			continue
		}
		if cnf.skipByte > 0 {
			_, err = f.Seek(int64(cnf.skipByte), os.SEEK_SET)
			if err != nil {
				fmt.Fprintf(errStream, "%s:err=%s\n", cnf.files[i], err)
				f.Close()
				continue
			}
		}

		err = scan(f, r, cnf.t, uint64(cnf.skipByte))
		if err != nil {
			fmt.Fprintf(errStream, "%s:err=%s\n", cnf.files[i], err)
		}

		// closing
		if cnf.countMode {
			fmt.Fprintf(o, "%d\n", r.count)
		}
		f.Close()
	}
}
