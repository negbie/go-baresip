# go-baresip

## Build Demo

sudo docker run --rm=true -itv $PWD:/mnt debian:buster-slim /mnt/build_docker_bin.sh

The above command will build a go-baresip-demo binary inside the go-baresip-demo folder.
On first run following files will be created if they do not exist:

* accounts
* config
* contacts
* current_contact

