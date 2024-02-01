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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestGenerateRequestBody(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		ghJSON          map[string]any
		jobJSON         map[string]any
		timestamp       time.Time
		location        time.Location
		wantMessageBody map[string]any
	}{
		{
			name: "test_success_workflow",
			ghJSON: map[string]any{
				"workflow":         "test-workflow",
				"ref":              "test-ref",
				"triggering_actor": "test-triggered_actor",
				"repository":       "test-repository",
				"run_id":           "test-run-id",
			},
			jobJSON: map[string]any{
				"status": "success",
			},
			timestamp: time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC),
			wantMessageBody: map[string]any{
				"cardsV2": map[string]any{
					"cardId": "createCardMessage",
					"card": map[string]any{
						"header": map[string]any{
							"title":    fmt.Sprintf("GitHub workflow %s", "success"),
							"subtitle": fmt.Sprintf("Workflow: <b>%s</b>", "test-workflow"),
							"imageUrl": "https://github.githubassets.com/favicons/favicon.png",
						},
						"sections": []any{
							map[string]any{
								"collapsible":               true,
								"uncollapsibleWidgetsCount": float64(1),
								"widgets": []map[string]any{
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"iconUrl": widgetRefIconURL,
											},
											"text": fmt.Sprintf("<b>Repo: </b> %s", "test-repository"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"iconUrl": "https://fonts.gstatic.com/s/i/short-term/release/googlesymbols/quick_reference/default/48px.svg",
											},
											"text": fmt.Sprintf("<b>Ref: </b> %s", "test-ref"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "PERSON",
											},
											"text": fmt.Sprintf("<b>Actor: </b> %s", "test-triggered_actor"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "CLOCK",
											},
											"text": fmt.Sprintf("<b>UTC: </b> %s", time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC).UTC().Format(time.RFC3339)),
										},
									},
									{
										"buttonList": map[string]any{
											"buttons": []any{
												map[string]any{
													"text": "Open workflow",
													"onClick": map[string]any{
														"openLink": map[string]any{
															"url": fmt.Sprintf("https://github.com/%s/actions/runs/%s",
																"test-repository", "test-run-id"),
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
			},
		},
		{
			name: "test_failed_workflow",
			ghJSON: map[string]any{
				"workflow":         "test-workflow",
				"ref":              "test-ref",
				"triggering_actor": "test-triggered_actor",
				"repository":       "test-repository",
				"run_id":           "test-run-id",
			},
			jobJSON: map[string]any{
				"status": "failure",
			},
			timestamp: time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC),
			wantMessageBody: map[string]any{
				"cardsV2": map[string]any{
					"cardId": "createCardMessage",
					"card": map[string]any{
						"header": map[string]any{
							"title":    fmt.Sprintf("GitHub workflow %s", "failure"),
							"subtitle": fmt.Sprintf("Workflow: <b>%s</b>", "test-workflow"),
							"imageUrl": "https://github.githubassets.com/favicons/favicon-failure.png",
						},
						"sections": []any{
							map[string]any{
								"collapsible":               true,
								"uncollapsibleWidgetsCount": float64(1),
								"widgets": []map[string]any{
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"iconUrl": widgetRefIconURL,
											},
											"text": fmt.Sprintf("<b>Repo: </b> %s", "test-repository"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"iconUrl": "https://fonts.gstatic.com/s/i/short-term/release/googlesymbols/quick_reference/default/48px.svg",
											},
											"text": fmt.Sprintf("<b>Ref: </b> %s", "test-ref"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "PERSON",
											},
											"text": fmt.Sprintf("<b>Actor: </b> %s", "test-triggered_actor"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "CLOCK",
											},
											"text": fmt.Sprintf("<b>UTC: </b> %s", time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC).UTC().Format(time.RFC3339)),
										},
									},
									{
										"buttonList": map[string]any{
											"buttons": []any{
												map[string]any{
													"text": "Open workflow",
													"onClick": map[string]any{
														"openLink": map[string]any{
															"url": fmt.Sprintf("https://github.com/%s/actions/runs/%s",
																"test-repository", "test-run-id"),
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
			},
		},
		{
			name: "test_issue_trigger",
			ghJSON: map[string]any{
				"workflow":         "test-workflow",
				"ref":              "test-ref",
				"triggering_actor": "test-triggered_actor",
				"repository":       "test-repository",
				"event_name":       "issues",
				"event": map[string]any{
					"action": "opened",
					"issue": map[string]any{
						"title":      "test-title",
						"created_at": time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC).UTC().Format(time.RFC3339),
						"html_url":   "https://foo.com",
					},
				},
			},
			jobJSON:   map[string]any{},
			timestamp: time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC),
			wantMessageBody: map[string]any{
				"cardsV2": map[string]any{
					"cardId": "createCardMessage",
					"card": map[string]any{
						"header": map[string]any{
							"title":    "A issue is opened",
							"subtitle": fmt.Sprintf("Issue title: <b>%s</b>", "test-title"),
							"imageUrl": "https://github.githubassets.com/favicons/favicon.png",
						},
						"sections": []any{
							map[string]any{
								"collapsible":               true,
								"uncollapsibleWidgetsCount": float64(1),
								"widgets": []map[string]any{
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"iconUrl": widgetRefIconURL,
											},
											"text": fmt.Sprintf("<b>Repo: </b> %s", "test-repository"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"iconUrl": "https://fonts.gstatic.com/s/i/short-term/release/googlesymbols/quick_reference/default/48px.svg",
											},
											"text": fmt.Sprintf("<b>Ref: </b> %s", "test-ref"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "PERSON",
											},
											"text": fmt.Sprintf("<b>Actor: </b> %s", "test-triggered_actor"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "CLOCK",
											},
											"text": fmt.Sprintf("<b>UTC: </b> %s", time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC).UTC().Format(time.RFC3339)),
										},
									},
									{
										"buttonList": map[string]any{
											"buttons": []any{
												map[string]any{
													"text": "Open issue",
													"onClick": map[string]any{
														"openLink": map[string]any{
															"url": "https://foo.com",
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
			},
		},
		{
			name: "test_release_trigger",
			ghJSON: map[string]any{
				"workflow":         "test-workflow",
				"ref":              "test-ref",
				"triggering_actor": "test-triggered_actor",
				"repository":       "test-repository",
				"event_name":       "release",
				"event": map[string]any{
					"action": "released",
					"release": map[string]any{
						"name":       "test-title",
						"created_at": time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC).UTC().Format(time.RFC3339),
						"html_url":   "https://foo.com",
					},
				},
			},
			jobJSON:   map[string]any{},
			timestamp: time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC),
			wantMessageBody: map[string]any{
				"cardsV2": map[string]any{
					"cardId": "createCardMessage",
					"card": map[string]any{
						"header": map[string]any{
							"title":    "A release is released",
							"subtitle": fmt.Sprintf("Release name: <b>%s</b>", "test-title"),
							"imageUrl": "https://github.githubassets.com/favicons/favicon.png",
						},
						"sections": []any{
							map[string]any{
								"collapsible":               true,
								"uncollapsibleWidgetsCount": float64(1),
								"widgets": []map[string]any{
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"iconUrl": widgetRefIconURL,
											},
											"text": fmt.Sprintf("<b>Repo: </b> %s", "test-repository"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"iconUrl": "https://fonts.gstatic.com/s/i/short-term/release/googlesymbols/quick_reference/default/48px.svg",
											},
											"text": fmt.Sprintf("<b>Ref: </b> %s", "test-ref"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "PERSON",
											},
											"text": fmt.Sprintf("<b>Actor: </b> %s", "test-triggered_actor"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "CLOCK",
											},
											"text": fmt.Sprintf("<b>UTC: </b> %s", time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC).UTC().Format(time.RFC3339)),
										},
									},
									{
										"buttonList": map[string]any{
											"buttons": []any{
												map[string]any{
													"text": "Open release",
													"onClick": map[string]any{
														"openLink": map[string]any{
															"url": "https://foo.com",
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
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotMessageBody, err := generateRequestBody(generateMessageBodyContent(tc.ghJSON, tc.jobJSON, tc.timestamp))
			if err != nil {
				t.Fatalf("failed to generate messag body %v", err)
			}

			wantMessageBodyByte, err := json.Marshal(tc.wantMessageBody)
			if err != nil {
				t.Fatalf("failed to marshal tc.wantMessageBody: %v", err)
			}

			if diff := cmp.Diff(wantMessageBodyByte, gotMessageBody); diff != "" {
				t.Errorf("messageBody got unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
