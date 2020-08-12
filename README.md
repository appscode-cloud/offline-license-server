# offline-license-server

AppsCode License server. We call it `offline` license server because once you have received the license file, no further connection is required with the license server. So, these licenses can be used within an air-gapped Kubernetes cluster.

## API Reference

### Cluster UID

We use the `uid` of the `kube-system` namespace as the Kubernetes cluster UID. Please run the command below to get the cluster uid for your cluster:

```console
kubectl get ns kube-system -o=jsonpath='{.metadata.uid}'
```

### License Issuer CA

The license issuer ca can be found here: https://licenses.appscode.com/certificates/ca.crt

### List of products

- kubedb-community
- kubedb-enterprise
- stash-community
- stash-enterprise

### Email Address Requirements

You must provide a valid non-disposable email to acquire license. For Enterprise products, you must provide a valid work email to acquire license.

### Register Email

Make the api call below with a valid email address. We are going to email you a token for the License server.

```console
curl -d "email=***" -X POST https://license-issuer.appscode.com/register
```

### Issue License

You can use the `{email, token}` to issue license using `curl` from command line. In the example below replace `***` with an actual token you have received in the email.

```console
# pass request body as application/x-www-form-urlencoded

curl -X POST \
  -d "name=***&email=***&product=kubedb-community&cluster=***&tos=true&token=***" \
  https://license-issuer.appscode.com/issue-license

# pass request body as application/json

curl -X POST -H "Content-Type: application/json" \
  -d '{"name":"***","email":"***","product":"kubedb-community","cluster":"***","tos":"true","token":"***"}' \
  https://license-issuer.appscode.com/issue-license

# pretty printed request json body
{
  "name": "***",
  "email": "***",
  "product": "kubedb-community",
  "cluster": "***",
  "tos": "true",
  "token": "***"
}
```

## Installation

These instructions are useful if you are trying to deploy your own license server or update an existing license server.

- Download pre-built binary into a server

```console
curl -fsSL -O https://github.com/appscode/offline-license-server/releases/download/v0.0.4/offline-license-server-linux-amd64
chmod +x offline-license-server-linux-amd64
mv offline-license-server-linux-amd64 /usr/local/bin/offline-license-server
```

- Install systemd service

```console
curl -fsSL -O https://github.com/appscode/offline-license-server/raw/v0.0.4/hack/systemd/offline-license-server.service
chmod +x offline-license-server.service

# 1. Copy Google cloud service account json key to /root/app/gcloud.json
# 2. Edit offline-license-server.service file to
#    - set MAILGUN_KEY
#    - add `--ssl`

mv offline-license-server.service /lib/systemd/system/offline-license-server.service
```

Now, you should be able to enable the service, start it, then monitor the logs by tailing the systemd journal:

```console
sudo systemctl enable offline-license-server.service
sudo systemctl start offline-license-server
sudo journalctl -f -u offline-license-server
```
