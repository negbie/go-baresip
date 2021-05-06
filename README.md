# go-baresip

Is a tiny wrapper around [baresip](https://github.com/baresip/baresip)

## Build Demo

sudo docker run --rm=true -itv $PWD:/mnt debian:buster-slim /mnt/build_docker_bin.sh

The above command will build a go-baresip-demo binary inside the go-baresip-demo folder.
On first run following files will be created if they do not exist:

* accounts
* config
* contacts
* current_contact
* uuid

## Basic Usage

```Go
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

        if err := gb.Dial("01234"); err != nil {
            log.Println(err)
        }
    }()

    err = gb.Run()
    if err != nil {
        log.Println(err)
    }
    defer gb.Close()
}
```