package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"gitee.com/openeuler/go-gitee/gitee"
	gitee_utils "github.com/TECH4DX/webhook-adapter/src/gitee-utils"
	"gopkg.in/go-playground/webhooks.v5/github"
	// "github.com/google/go-github/github"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Webhook adapter version: 0.10 \n")
	fmt.Fprint(w, "Webhook received! \n")

	eventType, _, payload, ok, statusCode := gitee_utils.ValidateWebhook(w, r)
	if !ok {
		fmt.Fprint(w, "Validate webhook failed, status code: ", statusCode)
		return
	}
	fmt.Fprint(w, "Validate webhook finished, status code: ", statusCode)

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
	// TODO: Implement pull request event adapter
	return
}

func handlePushEvent(i *gitee.PushEvent) {
	var g github.PushPayload
	apiURL := os.Getenv("REMOTE_WEBHOOK_URL")
	pushEventAdapter(i, &g)
	res, err := sentEventWebhook(&g, apiURL)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
	return
}

func handleIssueCommentEvent(i *gitee.NoteEvent) {
	// TODO: Implement issue comment event adapter
	return
}

func handlePRCommentEvent(i *gitee.NoteEvent) {
	// TODO: Implement PR comment event adapter
	return
}

func pushEventAdapter(i *gitee.PushEvent, g *github.PushPayload) {
	// TODO: make more attributes consistent
	// *(g.PushID) = int64(i.Pusher.Id)
	g.Repository.HTMLURL = i.Repository.HtmlUrl
	g.Ref = *i.Ref
	g.Repository.DefaultBranch = i.Project.DefaultBranch
}

func sentEventWebhook(g *github.PushPayload, apiURL string) (string, error) {
	var retries = 3
	var post, err = json.Marshal(g)
	if err != nil {
		return "Bad Github Push Event!", err
	}

	var payload = []byte(post)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	req.Close = true
	if err != nil {
		return "Bad Http Request!", err
	}

	header := http.Header{
		"X-Github-Event": []string{"push"},
	}
	req.Header = header
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	for retries > 0 {
		resp, err := client.Do(req)
		if err != nil {
			retries--
			fmt.Printf("POST action failed, get response:\n %+v , \n get error:%+v \n retring...\n", resp, err)
		} else {
			defer resp.Body.Close()
			fmt.Println("Response Status:", resp.Status)
			fmt.Println("Response Header:", resp.Header)
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("Response Body", string(body))
			return "POST Succeed!", nil
		}
	}
	return "POST failed after retries!", errors.New("post failed")
}

func checkRepository(payload []byte, rep *gitee.ProjectHook) error {
	if rep == nil {
		return fmt.Errorf("event repository is empty,payload: %s", string(payload))
	}
	return nil
}

func main() {
	http.HandleFunc("/", ServeHTTP)
	http.ListenAndServe(":8080", nil)
}
