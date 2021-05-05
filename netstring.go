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
	"io"
	"strconv"
	"strings"
)

const (
	lengthDelim byte = ':'
	dataDelim   byte = ','
)

type Reader struct {
	r *bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r)}
}

func (r *Reader) ReadNetstring() ([]byte, error) {
	length, err := r.r.ReadBytes(lengthDelim)
	if err != nil {
		return nil, err
	}
	l, err := strconv.Atoi(strings.TrimSuffix(string(length), string(lengthDelim)))
	if err != nil {
		return nil, err
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

func Decode(in []byte) ([][]byte, error) {
	rd := NewReader(bytes.NewReader(in))
	ret := make([][]byte, 0)
	for {
		d, err := rd.ReadNetstring()
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

type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (w *Writer) WriteNetstring(b []byte) error {
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

func Encode(in ...[]byte) ([]byte, error) {
	var buf bytes.Buffer
	wr := NewWriter(&buf)
	for _, d := range in {
		err := wr.WriteNetstring(d)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
