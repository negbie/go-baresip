package gobaresip

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
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

// Command will send a raw baresip command over ctrl_tcp.
func (b *Baresip) Command(command, params, token string) error {
	msg, err := json.Marshal(buildCommand(command, params, token))
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

// CommandAccept will accept incoming call
func (b *Baresip) CommandAccept(s ...string) error {
	c := "accept"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandAcceptdir will accept incoming call with audio and videodirection.
func (b *Baresip) CommandAcceptdir(s ...string) error {
	c := "acceptdir"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandAnswermode will set answer mode
func (b *Baresip) CommandAnswermode(s ...string) error {
	c := "answermode"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandAuplay will switch audio player
func (b *Baresip) CommandAuplay(s ...string) error {
	c := "auplay"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandAusrc will switch audio source
func (b *Baresip) CommandAusrc(s ...string) error {
	c := "ausrc"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandCallstat will show call status
func (b *Baresip) CommandCallstat(s ...string) error {
	c := "callstat"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandContact_next will set next contact
func (b *Baresip) CommandContact_next(s ...string) error {
	c := "contact_next"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandContact_prev will set previous contact
func (b *Baresip) CommandContact_prev(s ...string) error {
	c := "contact_prev"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandAutodial will dial number automatically
func (b *Baresip) CommandAutodial(s ...string) error {
	c := "autodial dial"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandAutodialdelay will set delay before auto dial [ms]
func (b *Baresip) CommandAutodialdelay(n ...int) error {
	c := "autodialdelay"
	if len(n) > 0 {
		return b.Command(c, strconv.Itoa(n[0]), "command_"+c+"_"+strconv.Itoa(n[0]))
	}
	return b.Command(c, "", "command_"+c)
}

// CommandDial will dial number
func (b *Baresip) CommandDial(s ...string) error {
	c := "dial"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandDialcontact will dial current contact
func (b *Baresip) CommandDialcontact(s ...string) error {
	c := "dialcontact"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandDialdir will dial with audio and videodirection
func (b *Baresip) CommandDialdir(s ...string) error {
	c := "dialdir"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandAutohangup will hangup call automatically
func (b *Baresip) CommandAutohangup(s ...string) error {
	c := "autohangup"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandAutohangupdelay will set delay before hangup [ms]
func (b *Baresip) CommandAutohangupdelay(n ...int) error {
	c := "autohangupdelay"
	if len(n) > 0 {
		return b.Command(c, strconv.Itoa(n[0]), "command_"+c+"_"+strconv.Itoa(n[0]))
	}
	return b.Command(c, "", "command_"+c)
}

// CommandHangup will hangup call
func (b *Baresip) CommandHangup(s ...string) error {
	c := "hangup"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandHangupall will hangup all calls with direction
func (b *Baresip) CommandHangupall(s ...string) error {
	c := "hangupall"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandInsmod will load module
func (b *Baresip) CommandInsmod(s ...string) error {
	c := "insmod"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandListcalls will list active calls
func (b *Baresip) CommandListcalls(s ...string) error {
	c := "listcalls"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandReginfo will list registration info
func (b *Baresip) CommandReginfo(s ...string) error {
	c := "reginfo"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandRmmod will unload module
func (b *Baresip) CommandRmmod(s ...string) error {
	c := "rmmod"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandSetadelay will set answer delay for outgoing call
func (b *Baresip) CommandSetadelay(s ...string) error {
	c := "setadelay"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandUadel will delete User-Agent
func (b *Baresip) CommandUadel(s ...string) error {
	c := "uadel"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandUadelall will delete all User-Agents
func (b *Baresip) CommandUadelall(s ...string) error {
	c := "uadelall"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandUafind will find User-Agent <aor>
func (b *Baresip) CommandUafind(s ...string) error {
	c := "uafind"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandUanew will create User-Agent
func (b *Baresip) CommandUanew(s ...string) error {
	c := "uanew"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandUareg will register <regint> [index]
func (b *Baresip) CommandUareg(s ...string) error {
	c := "uareg"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}

// CommandQuit will quit baresip
func (b *Baresip) CommandQuit(s ...string) error {
	c := "quit"
	if len(s) > 0 {
		return b.Command(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Command(c, "", "command_"+c)
}
