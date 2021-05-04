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

	gb, err := gobaresip.New("go-baresip", gobaresip.SetConfigPath("."), gobaresip.SetAudioPath("./sounds"))
	if err != nil {
		log.Println(err)
		return
	}

	client, err := CreateLokiClient(*lokiServer, 4, 2)
	if err != nil {
		log.Println(err)
	}

	defer client.Close()

	eChan := gb.GetEventChan()
	rChan := gb.GetResponseChan()

	go func() {
		for {
			select {
			case e := <-eChan:
				log.Println(e)
				client.Send(staticlokiLabel, e.Raw)
			case r := <-rChan:
				log.Println(r)
				client.Send(staticlokiLabel, r.Raw)
			}
		}
	}()

	go func() {
		time.Sleep(2 * time.Second)
		if err := gb.Reginfo(); err != nil {
			log.Println(err)
		}
	}()

	err = gb.Run()
	if err != nil {
		log.Println(err)
	}

}
