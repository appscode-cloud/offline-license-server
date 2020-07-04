# gh-ci-webhook

## Installation

- Download pre-built binary into a server

```console
curl -fsSL -O https://github.com/appscodelabs/gh-ci-webhook/releases/download/v0.0.7/gh-ci-webhook-linux-amd64
chmod +x gh-ci-webhook-linux-amd64
mv gh-ci-webhook-linux-amd64 /usr/local/bin/gh-ci-webhook
```

- Install systemd service

```console
curl -fsSL -O https://github.com/appscodelabs/gh-ci-webhook/raw/v0.0.7/hack/systemd/gh-ci-webhook.service
chmod +x gh-ci-webhook.service

# edit gh-ci-webhook.service file to add `--ssl --secret-key=<uuid>`

mv gh-ci-webhook.service /lib/systemd/system/gh-ci-webhook.service
```

Now, you should be able to enable the service, start it, then monitor the logs by tailing the systemd journal:

```console
sudo systemctl enable gh-ci-webhook.service
sudo systemctl start gh-ci-webhook
sudo journalctl -f -u gh-ci-webhook
```

## Configure Webhooks

## private repo
`https://gh-ci-webhook.appscode.ninja/payload?ci-repo=github.com/appscode-cloud/grafana-tester&actions=closed`

## public repo
`https://gh-ci-webhook.appscode.ninja/payload?pr-repo=github.com/appscode-cloud/private-repo`

Also, set the `<uuid>` passed to gh-ci-webhook.service as the secret key.
