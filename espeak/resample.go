/*
	Copyright (C) 2016 - 2018, Lefteris Zafiris <zaf@fastmail.com>
	This program is free software, distributed under the terms of
	the BSD 3-Clause License. See the LICENSE file
	at the top of the source tree.
*/

/*
Package resample implements resampling of PCM-encoded audio.
It uses the SoX Resampler library `libsoxr'.
To install make sure you have libsoxr installed, then run:
go get -u github.com/zaf/resample
The package warps an io.Reader in a Resampler that resamples and
writes all input data. Input should be RAW PCM encoded audio samples.
For usage details please see the code snippet in the cmd folder.
*/
package espeak

/*
#cgo LDFLAGS: ${SRCDIR}/libsoxr.a
#include <stdlib.h>
#include "soxr.h"
*/
import "C"
import (
	"encoding/binary"
	"errors"
	"io"
	"runtime"
	"unsafe"
)

const (
	// Quality settings
	Quick     = 0 // Quick cubic interpolation
	LowQ      = 1 // LowQ 16-bit with larger rolloff
	MediumQ   = 2 // MediumQ 16-bit with medium rolloff
	HighQ     = 4 // High quality
	VeryHighQ = 6 // Very high quality

	// Input formats
	F32 = 0 // 32-bit floating point PCM
	F64 = 1 // 64-bit floating point PCM
	I32 = 2 // 32-bit signed linear PCM
	I16 = 3 // 16-bit signed linear PCM

	byteLen = 8
)

// Resampler resamples PCM sound data.
type Resampler struct {
	resampler   C.soxr_t
	inRate      float64   // input sample rate
	outRate     float64   // output sample rate
	channels    int       // number of input channels
	frameSize   int       // frame size in bytes
	destination io.Writer // output data
}

var threads int

func init() {
	threads = runtime.NumCPU()
}

// NewResampler returns a pointer to a Resampler that implements an io.WriteCloser.
// It takes as parameters the destination data Writer, the input and output
// sampling rates, the number of channels of the input data, the input format
// and the quality setting.
func NewResampler(writer io.Writer, inputRate, outputRate float64, channels, format, quality int) (*Resampler, error) {
	var err error
	var size int
	if writer == nil {
		return nil, errors.New("io.Writer is nil")
	}
	if inputRate <= 0 || outputRate <= 0 {
		return nil, errors.New("Invalid input or output sampling rates")
	}
	if channels == 0 {
		return nil, errors.New("Invalid channels number")
	}
	if quality < 0 || quality > 6 {
		return nil, errors.New("Invalid quality setting")
	}
	switch format {
	case F64:
		size = 64 / byteLen
	case F32, I32:
		size = 32 / byteLen
	case I16:
		size = 16 / byteLen
	default:
		return nil, errors.New("Invalid format setting")
	}
	var soxr C.soxr_t
	var soxErr C.soxr_error_t
	// Setup soxr and create a stream resampler
	ioSpec := C.soxr_io_spec(C.soxr_datatype_t(format), C.soxr_datatype_t(format))
	qSpec := C.soxr_quality_spec(C.ulong(quality), 0)
	runtimeSpec := C.soxr_runtime_spec(C.uint(threads))
	soxr = C.soxr_create(C.double(inputRate), C.double(outputRate), C.uint(channels), &soxErr, &ioSpec, &qSpec, &runtimeSpec)
	if C.GoString(soxErr) != "" && C.GoString(soxErr) != "0" {
		err = errors.New(C.GoString(soxErr))
		C.free(unsafe.Pointer(soxErr))
		return nil, err
	}

	r := Resampler{
		resampler:   soxr,
		inRate:      inputRate,
		outRate:     outputRate,
		channels:    channels,
		frameSize:   size,
		destination: writer,
	}
	C.free(unsafe.Pointer(soxErr))
	return &r, err
}

// Reset permits reusing a Resampler rather than allocating a new one.
func (r *Resampler) Reset(writer io.Writer) (err error) {
	if r.resampler == nil {
		return errors.New("soxr resampler is nil")
	}
	r.destination = writer
	C.soxr_clear(r.resampler)
	return
}

