package main

import (
	"log"
	"time"

	gobaresip "github.com/negbie/go-baresip"
)

func main() {

	gb, err := gobaresip.New(gobaresip.SetConfigPath("."))
	if err != nil {
		log.Println(err)
		return
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
				log.Println(e)
			case r, ok := <-rChan:
				if !ok {
					continue
				}
				log.Println(r)
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
