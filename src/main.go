// Copyright 2023 The Authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abcxyz/pkg/cli"
)

const (
	githubContextEnvKey               = "GITHUB_CONTEXT"
	jobContextEnvKey                  = "JOB_CONTEXT"
	githubContextRefKey               = "ref"
	githubContextRepositoryKey        = "repository"
	githubContextTriggeringActorKey   = "triggering_actor"
	githubContextEventObjectActionKey = "action"
	githubContextEventNameKey         = "event_name"
	githubContextEventKey             = "event"
	githubContextEventURLKey          = "html_url"
	githubEventContenntCreatedAtKey   = "created_at"
	githubContextServerURLKey         = "server_url"
)

const (
	successHeaderIconURL = "https://github.githubassets.com/favicons/favicon.png"
	failureHeaderIconURL = "https://github.githubassets.com/favicons/favicon-failure.png"
	widgetRefIconURL     = "https://fonts.gstatic.com/s/i/short-term/release/googlesymbols/quick_reference/default/48px.svg"
)

var rootCmd = func() cli.Command {
	return &cli.RootCommand{
		Name: "send-google-chat-webhook",
		Commands: map[string]cli.CommandFactory{
			"chat": func() cli.Command {
				return &cli.RootCommand{
					Name:        "workflownotification",
					Description: "notification for workflow",
					Commands: map[string]cli.CommandFactory{
						"workflownotification": func() cli.Command {
							return &WorkflowNotificationCommand{}
						},
					},
				}
			},
		},
	}
}

type WorkflowNotificationCommand struct {
	cli.BaseCommand
	flagWebhookURL string
}

func (c *WorkflowNotificationCommand) Desc() string {
	return "Send a message to a Google Chat space"
}

