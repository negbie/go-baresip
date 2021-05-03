package gobaresip

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"time"
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

type Ctrl struct {
	conn         net.Conn
	responseChan chan ResponseMsg
	eventChan    chan EventMsg
}

func NewCtrl(addr string) *Ctrl {
	c := &Ctrl{
		responseChan: make(chan ResponseMsg, 100),
		eventChan:    make(chan EventMsg, 100),
	}

	go c.connectCtrl(addr)

	return c
}

func (c *Ctrl) connectCtrl(addr string) {
	var err error
	b := &Backoff{
		Max: 1 * time.Minute,
	}

	for {
		c.conn, err = net.Dial("tcp", addr)
		if err != nil {
			d := b.Duration()

			if d.Seconds() > 10 {
				fmt.Printf("can't connect to %s, exit!\n", addr)
				return
			}
			fmt.Printf("%s, reconnecting in %s\n", err, d)
			time.Sleep(d)
			continue
		}
		fmt.Printf("Connection to %s established\n", addr)
		break
	}

	c.Read()
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

func cmd(command, params, token string) *CommandMsg {
	return &CommandMsg{
		Command: command,
		Params:  params,
		Token:   token,
	}
}

func (c *Ctrl) Exec(command, params, token string) error {
	msg, err := json.Marshal(cmd(command, params, token))
	if err != nil {
		return err
	}

	_, err = c.conn.Write([]byte(fmt.Sprintf("%d:%s,", len(msg), msg)))
	if err != nil {
		return err
	}

	return nil
}

func (c *Ctrl) Close() {
	close(c.responseChan)
	close(c.eventChan)
}

func (c *Ctrl) HandleEvent(get func(e EventMsg)) {
	go func() {
		for {
			get(<-c.eventChan)
		}
	}()
}

func (c *Ctrl) HandleResponse(get func(r ResponseMsg)) {
	go func() {
		for {
			get(<-c.responseChan)
		}
	}()
}
