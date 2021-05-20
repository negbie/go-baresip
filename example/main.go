package main

import (
	"log"
	"time"

	gobaresip "github.com/negbie/go-baresip"
)

func main() {

	gb, err := gobaresip.New(
		gobaresip.SetAudioPath("sounds"),
		gobaresip.SetConfigPath("."),
		gobaresip.SetWsAddr("0.0.0.0:8080"),
	)
	if err != nil {
		log.Fatal(err)
	}

	eChan := gb.GetEventChan()
	rChan := gb.GetResponseChan()

	go func() {
		for {
			select {
			case e, ok := <-eChan:
				if !ok {
					continue
				}
				log.Println(string(e.RawJSON))
			case r, ok := <-rChan:
				if !ok {
					continue
				}
				log.Println(string(r.RawJSON))
			}
		}
	}()

	go func() {
		// Give baresip some time to init and register ua
		time.Sleep(1 * time.Second)

		if err := gb.CmdDial("012345"); err != nil {
			log.Println(err)
		}
	}()

	err = gb.Run()
	if err != nil {
		log.Println(err)
	}
	defer gb.Close()
}
