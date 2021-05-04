package gobaresip

// SetOption takes one or more option function and applies them in order to Baresip.
func (b *Baresip) SetOption(options ...func(*Baresip) error) error {
	for _, opt := range options {
		if err := opt(b); err != nil {
			return err
		}
	}
	return nil
}

// SetCtrlTCPAddr sets the ctrl_tcp modules address.
func SetCtrlTCPAddr(a string) func(*Baresip) error {
	return func(b *Baresip) error {
		b.ctrlAddr = a
		return nil
	}
}

// SetConfigPath sets the config path.
func SetConfigPath(p string) func(*Baresip) error {
	return func(b *Baresip) error {
		b.configPath = p
		return nil
	}
}

// SetAudioPath sets the audio path.
func SetAudioPath(p string) func(*Baresip) error {
	return func(b *Baresip) error {
		b.audioPath = p
		return nil
	}
}

// SetDebug sets the debug mode.
func SetDebug(d bool) func(*Baresip) error {
	return func(b *Baresip) error {
		b.debug = d
		return nil
	}
}
