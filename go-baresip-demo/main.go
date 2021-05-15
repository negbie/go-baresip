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
	wsAddr := flag.String("ws_address", "0.0.0.0:8080", "Loki HTTP server address")
	dial := flag.String("dial", "", "Dial SIP URI if it's not empty")
	repeatDial := flag.String("repeat_dial", "", "Repeat dial SIP URI if it's not empty")
	repeatDialInterval := flag.Int("repeat_dial_interval", 30, "Set repeat dial interval [s]")
	debug := flag.Bool("debug", false, "Set debug mode")
	flag.Parse()

	gb, err := gobaresip.New(
		gobaresip.SetConfigPath("."),
		gobaresip.SetAudioPath("./sounds"),
		gobaresip.SetDebug(*debug),
		gobaresip.SetWsAddr(*wsAddr),
	)
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
					loki.Send(staticlokiLabel, string(e.RawJSON))
				} else {
					log.Println(string(e.RawJSON))
				}

			case r, ok := <-rChan:
				if !ok {
					continue
				}
				if lerr == nil {
					loki.Send(staticlokiLabel, string(r.RawJSON))
				} else {
					log.Println(string(r.RawJSON))
				}
			}
		}
	}()

	go func() {
		// Give baresip some time to init and register ua
		time.Sleep(1 * time.Second)

		if *repeatDial != "" {
			if err := gb.CmdRepeatDialInterval(*repeatDialInterval); err != nil {
				log.Println(err)
			}
			if err := gb.CmdRepeatDial(*repeatDial); err != nil {
				log.Println(err)
			}
		} else if *dial != "" {
			if err := gb.CmdDial(*dial); err != nil {
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
