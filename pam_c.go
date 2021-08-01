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

int converse( pam_handle_t *pamh, int nargs, struct pam_message **message, struct pam_response **response )
{
  int retval;
  struct pam_conv *conv;
  retval = pam_get_item( pamh, PAM_CONV, (const void **) &conv );
  if ( retval == PAM_SUCCESS) {
    retval = conv -> conv (nargs, (const struct pam_message **) message, response, conv-> appdata_ptr );
  }
  return retval;
}

char *get_username(pam_handle_t *pamh)
{
  if (!pamh)
    return NULL;
  int pam_err = 0;
  const char *username;
  if ((pam_err = pam_get_user(pamh,&username,"login: "))!=PAM_SUCCESS ) {
		return NULL ;
  }
  return strdup(username);
}

char *get_rhost(pam_handle_t *pamh)
{
  if (!pamh)
    return NULL;
     int pam_err = 0;
    const char *pamRHost ;
    	if( (pam_err = pam_get_item(pamh, PAM_RHOST, (const void **) &pamRHost) != PAM_SUCCESS)){
        return NULL ;
      }
  return strdup(pamRHost);
}

char *get_tfaval(pam_handle_t *pamh, int flags) {
  int retval;
  char *tfaVAL ;
	struct pam_message msg[1],*pmsg[1];
  struct pam_response *resp;
  char *pam_auth_err = "pam_auth_err";
  char *pam_conv_err = "pam_conv_err";
	pmsg[0] = &msg[0] ;
	msg[0].msg_style = PAM_PROMPT_ECHO_ON ;
	msg->msg = "Enter your werbot code: " ;
	resp = NULL ;
	if( (retval = converse(pamh, 1 , pmsg, &resp))!=PAM_SUCCESS ) {
		return pam_conv_err ;
	}
	if( resp ) {
		if( (flags & PAM_DISALLOW_NULL_AUTHTOK) && resp[0].resp == NULL ) {
	    free( resp );
	    return pam_auth_err;
		}
		tfaVAL = resp[0].resp;
		resp[0].resp = NULL;
  } else {
		return pam_conv_err;
	}
  return tfaVAL;
}

int disable_ptrace() {
#ifdef __APPLE__
  return ptrace(PT_DENY_ATTACH, 0, 0, 0);
#elif __linux__
  return prctl(PR_SET_DUMPABLE, 0);
#endif
  return 1;
}
*/
import "C"

func disablePtrace() bool {
	return C.disable_ptrace() == C.int(0)
}