// Close clean-ups and frees memory. Should always be called when
// finished using the resampler.
func (r *Resampler) Close() (err error) {
	if r.resampler == nil {
		return errors.New("soxr resampler is nil")
	}
	C.soxr_delete(r.resampler)
	r.resampler = nil
	return
}

// Write resamples PCM sound data. Writes len(p) bytes from p to
// the underlying data stream, returns the number of bytes written
// from p (0 <= n <= len(p)) and any error encountered that caused
// the write to stop early.
func (r *Resampler) Write(p []byte) (i int, err error) {
	if r.resampler == nil {
		err = errors.New("soxr resampler is nil")
		return
	}
	if len(p) == 0 {
		return
	}
	if fragment := len(p) % (r.frameSize * r.channels); fragment != 0 {
		// Drop fragmented frames from the end of input data
		p = p[:len(p)-fragment]
	}
	framesIn := len(p) / r.frameSize / r.channels
	if framesIn == 0 {
		err = errors.New("Incomplete input frame data")
		return
	}
	framesOut := int(float64(framesIn) * (r.outRate / r.inRate))
	if framesOut == 0 {
		err = errors.New("Not enough input to generate output")
		return
	}
	dataIn := C.CBytes(p)
	dataOut := C.malloc(C.size_t(framesOut * r.channels * r.frameSize))
	var soxErr C.soxr_error_t
	var read, done C.size_t = 0, 0
	var written int
	for int(done) < framesOut {
		soxErr = C.soxr_process(r.resampler, C.soxr_in_t(dataIn), C.size_t(framesIn), &read, C.soxr_out_t(dataOut), C.size_t(framesOut), &done)
		if C.GoString(soxErr) != "" && C.GoString(soxErr) != "0" {
			err = errors.New(C.GoString(soxErr))
			goto cleanup
		}
		if int(read) == framesIn && int(done) < framesOut {
			// Indicate end of input to the resampler
			var d C.size_t = 0
			soxErr = C.soxr_process(r.resampler, C.soxr_in_t(nil), C.size_t(0), nil, C.soxr_out_t(dataOut), C.size_t(framesOut), &d)
			if C.GoString(soxErr) != "" && C.GoString(soxErr) != "0" {
				err = errors.New(C.GoString(soxErr))
				goto cleanup
			}
			done += d
			break
		}
	}
	written, err = r.destination.Write(C.GoBytes(dataOut, C.int(int(done)*r.channels*r.frameSize)))
	i = int(float64(written) * (r.inRate / r.outRate))
	// If we have read all input and flushed all output, avoid to report short writes due
	// to output frames missing because of downsampling or other odd reasons.
	if err == nil && framesIn == int(read) && framesOut == int(done) {
		i = len(p)
	}

cleanup:
	C.free(dataIn)
	C.free(dataOut)
	C.free(unsafe.Pointer(soxErr))
	return
}

// http://soundfile.sapp.org/doc/WaveFormat/

func littleEndianIntToHex(integer int, numberOfBytes int) (bytes []byte) {
	bytes = make([]byte, numberOfBytes)
	switch numberOfBytes {
	case 2:
		binary.LittleEndian.PutUint16(bytes, uint16(integer))
	case 4:
		binary.LittleEndian.PutUint32(bytes, uint32(integer))
	}
	return
}

func applyString(dst []byte, s string, numberOfBytes int) {
	copy(dst, []byte(s)[:numberOfBytes])
}

func applyLittleEndianInteger(dst []byte, i int, numberOfBytes int) {
	copy(dst, littleEndianIntToHex(i, numberOfBytes)[0:numberOfBytes])
}

type riffChunk struct {
	ChunkId   [4]byte
	ChunkSize [4]byte
	Format    [4]byte
}

func (rc *riffChunk) applyChunkId(chunkId string) {
	applyString(rc.ChunkId[:], chunkId, 4)
}

func (rc *riffChunk) applyChunkSize(chunkSize int) {
	applyLittleEndianInteger(rc.ChunkSize[:], chunkSize, 4)
}

func (rc *riffChunk) applyFormat(format string) {
	applyString(rc.Format[:], format, 4)
}

type fmtSubChunk struct {
	Subchunk1Id   [4]byte
	Subchunk1Size [4]byte
	AudioFormat   [2]byte
	NumChannels   [2]byte
	SampleRate    [4]byte
	ByteRate      [4]byte
	BlockAlign    [2]byte
	BitsPerSample [2]byte
}

