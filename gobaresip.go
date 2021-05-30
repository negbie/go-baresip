package gobaresip

/*
#cgo linux LDFLAGS: ${SRCDIR}/libbaresip/baresip/libbaresip.a
#cgo linux LDFLAGS: ${SRCDIR}/libbaresip/re/libre.a
#cgo linux LDFLAGS: ${SRCDIR}/libbaresip/rem/librem.a
#cgo linux LDFLAGS: ${SRCDIR}/libbaresip/opus/libopus.a
#cgo linux LDFLAGS: ${SRCDIR}/libbaresip/openssl/libssl.a
#cgo linux LDFLAGS: ${SRCDIR}/libbaresip/openssl/libcrypto.a
#cgo linux LDFLAGS: -ldl -lm

#include <stdint.h>
#include <stdlib.h>
#include <libbaresip/re/include/re.h>
#include <libbaresip/rem/include/rem.h>
#include <libbaresip/baresip/include/baresip.h>

static void signal_handler(int sig)
{
	static bool term = false;

	if (term) {
		module_app_unload();
		mod_close();
		exit(0);
	}

	term = true;

	info("terminated by signal %d\n", sig);

	ua_stop_all(false);
}

static void net_change_handler(void *arg)
{
	(void)arg;

	info("IP-address changed: %j\n",
	     net_laddr_af(baresip_network(), AF_INET));

	(void)uag_reset_transp(true, true);
}

static void set_net_change_handler()
{
	net_change(baresip_network(), 60, net_change_handler, NULL);
}

static void ua_exit_handler(void *arg)
{
	(void)arg;
	debug("ua exited -- stopping main runloop\n");

	//The main run-loop can be stopped now
	re_cancel();
}

static void set_ua_exit_handler()
{
	uag_set_exit_handler(ua_exit_handler, NULL);
}

int mainLoop(){
	return re_main(signal_handler);
}
*/
import "C"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

//ResponseMsg
type ResponseMsg struct {
	Response bool   `json:"response,omitempty"`
	Ok       bool   `json:"ok,omitempty"`
	Data     string `json:"data,omitempty"`
	Token    string `json:"token,omitempty"`
	RawJSON  []byte `json:"-"`
}

//EventMsg
type EventMsg struct {
	Event           bool   `json:"event,omitempty"`
	Type            string `json:"type,omitempty"`
	Class           string `json:"class,omitempty"`
	AccountAOR      string `json:"accountaor,omitempty"`
	Direction       string `json:"direction,omitempty"`
	PeerURI         string `json:"peeruri,omitempty"`
	PeerDisplayname string `json:"peerdisplayname,omitempty"`
	ID              string `json:"id,omitempty"`
	RemoteAudioDir  string `json:"remoteaudiodir,omitempty"`
	Param           string `json:"param,omitempty"`
	RawJSON         []byte `json:"-"`
}

type Baresip struct {
	userAgent      string
	ctrlAddr       string
	wsAddr         string
	configPath     string
	audioPath      string
	debug          bool
	ctrlConn       net.Conn
	ctrlConnAlive  uint32
	responseChan   chan ResponseMsg
	eventChan      chan EventMsg
	responseWsChan chan []byte
	eventWsChan    chan []byte
	ctrlStream     *reader
	autoCmd        ac
}

type ac struct {
	mux sync.RWMutex
	num map[string]int

	hangupGap uint32
}

