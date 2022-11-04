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
	"net"
	"net/http"
	"net/smtp"
	"os"
	"path"
	"time"

	"github.com/appscodelabs/offline-license-server/templates"
	"github.com/avct/uasurfer"
	"github.com/go-macaron/auth"
	"github.com/go-macaron/bindata"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/cors"
	"github.com/google/uuid"
	"github.com/oschwald/geoip2-golang"
	"github.com/pkg/errors"
	"github.com/zoom-lib-golang/zoom-lib-golang"
	"golang.org/x/crypto/acme/autocert"
	"gomodules.xyz/blobfs"
	"gomodules.xyz/cert"
	"gomodules.xyz/cert/certstore"
	. "gomodules.xyz/email-providers"
	freshsalesclient "gomodules.xyz/freshsales-client-go"
	gdrive "gomodules.xyz/gdrive-utils"
	listmonkclient "gomodules.xyz/listmonk-client-go"
	"gomodules.xyz/mailer"
	"gomodules.xyz/sets"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/api/youtube/v3"
	"gopkg.in/macaron.v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type Server struct {
	opts *Options

	certs      *certstore.CertStore
	fs         blobfs.Interface
	mg         *mailer.SMTPService
	freshsales *freshsalesclient.Client
	listmonk   *listmonkclient.Client
	geodb      *geoip2.Reader
	sch        *Scheduler

	driveClient *http.Client
	srvDrive    *drive.Service
	srvDoc      *docs.Service
	srvSheets   *sheets.Service
	sheet       *gdrive.Spreadsheet
	srvCalendar *calendar.Service
	srvYT       *youtube.Service

	zc               *zoom.Client
	zoomAccountEmail string

	blockedDomains sets.String
	blockedEmails  sets.String
}

func New(opts *Options) (*Server, error) {
	fs := blobfs.New("gs://" + opts.LicenseBucket)

	caCertPath := CACertificatesPath()
	issuerName := LicenseIssuerName
	if opts.Issuer != "" {
		caCertPath = path.Join(CACertificatesPath(), opts.Issuer)
		issuerName = opts.Issuer
	}
	certs, err := certstore.New(fs, caCertPath, issuerName)
	if err != nil {
		return nil, err
	}
	err = certs.InitCA()
	if err != nil {
		return nil, err
	}

	var geodb *geoip2.Reader
	if opts.GeoCityDatabase != "" {
		geodb, err = geoip2.Open(opts.GeoCityDatabase)
		if err != nil {
			return nil, err
		}
	}

	sch, err := NewScheduler(opts.TaskDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scheduler")
	}

	client, err := gdrive.DefaultClient(opts.GoogleCredentialDir, youtube.YoutubeReadonlyScope)
	if err != nil {
		return nil, err
	}

	srvDrive, err := drive.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	srvDoc, err := docs.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Docs client: %v", err)
	}

	sheetsService, err := sheets.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	sheet, err := gdrive.NewSpreadsheet(sheetsService, opts.LicenseSpreadsheetId) // Share this sheet with the service account email
	if err != nil {
		return nil, err
	}

	srvCalendar, err := calendar.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar gc: %v", err)
	}

	srvYT, err := youtube.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create YouTube client: %v", err)
	}

	smtpHost, _, err := net.SplitHostPort(opts.SMTPAddress)
	if err != nil {
		return nil, err
	}
	mg := &mailer.SMTPService{
		Address: opts.SMTPAddress,
		Auth:    smtp.PlainAuth("", opts.SMTPUsername, opts.SMTPPassword, smtpHost),
	}
	return &Server{
		opts:             opts,
		certs:            certs,
		fs:               fs,
		mg:               mg,
		sheet:            sheet,
		freshsales:       freshsalesclient.New(opts.freshsalesHost, opts.freshsalesAPIToken),
		listmonk:         listmonkclient.New(opts.listmonkHost, opts.listmonkUsername, opts.listmonkPassword),
		geodb:            geodb,
		sch:              sch,
		driveClient:      client,
		srvDrive:         srvDrive,
		srvDoc:           srvDoc,
		srvSheets:        sheetsService,
		srvCalendar:      srvCalendar,
		srvYT:            srvYT,
		zc:               zoom.NewClient(os.Getenv("ZOOM_API_KEY"), os.Getenv("ZOOM_API_SECRET")),
		zoomAccountEmail: os.Getenv("ZOOM_ACCOUNT_EMAIL"),
		blockedDomains:   sets.NewString(opts.BlockedDomains...),
		blockedEmails:    sets.NewString(opts.BlockedEmails...),
	}, nil
}

