#!/bin/sh

set -ex

#sudo apt install -y make gcc zlib1g-dev libssl-dev openssl git wget

mkdir -p libbaresip
cd libbaresip/
mkdir -p git
mkdir -p re
mkdir -p rem
mkdir -p baresip
cd git

my_base_modules="account contact autotest cons ctrl_tcp debug_cmd httpd menu ice stun turn serreg uuid stdio"
my_audio_modules="aubridge aufile auloop"
my_codec_modules="g711"
my_tls_modules="srtp"


if [ ! -d "re" ]; then
    git clone https://github.com/baresip/re.git
fi
cd re; make clean; make libre.a; cp libre.a ../../re; cd ..

if [ ! -d "rem" ]; then
    git clone https://github.com/baresip/rem.git
fi
cd rem; make clean; make librem.a; cp librem.a ../../rem; cd ..

if [ ! -d "baresip" ]; then
    git clone https://github.com/baresip/baresip.git
fi
if [ ! -d "baresip-apps" ]; then
    git clone https://github.com/baresip/baresip-apps.git
fi

cp -R baresip-apps/modules/autotest baresip/modules/
sed -i 's/$(BARESIP_MOD_MK)/mk\/mod.mk/g' baresip/modules/autotest/module.mk
sed -i '/auloop/a MODULES   += autotest' baresip/mk/modules.mk

cd baresip
    
make clean; make LIBRE_SO=../re LIBREM_PATH=../rem STATIC=1 libbaresip.a \
    MODULES="$my_base_modules $my_audio_modules $my_codec_modules $my_tls_modules"

cp libbaresip.a ../../baresip; cd ..
cp -R re/include ../re
cp -R rem/include ../rem
cp -R baresip/include ../baresip
cd ../..

cd espeak
if [ ! -d "espeak-ng" ]; then
    git clone https://github.com/espeak-ng/espeak-ng.git
fi
cd espeak-ng
./autogen.sh
./configure --without-async --without-mbrola --without-sonic --without-speechplayer
make
cp src/.libs/libespeak-ng.a ../
cp src/include/espeak-ng/speak_lib.h ../
make clean
cd ..

if [ ! -d "soxr-code" ]; then
    git clone https://git.code.sf.net/p/soxr/code soxr-code
fi
cd soxr-code
cmake -DWITH_OPENMP=OFF -DBUILD_SHARED_LIBS=OFF -DBUILD_TESTS=0 -DBUILD_EXAMPLES=0 .
make
cp src/libsoxr.a src/soxr.h ../
make clean
cd ..
