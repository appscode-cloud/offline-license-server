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
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"k8s.io/apimachinery/pkg/util/sets"
)

var knowTestEmails = sets.NewString("1gtm@appscode.com")
var skipEmailDomains = sets.NewString("appscode.com")

func RenderMail(src string, data interface{}) (string, string, error) {
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

func (s *Server) SendMail(recipient, subject, bodyText, bodyHtml string, attachments map[string][]byte) error {
	// The message object allows you to add attachments and Bcc recipients
	msg := s.mg.NewMessage(s.opts.MailSender, subject, bodyText, recipient)
	msg.SetHtml(bodyHtml)
	msg.AddBCC(s.opts.MailLicenseTracker)
	msg.SetReplyTo(s.opts.MailReplyTo)
	for filename, data := range attachments {
		msg.AddBufferAttachment(filename, data)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, _, err := s.mg.Send(ctx, msg)
	return err
}
