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
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	MailSender         = "license-issuer@appscode.ninja"
	MailLicenseTracker = "issued-license-tracker@appscode.com"
	MailReplyTo        = "support@appscode.com"
)

const LicenseIssuerName = "AppsCode Inc."

const DefaultTTLForEnterpriseProduct = 14 * 24 * time.Hour
const DefaultTTLForCommunityProduct = 365 * 24 * time.Hour

const LicenseBucket = "appscode-licenses"
const LicenseBucketURL = "gs://" + LicenseBucket
const GoogleApplicationCredentials = "/home/tamal/AppsCode/credentials/license-issuer@appscode-domains.json"

var supportedProducts = sets.NewString(
	"kubedb-community",
	"kubedb-enterprise",
	"stash-community",
	"stash-enterprise",
)

// https://email-verify.my-addr.com/list-of-most-popular-email-domains.php
var publicEmailDomains = sets.NewString(
	"gmail.com",
	"yahoo.com",
	"hotmail.com",
	"aol.com",
	"hotmail.co.uk",
	"hotmail.fr",
	"msn.com",
	"yahoo.fr",
	"wanadoo.fr",
	"orange.fr",
	"comcast.net",
	"yahoo.co.uk",
	"yahoo.com.br",
	"yahoo.co.in",
	"live.com",
	"rediffmail.com",
	"free.fr",
	"gmx.de",
	"web.de",
	"yandex.ru",
	"ymail.com",
	"libero.it",
	"outlook.com",
	"uol.com.br",
	"bol.com.br",
	"mail.ru",
	"cox.net",
	"hotmail.it",
	"sbcglobal.net",
	"sfr.fr",
	"live.fr",
	"verizon.net",
	"live.co.uk",
	"googlemail.com",
	"yahoo.es",
	"ig.com.br",
	"live.nl",
	"bigpond.com",
	"terra.com.br",
	"yahoo.it",
	"neuf.fr",
	"yahoo.de",
	"alice.it",
	"rocketmail.com",
	"att.net",
	"laposte.net",
	"facebook.com",
	"bellsouth.net",
	"yahoo.in",
	"hotmail.es",
	"charter.net",
	"yahoo.ca",
	"yahoo.com.au",
	"rambler.ru",
	"hotmail.de",
	"tiscali.it",
	"shaw.ca",
	"yahoo.co.jp",
	"sky.com",
	"earthlink.net",
	"optonline.net",
	"freenet.de",
	"t-online.de",
	"aliceadsl.fr",
	"virgilio.it",
	"home.nl",
	"qq.com",
	"telenet.be",
	"me.com",
	"yahoo.com.ar",
	"tiscali.co.uk",
	"yahoo.com.mx",
	"voila.fr",
	"gmx.net",
	"mail.com",
	"planet.nl",
	"tin.it",
	"live.it",
	"ntlworld.com",
	"arcor.de",
	"yahoo.co.id",
	"frontiernet.net",
	"hetnet.nl",
	"live.com.au",
	"yahoo.com.sg",
	"zonnet.nl",
	"club-internet.fr",
	"juno.com",
	"optusnet.com.au",
	"blueyonder.co.uk",
	"bluewin.ch",
	"skynet.be",
	"sympatico.ca",
	"windstream.net",
	"mac.com",
	"centurytel.net",
	"chello.nl",
	"live.ca",
	"aim.com",
	"bigpond.net.au",
)
