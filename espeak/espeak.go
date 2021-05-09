// Author Quentin Quaadgras
// Licensed under the the BSD style license, and the LGPL (Lesser GNU Public License);

package espeak

/*
#cgo linux LDFLAGS: ${SRCDIR}/libespeak-ng.a -lm

#include <stdlib.h>
#include <string.h>
#include <speak_lib.h>

void* user_data;
unsigned int *unique_identifier;
unsigned int samplestotal = 0;
int wavsamplerate;
char *wavefile=NULL;
FILE *f_wavfile = NULL;
int OpenWavFile(char *path, int rate);
void CloseWavFile();

int callback(short *wav, int numsamples, espeak_EVENT *events)
{
	int type;
	if(wav == NULL)
	{
		CloseWavFile();
		return(1);
	}
	if(f_wavfile == NULL){
		if(OpenWavFile(wavefile, wavsamplerate) != 0){
			return(1);
		}
	}
	if(numsamples > 0){
		samplestotal += numsamples;
		fwrite(wav,numsamples*2,1,f_wavfile);
	}
	return(0);
}

static void Write4Bytes(FILE *f, int value)
{
    int ix;
    for (ix = 0; ix < 4; ix++)
    {
        fputc(value & 0xff, f);
        value = value >> 8;
    }
}

int OpenWavFile(char *path, int rate)
{
    static unsigned char wave_hdr[44] = {
        'R', 'I', 'F', 'F', 0x24, 0xf0, 0xff, 0x7f, 'W', 'A', 'V', 'E', 'f', 'm', 't', ' ',
        0x10, 0, 0, 0, 1, 0, 1, 0, 9, 0x3d, 0, 0, 0x12, 0x7a, 0, 0,
        2, 0, 0x10, 0, 'd', 'a', 't', 'a', 0x00, 0xf0, 0xff, 0x7f};
    if (path == NULL)
        return (2);
    if (path[0] == 0)
        return (0);
    if (strcmp(path, "stdout") == 0)
        f_wavfile = stdout;
    else
        f_wavfile = fopen(path, "wb");
    if (f_wavfile == NULL)
    {
        fprintf(stderr, "Can't write to: '%s'\n", path);
        return (1);
    }
    fwrite(wave_hdr, 1, 24, f_wavfile);
    Write4Bytes(f_wavfile, rate);
    Write4Bytes(f_wavfile, rate * 2);
    fwrite(&wave_hdr[32], 1, 12, f_wavfile);
    return (0);
}

void CloseWavFile()
{
    unsigned int pos;
    if ((f_wavfile == NULL) || (f_wavfile == stdout))
        return;
    fflush(f_wavfile);
    pos = ftell(f_wavfile);
    fseek(f_wavfile, 4, SEEK_SET);
    Write4Bytes(f_wavfile, pos - 8);
    fseek(f_wavfile, 40, SEEK_SET);
    Write4Bytes(f_wavfile, pos - 44);
    fclose(f_wavfile);
    f_wavfile = NULL;
}

static int EspeakSynth(const char *text, unsigned int position, espeak_POSITION_TYPE position_type, unsigned int end_position, unsigned int flags)
{
	unsigned int size_t;
	size_t = strlen(text)+1;
	return (int) espeak_Synth( text, size_t, position, position_type, end_position, flags, unique_identifier, user_data );
}
*/
import "C"
import (
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

type EspeakAudioOutput int

const (
	AUDIO_OUTPUT_PLAYBACK       EspeakAudioOutput = 0
	AUDIO_OUTPUT_RETRIEVAL      EspeakAudioOutput = 1
	AUDIO_OUTPUT_SYNCHRONOUS    EspeakAudioOutput = 2
	AUDIO_OUTPUT_SYNCH_PLAYBACK EspeakAudioOutput = 3
)

type EspeakPositionType int

const (
	POS_CHARACTER EspeakPositionType = 1
	POS_WORD      EspeakPositionType = 2
	POS_SENTENCE  EspeakPositionType = 3
)

type EspeakSynthFlag int

const (
	FLAG_SSML          EspeakSynthFlag = 0x10
	FLAG_PHONEMES      EspeakSynthFlag = 0x100
	FLAG_ENDPAUSE      EspeakSynthFlag = 0x1000
	FLAG_KEEP_NAMEDATA EspeakSynthFlag = 0x2000
)

func Initialize(ngDataPath string) int {
	cpath := C.CString(ngDataPath)
	defer C.free(unsafe.Pointer(cpath))
	C.wavsamplerate = C.espeak_Initialize(C.espeak_AUDIO_OUTPUT(AUDIO_OUTPUT_SYNCHRONOUS), C.int(200), cpath, C.int(0))
	return int(C.wavsamplerate)
}

func SetVoiceByName(voice string) int {
	cvoice := C.CString(voice)
	defer C.free(unsafe.Pointer(cvoice))
	return int(C.espeak_SetVoiceByName(cvoice))
}

func Synth(text string, position uint, positionType EspeakPositionType, endPosition uint) int {
	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))
	return int(C.EspeakSynth(ctext, C.uint(position), C.espeak_POSITION_TYPE(positionType), C.uint(endPosition), C.uint(1)))
}

