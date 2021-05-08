package espeak

/*
#cgo linux CFLAGS: -I.
#cgo linux LDFLAGS: ${SRCDIR}/libespeak.a -lm

#include <stdlib.h>
#include <string.h>
#include <espeak.h>


struct findNextLanguage_ret
{
	const char *start;
	int len;
	const char *next;
};

static inline struct findNextLanguage_ret findNextLanguage(const char *data)
{
	struct findNextLanguage_ret ret;
	data++;
	ret.start = data;
	ret.len = 0;
	while (*data)
	{
		data++;
		ret.len++;
	}
	data++;
	ret.next = data;
	return ret;
}

static inline espeak_EVENT_TYPE eventType(const espeak_EVENT *event)
{
	return event->type;
}

struct eventID_ret
{
	int number;        // used for WORD and SENTENCE events.
	const char *name;  // used for MARK and PLAY events.  UTF8 string
	char string[8];    // used for phoneme names (UTF8). Terminated by a zero byte unless the name needs the full 8 bytes.
};

static inline struct eventID_ret eventID(const espeak_EVENT *event)
{
	struct eventID_ret ret = { 0 };
	switch (event->type)
	{
	case espeakEVENT_WORD:
	case espeakEVENT_SENTENCE:
		ret.number = event->id.number;
		break;
	case espeakEVENT_MARK:
	case espeakEVENT_PLAY:
		ret.name = event->id.name;
		break;
	case espeakEVENT_PHONEME:
		memcpy(ret.string, event->id.string, sizeof(ret.string));
		break;
	default:
		break;
	}
	return ret;
}

extern int synthCallback(short *wav, int numsamples, espeak_EVENT *events);

*/
import "C"
import (
	"time"
	"unsafe"
)

func init() {
	//lock.Lock()
	//defer lock.Unlock()

	C.espeak_ng_InitializePath(nil)

	var errCtx C.espeak_ng_ERROR_CONTEXT
	defer C.espeak_ng_ClearErrorContext(&errCtx)

	status := C.espeak_ng_Initialize(&errCtx)
	err := toErr(status)
	if err != nil {
		C.espeak_ng_PrintStatusCodeMessage(status, C.stderr, errCtx)
		panic(err)
	}

	status = C.espeak_ng_InitializeOutput(C.ENOUTPUT_MODE_SYNCHRONOUS, 0, nil)
	err = toErr(status)
	if err != nil {
		C.espeak_ng_PrintStatusCodeMessage(status, C.stderr, errCtx)
		panic(err)
	}

	C.espeak_SetSynthCallback((*C.t_espeak_callback)(C.synthCallback))
}

func toErr(status C.espeak_ng_STATUS) error {
	var errBuf [512]C.char

	if status == C.ENS_OK {
		return nil
	}

	C.espeak_ng_GetStatusCodeMessage(status, &errBuf[0], C.size_t(len(errBuf)))
	return &Error{
		Code:    uint32(status),
		Message: C.GoStringN(&errBuf[0], C.int(len(errBuf))),
	}
}

func getSampleRate() int {
	return int(C.espeak_ng_GetSampleRate())
}

func listVoices() []*Voice {
	var voices []*Voice

	for cVoices := C.espeak_ListVoices(nil); *cVoices != nil; cVoices = nextVoice(cVoices) {
		voices = append(voices, toVoice(*cVoices))
	}

	return voices
}

func nextVoice(p0 **C.espeak_VOICE) **C.espeak_VOICE {
	p1 := unsafe.Pointer(p0)
	p2 := uintptr(p1)
	p3 := p2 + unsafe.Sizeof((*C.espeak_VOICE)(nil))
	p4 := unsafe.Pointer(p3)
	return (**C.espeak_VOICE)(p4)
}

func toVoice(cVoice *C.espeak_VOICE) *Voice {
	return &Voice{
		Name:       C.GoString(cVoice.name),
		Languages:  toLanguages(cVoice.languages),
		Identifier: C.GoString(cVoice.identifier),
		Gender:     Gender(cVoice.gender),
		Age:        uint8(cVoice.age),
	}
}

