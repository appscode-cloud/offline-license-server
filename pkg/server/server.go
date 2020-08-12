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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/appscodelabs/offline-license-server/templates"
	"github.com/go-macaron/bindata"
	"github.com/go-macaron/binding"
	"github.com/google/uuid"
	"github.com/mailgun/mailgun-go/v4"
	"github.com/pkg/errors"
	"gocloud.dev/blob"
	"golang.org/x/crypto/acme/autocert"
	"gomodules.xyz/blobfs"
	"gomodules.xyz/cert"
	"gomodules.xyz/cert/certstore"
	emailproviders "gomodules.xyz/email-providers"
	"gopkg.in/macaron.v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Server struct {
	opts *Options

	certs *certstore.CertStore
	fs    *blobfs.BlobFS
	mg    mailgun.Mailgun
}

func New(opts *Options) (*Server, error) {
	fs := blobfs.New("gs://" + opts.LicenseBucket)
	certs, err := certstore.New(fs, CACertificatesPath(), LicenseIssuerName)
	if err != nil {
		return nil, err
	}
	err = certs.InitCA()
	if err != nil {
		return nil, err
	}
	return &Server{
		opts:  opts,
		certs: certs,
		fs:    fs,
		mg:    mailgun.NewMailgun(opts.MailgunDomain, opts.MailgunPrivateAPIKey),
	}, nil
}

func respond(ctx *macaron.Context, data []byte) {
	_, err := ctx.Write(data)
	if err != nil {
		panic(err)
	}
}

func (s *Server) Run() error {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Renderer(macaron.RenderOptions{
		TemplateFileSystem: bindata.Templates(bindata.Options{
			Asset:      templates.Asset,
			AssetDir:   templates.AssetDir,
			AssetNames: templates.AssetNames,
			Prefix:     "",
		}),
	}))
	// m.Use(macaron.Static("public"))
	m.Get("/", func(ctx *macaron.Context) {
		ctx.HTML(200, "index") // 200 is the response code.
	})

	m.Post("/register", binding.Bind(RegisterRequest{}), func(ctx *macaron.Context, info RegisterRequest) {
		// verify required fields are present
		err := s.HandleRegisterEmail(info)
		if err != nil {
			ctx.WriteHeader(http.StatusInternalServerError)
			respond(ctx, []byte(err.Error()))
			return
		}
		respond(ctx, []byte("Check your email for token"))
	})

	m.Post("/issue-license", binding.Bind(LicenseForm{}), func(ctx *macaron.Context, info LicenseForm) {
		if err := info.Validate(); err != nil {
			ctx.WriteHeader(http.StatusBadRequest)
			respond(ctx, []byte(err.Error()))
			return
		}

		err := s.HandleIssueLicense(ctx, info)
		if err != nil {
			ctx.WriteHeader(http.StatusInternalServerError)
			respond(ctx, []byte(err.Error()))
			return
		}
		// ctx.Write([]byte("Your license has been emailed!"))
	})

	if !s.opts.EnableSSL {
		addr := fmt.Sprintf(":%d", s.opts.Port)
		fmt.Println("Listening to addr", addr)
		return http.ListenAndServe(addr, m)
	}

	// ref:
	// - https://goenning.net/2017/11/08/free-and-automated-ssl-certificates-with-go/
	// - https://stackoverflow.com/a/40494806/244009
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(s.opts.CertDir),
		HostPolicy: autocert.HostWhitelist(s.opts.Hosts...),
		Email:      s.opts.CertEmail,
	}
	server := &http.Server{
		Addr:         ":https",
		Handler:      m,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	go func() {
		// does automatic http to https redirects
		err := http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		if err != nil {
			panic(err)
		}
	}()
	return server.ListenAndServeTLS("", "") //Key and cert are coming from Let's Encrypt
}

