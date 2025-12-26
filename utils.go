package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	apiEndpointPath = "/auth/agent/nix"
	tfaMethodTOTP   = "totp"
	tfaMethodU2F    = "U2F"
	httpTimeout     = 10 * time.Second
	logPath         = "/var/log/wpam.log"
)

var (
	logFile  *os.File
	logMutex sync.Mutex
	logInit  sync.Once
)

// serviceAgentLogin represents the authentication request structure
type serviceAgentLogin struct {
	ServiceID  string `json:"serviceID"`
	ServiceKey string `json:"serviceKey"`
	User       string `json:"user"`
	TfaMethod  string `json:"tfaMethod"`
	TotpCode   string `json:"totpCode"`
	UserIP     string `json:"userIP"`
	WID        string `json:"wID"`
}

// responseStruct represents the API response structure
type responseStruct struct {
	Status string      `json:"status"`
	Error  error       `json:"error,omitempty"`
	Reason string      `json:"reason,omitempty"`
	Intent string      `json:"intent,omitempty"`
	Data   interface{} `json:"data"`
}

// sendTfaReq sends TFA authentication request to Werbot API
// wID can be empty for account-only checks (will use user as wID)
func sendTfaReq(cfg *config, user, wID, totpCode, userIP string) bool {
	if user == "" || userIP == "" {
		writeLog(fmt.Sprintf("ERROR: Invalid parameters - user: %s, userIP: %s", user, userIP))
		return false
	}

	if wID == "" {
		wID = user
	}

	tfaMethod := tfaMethodU2F
	if totpCode != "" && len(totpCode) == totpCodeLength {
		tfaMethod = tfaMethodTOTP
	}

	reqData := serviceAgentLogin{
		ServiceID:  cfg.serviceID,
		ServiceKey: cfg.serviceKey,
		User:       user,
		WID:        wID,
		TfaMethod:  tfaMethod,
		TotpCode:   totpCode,
		UserIP:     userIP,
	}

	requestBody, err := json.Marshal(&reqData)
	if err != nil {
		writeLog(fmt.Sprintf("ERROR: Failed to marshal request data: %v", err))
		return false
	}

	urlPath := fmt.Sprintf("https://%s%s", cfg.serverURL, apiEndpointPath)

	if cfg.debug {
		logSafeRequest(urlPath, tfaMethod, reqData)
	}

	// Create HTTP client with timeout and TLS configuration
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.insecureSkipVerify,
		},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   httpTimeout,
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(requestBody))
	if err != nil {
		writeLog(fmt.Sprintf("ERROR: Failed to create HTTP request: %v", err))
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform HTTP request
	resp, err := client.Do(req)
	if err != nil {
		writeLog(fmt.Sprintf("ERROR: Failed to connect to Werbot server: %v", err))
		// If connection fails, check for offline user access
		if offlineUsersParse(user, cfg.offlineUsers) {
			writeLog(fmt.Sprintf("INFO: Allowing offline access for user %s", user))
			return true
		}
		return false
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		writeLog(fmt.Sprintf("ERROR: Failed to read response body: %v", err))
		return false
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		writeLog(fmt.Sprintf("ERROR: HTTP request failed with status %d: %s", resp.StatusCode, string(body)))
		return false
	}

	// Parse response
	var result responseStruct
	if err := json.Unmarshal(body, &result); err != nil {
		writeLog(fmt.Sprintf("ERROR: Failed to parse response body: %v", err))
		return false
	}

	if cfg.debug {
		logSafeResponse(result)
	}

	// Check authentication result
	if result.Status == "success" {
		return true
	}

	writeLog(fmt.Sprintf("WARN: Authentication failed for user %s - status: %s, reason: %s", user, result.Status, result.Reason))
	return false
}

// checkAccountAccess verifies account access without TFA (for account management)
func checkAccountAccess(cfg *config, user, userIP string) bool {
	if user == "" || userIP == "" {
		return false
	}
	return sendTfaReq(cfg, user, user, "", userIP)
}

// offlineUsersParse checks if username is in the comma-separated list of offline users
func offlineUsersParse(username, usernames string) bool {
	if usernames == "" || username == "" {
		return false
	}
	for _, v := range strings.Split(usernames, ",") {
		if strings.TrimSpace(v) == username {
			return true
		}
	}
	return false
}

// logSafeRequest logs request data with sensitive fields redacted
func logSafeRequest(urlPath, method string, reqData serviceAgentLogin) {
	reqData.ServiceKey = "[REDACTED]"
	reqData.TotpCode = "[REDACTED]"
	if body, err := json.Marshal(&reqData); err == nil {
		writeLog(fmt.Sprintf("DEBUG: Sending TFA request to %s | method: %s | data: %s", urlPath, method, string(body)))
	}
}

// logSafeResponse logs response data with sensitive fields redacted
func logSafeResponse(result responseStruct) {
	if dataMap, ok := result.Data.(map[string]interface{}); ok {
		safeData := make(map[string]interface{})
		for k, v := range dataMap {
			keyLower := strings.ToLower(k)
			if strings.Contains(keyLower, "key") || strings.Contains(keyLower, "token") {
				safeData[k] = "[REDACTED]"
			} else {
				safeData[k] = v
			}
		}
		result.Data = safeData
	}
	if body, err := json.Marshal(&result); err == nil {
		writeLog(fmt.Sprintf("DEBUG: Response received: %s", string(body)))
	}
}

// initLogFile initializes the log file (called once)
func initLogFile() {
	var err error
	logFile, err = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		// Fallback to stderr if log file cannot be opened
		fmt.Fprintf(os.Stderr, "ERROR: Failed to open log file %s: %v\n", logPath, err)
		logFile = nil
	}
}

// writeLog writes log messages to /var/log/wpam.log with thread-safe access
func writeLog(message string) {
	logInit.Do(initLogFile)

	logMutex.Lock()
	defer logMutex.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("%s - %s\n", timestamp, message)

	if logFile != nil {
		if _, err := logFile.WriteString(logEntry); err != nil {
			// Fallback to stderr if write fails
			fmt.Fprintf(os.Stderr, "ERROR: Failed to write to log file: %v\n", err)
			fmt.Fprintf(os.Stderr, "%s", logEntry)
		}
	} else {
		// Fallback to stderr if log file is not available
		fmt.Fprintf(os.Stderr, "%s", logEntry)
	}
}
