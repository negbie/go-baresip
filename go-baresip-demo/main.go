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

	lokiServer := flag.String("loki_server", "http://localhost:3100", "Loki HTTP address")
	flag.Parse()

	client, lerr := CreateLokiClient(*lokiServer, 4, 2)
	if lerr != nil {
		log.Println(lerr)
	}

	defer client.Close()

	gb := gobaresip.New("127.0.0.1:4444", ".", "./sounds", true)

	eChan := gb.GetEventChan()
	rChan := gb.GetResponseChan()

	go func() {
		for {
			select {
			case e := <-eChan:
				log.Println(e)
				if lerr == nil {
					client.Send(staticlokiLabel, e.Raw)
				}
			case r := <-rChan:
				log.Println(r)
				if lerr == nil {
					client.Send(staticlokiLabel, r.Raw)
				}
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
