/*
MIT License

Copyright (c) 2018 Justus Rossmeier

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package gobaresip

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

const (
	lengthDelim byte = ':'
	dataDelim   byte = ','
)

type reader struct {
	r *bufio.Reader
}

func newReader(r io.Reader) *reader {
	return &reader{r: bufio.NewReader(r)}
}

func (r *reader) readNetstring() ([]byte, error) {
	length, err := r.r.ReadBytes(lengthDelim)
	if err != nil {
		return nil, err
	}

	var l int
	for _, ch := range length {
		if ch != lengthDelim && ch != dataDelim {
			ch -= '0'
			if ch > 9 {
				return nil, fmt.Errorf("wrong netstring length character")
			}
			l = l*10 + int(ch)
		}
	}
	if l <= 0 {
		return nil, fmt.Errorf("wrong netstring length")
	}

	ret := make([]byte, l)
	_, err = io.ReadFull(r.r, ret)
	if err != nil {
		return nil, err
	}
	next, err := r.r.ReadByte()
	if err != nil && err != io.EOF {
		return nil, err
	}
	if next != dataDelim {
		r.r.UnreadByte()
	}
	return ret, nil
}

func decode(in []byte) ([][]byte, error) {
	rd := newReader(bytes.NewReader(in))
	ret := make([][]byte, 0)
	for {
		d, err := rd.readNetstring()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		ret = append(ret, d)
	}
	return ret, nil
}

type writer struct {
	w io.Writer
}

func newWriter(w io.Writer) *writer {
	return &writer{w: w}
}

func (w *writer) writeNetstring(b []byte) error {
	_, err := w.w.Write([]byte(strconv.Itoa(len(b))))
	if err != nil {
		return err
	}
	_, err = w.w.Write([]byte{lengthDelim})
	if err != nil {
		return err
	}
	_, err = w.w.Write(b)
	if err != nil {
		return err
	}
	_, err = w.w.Write([]byte{dataDelim})
	if err != nil {
		return err
	}
	return nil
}

func encode(in ...[]byte) ([]byte, error) {
	var buf bytes.Buffer
	wr := newWriter(&buf)
	for _, d := range in {
		err := wr.writeNetstring(d)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
