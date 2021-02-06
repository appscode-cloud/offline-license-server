package server

import (
	"fmt"

	"github.com/avct/uasurfer"
	"sigs.k8s.io/yaml"
)

func NewQuotationProcessFailedMailer(gen *QuotationGenerator, err error) Mailer {
	var src string

	info := struct {
		Lead     QuotationForm       `json:"lead"`
		UA       *uasurfer.UserAgent `json:"ua"`
		Location GeoLocation         `json:"location"`
		Err      string              `json:"error"`
	}{
		Lead:     gen.Lead,
		UA:       gen.UA,
		Location: gen.Location,
		Err:      err.Error(),
	}
	data, err := yaml.Marshal(info)
	if err != nil {
		src = fmt.Sprintf("%+v", info)
	} else {
		src = string(data)
	}

	return Mailer{
		Sender:          MailSales,
		BCC:             "",
		ReplyTo:         MailSales,
		Subject:         "[URGENT] Quotation request processing failed",
		Body:            src,
		params:          nil,
		AttachmentBytes: nil,
		EnableTracking:  false,
	}
}
