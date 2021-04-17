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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"moul.io/http2curl"
)

func SubscribeToMailingList(info LicenseForm) error {
	params := url.Values{}
	params.Add("email", info.Email)
	params.Add("name", info.Name)
	if plan, ok := supportedProducts[info.Product]; ok {
		if len(plan.MailingLists) == 0 {
			return nil
		}

		for _, listID := range plan.MailingLists {
			params.Add("l", listID)
		}
	}
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest(http.MethodPost, MailingListSubscriptionURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	command, _ := http2curl.GetCurlCommand(req)
	fmt.Println(command)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(buf.String())

	return nil
}
