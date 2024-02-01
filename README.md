# send-google-chat-webhook

This github action will enable users to send notification to google chat via github actions.

## Prerequisites
### Obtain a web hook from your google chat workspace.
1. Go to your chat space which you want to add a webhook.
2. At the top, click on the space title, select Apps & Integration.
3. Click Manage webhooks
4. Create a webhook or add another webhook if there is already one.
5. Copy the webhook URL that you intend to use for this github action.

For your own security purposes, we would suggest to store your webhook url in github secrets, and use `${{ secrets.WEBHOOK_URL}}` to get it's value.

## Usage

```yaml
jobs:
  job_id:
    # ...

    permissions:
      contents: 'read'
      id-token: 'write'

    steps:
    # ...

    - id: 'notify_google_chat'
      uses: 'google-github-actions/send-google-chat-webhook@v0.0.1'
      with:
        webhook_url: '${{ secrets.WEBHOOK_URL }}'
        mention: "<users/all>"
```

You can customize the condition for when you want this action is called..

```yaml
- id: 'notify google chat'
  if: ${{ inputs.fail_intentionally }}
  uses: 'google-github-actions/send-google-chat-webhook@v0.0.1'
  with:
    webhook_url: '${{ secrets.WEBHOOK_URL }}'
    mention: "<users/all>"
```

Helpful references:
* Messages and Cards
  * [Create, read, update, delete messages](https://developers.google.com/chat/api/guides/crudl/messages)
  * [Send a card message](https://developers.google.com/chat/api/guides/message-formats/cards)
  * [REST Resource: spaces.messages](https://developers.google.com/chat/api/reference/rest/v1/spaces.messages)
  * [Method: spaces.messages.create](https://developers.google.com/chat/api/reference/rest/v1/spaces.messages/create)
  * [Cards v2](https://developers.google.com/chat/api/reference/rest/v1/cards)
* abcxyz
  * [abcxyz/pkg/cli](https://pkg.go.dev/github.com/abcxyz/pkg/cli)
