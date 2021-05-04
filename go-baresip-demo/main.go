package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	gobaresip "github.com/negbie/go-baresip"
)

var staticlokiLabel = map[string]string{
	"job": "go-baresip",
}

func main() {

	lokiServer := flag.String("loki_server", "http://localhost:3100", "Loki HTTP server address")
	dial := flag.String("dial", "", "Dial SIP URI if it's not empty")
	repeatDialDuration := flag.String("repeat_dial_duration", "0s", "Repeats dial after this duration if it's more than 5s")
	flag.Parse()

	gb, err := gobaresip.New(gobaresip.SetConfigPath("."), gobaresip.SetAudioPath("./sounds"))
	if err != nil {
		log.Println(err)
		return
	}

	loki, lerr := NewLokiClient(*lokiServer, 4, 2)
	if lerr != nil {
		log.Println(lerr)
	}

	defer loki.Close()

	eChan := gb.GetEventChan()
	rChan := gb.GetResponseChan()

	go func() {
		for {
			select {
			case e := <-eChan:
				if lerr == nil {
					loki.Send(staticlokiLabel, e.Raw)
				} else {
					log.Println(e)
				}
			case r := <-rChan:
				if lerr == nil {
					loki.Send(staticlokiLabel, r.Raw)
				} else {
					log.Println(r)
				}
			}
		}
	}()

	go func() {
		time.Sleep(2 * time.Second)
		if *dial != "" && *repeatDialDuration != "" && !strings.HasPrefix(*repeatDialDuration, "0") {
			if d, err := time.ParseDuration(*repeatDialDuration); err == nil && d > time.Duration(5*time.Second) {
				ticker := time.NewTicker(d)
				defer ticker.Stop()
				for ; true; <-ticker.C {
					fmt.Println(time.Now())
					if err := gb.Dial(*dial); err != nil {
						log.Println(err)
					}
				}
			} else {
				log.Println("repeat_dial_duration must be higher than 5s and lower than 1d")
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
