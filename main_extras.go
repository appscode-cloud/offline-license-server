/*
Copyright AppsCode Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

//nolint
package main

import (
	"bytes"
	"flag"
	"fmt"

	"github.com/appscodelabs/offline-license-server/pkg/server"
	"github.com/appscodelabs/offline-license-server/pkg/verifier"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gomodules.xyz/blobfs/testing"
	"gomodules.xyz/cert/certstore"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func main_MD_HTML() {
	source := `Hi Tamal,
Thanks for your interest in stash. Here is the link to the license for Kubernetes cluster: xyz-abc

https://appscode.com

Regards,
AppsCode Team
`

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
}

func main_cluster_uid() {
	var kubeconfig string
	var master string

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.Parse()

	// creates the connection
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	clusterUID, err := verifier.GetClusterUID(clientset.CoreV1().Namespaces())
	if err != nil {
		klog.Fatal(err)
	}

	fmt.Println(clusterUID)
}

func main_cert() {
	// Verifying with a custom list of root certificates.

	const rootPEM = `
-----BEGIN CERTIFICATE-----
MIIC6DCCAdCgAwIBAgIBADANBgkqhkiG9w0BAQsFADAlMRYwFAYDVQQKEw1BcHBz
Q29kZSBJbmMuMQswCQYDVQQDEwJjYTAeFw0yMDA4MDkxMTE4MzJaFw0zMDA4MDcx
MTE4MzJaMCUxFjAUBgNVBAoTDUFwcHNDb2RlIEluYy4xCzAJBgNVBAMTAmNhMIIB
IjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtO3NKvg6Jk11RkYqDfkZwajY
/w8bHiq/DV5KjQ7h45BHzLxrd4XupZweRQR1MqMUVxH3sXagO6q7vGMWzuhBmC9e
7B67ZjRGt3z3B49Q6VFIop0NB2DoYWk6FsAK4Fp3jtIgXCMcFApdmPZZ20H3F+mq
KAaS1I6X5VXEr5II9qvncUO2a7O9Tb4H+xZsr1xdv0CuC3FevmfQLbFt5nQJTRNq
ukcugxvImsoF1WZ+d8cz65krnvMlC5uUyGDCqpyIh4Iy1sMssk/7MOVzgqGZ8ISa
f5Jv+3IzZL3rQQ7TGZMBBeMBvIES6bg5FmDvg/6Rgo5K9KifdFkKLFiMUSKvOQID
AQABoyMwITAOBgNVHQ8BAf8EBAMCAqQwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG
9w0BAQsFAAOCAQEAAYumKKA1aPF3MuLRvSnUWijvw5+Rdptc+AdMgNvONrpukZ26
BAGRhA5DkGupBCjkyMCIAcTUX+hgW8QKpucun4VSoMkW2x69Z6xfKxDhGlRk3PD1
jeHcDmdnzC864tQ8rdINd+D7RksdBP9aCWTkXlcBlkYimCkanmFbaz7YmmFvdnZs
LeqjtZmJeARBz7p59AT4Pn2NpdLYKjHP7AEFmVpoF7Z4y1cl0AwsINuVNM5++7El
YMJ9tfGRB4Sgj1BZKcvwmqY3RBOtgpY5P27w46WqxYhRrmP867GEWzeCSzm1jRAI
hqDntdQyIGPXtqiMjPjKUxUMCSsAGL3ZqrMe9Q==
-----END CERTIFICATE-----`

	const certPEM = `
-----BEGIN CERTIFICATE-----
MIIDSDCCAjCgAwIBAgIIHfVMa0LGIXswDQYJKoZIhvcNAQELBQAwJTEWMBQGA1UE
ChMNQXBwc0NvZGUgSW5jLjELMAkGA1UEAxMCY2EwHhcNMjAwODA5MTExODMyWhcN
MjEwODEwMDQzNzA1WjBJMRgwFgYDVQQKEw9zdGFzaC1jb21tdW5pdHkxLTArBgNV
BAMTJDdjYTk0ZTBlLTgyMmYtNDBhYS1iYWNmLTg3ZDIxZTUxZWE4ZDCCASIwDQYJ
KoZIhvcNAQEBBQADggEPADCCAQoCggEBAMNWNGPki2sElUJSKBHU6Uc3jCCi0bFF
urppVSUtq0omQ1QtD6hOxvMgm6LMniY+I5JSaybaazCSBx+Clcteo6cIGkj9p/u2
4ltXsBPG5guvqP2s9UQsuZMFTwwG7etxUlxTAzWYjOejT8gMsYo/4vTn8RFrpxLE
SjJp/NeDdzzTXqQxETZ9TnOXAsxIOJ7dGHl4dj0prTj9tYL//N8fVMb/whyGNYFv
5hCfOOE0987hwwas9LPpXpiMyX6fuxBEnMeZYh0d3ns5ksZAAx4Zm6jqendUfO9w
kS9iOO8BWvl+qo19dFjsJxkUT+dRcFti1ya5+nSJ6JnB+STMhhHf6Z0CAwEAAaNY
MFYwDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMCMC8GA1UdEQQo
MCaCJDdjYTk0ZTBlLTgyMmYtNDBhYS1iYWNmLTg3ZDIxZTUxZWE4ZDANBgkqhkiG
9w0BAQsFAAOCAQEAEIf4K8WI/V09XVm6sI/DqalB3kECAdZmtWDTdoesRxx4KXTY
g56UqjD+El7aQil5pEnL0nzQ6gkGuavQpbhoaHegYdMdIhxBujnO59NT0G4Km1UN
s18FIhOcU1Oxdcya++kksdOtRCAQq+bv46PIGVnoZF7CR+MXzSZpkMsvhrk7WhrU
fvq1j+4DLBMw2ATix/U6Wha9tkEnjbOK1hKka4F7J7wUExRcPWpzwsTlPmjgE/2f
gLTp7xSpUYBWzRcct02/XeRkZyQD9gxvrdyKot6GjEW5PTYCCczGw7hEG93Ubqhc
0R7xqzNhilws/ftTACY0h9ulUhrZFGDMu+um+Q==
-----END CERTIFICATE-----`

	err := verifier.VerifyLicense("stash-community", "7ca94e0e-822f-40aa-bacf-87d21e51ea8d", []byte(rootPEM), []byte(certPEM))
	if err != nil {
		panic(err)
	}
}

func main_BlobFS() {
	fs, err := testing.NewTestGCS(server.LicenseBucketURL, server.GoogleApplicationCredentials)
	if err != nil {
		panic(err)
	}
	store, err := certstore.New(fs, "certificates", "AppsCode Inc.")
	if err != nil {
		panic(err)
	}
	err = store.InitCA()
	if err != nil {
		panic(err)
	}
}

// ------------------------------
