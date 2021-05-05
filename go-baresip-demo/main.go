package main

import (
	"flag"
	"log"
	"time"

	gobaresip "github.com/negbie/go-baresip"
)

var staticlokiLabel = map[string]string{
	"job": "go-baresip",
}

func main() {

	lokiServer := flag.String("loki_server", "http://localhost:3100", "Loki HTTP server address")
	dial := flag.String("dial", "", "Dial SIP URI if it's not empty")
	autoDial := flag.String("auto_dial", "", "Auto dial SIP URI if it's not empty")
	autoDialDelay := flag.Int("auto_dial_delay", 5000, "Set delay before auto dial [ms]")
	autoHangup := flag.Bool("auto_hangup", true, "Set auto hangup")
	autoHangupDelay := flag.Int("auto_hangup_delay", 5000, "Set delay before auto hangup [ms]")
	debug := flag.Bool("debug", false, "Set debug mode")
	flag.Parse()

	gb, err := gobaresip.New(gobaresip.SetConfigPath("."), gobaresip.SetAudioPath("./sounds"), gobaresip.SetDebug(*debug))
	if err != nil {
		log.Println(err)
		return
	}

	loki, lerr := NewLokiClient(*lokiServer, 10, 4)
	if lerr != nil {
		log.Println(lerr)
	}

	defer loki.Close()

	eChan := gb.GetEventChan()
	rChan := gb.GetResponseChan()

	go func() {
		for {
			select {
			case e, ok := <-eChan:
				if !ok {
					continue
				}
				if lerr == nil {
					loki.Send(staticlokiLabel, e.Raw)
				} else {
					log.Println(e)
				}
			case r, ok := <-rChan:
				if !ok {
					continue
				}
				if lerr == nil {
					loki.Send(staticlokiLabel, r.Raw)
				} else {
					log.Println(r)
				}
			}
		}
	}()

	go func() {
		// Give baresip some time to init and register ua
		time.Sleep(1 * time.Second)

		if *autoHangup {
			if *autoHangupDelay >= 1000 {
				if err := gb.Autohangupdelay(*autoDialDelay); err != nil {
					log.Println(err)
				}
				if err := gb.Autohangup(); err != nil {
					log.Println(err)
				}
			} else {
				log.Println("auto_hangup_delay is too short")
			}
		}

		if *autoDial != "" {
			if *autoDialDelay >= 1000 {
				if err := gb.Autodialdelay(*autoDialDelay); err != nil {
					log.Println(err)
				}
				if err := gb.Autodial(*autoDial); err != nil {
					log.Println(err)
				}
			} else {
				log.Println("auto_dial_delay is too short")
			}
		} else if *dial != "" {
			if err := gb.Dial(*dial); err != nil {
				log.Println(err)
			}
		}
	}()

	err = gb.Run()
	if err != nil {
		log.Println(err)
	}
	defer gb.Close()
}