func (s *Server) HandleRegisterEmail(req RegisterRequest) error {
	domain := Domain(req.Email)
	token := uuid.New()

	if emailproviders.IsDisposableEmail(domain) {
		return fmt.Errorf("disposable email %s is not supported", req.Email)
	}

	if exists, err := s.fs.Exists(context.TODO(), EmailBannedPath(domain, req.Email)); err == nil && exists {
		return fmt.Errorf("email %s is banned", req.Email)
	}

	exists, err := s.fs.Exists(context.TODO(), EmailVerifiedPath(domain, req.Email))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("email is already verified")
	}

	err = s.fs.WriteFile(context.TODO(), EmailTokenPath(domain, req.Email, token.String()), []byte(time.Now().UTC().Format(time.RFC3339)))
	if err != nil {
		return err
	}

	{
		subject := "Token for AppsCode License server"
		src := `Hi,
Please use the token below to issue licenses with this email.

{{.Token}}

Regards,
AppsCode Team`
		data := struct {
			Token string
		}{
			token.String(),
		}

		bodyText, bodyHtml, err := RenderMail(src, data)
		if err != nil {
			return err
		}

		err = s.SendMail(req.Email, subject, bodyText, bodyHtml)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) HandleIssueLicense(ctx *macaron.Context, info LicenseForm) error {
	domain := Domain(info.Email)

	if emailproviders.IsDisposableEmail(domain) {
		return fmt.Errorf("disposable email %s is not supported", info.Email)
	}

	if exists, err := s.fs.Exists(context.TODO(), EmailBannedPath(domain, info.Email)); err == nil && exists {
		return fmt.Errorf("email %s is banned", info.Email)
	}
	if info.Token != "" {
		exists, err := s.fs.Exists(context.TODO(), EmailTokenPath(domain, info.Email, info.Token))
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("token is invalid")
		}
	}

	license, err := s.GetDomainLicense(domain, info.Product)
	if err != nil {
		return err
	}
	crtLicense, err := s.CreateLicense(*license, info.Cluster)
	if err != nil {
		return err
	}
	crtURL, err := s.fs.SignedURL(context.TODO(), LicenseCertPath(license.Domain, license.Product, info.Cluster), &blob.SignedURLOptions{
		Expiry: 24 * time.Hour,
	})
	if err != nil {
		return err
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	{
		// record request
		accesslog := struct {
			LicenseForm
			IP string
		}{
			info,
			GetIP(ctx.Req.Request),
		}
		// TODO: IP to location
		// https://github.com/oschwald/geoip2-golang

		data, err := json.MarshalIndent(accesslog, "", "  ")
		if err != nil {
			return err
		}
		err = s.fs.WriteFile(context.TODO(), ProductAccessLogPath(domain, info.Product, info.Cluster, timestamp), data)
		if err != nil {
			return err
		}
	}
	{
		// record request
		accesslog := struct {
			LicenseForm
			IP string
		}{
			info,
			GetIP(ctx.Req.Request),
		}
		// TODO: IP to location
		// https://github.com/oschwald/geoip2-golang

		data, err := json.MarshalIndent(accesslog, "", "  ")
		if err != nil {
			return err
		}
		err = s.fs.WriteFile(context.TODO(), EmailAccessLogPath(domain, info.Email, info.Product, timestamp), data)
		if err != nil {
			return err
		}
	}

	{
		subject := fmt.Sprintf("%s License for cluster %s", info.Product, info.Cluster)

		src := `Hi {{.Name}},
Thanks for your interest in {{.Product}}. Here is the link to the license for Kubernetes cluster: {{.Cluster}}

{{.CrtURL}}

Regards,
AppsCode Team`
		data := struct {
			LicenseForm
			CrtURL string
		}{
			info,
			crtURL,
		}
		bodyText, bodyHtml, err := RenderMail(src, data)
		if err != nil {
			return err
		}

		err = s.SendMail(info.Email, subject, bodyText, bodyHtml)
		if err != nil {
			return err
		}
	}

	// mark email as verified
	if exists, err := s.fs.Exists(context.TODO(), EmailVerifiedPath(domain, info.Email)); err == nil && !exists {
		err = s.fs.WriteFile(context.TODO(), EmailVerifiedPath(domain, info.Email), []byte(timestamp))
		if err != nil {
			return err
		}
	}

	{
		if info.Token != "" {
			respond(ctx, crtLicense)
		} else {
			respond(ctx, []byte("Your license has been emailed!"))
		}
	}

	return nil
}

func (s *Server) GetDomainLicense(domain string, product string) (*ProductLicense, error) {
	if !emailproviders.IsWorkEmail(domain) {
		if IsEnterpriseProduct(product) {
			return nil, apierrors.NewBadRequest("requires work email for enterprise license")
		}
		ttl := metav1.Duration{Duration: DefaultTTLForCommunityProduct}
		return &ProductLicense{
			Domain:  domain,
			Product: product,
			TTL:     &ttl,
		}, nil
	}

	exists, err := s.fs.Exists(context.TODO(), AgreementPath(domain, product))
	if err != nil {
		return nil, err
	}

	var opts ProductLicense
	if !exists {
		var ttl metav1.Duration
		if IsEnterpriseProduct(product) {
			ttl = metav1.Duration{Duration: DefaultTTLForEnterpriseProduct}
		} else {
			ttl = metav1.Duration{Duration: DefaultTTLForCommunityProduct}
		}
		opts = ProductLicense{
			Domain:  domain,
			Product: product,
			TTL:     &ttl,
		}
		data, err := json.MarshalIndent(opts, "", "  ")
		if err != nil {
			return nil, err
		}
		err = s.fs.WriteFile(context.TODO(), AgreementPath(domain, product), data)
		if err != nil {
			return nil, err
		}
	} else {
		data, err := s.fs.ReadFile(context.TODO(), AgreementPath(domain, product))
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &opts)
		if err != nil {
			return nil, err
		}
	}
	return &opts, nil
}

func (s *Server) CreateLicense(license ProductLicense, cluster string) ([]byte, error) {
	exists, err := s.fs.Exists(context.TODO(), LicenseCertPath(license.Domain, license.Product, cluster))
	if err != nil {
		return nil, err
	}
	if !exists {
		// agreement, TTL
		sans := cert.AltNames{
			DNSNames: []string{cluster},
		}
		cfg := Config{
			CommonName:   getCN(sans),
			Organization: []string{license.Product},
			AltNames:     sans,
			Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}
		if license.Agreement != nil {
			cfg.NotAfter = license.Agreement.ExpiryDate.UTC()
		} else if license.TTL != nil {
			cfg.NotAfter = time.Now().Add(license.TTL.Duration).UTC()
		} else {
			return nil, apierrors.NewInternalError(fmt.Errorf("Missing license TTL")) // this should never happen
		}

		key, err := cert.NewPrivateKey()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate private key")
		}
		crt, err := NewSignedCert(cfg, key, s.certs.CACert(), s.certs.CAKey())
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate client certificate")
		}

		err = s.fs.WriteFile(context.TODO(), LicenseCertPath(license.Domain, license.Product, cluster), cert.EncodeCertPEM(crt))
		if err != nil {
			return nil, err
		}
		err = s.fs.WriteFile(context.TODO(), LicenseKeyPath(license.Domain, license.Product, cluster), cert.EncodePrivateKeyPEM(key))
		if err != nil {
			return nil, err
		}
	}
	return s.fs.ReadFile(context.TODO(), LicenseCertPath(license.Domain, license.Product, cluster))
}
