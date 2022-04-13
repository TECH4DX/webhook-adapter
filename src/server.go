package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	gitee_utils "github.com/TECH4DX/webhook-adapter/src/gitee-utils"
	"gitee.com/openeuler/go-gitee/gitee"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Webhook adapter version: 0.55 \n")
	fmt.Fprint(w, "Webhook received !!! \n")
	eventType, _, payload, ok, statusCode := gitee_utils.ValidateWebhook(w, r)
	if !ok {
		fmt.Fprint(w, "Validate webhook failed, status code: %d", statusCode)
		return
	}
	fmt.Fprint(w, "Validate webhook finished, status code: %d", statusCode)

	switch eventType {
	case "Issue Hook":
		var ie gitee.IssueEvent
		if err := json.Unmarshal(payload, &ie); err != nil {
			return
		}
		if err := checkRepository(payload, ie.Repository); err != nil {
			return
		}
		go handleIssueEvent(&ie)
	case "Note Hook":
		var ne gitee.NoteEvent
		if err := json.Unmarshal(payload, &ne); err != nil {
			return
		}
		go handleCommentEvent(&ne)
	case "Merge Request Hook":
		var pre gitee.PullRequestEvent
		if err := json.Unmarshal(payload, &pre); err != nil {
			return
		}
		go handlePullRequestEvent(&pre)
	case "Push Hook":
		var pe gitee.PushEvent
		if err := json.Unmarshal(payload, &pe); err != nil {
			return
		}
		go handlePushEvent(&pe)
	default:
		return
	}
}

func handleIssueEvent(i *gitee.IssueEvent) {
	return
}

func handleCommentEvent(i *gitee.NoteEvent) {
	switch *(i.NoteableType) {
	case "Issue":
		go handleIssueCommentEvent(i)
	case "PullRequest":
		go handlePRCommentEvent(i)
	default:
		return
	}
}

func handlePullRequestEvent(i *gitee.PullRequestEvent) {
	return
}

func handlePushEvent(i *gitee.PushEvent) {
	return
}

func handleIssueCommentEvent(i *gitee.NoteEvent) {
	return
}

func handlePRCommentEvent(i *gitee.NoteEvent) {
	return
}

func checkRepository(payload []byte, rep *gitee.ProjectHook) error {
	if rep == nil {
		return fmt.Errorf("event repository is empty,payload: %s", string(payload))
	}
	return nil
}

func main() {
	http.HandleFunc("/", ServeHTTP)
	http.ListenAndServe(":8008", nil)
}
