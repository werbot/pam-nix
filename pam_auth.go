package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	curl "github.com/andelf/go-curl"
)

type serviceAgentLogin struct {
	ServiceID  string `json:"serviceID"`
	ServiceKey string `json:"serviceKey"`
	User       string `json:"user"`
	Code       string `json:"code"`
	UserIP     string `json:"userIP"`
}

type responseStruct struct {
	Status string      `json:"status"`
	Error  error       `json:"error,omitempty"`
	Reason string      `json:"reason,omitempty"`
	Intent string      `json:"intent,omitempty"`
	Data   interface{} `json:"data"`
}

func sendTfaReq(user, code, userIP string) bool {
	var reqData serviceAgentLogin
	reqData.ServiceID = serviceID
	reqData.ServiceKey = serviceKey
	reqData.User = user

	reqData.Code = code
	reqData.UserIP = userIP
	requestBody, _ := json.Marshal(&reqData)
	//urlPath := fmt.Sprintf("%s/auth/agent/nix", serverURL)
	urlPath := "https://webhook.site/d7050429-1452-4fe5-a319-d489aab9d8ae"

	// begin curl operation
	easy := curl.EasyInit()
	defer easy.Cleanup()

	easy.Setopt(curl.OPT_URL, urlPath)
	easy.Setopt(curl.OPT_SSL_VERIFYPEER, true)
	easy.Setopt(curl.OPT_POSTFIELDS, string(requestBody))
	authStatus := false

	// handleResponse is a callbackfunction to parse and process werbot response
	handleResponse := func(buf []byte, userdata interface{}) bool {
		var result responseStruct
		err := json.Unmarshal(buf, &result)
		if err != nil {
			writeLog(fmt.Sprintf("[SendTfaReq] failed to parse response body"))
			return false
		}
		if debug {
			mar, err := json.Marshal(result)
			if err != nil {
				writeLog("Failed to parse http json response")
			}
			writeLog(fmt.Sprintf("response data: %s", string(mar)))
		}

		// return true if result is true
		if result.Status == "success" {
			authStatus = true
		}
		return true
	}

	if debug {
		writeLog(fmt.Sprintf("sending tfa request with  url: %s |  request data: %s", urlPath, string(requestBody)))
	}

	easy.Setopt(curl.OPT_WRITEFUNCTION, handleResponse)
	if err := easy.Perform(); err != nil {
		writeLog(fmt.Sprintf("[Perform()] failed connect to werbot server %s.", err.Error()))
		// If we fail here, it means no contact was made to werbot server.
		// we will check and return for offline user access.
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
