package main

/*
#cgo LDFLAGS: -lpam -fPIC
#include <security/pam_appl.h>
#include <stdlib.h>

char *string_from_argv(int, char**);
char *get_host(pam_handle_t *pamh);
char *get_user(pam_handle_t *pamh);
int get_uid(char *user);
*/
import "C"
import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var (
	serverURL    = "https://api.werbot.com"
	serviceID    = "bef57548-dfdd-4aba-a545-50aa9e4f50db"
	serviceKey   = "7ba94fcbac95b794fc6efc25e1c23f0d6b"
	offlineUsers = "ubuntu"
	debug        = true

	matchTfa = regexp.MustCompile(`^[0-9]{6}$`)
)

//export pam_sm_authenticate
func pam_sm_authenticate(pamh *C.pam_handle_t, flags, argc C.int, argv **C.char) C.int {
	cUsername := C.get_user(pamh)
	if cUsername == nil {
		return C.PAM_USER_UNKNOWN
	}
	defer C.free(unsafe.Pointer(cUsername))

	uid := int(C.get_uid(cUsername))
	if uid < 0 {
		return C.PAM_USER_UNKNOWN
	}

	return pamAuthenticate(pamh, uid, C.GoString(cUsername), sliceFromArgv(argc, argv))
}

//export pam_sm_setcred
func pam_sm_setcred(pamh *C.pam_handle_t, flags, argc C.int, argv **C.char) C.int {
	return C.PAM_IGNORE
}

func pamAuthenticate(pamh *C.pam_handle_t, uid int, username string, argv []string) C.int {
	runtime.GOMAXPROCS(1)

	for _, arg := range argv {
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

	// get remote user ip address
	host := getUserHost(pamh)
	if debug {
		writeLog(fmt.Sprintf("host result: %s", host))
	}

	// get tfa code
	tfaVal, err := conversation(pamh, "Your werbot code: ")
	if err != nil || !matchTfa.MatchString(tfaVal) {
		return C.PAM_AUTH_ERR
	}
	if debug {
		writeLog(fmt.Sprintf("tfaVal result: %s", tfaVal))
	}

	//
	origEUID := os.Geteuid()
	if os.Getuid() != origEUID || origEUID == 0 {
		if !seteuid(uid) {
			if debug {
				writeLog(fmt.Sprintf("error dropping privs from %d to %d", origEUID, uid))
			}
			return C.PAM_AUTH_ERR
		}
		defer func() {
			if !seteuid(origEUID) {
				if debug {
					writeLog(fmt.Sprintf("error resetting uid to %d", origEUID))
				}
			}
		}()
	}

	user, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		if debug {
			writeLog(fmt.Sprintf("error looking for user %d", uid))
		}
		return C.PAM_AUTH_ERR
	}

	// call tfaReuqest flow here
	tfaResp := sendTfaReq(user.Username, tfaVal, host)
	if debug {
		writeLog(fmt.Sprintf("final result: %v", tfaResp))
	}

	if tfaResp == true {
		return C.PAM_SUCCESS
	}

	return C.PAM_AUTH_ERR
}

// writeLog is error or info log writer for wpam. writes log in /var/log/wpam.log
func writeLog(errval string) {
	logPath := "/var/log/wpam.log"
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	var buf []byte
	data := fmt.Sprintf("%s - %s\n", time.Now().String(), errval)
	buf = append(buf, []byte(data)...)
	_, err = file.Write(buf)
}

func main() {}
