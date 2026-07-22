package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Event struct {
	TypeOfEvent string `json:"type"` // eg CreateEvent
	Actor       `json:"actor"`
	Repo        `json:"repo"`
	Payload     json.RawMessage `json:"payload"`
	Public      bool            `json:"public"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Actor struct {
	Login string `json:"login"`
}

type Repo struct {
	Name string `json:"name"` // Repo name include username, eg. torvalds/linux
}

type CreatePayload struct {
	Ref          string `json:"ref"`      // name of ref
	RefType      string `json:"ref_type"` // eg "branch"
	MasterBranch string `json:"master_branch"`
	Description  string `json:"description"`
}

type DeletePayload struct {
	Ref     string `json:"ref"`      // name of ref
	RefType string `json:"ref_type"` // eg "branch"
}

type IssueCommentPayload struct {
	Action string `json:"action"`
	Issue  struct {
		URL string `json:"html_url"`
	} `json:"issue"`
}

type ForkPayload struct {
	Action string `json:"action"`
}

type PushPayload struct {
	Ref string `json:"ref"`
}

type PullRequestPayload struct {
	Action string `json:"action"`
	Number int    `json:"number"`
}

func (e *Event) FormatEvent() (string, error) {
	switch e.TypeOfEvent {
	case "CreateEvent":
		var payload CreatePayload
		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			return "", err
		}

		desc := func() string {
			if payload.Description == "" {
				return ""
			}
			return " Description: " + payload.Description
		}()

		if payload.Ref == payload.MasterBranch {
			return "- Created a new repository (" + e.Name + ")" + desc, nil
		}
		return fmt.Sprintf("- Created a new %v in %v (%v%v)", payload.RefType, e.Name, payload.Ref, desc), nil

	case "DeleteEvent":
		var payload DeletePayload
		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			return "", err
		}

		return fmt.Sprintf("- Deleted %v %v in %v", payload.RefType, payload.Ref, e.Name), nil

	case "IssueCommentEvent":
		var payload IssueCommentPayload
		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			return "", err
		}

		if payload.Action != "created" {
			return "Unknown", nil
		}

		return fmt.Sprintf("- Commented on a issue (%v)", payload.Issue.URL), nil

	case "ForkEvent":
		var payload ForkPayload
		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			return "", err
		}

		if payload.Action != "forked" {
			return "Unknown (fork)", nil
		}

		return fmt.Sprintf("- Forked a repository (%v)", e.Name), nil

	case "PushEvent":
		var payload PushPayload
		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			return "", err
		}

		return fmt.Sprintf("- Pushed commits to %v (on %v)", e.Name, payload.Ref), nil

	case "PullRequestEvent":
		var payload PullRequestPayload
		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			return "", err
		}

		return fmt.Sprintf("- %v pull request in %v (number %v)", strings.ToUpper(payload.Action[:1])+payload.Action[1:], e.Name, strconv.Itoa(payload.Number)), nil

	case "WatchEvent":
		return fmt.Sprintf("- Starred %v", e.Name), nil

	default:
		return "Unknown", nil
	}
}

func GetEventsFromURL(url string) ([]Event, error) {
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

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	return events, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Invalid use")
		os.Exit(1)
	}

	user := os.Args[1]
	events, err := GetEventsFromURL("https://api.github.com/users/" + user + "/events")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(events) == 0 {
		fmt.Println("The user doesn't have any activity")
	}

	for _, e := range events {
		action, err := e.FormatEvent()
		if err != nil {
			fmt.Println("error")
			continue
		}
		fmt.Println(action)
	}
}
