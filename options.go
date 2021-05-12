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
func SetCtrlTCPAddr(opt string) func(*Baresip) error {
	return func(b *Baresip) error {
		b.ctrlAddr = opt
		return nil
	}
}

// SetWsAddr sets the ws address.
func SetWsAddr(opt string) func(*Baresip) error {
	return func(b *Baresip) error {
		b.wsAddr = opt
		return nil
	}
}

// SetConfigPath sets the config path.
func SetConfigPath(opt string) func(*Baresip) error {
	return func(b *Baresip) error {
		b.configPath = opt
		return nil
	}
}

// SetAudioPath sets the audio path.
func SetAudioPath(opt string) func(*Baresip) error {
	return func(b *Baresip) error {
		b.audioPath = opt
		return nil
	}
}

// SetDebug sets the debug mode.
func SetDebug(opt bool) func(*Baresip) error {
	return func(b *Baresip) error {
		b.debug = opt
		return nil
	}
}

// SetUserAgent sets the UserAgent.
func SetUserAgent(opt string) func(*Baresip) error {
	return func(b *Baresip) error {
		b.userAgent = opt
		return nil
	}
}
