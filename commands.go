package gobaresip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

/*
  /about                           About box
  /accept                 a        Accept incoming call
  /acceptdir ..                    Accept incoming call with audio and videodirection.
  /answermode ..                   Set answer mode
  /aubitrate ..                    Set audio bitrate
  /audio_debug            A        Audio stream
  /auplay ..                       Switch audio player
  /ausrc ..                        Switch audio source
  /autodial ..                     Set auto dial command
  /autodialcancel                  Cancel auto dial
  /autodialdelay ..                Set delay before auto dial [ms]
  /autohangup ..                   Set auto hangup command
  /autohangupcancel                Cancel auto hangup
  /autohangupdelay ..              Set delay before hangup [ms]
  /autostat                        Print autotest status
  /callfind ..                     Find call
  /callstat               c        Call status
  /contact_next           >        Set next contact
  /contact_prev           <        Set previous contact
  /contacts               C        List contacts
  /dial ..                d ..     Dial
  /dialcontact            D        Dial current contact
  /dialdir ..                      Dial with audio and videodirection.
  /dnd ..                          Set Do not Disturb
  /hangup                 b        Hangup call
  /hangupall ..                    Hangup all calls with direction
  /help                   h        Help menu
  /hold                   x        Call hold
  /insmod ..                       Load module
  /line ..                @ ..     Set current call
  /listcalls              l        List active calls
  /medialdir ..                    Set local media direction
  /message ..             M ..     Message current contact
  /mute                   m        Call mute/un-mute
  /options ..             o ..     Options
  /quit                   q        Quit
  /reginfo                r        Registration info
  /reinvite               I        Send re-INVITE
  /resume                 X        Call resume
  /rmmod ..                        Unload module
  /setadelay ..                    Set answer delay for outgoing call
  /sndcode ..                      Send Code
  /statmode               S        Statusmode toggle
  /tlsissuer                       TLS certificate issuer
  /tlssubject                      TLS certificate subject
  /transfer ..            t ..     Transfer call
  /uadel ..                        Delete User-Agent
  /uadelall ..                     Delete all User-Agents
  /uafind ..                       Find User-Agent
  /uanew ..                        Create User-Agent
  /uareg ..                        UA register  [index]
  /video_debug            V        Video stream
  /videodir ..                     Set video direction
  /vidsrc ..                       Switch video source
*/

//CommandMsg struct for ctrl_tcp
type CommandMsg struct {
	Command string `json:"command,omitempty"`
	Params  string `json:"params,omitempty"`
	Token   string `json:"token,omitempty"`
}

func buildCommand(command, params, token string) *CommandMsg {
	return &CommandMsg{
		Command: command,
		Params:  params,
		Token:   token,
	}
}

// Cmd will send a raw baresip command over ctrl_tcp.
func (b *Baresip) Cmd(command, params, token string) error {
	msg, err := json.Marshal(buildCommand(command, params, token))
	if err != nil {
		return err
	}

	if atomic.LoadUint32(&b.ctrlConnAlive) == 0 {
		return fmt.Errorf("can't write command to closed tcp_ctrl connection")
	}

	b.ctrlConn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err = b.ctrlConn.Write([]byte(fmt.Sprintf("%d:%s,", len(msg), msg)))
	if err != nil {
		return err
	}

	return nil
}

// CmdAccept will accept incoming call
func (b *Baresip) CmdAccept() error {
	c := "accept"
	return b.Cmd(c, "", "cmd_"+c)
}

