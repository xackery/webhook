package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"net/http"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/xackery/webhook/config"
)

var (
	mu       sync.Mutex
	triggers []*Trigger
	Version  string
)

// Trigger represents a webhook trigger
type Trigger struct {
	event      *config.Event
	gitWebhook *github.Webhook
}

func main() {
	err := run()
	if err != nil {
		fmt.Println("Failed to run:", err)
		for {
			time.Sleep(time.Second)
		}
	}
}

func run() error {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	isValid := true
	for _, evt := range cfg.Events {
		if evt.WebhookToken == "" {
			fmt.Println(evt.Name, "Webhook token not set")
			isValid = false
			continue
		}
		if evt.DiscordWebhook == "" {
			fmt.Println(evt.Name, "Discord webhook not set")
			isValid = false
			continue
		}
		gitWebhook, err := github.New(github.Options.Secret(evt.WebhookToken))
		if err != nil {
			return fmt.Errorf(evt.Name, "github new: %w", err)
		}
		trigger := &Trigger{
			event:      &evt,
			gitWebhook: gitWebhook,
		}
		triggers = append(triggers, trigger)
	}
	if !isValid {
		return fmt.Errorf("invalid configuration")
	}

	fmt.Println("webhook", Version, "listening on :3000")
	http.HandleFunc("/webhooks", hookRequest)
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

func hookRequest(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	for _, trigger := range triggers {

		payload, err := trigger.gitWebhook.Parse(r, github.PushEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				fmt.Printf("Ignoring event %s %s: %s\n", trigger.event.Name, github.Event(r.Header.Get("X-GitHub-Event")), err)
				continue
			}
			fmt.Printf("Failed to parse hook %s: %s", trigger.event.Name, err)
			continue
		}

		switch payload.(type) {
		case github.PingPayload:
			fmt.Println("Ping event received")
		case github.PushPayload:
			//req.Repository.Name
			//repoName := req.Repository.Name
			// if req.CheckRun.Status != "completed" {
			// 	fmt.Printf("Ignoring incomplete check run for %s: %s\n", repoName, req.CheckRun.Status)
			// 	return
			// }
			// if req.CheckRun.Conclusion != "success" {
			// 	sendDiscordMessage(trigger, fmt.Sprintf("Failed to deploy %s %s: Github Action resulted in %s", trigger.event.Name, repoName, req.CheckRun.Conclusion))
			// 	fmt.Printf("Ignoring failed check run for %s: %s\n", repoName, req.CheckRun.Conclusion)
			// 	return
			// }

			go deploy(trigger)
		}
	}
}

func deploy(trigger *Trigger) {
	start := time.Now()
	fmt.Println("Deploying:", trigger.event.Name)
	result, err := onDeploy(trigger)
	if err != nil {
		result += fmt.Sprintf("Failed to deploy %s: %s in %0.2f seconds\n", trigger.event.Name, err, time.Since(start).Seconds())
	} else {
		result += fmt.Sprintf("Deployed %s successfully in %0.2f seconds", trigger.event.Name, time.Since(start).Seconds())
	}
	fmt.Println(result)

	err = sendDiscordMessage(trigger, result)
	if err != nil {
		fmt.Println("Failed to send discord message:", err)
	}

}

func onDeploy(trigger *Trigger) (string, error) {

	err := os.Chdir(trigger.event.Path)
	if err != nil {
		return "", fmt.Errorf("chdir %s: %w", trigger.event.Path, err)
	}

	// buffer both stdout and stderr
	var out bytes.Buffer
	cmd := exec.Command(trigger.event.Command, trigger.event.Args...)
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		result := fmt.Sprintf("**Deploying %s FAILED**:\n```\n", trigger.event.Name)
		result += out.String()
		result += "\n```"
		fmt.Println("make deploy failed:")
		fmt.Println(out.String())
		return result, fmt.Errorf("make deploy: %w", err)
	}

	fmt.Println(out.String())
	return "", nil // change "" to out.String() to send output to discord
}

type discordWebhook struct {
	Content string `json:"content"`
}

func sendDiscordMessage(trigger *Trigger, message string) error {
	payload := discordWebhook{
		Content: message,
	}

	w := &bytes.Buffer{}
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		return fmt.Errorf("encode payload: %w", err)
	}

	req, err := http.NewRequest("POST", trigger.event.DiscordWebhook, w)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	return nil
}
