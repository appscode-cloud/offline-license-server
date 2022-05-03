# Development Notes

```bash
# on development machine
make build OS=linux ARCH=amd64
scp bin/offline-license-server-linux-amd64 root@license-issuer.appscode.com:/root


# on production server
> ssh root@license-issuer.appscode.com

chmod +x offline-license-server-linux-amd64
mv offline-license-server-linux-amd64 /usr/local/bin/offline-license-server
sudo systemctl restart offline-license-server
```
