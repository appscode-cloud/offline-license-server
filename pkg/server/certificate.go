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
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"
)

func getCN(sans AltNames) string {
	if len(sans.DNSNames) > 0 {
		return sans.DNSNames[0]
	}
	if len(sans.IPs) > 0 {
		return sans.IPs[0].String()
	}
	return ""
}

type AltNames struct {
	DNSNames       []string
	IPs            []net.IP
	EmailAddresses []string
}

// Config contains the basic fields required for creating a certificate
type Config struct {
	CommonName          string
	Organization        []string
	AltNames            AltNames
	Usages              []x509.ExtKeyUsage
	NotBefore, NotAfter time.Time // Validity bounds.
}

// NewSignedCert creates a signed certificate using the given CA certificate and key
func NewSignedCert(cfg Config, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, errors.New("must specify a CommonName")
	}
	if len(cfg.Usages) == 0 {
		return nil, errors.New("must specify at least one ExtKeyUsage")
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:       cfg.AltNames.DNSNames,
		IPAddresses:    cfg.AltNames.IPs,
		EmailAddresses: cfg.AltNames.EmailAddresses,
		SerialNumber:   serial,
		NotBefore:      cfg.NotBefore,
		NotAfter:       cfg.NotAfter,
		KeyUsage:       x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:    cfg.Usages,
	}
	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}