func (s *Server) Close() {
	if s.geodb != nil {
		_ = s.geodb.Close()
	}
	if s.sch != nil {
		_ = s.sch.Close()
	}
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
	m.Use(cache.Cacher())
	m.Use(cors.CORS(cors.Options{
		Section:          "",
		Scheme:           "*",
		AllowDomain:      []string{"*"}, //{"appscode.com", "kubedb.com", "stash.run", "kubevault.com", "kubeform.cloud"},
		AllowSubdomain:   true,
		Methods:          []string{http.MethodGet, http.MethodPost},
		MaxAgeSeconds:    600,
		AllowCredentials: false,
	}))
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
		ctx.Data["Product"] = ctx.Query("p")
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

	m.Get("/_/pricing/", auth.Basic(os.Getenv("APPSCODE_SALES_USERNAME"), os.Getenv("APPSCODE_SALES_PASSWORD")), func(ctx *macaron.Context) {
		product := ctx.Query("p")
		if product != "" && IsPAYGProduct(product) {
			ctx.Error(http.StatusBadRequest, fmt.Sprintf("unknown product: %s", product))
			return
		}
		ctx.Data["Product"] = product
		ctx.HTML(200, "pricing") // 200 is the response code.
	})
	m.Post("/_/pricing/", binding.Bind(QuotationForm{}), func(ctx *macaron.Context, lead QuotationForm) {
		if err := lead.Validate(); err != nil {
			ctx.WriteHeader(http.StatusBadRequest)
			respond(ctx, []byte(err.Error()))
			return
		}

		err := s.HandleEmailQuotation(ctx, lead)
		if err != nil {
			ctx.WriteHeader(http.StatusInternalServerError)
			respond(ctx, []byte(err.Error()))
			return
		}
	})

	m.Get("/_/eula/", auth.Basic(os.Getenv("APPSCODE_SALES_USERNAME"), os.Getenv("APPSCODE_SALES_PASSWORD")), func(ctx *macaron.Context) {
		ctx.HTML(200, "eula") // 200 is the response code.
	})
	m.Post("/_/eula/", binding.Bind(EULAInfo{}), func(ctx *macaron.Context, form EULAInfo) {
		if err := form.Complete(); err != nil {
			ctx.WriteHeader(http.StatusBadRequest)
			respond(ctx, []byte(err.Error()))
			return
		}
		if err := form.Validate(); err != nil {
			ctx.WriteHeader(http.StatusBadRequest)
			respond(ctx, []byte(err.Error()))
			return
		}

		folderId, err := s.GenerateEULA(&form)
		if err != nil {
			ctx.WriteHeader(http.StatusInternalServerError)
			respond(ctx, []byte(err.Error()))
			return
		}
		ctx.Redirect(fmt.Sprintf("https://drive.google.com/drive/folders/%s", folderId))
	})

	m.Get("/_/offerletter/", auth.Basic(os.Getenv("APPSCODE_SALES_USERNAME"), os.Getenv("APPSCODE_SALES_PASSWORD")), func(ctx *macaron.Context) {
		ctx.HTML(200, "offerletter") // 200 is the response code.
	})
	m.Post("/_/offerletter/", binding.Bind(CandidateInfo{}), func(ctx *macaron.Context, form CandidateInfo) {
		form.Complete()
		if err := form.Validate(); err != nil {
			ctx.WriteHeader(http.StatusBadRequest)
			respond(ctx, []byte(err.Error()))
			return
		}

		folderId, err := s.GenerateOfferLetter(&form)
		if err != nil {
			ctx.WriteHeader(http.StatusInternalServerError)
			respond(ctx, []byte(err.Error()))
			return
		}
		ctx.Redirect(fmt.Sprintf("https://drive.google.com/drive/folders/%s", folderId))
	})

	s.RegisterWebinarAPI(m)
	s.RegisterNewsAPI(m)
	// m.Post("/_/webhooks/mailgun/", s.HandleMailgunWebhook)

	s.RegisterYoutubeAPI(m)
	s.RegisterQAAPI(m)

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

	if s.opts.EnableDripCampaign {
		go func() {
			if err := mailer.RunCampaigns(context.TODO(),
				NewCommunitySignupCampaign(s.srvSheets, s.mg),
				NewEnterpriseSignupCampaign(s.srvSheets, s.mg),
				NewEnterpriseFirstTimeCampaign(s.srvSheets, s.mg),
			); err != nil {
				panic(err)
			}
		}()
	}
	go func() {
		if err := s.sch.Cleanup(s.RevokePermission); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		}
	}()
	go func() {
		// does automatic http to https redirects
		err := http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		if err != nil {
			panic(err)
		}
	}()
	return server.ListenAndServeTLS("", "") // Key and cert are coming from Let's Encrypt
}

