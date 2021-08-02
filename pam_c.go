package main

/*
#include <security/pam_appl.h>
#include <security/pam_modules.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#ifdef __APPLE__
  #include <sys/ptrace.h>
#elif __linux__
  #include <sys/prctl.h>
#endif

typedef const struct pam_message** MsgsT;
typedef struct pam_response* RespsT;

char *string_from_argv(int i, char **argv) {
    return strdup(argv[i]);
}

char *get_username(pam_handle_t *pamh){
    if (!pamh)
        return NULL;
    int pam_err = 0;
    const char *username;
    if ((pam_err = pam_get_user(pamh, &username, "login: ")) != PAM_SUCCESS)
        return NULL;
    return strdup(username);
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
