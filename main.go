package main

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
		mod_close();
		exit(0);
	}

	term = true;

	info("terminated by signal %d\n", sig);

	ua_stop_all(false);
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
	"net"
	"time"
	"unsafe"
)

func Start(configPath, audioPath string) (err C.int) {

	ua := C.CString("baresip")
	defer C.free(unsafe.Pointer(ua))

	err = C.libre_init()
	if err != 0 {
		fmt.Printf("libre init failed with error code %d\n", err)
		return End(err)
	}

	if configPath != "" {
		cp := C.CString(configPath)
		defer C.free(unsafe.Pointer(cp))
		C.conf_path_set(cp)
	}

	err = C.conf_configure()
	if err != 0 {
		fmt.Printf("baresip configure failed with error code %d\n", err)
		return End(err)
	}

	// Top-level baresip struct init must be done AFTER configuration is complete.
	err = C.baresip_init(C.conf_config())
	if err != 0 {
		fmt.Printf("baresip main init failed with error code %d\n", err)
		return End(err)
	}

	if audioPath != "" {
		ap := C.CString(audioPath)
		defer C.free(unsafe.Pointer(ap))
		C.play_set_path(C.baresip_player(), ap)
	}

	err = C.ua_init(ua, 1, 1, 1)
	if err != 0 {
		fmt.Printf("baresip ua init failed with error code %d\n", err)
		return End(err)
	}

	err = C.conf_modules()
	if err != 0 {
		fmt.Printf("baresip load modules failed with error code %d\n", err)
		return End(err)
	}

	//C.sys_daemon()
	//C.uag_enable_sip_trace(1)
	//C.log_enable_debug(1)
	/*
		ua_eprm := C.CString("")
		defer C.free(unsafe.Pointer(ua_eprm))
		err = C.uag_set_extra_params(ua_eprm)
	*/

	err = C.mainLoop()
	if err != 0 {
		fmt.Printf("baresip main loop failed with error code %d\n", err)
		return End(err)
	}
	return err
}

func End(err C.int) C.int {

	C.ua_stop_all(1)

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

func main() {
	go StartCtrl()
	Start(".", "./sounds")
	End(0)
}

func StartCtrl() {
	time.Sleep(2 * time.Second)
	c, err := NewCtrl()
	if err != nil {
		return
	}

	go c.GetEvent(func(e EventMsg) {
		fmt.Println(e)
	})

	go c.GetResponse(func(r ResponseMsg) {
		fmt.Println(r)
	})
}

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

//Ctrl
type Ctrl struct {
	conn         net.Conn
	commandChan  chan string
	responseChan chan ResponseMsg
	eventChan    chan EventMsg
}

//NewCtrl
func NewCtrl() (*Ctrl, error) {
	c := &Ctrl{
		commandChan:  make(chan string, 100),
		responseChan: make(chan ResponseMsg, 100),
		eventChan:    make(chan EventMsg, 100),
	}

	var err error
	c.conn, err = connectCtrl("127.0.0.1:4444")
	if err != nil {
		return nil, err
	}

	go c.Read()
	return c, nil
}

func connectCtrl(a string) (net.Conn, error) {
	return net.DialTimeout("tcp", a, 10*time.Second)
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

func (c *Ctrl) Read() {
	defer c.conn.Close()
	scanner := bufio.NewScanner(c.conn)
	scanner.Split(eventSplitFunc)
	for {
		ok := scanner.Scan()
		if !ok {
			fmt.Printf("scanner end\n")
			break
		}

		msg := scanner.Bytes()

		if bytes.Contains(msg, []byte("\"event\":true")) {
			var e EventMsg
			err := json.Unmarshal(msg, &e)
			if err != nil {
				fmt.Println(err, string(msg))
			}
			c.eventChan <- e
		} else if bytes.Contains(msg, []byte("\"response\":true")) {
			var r ResponseMsg
			err := json.Unmarshal(bytes.Replace(msg, []byte("\\n"), []byte(""), -1), &r)
			if err != nil {
				fmt.Println(err, string(msg))
			}
			c.responseChan <- r
		}
	}

	if scanner.Err() != nil {
		fmt.Printf("scanner error: %s\n", scanner.Err())
	}
}

func Cmd(command, params, token string) *CommandMsg {
	return &CommandMsg{
		Command: command,
		Params:  params,
		Token:   token,
	}
}

func (c *Ctrl) Write(cmd *CommandMsg) error {
	msg, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	_, err = c.conn.Write([]byte(fmt.Sprintf("%d:%s,", len(msg), msg)))
	if err != nil {
		return err
	}

	return nil
}

func (c *Ctrl) GetEvent(get func(e EventMsg)) {
	for {
		get(<-c.eventChan)
	}
}

func (c *Ctrl) GetResponse(get func(r ResponseMsg)) {
	for {
		get(<-c.responseChan)
	}
}
