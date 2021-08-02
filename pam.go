package main

/*
#cgo LDFLAGS: -lpam -fPIC
#include <security/pam_appl.h>
#include <stdlib.h>

char *string_from_argv(int, char**);
char *get_username(pam_handle_t *pamh);
char *get_rhost(pam_handle_t *pamh);
*/
import "C"

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

var (
	serverURL    = "https://api.werbot.com"
	serviceID    = "bef57548-dfdd-4aba-a545-50aa9e4f50db"
	serviceKey   = "7ba94fcbac95b794fc6efc25e1c23f0d6b"
	offlineUsers = ""
	debug        = true
)

//export pam_sm_authenticate
func pam_sm_authenticate(pamh *C.pam_handle_t, flags, argc C.int, argv **C.char) C.int {
	runtime.GOMAXPROCS(1)

	for _, arg := range sliceFromArgv(argc, argv) {
		opt := strings.SplitN(arg, "=", 2)
		switch opt[0] {
		case "server_url":
			serverURL = opt[1]
		case "service_id":
			serviceID = opt[1]
		case "service_key":
			serviceKey = opt[1]
		case "ofline_users":
			offlineUsers = opt[1]
		case "debug":
			t, _ := strconv.ParseBool(opt[1])
			debug = t
		}
	}

	// get username
	cUsername := C.get_username(pamh)
	if cUsername == nil {
		return C.PAM_USER_UNKNOWN
	}
	defer C.free(unsafe.Pointer(cUsername))

	// get remote user ip address
	cRHost := C.get_rhost(pamh)
	if cRHost == nil {
		return C.PAM_USER_UNKNOWN
	}
	defer C.free(unsafe.Pointer(cRHost))

	if debug {
		rh := fmt.Sprintf("cRHost result: %s", C.GoString(cRHost))
		writeLog(rh)
	}

	// get tfaval code
	cTfaval, err := conversation(pamh, "Enter your werbot code: ")
	if err != nil {
		return C.PAM_AUTH_ERR
	}

	if debug {
		tf := fmt.Sprintf("cTfaval result: %s", cTfaval)
		writeLog(tf)
	}

	// call tfaReuqest flow here
	tfaResp := sendTfaReq(C.GoString(cUsername), cTfaval, C.GoString(cRHost))
	if debug {
		f := fmt.Sprintf("final result: %v", tfaResp)
		writeLog(f)
	}

	if tfaResp == true {
		return C.PAM_SUCCESS
	}

	return C.PAM_AUTH_ERR
}

//export pam_sm_setcred
func pam_sm_setcred(pamh *C.pam_handle_t, flags, argc C.int, argv **C.char) C.int {
	return C.PAM_IGNORE
}

func main() {}
