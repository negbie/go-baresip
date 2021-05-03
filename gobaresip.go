package gobaresip

/*
#cgo linux CFLAGS: -I.
#cgo linux LDFLAGS: ${SRCDIR}/libbaresip/baresip/libbaresip.a ${SRCDIR}/libbaresip/re/libre.a ${SRCDIR}/libbaresip/rem/librem.a -ldl -lm -lcrypto -lssl -lz

#include <stdint.h>
#include <stdlib.h>
#include <libbaresip/re/include/re.h>
#include <libbaresip/rem/include/rem.h>
#include <libbaresip/baresip/include/baresip.h>


static void signal_handler(int sig)
{
	static bool term = false;

	if (term) {
		module_app_unload();
		mod_close();
		exit(0);
	}

	term = true;

	info("terminated by signal %d\n", sig);

	ua_stop_all(false);
}

static void net_change_handler(void *arg)
{
	(void)arg;

	info("IP-address changed: %j\n",
	     net_laddr_af(baresip_network(), AF_INET));

	(void)uag_reset_transp(true, true);
}

static void set_net_change_handler()
{
	net_change(baresip_network(), 60, net_change_handler, NULL);
}

static void ua_exit_handler(void *arg)
{
	(void)arg;
	debug("ua exited -- stopping main runloop\n");

	//The main run-loop can be stopped now
	re_cancel();
}

static void set_ua_exit_handler()
{
	uag_set_exit_handler(ua_exit_handler, NULL);
}

int mainLoop(){
	return re_main(signal_handler);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func Run(configPath, audioPath string) (err C.int) {

	ua := C.CString("go-baresip")
	defer C.free(unsafe.Pointer(ua))

	err = C.libre_init()
	if err != 0 {
		fmt.Printf("libre init failed with error code %d\n", err)
		return end(err)
	}

	if configPath != "" {
		cp := C.CString(configPath)
		defer C.free(unsafe.Pointer(cp))
		C.conf_path_set(cp)
	}

	err = C.conf_configure()
	if err != 0 {
		fmt.Printf("baresip configure failed with error code %d\n", err)
		return end(err)
	}

	// Top-level baresip struct init must be done AFTER configuration is complete.
	err = C.baresip_init(C.conf_config())
	if err != 0 {
		fmt.Printf("baresip main init failed with error code %d\n", err)
		return end(err)
	}

	if audioPath != "" {
		ap := C.CString(audioPath)
		defer C.free(unsafe.Pointer(ap))
		C.play_set_path(C.baresip_player(), ap)
	}

	err = C.ua_init(ua, 1, 1, 1)
	if err != 0 {
		fmt.Printf("baresip ua init failed with error code %d\n", err)
		return end(err)
	}

	C.set_net_change_handler()
	C.set_ua_exit_handler()

	err = C.conf_modules()
	if err != 0 {
		fmt.Printf("baresip load modules failed with error code %d\n", err)
		return end(err)
	}

	//C.sys_daemon()
	//C.uag_enable_sip_trace(1)
	//C.log_enable_debug(1)
	/*
		ua_eprm := C.CString("")
		defer C.free(unsafe.Pointer(ua_eprm))
		err = C.uag_set_extra_params(ua_eprm)
	*/

	return end(C.mainLoop())
}

func end(err C.int) C.int {
	if err != 0 {
		C.ua_stop_all(1)
	}

	C.ua_close()
	C.module_app_unload()
	C.conf_close()

	C.baresip_close()

	// Modules must be unloaded after all application activity has stopped.
	C.mod_close()

	C.libre_close()

	// Check for memory leaks
	C.tmr_debug()
	C.mem_debug()

	return err
}
