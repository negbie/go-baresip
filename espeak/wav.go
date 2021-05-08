package espeak

import (
	"encoding/binary"
	"fmt"
	"io"
)

// WriteTo writes the Samples in this Context to an io.Writer in WAV format.
func (ctx *Context) WriteTo(w io.Writer) (int64, error) {
	check32 := func(n int) int32 {
		if n < 0 {
			panic(fmt.Sprintf("espeak: unexpected negative number in wav: %d (possible overflow?)", n))
		}

		if int(int32(n)) != n {
			panic(fmt.Sprintf("espeak: value %d in wav header overflows int32", n))
		}

		return int32(n)
	}

	// based on https://gist.github.com/Jon-Schneider/8b7c53d27a7a13346a643dac9c19d34f
	var header struct {
		RiffHeader [4]byte
		WavSize    int32
		WaveHeader [4]byte

		FmtHeader       [4]byte
		FmtChunkSize    int32
		AudioFormat     int16
		NumChannels     int16
		SampleRate      int32
		ByteRate        int32
		SampleAlignment int16
		BitDepth        int16

		DataHeader [4]byte
		DataBytes  int32
	}

	// RIFF header
	header.RiffHeader = [...]byte{'R', 'I', 'F', 'F'}
	header.WavSize = check32(len(ctx.Samples)*2 + binary.Size(header) - 8)
	header.WaveHeader = [...]byte{'W', 'A', 'V', 'E'}

	// Format header
	header.FmtHeader = [...]byte{'f', 'm', 't', ' '}
	header.FmtChunkSize = 16
	header.AudioFormat = 1
	header.NumChannels = 1
	header.SampleRate = check32(SampleRate())
	header.ByteRate = check32(SampleRate() * 2)
	header.SampleAlignment = 2
	header.BitDepth = 16

	// Data
	header.DataHeader = [...]byte{'d', 'a', 't', 'a'}
	header.DataBytes = check32(len(ctx.Samples) * 2)

	cw := countWriter{w: w}
	cw.check(binary.Write(&cw, binary.LittleEndian, &header))
	cw.check(binary.Write(&cw, binary.LittleEndian, ctx.Samples))

	return cw.n, cw.err
}

type countWriter struct {
	w   io.Writer
	n   int64
	err error
}

func (cw *countWriter) Write(p []byte) (int, error) {
	if cw.err != nil {
		return 0, cw.err
	}

	n, err := cw.w.Write(p)
	cw.n += int64(n)
	cw.err = err
	return n, err
}

func (cw *countWriter) check(err error) {
	if cw.err == nil {
		cw.err = err
	}
}