// CmdAcceptdir will accept incoming call with audio and videodirection.
func (b *Baresip) CmdAcceptdir(s string) error {
	c := "acceptdir"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdAnswermode will set answer mode
func (b *Baresip) CmdAnswermode(s string) error {
	c := "answermode"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdAuplay will switch audio player
func (b *Baresip) CmdAuplay(s string) error {
	c := "auplay"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdAusrc will switch audio source
func (b *Baresip) CmdAusrc(s string) error {
	c := "ausrc"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdCallstat will show call status
func (b *Baresip) CmdCallstat() error {
	c := "callstat"
	return b.Cmd(c, "", "cmd_"+c)
}

// CmdContact_next will set next contact
func (b *Baresip) CmdContact_next() error {
	c := "contact_next"
	return b.Cmd(c, "", "cmd_"+c)
}

// CmdContact_prev will set previous contact
func (b *Baresip) CmdContact_prev() error {
	c := "contact_prev"
	return b.Cmd(c, "", "cmd_"+c)
}

// CmdAutodial will dial number automatically
func (b *Baresip) CmdAutodial(s string) error {
	c := "autodial dial"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdAutodialdelay will set delay before auto dial [ms]
func (b *Baresip) CmdAutodialdelay(n int) error {
	c := "autodialdelay"
	return b.Cmd(c, strconv.Itoa(n), "cmd_"+c+"_"+strconv.Itoa(n))
}

// CmdDial will dial number
func (b *Baresip) CmdDial(s string) error {
	c := "dial"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdDialcontact will dial current contact
func (b *Baresip) CmdDialcontact() error {
	c := "dialcontact"
	return b.Cmd(c, "", "cmd_"+c)
}

// CmdDialdir will dial with audio and videodirection
func (b *Baresip) CmdDialdir(s string) error {
	c := "dialdir"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdAutohangup will hangup call automatically
func (b *Baresip) CmdAutohangup() error {
	c := "autohangup"
	return b.Cmd(c, "hangup", "cmd_"+c)
}

// CmdAutohangupdelay will set delay before hangup [ms]
func (b *Baresip) CmdAutohangupdelay(n int) error {
	c := "autohangupdelay"
	return b.Cmd(c, strconv.Itoa(n), "cmd_"+c+"_"+strconv.Itoa(n))
}

// CmdHangup will hangup call
func (b *Baresip) CmdHangup() error {
	c := "hangup"
	return b.Cmd(c, "", "cmd_"+c)
}

// CmdHangupID will hangup call with Call-ID
func (b *Baresip) CmdHangupID(callID string) error {
	c := "hangup"
	return b.Cmd(c, callID, "cmd_"+c)
}

// CmdHangupall will hangup all calls with direction
func (b *Baresip) CmdHangupall(s string) error {
	c := "hangupall"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdInsmod will load module
func (b *Baresip) CmdInsmod(s string) error {
	c := "insmod"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdListcalls will list active calls
func (b *Baresip) CmdListcalls() error {
	c := "listcalls"
	return b.Cmd(c, "", "cmd_"+c)
}

// CmdReginfo will list registration info
func (b *Baresip) CmdReginfo() error {
	c := "reginfo"
	return b.Cmd(c, "", "cmd_"+c)
}

// CmdRmmod will unload module
func (b *Baresip) CmdRmmod(s string) error {
	c := "rmmod"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdSetadelay will set answer delay for outgoing call
func (b *Baresip) CmdSetadelay(n int) error {
	c := "setadelay"
	return b.Cmd(c, strconv.Itoa(n), "cmd_"+c+"_"+strconv.Itoa(n))
}

// CmdUadel will delete User-Agent
func (b *Baresip) CmdUadel(s string) error {
	c := "uadel"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdUadelall will delete all User-Agents
func (b *Baresip) CmdUadelall() error {
	c := "uadelall"
	return b.Cmd(c, "", "cmd_"+c)
}

// CmdUafind will find User-Agent <aor>
func (b *Baresip) CmdUafind(s string) error {
	c := "uafind"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdUanew will create User-Agent
func (b *Baresip) CmdUanew(s string) error {
	c := "uanew"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdUareg will register <regint> [index]
func (b *Baresip) CmdUareg(s string) error {
	c := "uareg"
	return b.Cmd(c, s, "cmd_"+c+"_"+s)
}

// CmdQuit will quit baresip
func (b *Baresip) CmdQuit() error {
	c := "quit"
	return b.Cmd(c, "", "cmd_"+c)
}

func (b *Baresip) CmdWs(raw []byte) error {
	m := strings.Split(string(bytes.TrimSpace(raw)), " ")
	if len(m) < 1 {
		return nil
	}

	m[0] = strings.ToLower(m[0])
	if m[0] == "quit" {
		return nil
	} else if m[0] == "line" {
		return nil
	} else if m[0] == "uadelall" {
		return nil
	}

	if len(m) == 2 && m[0] == "repeatdial" {
		b.CmdRepeatDial(m[1])
	} else if len(m) == 2 && m[0] == "repeatdialinterval" {
		if n, err := strconv.Atoi(m[1]); err == nil {
			b.CmdRepeatDialInterval(n)
		}
	} else if m[0] == "repeatdialcancel" {
		b.CmdRepeatDialCancel()
	} else if m[0] == "repeatdialinfo" {
		b.CmdRepeatDialInfo()
	} else if len(m) == 1 {
		b.Cmd(m[0], "", "cmd_"+m[0])
	} else if len(m) == 2 {
		b.Cmd(m[0], m[1], "cmd_"+m[0])
	}
	return nil
}

func (b *Baresip) CmdRepeatDial(s string) error {
	b.autotest.uris += s + ","
	uris := strings.Split(s, ",")
	atomic.StoreUint32(&b.autotest.cancel, 0)

	go func(u []string) {
		for {
			for _, v := range u {
				if atomic.LoadUint32(&b.autotest.cancel) == 1 {
					return
				}
				b.CmdDial(strings.TrimSpace(v))
			}
			d := int(atomic.LoadUint32(&b.autotest.interval))
			time.Sleep(time.Duration(d) * time.Second)
		}
	}(uris)

	return b.Cmd("repeatdialinfo", "", "cmd_repeatdial")
}

func (b *Baresip) CmdRepeatDialInfo() error {
	return b.Cmd("repeatdialinfo", "", "cmd_repeatdialinfo")
}

func (b *Baresip) CmdRepeatDialInterval(n int) error {
	if n < 5 {
		n = 5
	}
	atomic.StoreUint32(&b.autotest.interval, uint32(n))
	return b.Cmd("repeatdialinfo", "", "cmd_repeatdialinterval")
}

func (b *Baresip) CmdRepeatDialCancel() error {
	atomic.StoreUint32(&b.autotest.cancel, 1)
	b.autotest.uris = ""
	b.CmdHangupall("all")
	return b.Cmd("repeatdialinfo", "", "cmd_repeatdialcancel")
}
