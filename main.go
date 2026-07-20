package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type EventType int

const (
	Create EventType = iota
	Delete
	Issues
	IssueComment
	Fork
	Push
	PullRequest
	Watch
)

type Event struct {
	TypeOfEvent string `json:"type"` // eg CreateEvent
	EventType
	Actor     `json:"actor"`
	Repo      `json:"repo"`
	Payload   json.RawMessage `json:"payload"`
	Public    bool            `json:"public"`
	CreatedAt time.Time       `json:"created_at"`
}

type Actor struct {
	Login string `json:"login"`
}

type Repo struct {
	Name string `json:"name"` // Repo name include username, eg. torvalds/linux
}

func (e *Event) FormatEvent() string {
	switch e.TypeOfEvent {
	case "CreateEvent":
		e.EventType = Create

		var payload struct {
			Ref          string `json:"ref"`      // name of ref
			RefType      string `json:"ref_type"` // eg "branch"
			MasterBranch string `json:"master_branch"`
			Description  string `json:"description"`
		}

		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			panic(err)
		}

		desc := func() string {
			if payload.Description == "" {
				return ""
			}
			return " Description: " + payload.Description
		}()

		if payload.Ref == payload.MasterBranch {
			return "- Created a new repository (" + e.Name + ")" + desc
		}
		return "- Created a new " + payload.RefType + " in " + e.Name + " (" + payload.Ref + desc + ")"

	case "DeleteEvent":
		e.EventType = Delete

		var payload struct {
			Ref     string `json:"ref"`      // name of ref
			RefType string `json:"ref_type"` // eg "branch"
		}

		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			panic(err)
		}

		return "- Deleted " + payload.RefType + " " + payload.Ref + " in " + e.Name

	case "IssueCommentEvent":
		e.EventType = IssueComment

		var payload struct {
			Action string `json:"action"`
			Issue  struct {
				URL string `json:"html_url"`
			} `json:"issue"`
		}

		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			panic(err)
		}

		if payload.Action != "created" {
			return "Unknown"
		}

		return "- Commented on a issue (" + payload.Issue.URL + ")"

	case "ForkEvent":
		e.EventType = Fork

		var payload struct {
			Action string `json:"action"`
		}

		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			panic(err)
		}

		if payload.Action != "forked" {
			return "Unknown (fork)"
		}

		return "- Forked a repository (" + e.Name + ")"

	case "PushEvent":
		e.EventType = Push

		var payload struct {
			Ref string `json:"ref"`
		}

		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			panic(err)
		}

		return "- Pushed commits to " + e.Name + " (on " + payload.Ref + ")"

	case "PullRequestEvent":
		e.EventType = PullRequest

		var payload struct {
			Action string `json:"action"`
			Number int    `json:"number"`
		}

		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			panic(err)
		}

		return "- " + string(payload.Action[0]-('a'-'A')) + payload.Action[1:] + " pull request in " + e.Name + " (number " + strconv.Itoa(payload.Number) + ")"

	case "WatchEvent":
		e.EventType = Watch

		return "- Starred " + e.Name

	default:
		return "Unknown"
	}
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
	if len(jsonData) == 2 {
		fmt.Println("The user doesn't have any activity in the last 30 days")
	}

	var events []Event
	if err := json.Unmarshal(jsonData, &events); err != nil {
		panic(err)
	}

	for _, e := range events {
		fmt.Println(e.FormatEvent())
	}
}
