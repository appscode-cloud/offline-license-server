/*
Copyright AppsCode Inc. and Contributors

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

package mailer

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/pkg/errors"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gomodules.xyz/email"
	"google.golang.org/api/drive/v3"
)

type SMTPService struct {
	Address string
	Auth    smtp.Auth
}

func NewSMTPServiceFromEnv() (*SMTPService, error) {
	addr := os.Getenv("SMTP_ADDRESS")
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	return &SMTPService{
		Address: addr,
		Auth:    smtp.PlainAuth("", os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"), host),
	}, nil
}

type Mailer struct {
	Sender  string
	BCC     string
	ReplyTo string

	Subject string
	Body    string
	Params  interface{}

	AttachmentBytes map[string][]byte
	GDriveFiles     map[string]string
	GoogleDocIds    map[string]string
}

func (m *Mailer) renderSubject() (string, error) {
	subject := m.Subject
	if m.Params != nil && strings.Contains(subject, `{{`) {
		var sub bytes.Buffer
		tpl := template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(subject))
		err := tpl.Execute(&sub, m.Params)
		if err != nil {
			return "", err
		}
		subject = sub.String()
	}
	return subject, nil
}

func (m *Mailer) renderMail(src string, params interface{}) (string, string, error) {
	var bodyText bytes.Buffer

	if params == nil {
		// by pass if there is no params passed for rendering
		bodyText.WriteString(src)
	} else {
		tpl := template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(src))
		err := tpl.Execute(&bodyText, params)
		if err != nil {
			return "", "", err
		}
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

func (m *Mailer) SendMail(mg *SMTPService, recipient, cc string, srv *drive.Service) error {
	subject, bodyText, bodyHtml, err := m.Render()
	if err != nil {
		return err
	}

	// The message object allows you to add attachments and Bcc recipients
	msg := email.NewEmail()
	msg.From = m.Sender
	msg.To = []string{recipient}
	msg.Subject = subject
	msg.Text = []byte(bodyText)
	msg.HTML = []byte(bodyHtml)
	if cc != "" {
		for _, e := range strings.Split(cc, ",") {
			msg.Cc = append(msg.Cc, strings.TrimSpace(e))
		}
	}
	if m.BCC != "" {
		for _, e := range strings.Split(m.BCC, ",") {
			msg.Bcc = append(msg.Bcc, strings.TrimSpace(e))
		}
	}
	if m.ReplyTo != "" {
		msg.ReplyTo = []string{m.ReplyTo}
	}

	for filename, data := range m.AttachmentBytes {
		mtype := mimetype.Detect(data)
		if _, err := msg.Attach(bytes.NewReader(data), filename, mtype.String()); err != nil {
			return errors.Wrapf(err, "failed to attach file %q", filename)
		}
	}

	for f, docId := range m.GoogleDocIds {
		filename := filepath.Join(os.TempDir(), recipient, f)
		err := ExportPDF(srv, docId, filename)
		if err != nil {
			return err
		}
		if _, err := msg.AttachFile(filename); err != nil {
			return errors.Wrapf(err, "failed to attach file %q", filename)
		}
	}

	for f, docId := range m.GDriveFiles {
		filename := filepath.Join(os.TempDir(), recipient, f)
		err := DownloadFile(srv, docId, filename)
		if err != nil {
			return err
		}
		if _, err := msg.AttachFile(filename); err != nil {
			return errors.Wrapf(err, "failed to attach file %q", filename)
		}
	}

	// Send the message with a 10 second timeout
	return msg.Send(mg.Address, mg.Auth)
}

func (m *Mailer) Render() (subject string, bodyText string, bodyHtml string, err error) {
	subject, err = m.renderSubject()
	if err != nil {
		return
	}

	bodyText, bodyHtml, err = m.renderMail(m.Body, m.Params)
	return
}

func ExportPDF(srvDrive *drive.Service, docId, filename string) error {
	resp, err := srvDrive.Files.Export(docId, "application/pdf").Download()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}
	// filename := filepath.Join(gen.cfg.OutDir, FolderName(gen.cfg.Email), docName+".pdf")
	err = os.MkdirAll(filepath.Dir(filename), 0o755)
	if err != nil {
		return err
	}
	fmt.Println("writing file:", filename)
	return ioutil.WriteFile(filename, buf.Bytes(), 0o644)
}

func DownloadFile(srvDrive *drive.Service, docId, filename string) error {
	resp, err := srvDrive.Files.Get(docId).Download()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}
	// filename := filepath.Join(gen.cfg.OutDir, FolderName(gen.cfg.Email), docName+".pdf")
	err = os.MkdirAll(filepath.Dir(filename), 0o755)
	if err != nil {
		return err
	}
	fmt.Println("writing file:", filename)
	return ioutil.WriteFile(filename, buf.Bytes(), 0o644)
}
