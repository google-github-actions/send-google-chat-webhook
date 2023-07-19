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
	"net/http"
	"os"
	"time"

	"github.com/abcxyz/pkg/cli"
)

const (
	githubContextEnv = "GITHUB_CONTEXT"
	jobContextEnv    = "JOB_CONTEXT"
)

type WorkflowNotificationCommand struct {
	cli.BaseCommand
	flagWebhookUrl string
}

func (c *WorkflowNotificationCommand) Desc() string {
	return "Send a message to a Chat space"
}

func (c *WorkflowNotificationCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

  The chat command sends messages to Chat spaces.
`
}

func (c *WorkflowNotificationCommand) Flags() *cli.FlagSet {
	set := cli.NewFlagSet()

	f := set.NewSection("Chat space options")

	f.StringVar(&cli.StringVar{
		Name:    "webhook_url",
		Aliases: []string{"w"},
		Example: "https://chat.goog...",
		Default: "",
		Target:  &c.flagWebhookUrl,
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

	ghJsonStr := os.Getenv(githubContextEnv)
	if ghJsonStr == "" {
		return fmt.Errorf("environment var %s not set", githubContextEnv)
	}
	jobJsonStr := os.Getenv(jobContextEnv)
	if jobJsonStr == "" {
		return fmt.Errorf("environment var %s not set", jobContextEnv)
	}

	ghJson := map[string]any{}
	jobJson := map[string]any{}
	if err := json.Unmarshal([]byte(ghJsonStr), &ghJson); err != nil {
		return fmt.Errorf("failed unmarshaling %s: %w", githubContextEnv, err)
	}
	if err := json.Unmarshal([]byte(jobJsonStr), &jobJson); err != nil {
		return fmt.Errorf("failed unmarshaling %s: %w", jobContextEnv, err)
	}

	b, err := messageBody(ghJson, jobJson)
	if err != nil {
		return fmt.Errorf("failed to generate message body: %w", err)
	}

	url := c.flagWebhookUrl
	fmt.Println("url: ", url)

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("creating http request failed: %w", err)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("sending http request failed: %w", err)
	}
	fmt.Println("resp: ", resp)
	defer resp.Body.Close()

	return nil
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func realMain() error {
	rootCmd := func() cli.Command {
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

	cmd := rootCmd()
	// Help output is written to stderr by default. Redirect to stdout so the
	// "Output" assertion works.
	cmd.SetStderr(os.Stdout)

	ctx := context.Background()
	err := cmd.Run(ctx, os.Args[1:])
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	return nil
}

func messageBody(ghJson, jobJson map[string]any) ([]byte, error) {
	timezoneLoc, _ := time.LoadLocation("America/Los_Angeles")

	var iconUrl string
	switch jobJson["status"] {
	case "success":
		iconUrl = "https://github.githubassets.com/favicons/favicon.png"
	default:
		iconUrl = "https://github.githubassets.com/favicons/favicon-failure.png"
	}

	jsonData := map[string]any{
		"cardsV2": map[string]any{
			"cardId": "createCardMessage",
			"card": map[string]any{
				"header": map[string]any{
					"title":    fmt.Sprintf("GitHub workflow %s", jobJson["status"]),
					"subtitle": fmt.Sprintf("Workflow: <b>%s</b>", ghJson["workflow"]),
					"imageUrl": iconUrl,
				},
				"sections": []any{
					map[string]any{
						// "header":                    "This is the section header",
						"collapsible":               true,
						"uncollapsibleWidgetsCount": 1,
						"widgets": []map[string]any{
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"iconUrl": "https://fonts.gstatic.com/s/i/short-term/release/googlesymbols/quick_reference/default/48px.svg",
									},
									"text": fmt.Sprintf("<b>Ref:</b> %s", ghJson["ref"]),
								},
							},
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"knownIcon": "PERSON",
									},
									"text": fmt.Sprintf("<b>Run by:</b> %s", ghJson["triggering_actor"]),
								},
							},
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"knownIcon": "CLOCK",
									},
									"text": fmt.Sprintf("<b>Pacific:</b> %s", time.Now().In(timezoneLoc).Format(time.DateTime)),
								},
							},
							{
								"decoratedText": map[string]any{
									"startIcon": map[string]any{
										"knownIcon": "CLOCK",
									},
									"text": fmt.Sprintf("<b>UTC:</b> %s", time.Now().UTC().Format(time.DateTime)),
								},
							},
							{
								"buttonList": map[string]any{
									"buttons": []any{
										map[string]any{
											"text": "Open",
											"onClick": map[string]any{
												"openLink": map[string]any{
													"url": fmt.Sprintf("https://github.com/%s/actions/runs/%s",
														ghJson["repository"], ghJson["run_id"]),
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

	return json.Marshal(jsonData)
}
