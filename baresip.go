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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"time"
	"unsafe"
)

//CommandMsg
type CommandMsg struct {
	Command string `json:"command,omitempty"`
	Params  string `json:"params,omitempty"`
	Token   string `json:"token,omitempty"`
}

//ResponseMsg
type ResponseMsg struct {
	Response bool   `json:"response,omitempty"`
	Ok       bool   `json:"ok,omitempty"`
	Data     string `json:"data,omitempty"`
	Token    string `json:"token,omitempty"`
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
}

type Baresip struct {
	addr         string
	configPath   string
	audioPath    string
	conn         net.Conn
	connAlive    uint32
	responseChan chan ResponseMsg
	eventChan    chan EventMsg
}

func New(addr, configPath, audioPath string) *Baresip {
	b := &Baresip{
		addr:         addr,
		configPath:   configPath,
		audioPath:    audioPath,
		responseChan: make(chan ResponseMsg, 100),
		eventChan:    make(chan EventMsg, 100),
	}

	go b.connectCtrl()

	return b
}

func (b *Baresip) connectCtrl() {
	var err error

	for i := 0; i < 10; i++ {
		b.conn, err = net.Dial("tcp", b.addr)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		log.Printf("Connection to %s established\n", b.addr)
		break
	}

	if b.conn == nil {
		atomic.StoreUint32(&b.connAlive, 0)
		log.Printf("can't connect to %s, exit!\n", b.addr)
		return
	}

	atomic.StoreUint32(&b.connAlive, 1)

	b.read()
}

func eventSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.Index(data, []byte("\"}")); i != -1 {
		if j := bytes.Index(data, []byte("{\"")); j != -1 {
			return i + 2, data[j : i+2], nil
		}
	}

	if atEOF {
		return len(data), data, nil
	}

	return
}

func (b *Baresip) read() {
	defer b.conn.Close()
	defer atomic.StoreUint32(&b.connAlive, 0)
	scanner := bufio.NewScanner(b.conn)
	scanner.Split(eventSplitFunc)
	for {
		ok := scanner.Scan()
		if !ok {
			log.Printf("scanner end\n")
			break
		}

		msg := scanner.Bytes()

		if bytes.Contains(msg, []byte("\"event\":true")) {
			var e EventMsg
			err := json.Unmarshal(msg, &e)
			if err != nil {
				log.Println(err, string(msg))
			}
			b.eventChan <- e
		} else if bytes.Contains(msg, []byte("\"response\":true")) {
			var r ResponseMsg
			err := json.Unmarshal(bytes.Replace(msg, []byte("\\n"), []byte(""), -1), &r)
			if err != nil {
				log.Println(err, string(msg))
			}
			b.responseChan <- r
		}
	}

	if scanner.Err() != nil {
		log.Printf("scanner error: %s\n", scanner.Err())
	}
}

func cmd(command, params, token string) *CommandMsg {
	return &CommandMsg{
		Command: command,
		Params:  params,
		Token:   token,
	}
}

// Exec sends a command over ctrl_tcp to baresip
func (b *Baresip) Exec(command, params, token string) error {
	msg, err := json.Marshal(cmd(command, params, token))
	if err != nil {
		return err
	}

	if atomic.LoadUint32(&b.connAlive) == 0 {
		return fmt.Errorf("can't write to closed tcp_ctrl connection")
	}

	_, err = b.conn.Write([]byte(fmt.Sprintf("%d:%s,", len(msg), msg)))
	if err != nil {
		return err
	}

	return nil
}

func (b *Baresip) Close() {
	close(b.responseChan)
	close(b.eventChan)
}

func (b *Baresip) handleEvent(get func(e EventMsg)) {
	go func() {
		for {
			get(<-b.eventChan)
		}
	}()
}

func (b *Baresip) handleResponse(get func(r ResponseMsg)) {
	go func() {
		for {
			get(<-b.responseChan)
		}
	}()
}

// ReadChan returns the receive-only EventMsg channel for reading data
func (b *Baresip) EventChan() <-chan EventMsg {
	return b.eventChan
}

// ResponseChan returns the receive-only ResponseMsg channel for reading data
func (b *Baresip) ResponseChan() <-chan ResponseMsg {
	return b.responseChan
}

// Run a baresip instance
func (b *Baresip) Run() (err C.int) {

	ua := C.CString("go-baresip")
	defer C.free(unsafe.Pointer(ua))

	err = C.libre_init()
	if err != 0 {
		log.Printf("libre init failed with error code %d\n", err)
		return b.end(err)
	}

	C.log_enable_stdout(0)

	cp := C.CString(b.configPath)
	defer C.free(unsafe.Pointer(cp))
	C.conf_path_set(cp)

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

	ap := C.CString(b.audioPath)
	defer C.free(unsafe.Pointer(ap))
	C.play_set_path(C.baresip_player(), ap)

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

	ct := C.CString("ctrl_tcp")
	defer C.free(unsafe.Pointer(ct))
	C.module_load(cp, ct)

	//C.sys_daemon()
	//C.uag_enable_sip_trace(1)
	//C.log_enable_debug(1)
	/*
		ua_eprm := C.CString("")
		defer C.free(unsafe.Pointer(ua_eprm))
		err = C.uag_set_extra_params(ua_eprm)
	*/

	return b.end(C.mainLoop())
}

func (b *Baresip) end(err C.int) C.int {
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

	return err
}
