package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type EventType int

const (
	Create EventType = iota
	Delete
	Discussion
	Issues
	IssueComment
	Fork
	Push
	PullRequest
	Release
	Watch
)

type Event struct {
	User string
	Repo string
}

func GetJSONFromURL(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Returned status: %v", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)

	var buff bytes.Buffer
	err = json.Indent(&buff, body, "", "  ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return buff.Bytes(), nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Invalid use")
		os.Exit(1)
	}

	user := os.Args[1]
	jsonData, err := GetJSONFromURL("https://api.github.com/users/" + user + "/events")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}
