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

package server

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/mailgun/mailgun-go/v4"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"google.golang.org/api/drive/v3"
)

type Mailer struct {
	Sender  string
	BCC     string
	ReplyTo string

	Subject string
	Body    string
	params  interface{}

	AttachmentBytes map[string][]byte
	GDriveFiles     map[string]string
	GoogleDocIds    map[string]string
}

func (m *Mailer) renderMail(src string, data interface{}) (string, string, error) {
	tpl := template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(src))

	var bodyText bytes.Buffer
	err := tpl.Execute(&bodyText, data)
	if err != nil {
		return "", "", err
	}

	var bodyHtml bytes.Buffer
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
	if err := md.Convert(bodyText.Bytes(), &bodyHtml); err != nil {
		return "", "", err
	}
	return bodyText.String(), bodyHtml.String(), nil
}

func (m *Mailer) SendMail(mg mailgun.Mailgun, recipient string, srv *drive.Service) error {
	bodyText, bodyHtml, err := m.renderMail(m.Body, m.params)
	if err != nil {
		return err
	}

	// The message object allows you to add attachments and Bcc recipients
	msg := mg.NewMessage(m.Sender, m.Subject, bodyText, recipient)
	msg.AddBCC(m.BCC)
	msg.SetReplyTo(m.ReplyTo)

	msg.SetTracking(true)
	msg.SetTrackingClicks(true)
	msg.SetTrackingOpens(true)

	msg.SetHtml(bodyHtml)
	for filename, data := range m.AttachmentBytes {
		msg.AddBufferAttachment(filename, data)
	}

	for f, docId := range m.GoogleDocIds {
		filename := filepath.Join(os.TempDir(), recipient, f)
		err := ExportPDF(srv, docId, filename)
		if err != nil {
			return err
		}
		msg.AddAttachment(filename)
	}

	for f, docId := range m.GDriveFiles {
		filename := filepath.Join(os.TempDir(), recipient, f)
		err := DownloadFile(srv, docId, filename)
		if err != nil {
			return err
		}
		msg.AddAttachment(filename)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, _, err = mg.Send(ctx, msg)
	return err
}
