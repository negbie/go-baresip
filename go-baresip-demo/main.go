package main

import (
	"log"
	"time"

	gobaresip "github.com/negbie/go-baresip"
)

func main() {

	gb := gobaresip.New("127.0.0.1:4444", ".", "./sounds")

	eChan := gb.EventChan()
	rChan := gb.ResponseChan()

	go func() {
		for {
			select {
			case e := <-eChan:
				log.Println(e)

			case r := <-rChan:
				log.Println(r)

			}
		}
	}()

	go func() {
		time.Sleep(2 * time.Second)

		if err := gb.Reginfo(); err != nil {
			log.Println(err)
		}
	}()

	err := gb.Run()
	if err != 0 {
		log.Println(err)
	}
}
