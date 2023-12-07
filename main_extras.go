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

// nolint
package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"go.bytebuilders.dev/offline-license-server/pkg/server"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	gdrive "gomodules.xyz/gdrive-utils"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"k8s.io/klog/v2"
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

func main_sheets() {
	client, err := gdrive.DefaultClient(".")
	if err != nil {
		klog.Fatalln(err)
	}

	srvSheets, err := sheets.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		klog.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	si, err := gdrive.NewSpreadsheet(srvSheets, "1evwv2ON94R38M-Lkrw8b6dpVSkRYHUWsNOuI7X0_-zA") // Share this sheet with the service account email
	if err != nil {
		klog.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	info := server.LogEntry{
		LicenseForm: server.LicenseForm{
			Name:         "Fahim Abrar",
			Email:        "fahimabrar@appscode.com",
			ProductAlias: "kubeform-community",
			Cluster:      "bad94a42-0210-4c81-b07a-99bae529ec14",
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	err = server.LogLicense(si, &info)
	if err != nil {
		klog.Fatal(err)
	}
}
