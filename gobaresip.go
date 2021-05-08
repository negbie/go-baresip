package gobaresip

/*
#cgo linux CFLAGS: -I.
#cgo linux LDFLAGS: ${SRCDIR}/libbaresip/baresip/libbaresip.a ${SRCDIR}/libbaresip/re/libre.a ${SRCDIR}/libbaresip/rem/librem.a -ldl -lm -lcrypto -lssl -lz

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
	"strings"
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
	Raw      string
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
	Level           string
	Raw             string
}

type Baresip struct {
	userAgent    string
	ctrlAddr     string
	configPath   string
	audioPath    string
	debug        bool
	conn         net.Conn
	connAlive    uint32
	responseChan chan ResponseMsg
	eventChan    chan EventMsg
	ctrlStream   *reader
}

func New(options ...func(*Baresip) error) (*Baresip, error) {
	b := &Baresip{
		ctrlAddr:     "127.0.0.1:4444",
		responseChan: make(chan ResponseMsg, 100),
		eventChan:    make(chan EventMsg, 100),
	}

	if err := b.SetOption(options...); err != nil {
		return nil, err
	}

	if b.userAgent == "" {
		b.userAgent = "go-baresip"
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
	b.conn, err = net.Dial("tcp", b.ctrlAddr)
	if err != nil {
		atomic.StoreUint32(&b.connAlive, 0)
		return fmt.Errorf("%v: please make sure ctrl_tcp is enabled", err)
	}

	b.ctrlStream = newReader(b.conn)

	atomic.StoreUint32(&b.connAlive, 1)
	return nil
}

func (b *Baresip) read() {
	for {
		msg, err := b.ctrlStream.readNetstring()
		if err != nil {
			log.Println(err)
			return
		}

		if atomic.LoadUint32(&b.connAlive) == 0 {
			return
		}

		if bytes.Contains(msg, []byte("\"event\":true")) {
			var e EventMsg
			e.Raw = string(msg)
			err := json.Unmarshal(msg, &e)
			if err != nil {
				log.Println(err, string(msg))
			}

			cc := e.Type == "CALL_CLOSED"
			e.Level = "info"

			if cc && e.ID == "" {
				e.Level = "warning"
			} else if cc && strings.HasPrefix(e.Param, "4") {
				e.Level = "warning"
			} else if cc && strings.HasPrefix(e.Param, "5") {
				e.Level = "error"
			} else if cc && strings.HasPrefix(e.Param, "6") {
				e.Level = "error"
			} else if strings.Contains(e.Type, "FAIL") || strings.Contains(e.Type, "ERROR") {
				e.Level = "warning"
			}

			b.eventChan <- e
		} else if bytes.Contains(msg, []byte("\"response\":true")) {
			if bytes.Contains(msg, []byte("keep_active_ping")) {
				continue
			}

			var r ResponseMsg
			r.Raw = string(msg)
			err := json.Unmarshal(bytes.Replace(msg, []byte("\\n"), []byte(""), -1), &r)
			if err != nil {
				log.Println(err, string(msg))
			}
			b.responseChan <- r
		}
	}
}

func (b *Baresip) Close() {
	atomic.StoreUint32(&b.connAlive, 0)
	if b.conn != nil {
		b.conn.Close()
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
		time.Sleep(500 * time.Millisecond)
		b.Command("listcalls", "", "keep_active_ping")
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
