package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestGenerateMessageBody(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		ghJson          map[string]any
		jobJson         map[string]any
		timestamp       time.Time
		location        time.Location
		wantMessageBody map[string]any
	}{
		{
			name: "test_success_workflow",
			ghJson: map[string]any{
				"workflow":         "test-workflow",
				"ref":              "test-ref",
				"triggering_actor": "test-triggered_actor",
				"repository":       "test-repository",
				"run_id":           "test-run-id",
			},
			jobJson: map[string]any{
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
												"iconUrl": "https://fonts.gstatic.com/s/i/short-term/release/googlesymbols/quick_reference/default/48px.svg",
											},
											"text": fmt.Sprintf("<b>Ref:</b> %s", "test-ref"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "PERSON",
											},
											"text": fmt.Sprintf("<b>Run by:</b> %s", "test-triggered_actor"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "CLOCK",
											},
											"text": fmt.Sprintf("<b>Pacific:</b> %s", time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC).In(time.FixedZone("UTC-8", -7*60*60)).Format(time.DateTime)),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "CLOCK",
											},
											"text": fmt.Sprintf("<b>UTC:</b> %s", time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC).UTC().Format(time.DateTime)),
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
			ghJson: map[string]any{
				"workflow":         "test-workflow",
				"ref":              "test-ref",
				"triggering_actor": "test-triggered_actor",
				"repository":       "test-repository",
				"run_id":           "test-run-id",
			},
			jobJson: map[string]any{
				"status": "xxx",
			},
			timestamp: time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC),
			wantMessageBody: map[string]any{
				"cardsV2": map[string]any{
					"cardId": "createCardMessage",
					"card": map[string]any{
						"header": map[string]any{
							"title":    fmt.Sprintf("GitHub workflow %s", "xxx"),
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
												"iconUrl": "https://fonts.gstatic.com/s/i/short-term/release/googlesymbols/quick_reference/default/48px.svg",
											},
											"text": fmt.Sprintf("<b>Ref:</b> %s", "test-ref"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "PERSON",
											},
											"text": fmt.Sprintf("<b>Run by:</b> %s", "test-triggered_actor"),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "CLOCK",
											},
											"text": fmt.Sprintf("<b>Pacific:</b> %s", time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC).In(time.FixedZone("UTC-8", -7*60*60)).Format(time.DateTime)),
										},
									},
									{
										"decoratedText": map[string]any{
											"startIcon": map[string]any{
												"knownIcon": "CLOCK",
											},
											"text": fmt.Sprintf("<b>UTC:</b> %s", time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC).UTC().Format(time.DateTime)),
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
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			gotMessageBody, err := generateMessageBody(tc.ghJson, tc.jobJson, tc.timestamp)
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
