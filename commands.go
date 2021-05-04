package gobaresip

/*
/accept             a        Accept incoming call
/acceptdir ..                Accept incoming call with audio and videodirection
/answermode ..               Set answer mode
/auplay ..                   Switch audio player
/ausrc ..                    Switch audio source
/callstat           c        Call status
/contact_next       >        Set next contact
/contact_prev       <        Set previous contact
/contacts           C        List contacts
/dial ..            d ..     Dial
/dialcontact        D        Dial current contact
/dialdir ..                  Dial with audio and videodirection.
/dnd ..                      Set Do not Disturb
/hangup             b        Hangup call
/hangupall ..                Hangup all calls with direction
/help               h        Help menu
/insmod ..                   Load module
/listcalls          l        List active calls
/message ..         M ..     Message current contact
/options ..         o ..     Options
/quit               q        Quit
/reginfo            r        Registration info
/rmmod ..                    Unload module
/setadelay ..                Set answer delay for outgoing call
/tlsissuer                   TLS certificate issuer
/tlssubject                  TLS certificate subject
/uadel ..                    Delete User-Agent
/uadelall ..                 Delete all User-Agents
/uafind ..                   Find User-Agent <aor>
/uanew ..                    Create User-Agent
/uareg ..                    UA register <regint> [index]
/vidsrc ..                   Switch video source asdf
*/

// Accept will accept incoming call
func (b *Baresip) Accept(s ...string) error {
	c := "accept"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Acceptdir will accept incoming call with audio and videodirection.
func (b *Baresip) Acceptdir(s ...string) error {
	c := "acceptdir"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Answermode will set answer mode
func (b *Baresip) Answermode(s ...string) error {
	c := "answermode"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Auplay will switch audio player
func (b *Baresip) Auplay(s ...string) error {
	c := "auplay"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Ausrc will switch audio source
func (b *Baresip) Ausrc(s ...string) error {
	c := "ausrc"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Callstat will show call status
func (b *Baresip) Callstat(s ...string) error {
	c := "callstat"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Contact_next will set next contact
func (b *Baresip) Contact_next(s ...string) error {
	c := "contact_next"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Contact_prev will set previous contact
func (b *Baresip) Contact_prev(s ...string) error {
	c := "contact_prev"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Dial will dial number
func (b *Baresip) Dial(s ...string) error {
	c := "dial"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Dialcontact will dial current contact
func (b *Baresip) Dialcontact(s ...string) error {
	c := "dialcontact"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Dialdir will dial with audio and videodirection
func (b *Baresip) Dialdir(s ...string) error {
	c := "dialdir"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Hangup will hangup call
func (b *Baresip) Hangup(s ...string) error {
	c := "hangup"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Hangupall will hangup all calls with direction
func (b *Baresip) Hangupall(s ...string) error {
	c := "hangupall"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Insmod will load module
func (b *Baresip) Insmod(s ...string) error {
	c := "insmod"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Listcalls will list active calls
func (b *Baresip) Listcalls(s ...string) error {
	c := "listcalls"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Reginfo will list registration info
func (b *Baresip) Reginfo(s ...string) error {
	c := "reginfo"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Rmmod will unload module
func (b *Baresip) Rmmod(s ...string) error {
	c := "rmmod"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Setadelay will set answer delay for outgoing call
func (b *Baresip) Setadelay(s ...string) error {
	c := "setadelay"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Uadel will delete User-Agent
func (b *Baresip) Uadel(s ...string) error {
	c := "uadel"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Uadelall will delete all User-Agents
func (b *Baresip) Uadelall(s ...string) error {
	c := "uadelall"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Uafind will find User-Agent <aor>
func (b *Baresip) Uafind(s ...string) error {
	c := "uafind"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Uanew will create User-Agent
func (b *Baresip) Uanew(s ...string) error {
	c := "uanew"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Uareg will register <regint> [index]
func (b *Baresip) Uareg(s ...string) error {
	c := "uareg"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}

// Quit will quit baresip
func (b *Baresip) Quit(s ...string) error {
	c := "quit"
	if len(s) > 0 {
		return b.Exec(c, s[0], "command_"+c+"_"+s[0])
	}
	return b.Exec(c, "", "command_"+c)
}
