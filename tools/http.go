package tools

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
)

func PostToDiscord(jsonData []byte) {

	webhook := viper.GetString("ctftime.discord-webhook")
	client := &http.Client{}
	req, _ := http.NewRequest("POST", webhook, bytes.NewBuffer(jsonData))

	//Prepare headers. Discord does not accept when not receiving a user-agent
	req.Header = http.Header{
		"user-agent":   []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.87 Safari/537.36"},
		"Content-Type": []string{"application/json; charset=UTF-8"},
	}
	//Send Request
	resp, _ := client.Do(req)

	//Check for status code 200 else exit program
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		WarningLogger.Printf("%s %s returned HTTP %d\n", resp.Request.Method, resp.Request.URL, resp.StatusCode)
		os.Exit(0)
	}

	defer resp.Body.Close()

	InfoLogger.Printf("[*] Status code: %d %s\n", resp.StatusCode, jsonData)
	//Sleep for not throttle the Discord webhook else it returns 429
	time.Sleep(1 * time.Second)
}

func GetDataFromUrl(baseURL string) ([]byte, error) {

	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseURL, nil)

	//Prepare headers. Some websites want a user-agent else it returns an error.
	req.Header = http.Header{
		"user-agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.87 Safari/537.36"}, //add user agent else website returns a 403
	}

	resp, _ := client.Do(req)

	//Check for status code 200 else exit program
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		WarningLogger.Printf("%s %s returned HTTP %d\n", resp.Request.Method, resp.Request.URL, resp.StatusCode)
		os.Exit(0)
	}
	defer resp.Body.Close() //close all connections

	body, _ := io.ReadAll(resp.Body)
	return body, nil
}