func (c *WorkflowNotificationCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

  The chat command sends messages to Google Chat spaces.
`
}

func (c *WorkflowNotificationCommand) Flags() *cli.FlagSet {
	set := c.NewFlagSet()

	f := set.NewSection("COMMAND OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "webhook-url",
		Example: "https://chat.googleapis.com/v1/spaces/<SPACE_ID>/messages?key=<KEY>&token=<TOKEN>",
		Target:  &c.flagWebhookURL,
		Usage:   `Webhook URL from google chat`,
	})

	return set
}

func (c *WorkflowNotificationCommand) Run(ctx context.Context, args []string) error {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	args = f.Args()
	if len(args) != 0 {
		return fmt.Errorf("expected 0 arguments, got %q", args)
	}

	ghJSONStr := c.GetEnv(githubContextEnvKey)
	if ghJSONStr == "" {
		return fmt.Errorf("environment var %s not set", githubContextEnvKey)
	}
	jobJSONStr := c.GetEnv(jobContextEnvKey)
	if jobJSONStr == "" {
		return fmt.Errorf("environment var %s not set", jobContextEnvKey)
	}

	ghJSON := map[string]any{}
	jobJSON := map[string]any{}
	if err := json.Unmarshal([]byte(ghJSONStr), &ghJSON); err != nil {
		return fmt.Errorf("failed unmarshaling %s: %w", githubContextEnvKey, err)
	}
	if err := json.Unmarshal([]byte(jobJSONStr), &jobJSON); err != nil {
		return fmt.Errorf("failed unmarshaling %s: %w", jobContextEnvKey, err)
	}

	b, err := generateRequestBody(generateMessageBodyContent(ghJSON, jobJSON, time.Now()))
	if err != nil {
		return fmt.Errorf("failed to generate message body: %w", err)
	}

	url := c.flagWebhookURL

	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("creating http request failed: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("sending http request failed: %w", err)
	}
	defer resp.Body.Close()

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read")
		}
		bodyString := string(bodyBytes)
		return fmt.Errorf("unexpected HTTP status code %d (%s)\n got body: %s", got, http.StatusText(got), bodyString)
	}

	return nil
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer done()

	if err := realMain(ctx); err != nil {
		done()
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func realMain(ctx context.Context) error {
	return rootCmd().Run(ctx, os.Args[1:]) //nolint:wrapcheck // Want passthrough
}

// messageBodyContent defines the necessary fields for generating the request body.
type messageBodyContent struct {
	title           string
	subtitle        string
	ref             string
	triggeringActor string
	timestamp       string
	clickURL        string
	headerIconURL   string
	eventName       string
	repo            string
}

// generateMessageBodyContent returns messageBodyContent for generating the request body.
// using currentTimestamp as a input is for easier testing on default case.
func generateMessageBodyContent(ghJSON, jobJSON map[string]any, currentTimeStamp time.Time) *messageBodyContent {
	event, ok := ghJSON[githubContextEventKey].(map[string]any)
	if !ok {
		event = map[string]any{}
	}
	eventName := getMapFieldStringValue(ghJSON, githubContextEventNameKey)
	switch eventName {
	case "issues":
		issueContent, ok := event["issue"].(map[string]any)
		if !ok {
			issueContent = map[string]any{}
		}
		return &messageBodyContent{
			title:           fmt.Sprintf("A issue is %s", getMapFieldStringValue(event, githubContextEventObjectActionKey)),
			subtitle:        fmt.Sprintf("Issue title: <b>%s</b>", getMapFieldStringValue(issueContent, "title")),
			ref:             getMapFieldStringValue(ghJSON, githubContextRefKey),
			triggeringActor: getMapFieldStringValue(ghJSON, githubContextTriggeringActorKey),
			timestamp:       getMapFieldStringValue(issueContent, githubEventContenntCreatedAtKey),
			clickURL:        getMapFieldStringValue(issueContent, githubContextEventURLKey),
			eventName:       "issue",
			repo:            getMapFieldStringValue(ghJSON, githubContextRepositoryKey),
			headerIconURL:   successHeaderIconURL,
		}
	case "release":
		releaseContent, ok := event["release"].(map[string]any)
		if !ok {
			releaseContent = map[string]any{}
		}
		return &messageBodyContent{
			title:           fmt.Sprintf("A release is %s", getMapFieldStringValue(event, githubContextEventObjectActionKey)),
			subtitle:        fmt.Sprintf("Release name: <b>%s</b>", getMapFieldStringValue(releaseContent, "name")),
			ref:             getMapFieldStringValue(ghJSON, githubContextRefKey),
			triggeringActor: getMapFieldStringValue(ghJSON, githubContextTriggeringActorKey),
			timestamp:       getMapFieldStringValue(releaseContent, githubEventContenntCreatedAtKey),
			clickURL:        getMapFieldStringValue(releaseContent, githubContextEventURLKey),
			eventName:       "release",
			repo:            getMapFieldStringValue(ghJSON, githubContextRepositoryKey),
			headerIconURL:   successHeaderIconURL,
		}
	default:
		res := &messageBodyContent{
			title:           fmt.Sprintf("GitHub workflow %s", getMapFieldStringValue(jobJSON, "status")),
			subtitle:        fmt.Sprintf("Workflow: <b>%s</b>", getMapFieldStringValue(ghJSON, "workflow")),
			ref:             getMapFieldStringValue(ghJSON, githubContextRefKey),
			triggeringActor: getMapFieldStringValue(ghJSON, githubContextTriggeringActorKey),
			// The key for getting timestamp is different in differnet triggering event
			// a simple work around is using the new timestamp.
			timestamp: currentTimeStamp.UTC().Format(time.RFC3339),
			clickURL:  fmt.Sprintf("%s/%s/actions/runs/%s", getMapFieldStringValue(ghJSON, githubContextServerURLKey), getMapFieldStringValue(ghJSON, githubContextRepositoryKey), getMapFieldStringValue(ghJSON, "run_id")),
			eventName: "workflow",
			repo:      getMapFieldStringValue(ghJSON, githubContextRepositoryKey),
		}
		v, ok := jobJSON["status"]
		if !ok || v == "failure" || v == "canceled" {
			res.headerIconURL = failureHeaderIconURL
		} else {
			res.headerIconURL = successHeaderIconURL
		}
		return res
	}
}

// generateRequestBody returns the body of the request.
func generateRequestBody(m *messageBodyContent) ([]byte, error) {
	jsonData := map[string]any{
		"cardsV2": map[string]any{
			"cardId": "createCardMessage",
			"card": map[string]any{
				"header": map[string]any{
					"title":    m.title,
					"subtitle": m.subtitle,
					"imageUrl": m.headerIconURL,
				},
				"sections": []any{
					map[string]any{
						"collapsible":               true,
						"uncollapsibleWidgetsCount": 1,
						"widgets": []map[string]any{
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"iconUrl": widgetRefIconURL,
									},
									"text": fmt.Sprintf("<b>Repo: </b> %s", m.repo),
								},
							},
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"iconUrl": widgetRefIconURL,
									},
									"text": fmt.Sprintf("<b>Ref: </b> %s", m.ref),
								},
							},
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"knownIcon": "PERSON",
									},
									"text": fmt.Sprintf("<b>Actor: </b> %s", m.triggeringActor),
								},
							},
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"knownIcon": "CLOCK",
									},
									"text": fmt.Sprintf("<b>UTC: </b> %s", m.timestamp),
								},
							},
							{
								"buttonList": map[string]any{
									"buttons": []any{
										map[string]any{
											"text": fmt.Sprintf("Open %s", m.eventName),
											"onClick": map[string]any{
												"openLink": map[string]any{
													"url": m.clickURL,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	res, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error marshal jsonData: %w", err)
	}
	return res, nil
}

// getMapFieldStringValue get value from a map[sting]any map.
// And convert it into string type. Return empty if the conversion failed.
// The keys should all exist as they are popluated by github, to simple the
// code on unnecessary error handling, a empty string is returned.
func getMapFieldStringValue(m map[string]any, key string) string {
	v, ok := m[key].(string)
	if !ok {
		v = ""
	}
	return v
}