func SynthFlags(text string, position uint, positionType EspeakPositionType, endPosition uint, flags EspeakSynthFlag) int {
	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))

	// Ensure that one of the espeakCHARS flags is set. If not, set UTF8.
	if flags&0x7 == 0 {
		flags |= 1
	}

	return int(C.EspeakSynth(ctext, C.uint(position), C.espeak_POSITION_TYPE(positionType), C.uint(endPosition), C.uint(flags)))
}

func Save(textInput, wavOutput string) int {
	outfile := checkWav(wavOutput)
	if _, err := os.Stat(filepath.Dir(outfile)); os.IsNotExist(err) {
		return -1
	}

	C.wavefile = C.CString(outfile)
	defer C.free(unsafe.Pointer(C.wavefile))

	C.espeak_SetSynthCallback((*C.t_espeak_callback)(C.callback))

	ctext := C.CString(textInput)
	defer C.free(unsafe.Pointer(ctext))

	if x := int(C.EspeakSynth(ctext, C.uint(0), C.espeak_POSITION_TYPE(0), C.uint(0), C.uint(1))); x != 0 {
		return x
	}

	return int(C.espeak_Synchronize())
}

func Sync() int {
	return int(C.espeak_Synchronize())
}

type EspeakParameter int

const (
	SILENCE     EspeakParameter = 0
	RATE        EspeakParameter = 1
	VOLUME      EspeakParameter = 2
	PITCH       EspeakParameter = 3
	RANGE       EspeakParameter = 4
	PUNCTUATION EspeakParameter = 5
	CAPITALS    EspeakParameter = 6
	WORDGAP     EspeakParameter = 7
	OPTIONS     EspeakParameter = 8
	INTONATION  EspeakParameter = 9

	RESERVED1      EspeakParameter = 10
	RESERVED2      EspeakParameter = 11
	EMPHASIS       EspeakParameter = 12
	LINELENGTH     EspeakParameter = 13
	VOICETYPE      EspeakParameter = 14
	N_SPEECH_PARAM EspeakParameter = 15
)

func SetParameter(parameter EspeakParameter, value int, relative int) int {
	return int(C.espeak_SetParameter(C.espeak_PARAMETER(parameter), C.int(value), C.int(relative)))
}

func GetParameter(parameter EspeakParameter) int {
	return int(C.espeak_GetParameter(C.espeak_PARAMETER(parameter), 1))
}

func GetDefaultParameter(parameter EspeakParameter) int {
	return int(C.espeak_GetParameter(C.espeak_PARAMETER(parameter), 0))
}

func Cancel() int {
	return int(C.espeak_Cancel())
}

func Terminate() int {
	return int(C.espeak_Terminate())
}

func IsPlaying() int {
	return int(C.espeak_IsPlaying())
}

func checkWav(s string) string {
	s = strings.ToLower(s)
	for s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	if !strings.HasSuffix(s, ".wav") {
		s += ".wav"
	}
	return s
}
