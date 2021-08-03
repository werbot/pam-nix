package main

/*
#include <pwd.h>
#include <security/pam_appl.h>
#include <security/pam_modules.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

#ifdef __APPLE__
  #include <sys/ptrace.h>
#elif __linux__
  #include <sys/prctl.h>
#endif
#define PAM_CONST const

typedef const struct pam_message** MsgsT;
typedef struct pam_response* RespsT;

char *string_from_argv(int i, char **argv) {
  return strdup(argv[i]);
}

char *get_user(pam_handle_t *pamh) {
  if (!pamh)
    return NULL;
  int pam_err = 0;
  const char *user;
  if ((pam_err = pam_get_item(pamh, PAM_USER, (const void**)&user)) != PAM_SUCCESS)
    return NULL;
  return strdup(user);
}

int get_uid(char *user) {
  if (!user)
    return -1;
  struct passwd pw, *result;
  char buf[8192];
  int i = getpwnam_r(user, &pw, buf, sizeof(buf), &result);
  if (!result || i != 0)
    return -1;
  return pw.pw_uid;
}

int change_euid(int uid) {
  return seteuid(uid);
}

char *get_host(pam_handle_t *pamh){
    if (!pamh)
        return NULL;
    int pam_err = 0;
    const char *pamRHost;
    if ((pam_err = pam_get_item(pamh, PAM_RHOST, (const void **)&pamRHost) != PAM_SUCCESS))
        return NULL;
    return strdup(pamRHost);
}

int do_conv(pam_handle_t* pamh, char* msg, RespsT* resp) {
	int err;
	struct pam_message _msg = { .msg_style = PAM_PROMPT_ECHO_ON, .msg = msg };
	struct pam_message *msgs = &_msg;
	struct pam_conv* conv;
	err = pam_get_item(pamh, PAM_CONV, (const void**)&conv);
	if(err != PAM_SUCCESS)
		return err;
	return conv->conv(1, (const MsgsT)&msgs, resp, conv->appdata_ptr);
}

int disable_ptrace(){
#ifdef __APPLE__
    return ptrace(PT_DENY_ATTACH, 0, 0, 0);
#elif __linux__
    return prctl(PR_SET_DUMPABLE, 0);
#endif
    return 1;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func seteuid(uid int) bool {
	return C.change_euid(C.int(uid)) == C.int(0)
}

func disablePtrace() bool {
	return C.disable_ptrace() == C.int(0)
}

func conversation(pamh *C.pam_handle_t, msg string) (string, error) {
	var resp C.RespsT
	code := C.do_conv(pamh, C.CString(msg), &resp)
	if code != C.PAM_SUCCESS || resp == nil {
		return "", fmt.Errorf("PAM_CONV_ERR")
	}
	ret := C.GoString((*resp).resp)
	C.free(unsafe.Pointer((*resp).resp))
	return ret, nil
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

func getUserHost(pamh *C.pam_handle_t) string {
	host := C.get_host(pamh)
	defer C.free(unsafe.Pointer(host))
	return C.GoString(host)
}