func (c *fmtSubChunk) applySubchunk1Id(subchunk1Id string) {
	applyString(c.Subchunk1Id[:], subchunk1Id, 4)
}

func (c *fmtSubChunk) applySubchunk1Size(subchunk1Size int) {
	applyLittleEndianInteger(c.Subchunk1Size[:], subchunk1Size, 4)
}

func (c *fmtSubChunk) applyAudioFormat(audioFormat int) {
	applyLittleEndianInteger(c.AudioFormat[:], audioFormat, 2)
}

func (c *fmtSubChunk) applyNumChannels(numChannels int) {
	applyLittleEndianInteger(c.NumChannels[:], numChannels, 2)
}

func (c *fmtSubChunk) applySampleRate(sampleRate int) {
	applyLittleEndianInteger(c.SampleRate[:], sampleRate, 4)
}

func (c *fmtSubChunk) applyByteRate(byteRate int) {
	applyLittleEndianInteger(c.ByteRate[:], byteRate, 4)
}

func (c *fmtSubChunk) applyBlockAlign(blockAlign int) {
	applyLittleEndianInteger(c.BlockAlign[:], blockAlign, 2)
}

func (c *fmtSubChunk) applyBitsPerSample(bitsPerSample int) {
	applyLittleEndianInteger(c.BitsPerSample[:], bitsPerSample, 2)
}

type dataSubChunk struct {
	Subchunk2Id   [4]byte
	Subchunk2Size [4]byte
}

func (c *dataSubChunk) applySubchunk2Id(subchunk2Id string) {
	applyString(c.Subchunk2Id[:], subchunk2Id, 4)
}

func (c *dataSubChunk) applySubchunk2Size(subchunk2Size int) {
	applyLittleEndianInteger(c.Subchunk2Size[:], subchunk2Size, 4)
}

func AddWavHeader(pcm []byte, channels int, sampleRate int, bitsPerSample int) (wav []byte, err error) {
	if channels != 1 && channels != 2 {
		return wav, errors.New("invalid_channels_value")
	}
	if sampleRate != 8000 && sampleRate != 16000 {
		return wav, errors.New("invalid_sample_rate_value")
	}
	if bitsPerSample != 8 && bitsPerSample != 16 {
		return wav, errors.New("invalid_bits_per_sample_value")
	}

	pcmLength := len(pcm)
	subchunk1Size := 16
	subchunk2Size := pcmLength
	chunkSize := 4 + (8 + subchunk1Size) + (8 + subchunk2Size)

	rc := riffChunk{}
	rc.applyChunkId("RIFF")
	rc.applyChunkSize(chunkSize)
	rc.applyFormat("WAVE")

	fsc := fmtSubChunk{}
	fsc.applySubchunk1Id("fmt ")
	fsc.applySubchunk1Size(subchunk1Size)
	fsc.applyAudioFormat(1)
	fsc.applyNumChannels(channels)
	fsc.applySampleRate(sampleRate)
	fsc.applyByteRate(sampleRate * channels * bitsPerSample / 8)
	fsc.applyBlockAlign(channels * bitsPerSample / 8)
	fsc.applyBitsPerSample(bitsPerSample)

	dsc := dataSubChunk{}
	dsc.applySubchunk2Id("data")
	dsc.applySubchunk2Size(subchunk2Size)

	wav = make([]byte, 0, pcmLength+8)

	wav = append(wav, rc.ChunkId[:]...)
	wav = append(wav, rc.ChunkSize[:]...)
	wav = append(wav, rc.Format[:]...)

	wav = append(wav, fsc.Subchunk1Id[:]...)
	wav = append(wav, fsc.Subchunk1Size[:]...)
	wav = append(wav, fsc.AudioFormat[:]...)
	wav = append(wav, fsc.NumChannels[:]...)
	wav = append(wav, fsc.SampleRate[:]...)
	wav = append(wav, fsc.ByteRate[:]...)
	wav = append(wav, fsc.BlockAlign[:]...)
	wav = append(wav, fsc.BitsPerSample[:]...)

	wav = append(wav, dsc.Subchunk2Id[:]...)
	wav = append(wav, dsc.Subchunk2Size[:]...)
	wav = append(wav, pcm...)

	return wav, err
}