func New(options ...func(*Baresip) error) (*Baresip, error) {
	b := &Baresip{
		responseChan: make(chan ResponseMsg, 100),
		eventChan:    make(chan EventMsg, 100),
	}

	if err := b.SetOption(options...); err != nil {
		return nil, err
	}

	if b.audioPath == "" {
		b.audioPath = "."
	}
	if b.configPath == "" {
		b.configPath = "."
	}
	if b.ctrlAddr == "" {
		b.ctrlAddr = "127.0.0.1:4444"
	}
	if b.userAgent == "" {
		b.userAgent = "go-baresip"
	}

	b.autoCmd.num = make(map[string]int)

	if b.wsAddr != "" {
		b.responseWsChan = make(chan []byte, 100)
		b.eventWsChan = make(chan []byte, 100)

		h := newWsHub(b)
		go h.run()

		http.HandleFunc("/", serveRoot)
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			serveWs(h, w, r)
		})
		go http.ListenAndServe(b.wsAddr, nil)
	}

	if err := b.setup(); err != nil {
		return nil, err
	}

	// Simple solution for this https://github.com/baresip/baresip/issues/584
	go b.keepActive()

	return b, nil
}

func (b *Baresip) connectCtrl() error {
	var err error
	b.ctrlConn, err = net.Dial("tcp", b.ctrlAddr)
	if err != nil {
		atomic.StoreUint32(&b.ctrlConnAlive, 0)
		return fmt.Errorf("%v: please make sure ctrl_tcp is enabled", err)
	}

	b.ctrlStream = newReader(b.ctrlConn)

	atomic.StoreUint32(&b.ctrlConnAlive, 1)
	return nil
}

func (b *Baresip) read() {
	for {
		if atomic.LoadUint32(&b.ctrlConnAlive) == 0 {
			break
		}

		msg, err := b.ctrlStream.readNetstring()
		if err != nil {
			log.Println(err)
			break
		}

		if bytes.Contains(msg, []byte("\"event\":true")) {
			if bytes.Contains(msg, []byte(",end of file")) {
				msg = bytes.Replace(msg, []byte("AUDIO_ERROR"), []byte("AUDIO_EOF"), 1)
			}

			var e EventMsg
			e.RawJSON = msg

			err := json.Unmarshal(e.RawJSON, &e)
			if err != nil {
				log.Println(err, string(e.RawJSON))
				continue
			}

			b.eventChan <- e
			if b.wsAddr != "" {
				select {
				case b.eventWsChan <- appendByte(e.RawJSON, []byte("\n")):
				default:
				}
			}
		} else if bytes.Contains(msg, []byte("\"response\":true")) {
			if bytes.Contains(msg, []byte("keep_active_ping")) {
				continue
			}

			var r ResponseMsg
			r.RawJSON = msg

			err := json.Unmarshal(r.RawJSON, &r)
			if err != nil {
				log.Println(err, string(r.RawJSON))
				continue
			}

			if strings.HasPrefix(r.Token, "cmd_dial") {
				if d := atomic.LoadUint32(&b.autoCmd.hangupGap); d > 0 {
					if id := findID([]byte(r.Data)); len(id) > 1 {
						go func() {
							time.Sleep(time.Duration(d) * time.Second)
							b.CmdHangupID(id)
						}()
					}
				}
			}

			if strings.HasPrefix(r.Token, "cmd_auto") {
				r.Ok = true
				b.autoCmd.mux.RLock()
				r.Data = fmt.Sprintf("dial%v;hangupgap=%d",
					b.autoCmd.num,
					atomic.LoadUint32(&b.autoCmd.hangupGap),
				)
				b.autoCmd.mux.RUnlock()
				r.Data = strings.Replace(r.Data, " ", ",", -1)
				r.Data = strings.Replace(r.Data, ":", ";autodialgap=", -1)
				rj, err := json.Marshal(r)
				if err != nil {
					log.Println(err, r.Data)
					continue
				}
				r.RawJSON = rj
			}

			b.responseChan <- r
			if b.wsAddr != "" {
				select {
				case b.responseWsChan <- appendByte(r.RawJSON, []byte("\n")):
				default:
				}
			}
		}
	}
}

func appendByte(prefix []byte, suffix []byte) []byte {
	b := make([]byte, len(prefix)+len(suffix))
	n := copy(b, prefix)
	copy(b[n:], suffix)
	return b
}

