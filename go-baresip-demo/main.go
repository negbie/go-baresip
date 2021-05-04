package main

import (
	"log"
	"os"
	"time"

	gobaresip "github.com/negbie/go-baresip"
)

func main() {
	logFile := "go-baresip-demo.log"
	f, _ := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()

	log.SetOutput(&gobaresip.Logger{
		Filename:   logFile,
		MaxSize:    14, // mb
		MaxBackups: 7,
		MaxAge:     21, //days
		Compress:   true,
	})

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
