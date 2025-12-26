package main

/*
#cgo LDFLAGS: -lpam -fPIC
#include <stdlib.h>
#include <security/pam_appl.h>

char *string_from_argv(int, char**);
char *get_username(pam_handle_t *pamh);
char *get_rhost(pam_handle_t *pamh);
char *get_wID(pam_handle_t *pamh, int flags);
char *get_tfaval(pam_handle_t *pamh, int flags);
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"
)

const (
	defaultServerURL = "api.werbot.com"
	defaultServiceID = "1"
	defaultServiceKey = "2"
	totpCodeLength   = 6
)

// config holds the PAM module configuration
type config struct {
	serverURL          string
	serviceID          string
	serviceKey         string
	offlineUsers       string
	insecureSkipVerify bool
	debug              bool
}

// initConfig initializes configuration with defaults
func initConfig() *config {
	return &config{
		serverURL:          defaultServerURL,
		serviceID:          defaultServiceID,
		serviceKey:         defaultServiceKey,
		offlineUsers:       "",
		insecureSkipVerify: false,
		debug:              false, // Security: debug disabled by default
	}
}

func sliceFromArgv(argc C.int, argv **C.char) []string {
	r := make([]string, 0, argc)
	for i := 0; i < int(argc); i++ {
		s := C.string_from_argv(C.int(i), argv)
		defer C.free(unsafe.Pointer(s))
		r = append(r, C.GoString(s))
	}
	return r
}

// parseBool converts string to boolean (supports "true"/"1" and "false"/"0")
func parseBool(s string) bool {
	return s == "true" || s == "1"
}

// getCString safely retrieves and converts C string to Go string
func getCString(cStr *C.char, fieldName string) string {
	if cStr == nil {
		writeLog(fmt.Sprintf("ERROR: Failed to get %s from PAM", fieldName))
		return ""
	}
	defer C.free(unsafe.Pointer(cStr))
	return C.GoString(cStr)
}

// parseConfig parses PAM arguments and returns configuration
func parseConfig(argc C.int, argv **C.char) *config {
	cfg := initConfig()
	for _, arg := range sliceFromArgv(argc, argv) {
		opt := strings.SplitN(arg, "=", 2)
		if len(opt) != 2 {
			continue
		}
		switch opt[0] {
		case "server_url":
			cfg.serverURL = opt[1]
		case "service_id":
			cfg.serviceID = opt[1]
		case "service_key":
			cfg.serviceKey = opt[1]
		case "offline_users":
			cfg.offlineUsers = opt[1]
		case "insecure_skip_verify":
			cfg.insecureSkipVerify = parseBool(opt[1])
		case "debug":
			cfg.debug = parseBool(opt[1])
		}
	}
	return cfg
}

//export pam_sm_authenticate
func pam_sm_authenticate(pamh *C.pam_handle_t, flags, argc C.int, argv **C.char) C.int {
	cfg := parseConfig(argc, argv)
	if cfg.serverURL == "" || cfg.serviceID == "" || cfg.serviceKey == "" {
		writeLog("ERROR: Invalid configuration - missing required parameters")
		return C.PAM_AUTH_ERR
	}

	// Get username
	username := getCString(C.get_username(pamh), "username")
	if username == "" {
		return C.PAM_USER_UNKNOWN
	}

	// Get remote user IP address
	cRHost := C.get_rhost(pamh)
	if cRHost == nil {
		writeLog(fmt.Sprintf("WARN: Failed to get remote host for user %s", username))
		return C.PAM_USER_UNKNOWN
	}
	defer C.free(unsafe.Pointer(cRHost))
	userIP := C.GoString(cRHost)

	// Get Werbot ID
	cWID := C.get_wID(pamh, flags)
	if cWID == nil {
		return C.PAM_USER_UNKNOWN
	}
	defer C.free(unsafe.Pointer(cWID))
	wID := C.GoString(cWID)

	// Check for PAM errors in wID
	if errCode := checkPAMError(wID); errCode != C.PAM_SUCCESS {
		return errCode
	}

	// Get TFA value (TOTP code or empty for U2F)
	cTfaval := C.get_tfaval(pamh, flags)
	defer C.free(unsafe.Pointer(cTfaval))
	totpCode := C.GoString(cTfaval)

	// Check for PAM errors in TFA value
	if errCode := checkPAMError(totpCode); errCode != C.PAM_SUCCESS {
		return errCode
	}

	// Call TFA request flow
	if sendTfaReq(cfg, username, wID, totpCode, userIP) {
		return C.PAM_SUCCESS
	}
	return C.PAM_AUTH_ERR
}

//export pam_sm_setcred
func pam_sm_setcred(pamh *C.pam_handle_t, flags, argc C.int, argv **C.char) C.int {
	return C.PAM_IGNORE
}

//export pam_sm_acct_mgmt
func pam_sm_acct_mgmt(pamh *C.pam_handle_t, flags, argc C.int, argv **C.char) C.int {
	cfg := parseConfig(argc, argv)
	if cfg.serverURL == "" || cfg.serviceID == "" || cfg.serviceKey == "" {
		return C.PAM_SUCCESS
	}

	username := getCString(C.get_username(pamh), "username")
	if username == "" {
		return C.PAM_USER_UNKNOWN
	}

	cRHost := C.get_rhost(pamh)
	if cRHost == nil {
		return C.PAM_SUCCESS
	}
	defer C.free(unsafe.Pointer(cRHost))
	userIP := C.GoString(cRHost)

	if checkAccountAccess(cfg, username, userIP) {
		return C.PAM_SUCCESS
	}
	return C.PAM_AUTH_ERR
}

// checkPAMError checks if the string from C function is a PAM error code
// and returns the corresponding PAM error constant, or PAM_SUCCESS if no error
func checkPAMError(err string) C.int {
	switch err {
	case "pam_auth_err":
		return C.PAM_AUTH_ERR
	case "pam_conv_err", "cr":
		return C.PAM_CONV_ERR
	}
	return C.PAM_SUCCESS
}

func main() {}