func findID(data []byte) string {
	if posA := bytes.Index(data, []byte("call id: ")); posA > 0 {
		if posB := bytes.Index(data[posA:], []byte("\n")); posB > 0 {
			l := len("call id: ")
			return string(data[posA+l : posA+posB])
		}
	}
	return ""
}

func (b *Baresip) Close() {
	atomic.StoreUint32(&b.ctrlConnAlive, 0)
	if b.ctrlConn != nil {
		b.ctrlConn.Close()
	}
	close(b.responseChan)
	close(b.eventChan)
}

// GetEventChan returns the receive-only EventMsg channel for reading data.
func (b *Baresip) GetEventChan() <-chan EventMsg {
	return b.eventChan
}

// GetResponseChan returns the receive-only ResponseMsg channel for reading data.
func (b *Baresip) GetResponseChan() <-chan ResponseMsg {
	return b.responseChan
}

func (b *Baresip) keepActive() {
	for {
		if atomic.LoadUint32(&b.ctrlConnAlive) == 0 {
			break
		}
		time.Sleep(1 * time.Second)
		b.Cmd("about", "", "keep_active_ping")
	}
}

// setup a baresip instance
func (b *Baresip) setup() error {

	ua := C.CString(b.userAgent)
	defer C.free(unsafe.Pointer(ua))

	C.sys_coredump_set(1)

	err := C.libre_init()
	if err != 0 {
		log.Printf("libre init failed with error code %d\n", err)
		return b.end(err)
	}

	if b.debug {
		C.log_enable_stdout(1)
	} else {
		C.log_enable_stdout(0)
	}

	if b.configPath != "" {
		cp := C.CString(b.configPath)
		defer C.free(unsafe.Pointer(cp))
		C.conf_path_set(cp)
	}

	err = C.conf_configure()
	if err != 0 {
		log.Printf("baresip configure failed with error code %d\n", err)
		return b.end(err)
	}

	// Top-level baresip struct init must be done AFTER configuration is complete.
	err = C.baresip_init(C.conf_config())
	if err != 0 {
		log.Printf("baresip main init failed with error code %d\n", err)
		return b.end(err)
	}

	if b.audioPath != "" {
		ap := C.CString(b.audioPath)
		defer C.free(unsafe.Pointer(ap))
		C.play_set_path(C.baresip_player(), ap)
	}

	err = C.ua_init(ua, 1, 1, 1)
	if err != 0 {
		log.Printf("baresip ua init failed with error code %d\n", err)
		return b.end(err)
	}

	C.set_net_change_handler()
	C.set_ua_exit_handler()

	err = C.conf_modules()
	if err != 0 {
		log.Printf("baresip load modules failed with error code %d\n", err)
		return b.end(err)
	}

	if b.debug {
		C.log_enable_debug(1)
		C.uag_enable_sip_trace(1)
	} else {
		C.log_enable_debug(0)
		C.uag_enable_sip_trace(0)
	}

	/*
		ua_eprm := C.CString("")
		defer C.free(unsafe.Pointer(ua_eprm))
		err = C.uag_set_extra_params(ua_eprm)
	*/

	if err := b.connectCtrl(); err != nil {
		b.end(1)
		return err
	}

	return nil
}

// Run a baresip instance
func (b *Baresip) Run() error {
	go b.read()
	err := b.end(C.mainLoop())
	if err.Error() == "0" {
		return nil
	}
	return err
}

func (b *Baresip) end(err C.int) error {
	if err != 0 {
		C.ua_stop_all(1)
	}

	C.ua_close()
	C.module_app_unload()
	C.conf_close()

	C.baresip_close()

	// Modules must be unloaded after all application activity has stopped.
	C.mod_close()

	C.libre_close()

	// Check for memory leaks
	C.tmr_debug()
	C.mem_debug()

	return fmt.Errorf("%d", err)
}
