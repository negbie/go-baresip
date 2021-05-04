package main

import (
	"fmt"
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
				fmt.Println(e)

			case r := <-rChan:
				fmt.Println(r)

			}
		}
	}()

	go func() {
		time.Sleep(2 * time.Second)
		if err := gb.Exec("listcalls", "", "asdf"); err != nil {
			fmt.Println(err)
		}
	}()

	err := gb.Run()
	if err != 0 {
		fmt.Println(err)
	}
}