func toLanguages(data *C.char) []Language {
	var languages []Language

	for {
		priority := uint8(*data)
		if priority == 0 {
			return languages
		}

		ret := C.findNextLanguage(data)
		languages = append(languages, Language{
			Priority: priority,
			Name:     C.GoStringN(ret.start, ret.len),
		})
		data = ret.next
	}
}

func setRate(rate int) error {
	return toErr(C.espeak_ng_SetParameter(C.espeakRATE, C.int(rate), 0))
}

func setVolume(volume int) error {
	return toErr(C.espeak_ng_SetParameter(C.espeakVOLUME, C.int(volume), 0))
}

func setPitch(pitch int) error {
	return toErr(C.espeak_ng_SetParameter(C.espeakPITCH, C.int(pitch), 0))
}

func setTone(tone int) error {
	return toErr(C.espeak_ng_SetParameter(C.espeakRANGE, C.int(tone), 0))
}

func setVoice(name, language string, gender Gender, age, variant uint8) error {
	var voice C.espeak_VOICE
	if name == "" {
		voice.name = nil
	} else {
		cName := C.CString(name)
		defer C.free(unsafe.Pointer(cName))
		voice.name = cName
	}

	if name != "" && language == "" && gender == Unknown && age == 0 && variant == 0 {
		return toErr(C.espeak_ng_SetVoiceByName(voice.name))
	}

	if language == "" {
		voice.languages = nil
	} else {
		cLanguage := C.CString(language)
		defer C.free(unsafe.Pointer(cLanguage))
		voice.languages = cLanguage
	}

	voice.gender = C.uchar(gender)
	voice.age = C.uchar(age)
	voice.variant = C.uchar(variant)

	return toErr(C.espeak_ng_SetVoiceByProperties(&voice))
}

var synthCtx *Espeak

//export synthCallback
func synthCallback(wav *C.short, numsamples C.int, events *C.espeak_EVENT) C.int {
	for i := C.int(0); i < numsamples; i++ {
		pwav := uintptr(unsafe.Pointer(wav)) + uintptr(i)*unsafe.Sizeof(C.short(0))
		synthCtx.Samples = append(synthCtx.Samples, int16(*(*C.short)(unsafe.Pointer(pwav))))
	}

	for C.eventType(events) != C.espeakEVENT_LIST_TERMINATED {
		if e := toEvent(events); e != nil {
			synthCtx.Events = append(synthCtx.Events, e)
		}

		events = (*C.espeak_EVENT)(unsafe.Pointer(uintptr(unsafe.Pointer(events)) + unsafe.Sizeof(C.espeak_EVENT{})))
	}

	return 0 // continue synthesis
}

func toEvent(event *C.espeak_EVENT) *SynthEvent {
	var synthEvent SynthEvent

	id := C.eventID(event)

	switch C.eventType(event) {
	case C.espeakEVENT_WORD:
		synthEvent.Type = EventWord
		synthEvent.Number = int(id.number)
	case C.espeakEVENT_SENTENCE:
		synthEvent.Type = EventSentence
		synthEvent.Number = int(id.number)
	case C.espeakEVENT_MARK:
		synthEvent.Type = EventMark
		synthEvent.Name = C.GoString(id.name)
	case C.espeakEVENT_PLAY:
		synthEvent.Type = EventPlay
		synthEvent.Name = C.GoString(id.name)
	case C.espeakEVENT_END:
		synthEvent.Type = EventEnd
	case C.espeakEVENT_MSG_TERMINATED:
		synthEvent.Type = EventMsgTerminated
	case C.espeakEVENT_PHONEME:
		synthEvent.Type = EventPhoneme
		synthEvent.Phoneme = C.GoStringN(&id.string[0], C.int(len(id.string)))
	default:
		return nil
	}

	synthEvent.TextPosition = int(event.text_position)
	synthEvent.Length = int(event.length)
	synthEvent.AudioPosition = time.Duration(event.audio_position) * time.Millisecond

	return &synthEvent
}

func synthesize(text string, e *Espeak) error {
	synthCtx = e
	defer func() {
		synthCtx = nil
	}()

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	return toErr(C.espeak_ng_Synthesize(unsafe.Pointer(cText), 0, 0, C.POS_CHARACTER, 0, C.espeakCHARS_UTF8|C.espeakSSML, nil, nil))
}
