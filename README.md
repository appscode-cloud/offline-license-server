# offline-license-server

## Installation

- Download pre-built binary into a server

```console
curl -fsSL -O https://github.com/appscodelabs/offline-license-server/releases/download/v0.0.9/offline-license-server-linux-amd64
chmod +x offline-license-server-linux-amd64
mv offline-license-server-linux-amd64 /usr/local/bin/offline-license-server
```

- Install systemd service

```console
curl -fsSL -O https://github.com/appscodelabs/offline-license-server/raw/v0.0.9/hack/systemd/offline-license-server.service
chmod +x offline-license-server.service

# edit offline-license-server.service file to add `--ssl --secret-key=<uuid>`

mv offline-license-server.service /lib/systemd/system/offline-license-server.service
```

Now, you should be able to enable the service, start it, then monitor the logs by tailing the systemd journal:

```console
sudo systemctl enable offline-license-server.service
sudo systemctl start offline-license-server
sudo journalctl -f -u offline-license-server
```

## Verify Email

```
curl -d "email=tamal@appscode.com" -X POST http://localhost:4000/register
```