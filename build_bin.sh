#!/bin/sh

set -ex

#sudo apt install -y make gcc zlib1g-dev libssl-dev openssl git wget

rm -rf libbaresip
rm -f go-baresip
mkdir libbaresip
cd libbaresip/
mkdir git
mkdir re
mkdir rem
mkdir baresip
cd git

my_base_modules="debug_cmd menu ice stun turn uuid account contact"
my_audio_modules="aufile auloop"
my_codec_modules="g711"
#my_ui_modules="stdio cons httpd"
my_ctrl_modules="ctrl_tcp"
my_tls_modules="srtp"

git clone https://github.com/baresip/re.git
cd re; make libre.a; cp libre.a ../../re; cd ..

git clone https://github.com/baresip/rem.git
cd rem; make librem.a; cp librem.a ../../rem; cd ..

git clone https://github.com/baresip/baresip.git
cd baresip
    
make LIBRE_SO=../re LIBREM_PATH=../rem STATIC=1 libbaresip.a \
    MODULES="$my_base_modules $my_audio_modules $my_codec_modules $my_ui_modules $my_ctrl_modules $my_tls_modules"

cp libbaresip.a ../../baresip; cd ..
mv re/include ../re
mv rem/include ../rem
mv baresip/include ../baresip
cd ..; rm -rf git

