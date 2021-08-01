package main

/*
#cgo LDFLAGS: -lpam -fPIC
#include <stdlib.h>
#include <security/pam_appl.h>
char *get_username(pam_handle_t *pamh);
char *get_rhost(pam_handle_t *pamh);
char *get_tfaval(pam_handle_t *pamh, int flags);
*/
import "C"

import (
	"fmt"
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

	// get tfaval. code or empty for u2f.
	cTfaval := C.get_tfaval(pamh, flags)
	defer C.free(unsafe.Pointer(cTfaval))

	check, cErr := checkAndReturnPAMerr(C.GoString(cTfaval))
	if check == true {
		return cErr
	}

	if debug {
		tf := fmt.Sprintf("cTfaval result: %s", C.GoString(cTfaval))
		writeLog(tf)
	}

	// call tfaReuqest flow here
	tfaResp := sendTfaReq(C.GoString(cUsername), C.GoString(cTfaval), C.GoString(cRHost))
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

// since we return char * from c function, we use this function to check and return PAM module error
func checkAndReturnPAMerr(err string) (bool, C.int) {
	switch err {
	case "pam_auth_err":
		return true, C.PAM_AUTH_ERR
	case "pam_conv_err":
		return true, C.PAM_CONV_ERR
	case "cr":
		return true, C.PAM_CONV_ERR
	}

	return false, C.PAM_SUCCESS
}

func main() {}