func (s *Server) HandleRegisterEmail(req RegisterRequest) error {
	domain := Domain(req.Email)
	token := uuid.New()

	if IsDisposableEmail(domain) {
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
		params := struct {
			Token string
		}{
			token.String(),
		}

		mailer := NewRegistrationMailer(params)
		err = mailer.SendMail(s.mg, req.Email, "", nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) HandleIssueLicense(ctx *macaron.Context, info LicenseForm) error {
	domain := Domain(info.Email)

	if IsDisposableEmail(domain) {
		return fmt.Errorf("disposable email %s is not supported", info.Email)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	if s.blockedDomains.Has(domain) || s.blockedEmails.Has(info.Email) {
		mailer := NewBlockedLicenseMailer(LicenseMailData{
			LicenseForm: info,
		})
		err := mailer.SendMail(s.mg, MailSales, info.CC, nil)
		if err != nil {
			return err
		}
		err = s.recordLicenseEvent(ctx, info, timestamp, EventTypeLicenseBlocked)
		if err != nil {
			return err
		}
		return errors.New("Please contact support@appscode.com to acquire license. Thanks!")
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
	crtLicense, err := s.CreateOrRetrieveLicense(info, *license, info.Cluster)
	if err != nil {
		return err
	}

	if !skipEmailDomains.Has(Domain(info.Email)) {
		// nolint:errcheck
		go func() (err error) {
			defer func() {
				if err != nil {
					klog.ErrorS(err, "failed record license download event", "email", info.Email)
				}
			}()

			dcEnt := NewEnterpriseSignupCampaign(s.srvSheets, s.mg)
			audEnt, err := dcEnt.ListAudiences()
			if err != nil {
				return err
			}

			dcComm := NewCommunitySignupCampaign(s.srvSheets, s.mg)
			audCom, err := dcComm.ListAudiences()
			if err != nil {
				return err
			}

			params := SignupCampaignData{
				Name:                info.Name,
				Cluster:             info.Cluster,
				Product:             info.Product,
				ProductDisplayName:  SupportedProducts[info.Product].DisplayName,
				IsEnterpriseProduct: IsEnterpriseProduct(info.Product),
				TwitterHandle:       SupportedProducts[info.Product].TwitterHandle,
				QuickstartLink:      SupportedProducts[info.Product].QuickstartLink,
			}

			var dc *mailer.DripCampaign
			if params.IsEnterpriseProduct {
				if !audEnt.Has(info.Email) && !audCom.Has(info.Email) {
					dc = dcEnt
				} else {
					dc = NewEnterpriseFirstTimeCampaign(s.srvSheets, s.mg)
				}
			} else {
				if !audCom.Has(info.Email) && !audEnt.Has(info.Email) {
					dc = dcComm
				}
			}
			if dc != nil {
				fmt.Printf("New user: %s\n", info.Email)
				data, err := json.Marshal(params)
				if err != nil {
					return err
				}
				err = dc.AddContact(mailer.Contact{
					Email: info.Email,
					Data:  string(data),
				})
				if err != nil {
					return err
				}
			}

			//mailer := NewWelcomeMailer(info)
			//err = mailer.SendMail(s.mg, info.Email, info.CC, nil)
			//if err != nil {
			//	return err
			//}

			err = s.recordLicenseEvent(ctx, info, timestamp, EventTypeLicenseIssued)
			return
		}()
	}

	{
		// avoid sending emails for know test emails
		if !knowTestEmails.Has(info.Email) {
			mailer := NewLicenseMailer(LicenseMailData{
				LicenseForm: info,
				License:     string(crtLicense),
			})
			mailer.AttachmentBytes = map[string][]byte{
				fmt.Sprintf("%s-license-%s.txt", info.Product, info.Cluster): crtLicense,
			}
			err = mailer.SendMail(s.mg, info.Email, info.CC, nil)
			if err != nil {
				return err
			}
		}
	}

	{
		if info.Token != "" {
			// mark email as verified
			if exists, err := s.fs.Exists(context.TODO(), EmailVerifiedPath(domain, info.Email)); err == nil && !exists {
				err = s.fs.WriteFile(context.TODO(), EmailVerifiedPath(domain, info.Email), []byte(timestamp))
				if err != nil {
					return err
				}
			}
			respond(ctx, crtLicense)
		} else {
			respond(ctx, []byte("Your license has been emailed!"))
		}
	}

	return nil
}

func (s *Server) recordLicenseEvent(ctx *macaron.Context, info LicenseForm, timestamp string, event LicenseEventType) error {
	domain := Domain(info.Email)

	// record request
	accesslog := LogEntry{
		LicenseForm: info,
		GeoLocation: GeoLocation{
			IP: GetIP(ctx.Req.Request),
		},
		Timestamp: timestamp,
		UA:        uasurfer.Parse(ctx.Req.UserAgent()),
	}
	DecorateGeoData(s.geodb, &accesslog.GeoLocation)

	data, err := json.MarshalIndent(accesslog, "", "  ")
	if err != nil {
		return err
	}

	err = s.fs.WriteFile(context.TODO(), ProductAccessLogPath(domain, info.Product, info.Cluster, timestamp), data)
	if err != nil {
		return err
	}

	err = s.fs.WriteFile(context.TODO(), EmailAccessLogPath(domain, info.Email, info.Product, timestamp), data)
	if err != nil {
		return err
	}

	err = LogLicense(s.sheet, accesslog)
	if err != nil {
		return err
	}

	if len(SupportedProducts[info.Product].MailingLists) > 0 {
		err = s.listmonk.SubscribeToList(listmonkclient.SubscribeRequest{
			Email:        info.Email,
			Name:         info.Name,
			MailingLists: SupportedProducts[info.Product].MailingLists,
		})
		if err != nil {
			return err
		}
	}

	return s.noteEventLicenseIssued(accesslog, event)
}

func (s *Server) GetDomainLicense(domain string, product string) (*ProductLicense, error) {
	if !IsWorkEmail(domain) {
		if IsEnterpriseProduct(product) {
			return nil, apierrors.NewBadRequest("Please provide work email to issue license for Enterprise products.")
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

func (s *Server) CreateOrRetrieveLicense(info LicenseForm, license ProductLicense, cluster string) ([]byte, error) {
	// Return existing license for enterprise products
	if IsEnterpriseProduct(license.Product) {
		exists, err := s.fs.Exists(context.TODO(), LicenseCertPath(license.Domain, license.Product, cluster))
		if err != nil {
			return nil, err
		}
		if exists {
			return s.fs.ReadFile(context.TODO(), LicenseCertPath(license.Domain, license.Product, cluster))
		}
	}
	return s.CreateLicense(info, license, cluster, nil)
}

func (s *Server) CreateLicense(info LicenseForm, license ProductLicense, cluster string, ff FeatureFlags) ([]byte, error) {
	// agreement, TTL
	sans := AltNames{
		DNSNames: []string{cluster},
		EmailAddresses: []string{
			fmt.Sprintf("%s <%s>", info.Name, info.Email),
			info.Email,
		},
	}
	cfg := Config{
		CommonName:         getCN(sans),
		Country:            SupportedProducts[license.Product].ProductLine,
		Province:           SupportedProducts[license.Product].TierName,
		Organization:       SupportedProducts[license.Product].Features,
		OrganizationalUnit: license.Product, // plan
		Locality:           ff.ToSlice(),
		AltNames:           sans,
		Usages:             []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	now := time.Now()
	cfg.NotBefore = now
	if license.Agreement != nil {
		cfg.NotAfter = license.Agreement.ExpiryDate.UTC()
	} else if license.TTL != nil {
		cfg.NotAfter = now.Add(license.TTL.Duration).UTC()
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

	return cert.EncodeCertPEM(crt), nil
}

func LogLicense(si *gdrive.Spreadsheet, info LogEntry) error {
	const sheetName = "License Issue Log"

	sheetId, err := si.EnsureSheet(sheetName, LogEntry{}.Headers())
	if err != nil {
		return err
	}
	return si.AppendRowData(sheetId, info.Data(), false)
}
