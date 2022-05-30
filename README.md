# offline-license-server

AppsCode License server. We call it `offline` license server because once you have received the license file, no further connection is required with the license server. So, these licenses can be used within an air-gapped Kubernetes cluster.

## License Validity

A License file is a valid for a given Kubernetes cluster. For the `community` edition, you will receive a license that is valid for 1 year. For the `enterprise` edition, this server will issue a 30 day trial license. If you are interested in purchasing Enterprise license, please contact us via sales@appscode.com for further discussion. You can also set up a meeting via our [calendly link](https://calendly.com/appscode/intro).

## Cluster UID

We use the `uid` of the `kube-system` namespace as the Kubernetes cluster UID. Please run the command below to get the cluster uid for your cluster:

```bash
kubectl get ns kube-system -o=jsonpath='{.metadata.uid}'
```

## License Issuer CA

The license issuer ca can be found here: https://licenses.appscode.com/certificates/ca.crt

## Email Address Requirements

You must provide a valid non-disposable email to acquire license. For Enterprise products, you must provide a valid work email to acquire license.

## API Reference

License issuing process can be automated in 2 steps:

### 1. Register Email

This is an one time set up step where you register with a valid email address. Make the api call below with a valid email address. We are going to email you a token for the License server.

```bash
curl -d "email=***" -X POST https://license-issuer.appscode.com/register
```

### 2. Issue License

Now every time you need a new license, use the token from the previous step to make an api call to our license server. You can use the `{email, token}` to issue license using `curl` from command line. In the example below replace `***` with an actual token you have received in the email.

```bash
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
**List of products**

 - kubedb-enterprise
 - kubedb-community
 - stash-enterprise
 - stash-community
 - kubevault-enterprise
 - kubevault-community
 - kubeform-enterprise
 - kubeform-community
 - voyager-enterprise
 - voyager-community
 - console-enterprise
 - auditor-enterprise
 - panopticon-enterprise

## Installation

These instructions are useful if you are trying to deploy your own license server or update an existing license server.

- Download pre-built binary into a server

```bash
curl -fsSL -O https://github.com/appscode/offline-license-server/releases/download/v0.0.32/offline-license-server-linux-amd64
chmod +x offline-license-server-linux-amd64
mv offline-license-server-linux-amd64 /usr/local/bin/offline-license-server
sudo systemctl restart offline-license-server
```

- Install systemd service

```bash
curl -fsSL -O https://github.com/appscode/offline-license-server/raw/v0.0.32/hack/systemd/offline-license-server.service
chmod +x offline-license-server.service

# 1. Copy Google cloud service account json key to /root/app/gcloud.json
# 2. Edit offline-license-server.service file to
#    - set MG_API_KEY
#    - set APPSCODE_SALES_USERNAME
#    - set APPSCODE_SALES_PASSWORD
#    - add `--ssl`
#    - add --spreadsheet-id=1evwv2ON94R38M-Lkrw8b6dpVSkRYHUWsNOuI7X0_-zA --geo-city-database-file=/root/maxmind/GeoLite2-City.mmdb

mv offline-license-server.service /lib/systemd/system/offline-license-server.service
```

Now, you should be able to enable the service, start it, then monitor the logs by tailing the systemd journal:

```bash
sudo systemctl enable offline-license-server.service
sudo systemctl start offline-license-server
sudo journalctl -f -u offline-license-server
```

## Issue Extended License

```bash
offline-license-server issue-full-license \
  --email= \
  --name= \
  --product= \
  --cluster= \
  --duration=(P30D | P1M | P1Y)

# --duration flag used https://pkg.go.dev/github.com/rickb777/date/period for parsing duration.
```

### Generate Quotation

```bash
offline-license-server quotation generate \
  --lead.email='****' \
  --lead.name='***' \
  --lead.title='***' \
  --lead.company='***' \
  --lead.telephone='***' \
  --template-doc-id=***
```

## Webinar signup

```bash
curl -X POST \
  -d "first_name=Tamal&last_name=Saha&phone=+1-1234567890&job_title=CEO&work_email=tamal@appscode.com&company=AppsCode&cluster_provider=aws&experience_level=tried&marketing_reach=word" \
  http://localhost:4000/_/webinars/2021-3-15/register
```

## Test configure

```
offline-license-server qa configure \
  --test.config-doc-id=1KB_Efi9jQcJ0_tCRF4fSLc6TR7QxaBKg05cKXAwbC9E \
  --test.qa-template-doc-id=16Ff6Lum3F6IeyAEy3P5Xy7R8CITIZRjdwnsRwBg9rD4 \
  --test.days-to-take-test=3 \
  --test.duration=60m
```
