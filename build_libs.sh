#!/bin/sh

#sudo apt update
#sudo apt install -y autoconf automake libtool pkg-config make cmake gcc zlib1g-dev libssl-dev openssl git wget

mkdir -p libbaresip
cd libbaresip/
mkdir -p git
mkdir -p re
mkdir -p rem
mkdir -p baresip
mkdir -p alsa
mkdir -p opus
mkdir -p openssl/include

my_base_modules="account b2bua contact cons ctrl_tcp debug_cmd echo httpd menu natpmp ice stun turn serreg uuid stdio"
my_audio_modules="alsa aubridge aufile auloop ausine"
my_codec_modules="g711 g722 opus"
my_tls_modules="dtls_srtp srtp"

opus="1.3.1"
openssl="1.1.1k"
alsa="1.2.4"

cd git

if [ ! -d "re" ]; then
    git clone https://github.com/baresip/re.git
fi
cd re; make clean; make -j16 USE_ZLIB= RELEASE=1 libre.a; cp libre.a ../../re; cd ..

if [ ! -d "rem" ]; then
    git clone https://github.com/baresip/rem.git
fi
cd rem; make clean; make -j16 USE_ZLIB= RELEASE=1 librem.a; cp librem.a ../../rem; cd ..

if [ ! -d "openssl-${openssl}" ]; then
    wget https://www.openssl.org/source/openssl-${openssl}.tar.gz
    tar -xzf openssl-${openssl}.tar.gz
fi
cd openssl-${openssl}; ./config no-shared; make clean; make -j16 build_libs; cd ..
cp openssl-${openssl}/*.a ../openssl
cp openssl-${openssl}/include/openssl/*.h ../openssl/include
sed -ri 's/<openssl\/(.*)>/"\1"/g' ../openssl/include/*

if [ ! -d "opus-${opus}" ]; then
    wget "https://archive.mozilla.org/pub/opus/opus-${opus}.tar.gz"
    tar -xzf opus-${opus}.tar.gz
fi
cd opus-${opus}; ./configure; make clean; make -j16; cd ..
mkdir opus
mkdir -p my_include/opus
cp opus-${opus}/.libs/libopus.a ../opus; cp opus-${opus}/.libs/libopus.a opus
cp opus-${opus}/include/*.h my_include/opus


if [ ! -d "alsa-lib-${alsa}" ]; then
    wget "https://github.com/alsa-project/alsa-lib/archive/refs/tags/v${alsa}.tar.gz" -O alsa-lib-${alsa}.tar.gz
    tar -xzf alsa-lib-${alsa}.tar.gz
fi
cd alsa-lib-${alsa};
libtoolize --force --copy --automake
aclocal
autoheader
automake --foreign --copy --add-missing
autoconf
./configure --enable-shared=no --enable-static=yes
make clean; make -j16; cd ..
mkdir alsa
mkdir -p my_include/alsa
cp alsa-lib-${alsa}/src/.libs/libasound.a ../alsa; cp alsa-lib-${alsa}/src/.libs/libasound.a alsa
cp alsa-lib-${alsa}/include/*.h my_include/alsa

if [ ! -d "baresip" ]; then
    git clone https://github.com/baresip/baresip.git
fi
cd baresip
rm -rf modules/g722
cp -ap ../../../g722 modules/

make clean; make -j16 LIBRE_SO=../re LIBREM_PATH=../rem USE_ZLIB= RELEASE=1 STATIC=1 libbaresip.a \
    MODULES="$my_base_modules $my_audio_modules $my_codec_modules $my_tls_modules" \
    EXTRA_CFLAGS="-I ../my_include" EXTRA_LFLAGS="-L ..opus -L ..alsa"

cp libbaresip.a ../../baresip; cd ..
cp -R re/include ../re
cp -R rem/include ../rem
cp -R baresip/include ../baresip
cd ../..

: '
cd espeak
if [ ! -d "espeak-ng" ]; then
    git clone https://github.com/espeak-ng/espeak-ng.git
fi
cd espeak-ng
./autogen.sh
./configure --without-async --without-mbrola --without-sonic --without-speechplayer
make -j16
cp src/.libs/libespeak-ng.a ../
cp src/include/espeak-ng/speak_lib.h ../
make clean
cd ..

if [ ! -d "soxr-code" ]; then
    git clone https://git.code.sf.net/p/soxr/code soxr-code
fi
cd soxr-code
cmake -DWITH_OPENMP=OFF -DBUILD_SHARED_LIBS=OFF -DBUILD_TESTS=0 -DBUILD_EXAMPLES=0 .
make -j16
cp src/libsoxr.a src/soxr.h ../
make clean
cd ..
'