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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"time"

	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	"github.com/gocarina/gocsv"
	"github.com/pkg/errors"
	csvtypes "gomodules.xyz/encoding/csv/types"
	gdrive "gomodules.xyz/gdrive-utils"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/macaron.v1"
)

const (
	MailCareer = "career+qa@appscode.com"
)

type ConfigType string

const (
	ConfigTypeQuestion ConfigType = "QuestionConfig"
)

type QuestionConfig struct {
	ConfigType            ConfigType        `json:"configType" csv:"Config Type"`
	TestName              string            `json:"testName" csv:"Test Name"`
	QuestionTemplateDocId string            `json:"questionTemplateDocId" csv:"Question Template Doc Id"`
	StartDate             csvtypes.Date     `json:"startDate" csv:"Start Date"`
	EndDate               csvtypes.Date     `json:"endDate" csv:"End Date"`
	Duration              csvtypes.Duration `json:"duration"  csv:"Duration"`
}

const (
	ProjectConfigSheet = "config"
	ProjectTestSheet   = "test"
)

func SaveConfig(svcSheets *sheets.Service, configDocId string, cfg QuestionConfig) error {
	w := gdrive.NewRowWriter(svcSheets, configDocId, ProjectConfigSheet, &gdrive.Predicate{
		Header: "Config Type",
		By: func(column []any) (int, error) {
			for i, v := range column {
				if v.(string) == string(cfg.ConfigType) {
					return i, nil
				}
			}
			return -1, io.EOF
		},
	})

	data := []*QuestionConfig{
		&cfg,
	}
	return gocsv.MarshalCSV(data, w)
}

func loadConfig(svcSheets *sheets.Service, configDocId string) (*QuestionConfig, error) {
	r, err := gdrive.NewRowReader(svcSheets, configDocId, ProjectConfigSheet, &gdrive.Predicate{
		Header: "Config Type",
		By: func(column []any) (int, error) {
			for i, v := range column {
				if v.(string) == string(ConfigTypeQuestion) {
					return i, nil
				}
			}
			return -1, io.EOF
		},
	})
	if err == io.EOF {
		return nil, errors.New("Question Config not found!")
	} else if err != nil {
		return nil, err
	}

	configs := []*QuestionConfig{}
	if err := gocsv.UnmarshalCSV(r, &configs); err != nil { // Load clients from file
		return nil, err
	}
	return configs[0], nil
}

func LoadConfig(svcSheets *sheets.Service, c cache.Cache, configDocId string) (*QuestionConfig, error) {
	key := fmt.Sprintf("api/LoadConfig/%s", configDocId)
	out := c.Get(key)
	if out == nil {
		cfg, err := loadConfig(svcSheets, configDocId)
		if err != nil {
			return nil, err
		}
		out = cfg
		_ = c.Put(key, out, 10*60) // cache for 10 minutes
	} else {
		log.Println(key, "found")
	}
	return out.(*QuestionConfig), nil
}

type TestAnswer struct {
	Email     string             `json:"email" csv:"Email"`
	DocId     string             `json:"docId"  csv:"Doc Id"`
	StartDate csvtypes.Timestamp `json:"startDate" csv:"Start Date"`
	EndDate   csvtypes.Timestamp `json:"endDate" csv:"End Date"`
	IP        string             `json:"ip,omitempty" csv:"IP"`
	City      string             `json:"city,omitempty" csv:"City"`
	Country   string             `json:"country,omitempty" csv:"Country"`
}

func SaveTestAnswer(svcSheets *sheets.Service, configDocId string, ans TestAnswer) error {
	w := gdrive.NewRowWriter(svcSheets, configDocId, ProjectTestSheet, &gdrive.Predicate{
		Header: "Email",
		By: func(column []any) (int, error) {
			for i, v := range column {
				if v.(string) == ans.Email {
					return i, nil
				}
			}
			return -1, io.EOF
		},
	})

	data := []*TestAnswer{
		&ans,
	}
	return gocsv.MarshalCSV(data, w)
}

func LoadTestAnswer(svcSheets *sheets.Service, configDocId, email string) (*TestAnswer, error) {
	r, err := gdrive.NewRowReader(svcSheets, configDocId, ProjectTestSheet, &gdrive.Predicate{
		Header: "Email",
		By: func(column []any) (int, error) {
			for i, v := range column {
				if v.(string) == email {
					return i, nil
				}
			}
			return -1, io.EOF
		},
	})
	//if err == io.EOF {
	//	return nil, errors.Errorf("%s has not started the test yet!", email)
	//} else
	if err != nil {
		return nil, err
	}

	answers := []*TestAnswer{}
	if err := gocsv.UnmarshalCSV(r, &answers); err != nil {
		return nil, err
	}
	return answers[0], nil
}

func (s *Server) startTest(c cache.Cache, ip string, configDocId, email string) error {
	// already submitted
	// started and x min left to finish the test, redirect, embed
	// did not start, copy file, stat clock

	now := time.Now()

	cfg, err := LoadConfig(s.srvSheets, c, configDocId)
	if err != nil {
		return err
	}

	if now.After(cfg.EndDate.Time) {
		return fmt.Errorf("time passed for this test")
	}
	ans, err := LoadTestAnswer(s.srvSheets, configDocId, email)
	if err != nil && err != io.EOF {
		return err
	}
	if err == nil {
		if now.After(ans.EndDate.Time) {
			return fmt.Errorf("%s passed after test has ended", time.Since(ans.EndDate.Time))
		}
	} else {
		location := GeoLocation{
			IP: ip,
		}
		DecorateGeoData(s.geodb, &location)

		ans = &TestAnswer{
			Email:     email,
			DocId:     "",
			StartDate: csvtypes.Timestamp{Time: now},
			EndDate:   csvtypes.Timestamp{Time: now.Add(cfg.Duration.Duration)},
			IP:        location.IP,
			City:      location.City,
			Country:   location.Country,
		}

		folderId, err := gdrive.GetFolderId(s.srvDrive, configDocId, path.Join("candidates", email))
		if err != nil {
			return err
		}
		docName := fmt.Sprintf("%s - %s %s", email, cfg.TestName, ans.StartDate.Format("2006-01-02"))
		replacements := map[string]string{
			"{{email}}":    email,
			"{{duration}}": fmt.Sprintf("%d minutes", int(cfg.Duration.Minutes())),
		}
		if tz, err := time.LoadLocation(location.Timezone); err == nil {
			replacements["{{start-time}}"] = ans.StartDate.In(tz).Format(time.RFC1123)
			replacements["{{end-time}}"] = ans.EndDate.In(tz).Format(time.RFC1123)
		} else {
			replacements["{{start-time}}"] = ans.StartDate.Format(time.RFC1123)
			replacements["{{end-time}}"] = ans.EndDate.Format(time.RFC1123)
		}

		docId, err := gdrive.CopyDoc(
			s.srvDrive, s.srvDoc, cfg.QuestionTemplateDocId, folderId, docName, replacements)
		if err != nil {
			return err
		}
		ans.DocId = docId

		err = SaveTestAnswer(s.srvSheets, configDocId, *ans)
		if err != nil {
			return err
		}
	}
	// fmt.Printf("%s left to take the test!\n", time.Until(ans.EndDate.Time))
	// fmt.Printf("https://docs.google.com/document/d/%s/edit\n", ans.DocId)

	_, err = gdrive.AddPermission(s.srvDrive, ans.DocId, email, "writer")
	if err != nil {
		return err
	}

	args, err := json.Marshal(ans)
	if err != nil {
		return err
	}
	fn := func(b []byte) error {
		err := s.RevokePermission(b)
		if err != nil {
			log.Printf("failed to revoke permission for docId %s, email %s, err: %v", ans.DocId, ans.Email, err)
		} else {
			log.Printf("revoked permission for docId %s, email %s", ans.DocId, ans.Email)
		}
		return err
	}
	err = s.sch.Schedule(ans.EndDate.Time, fn, args)
	if err != nil {
		return errors.Wrapf(err, "failed to schedule revoke permission task for docId %s, email %s", ans.DocId, ans.Email)
	}

	// mail career
	mailer := NewTestStartedMailer(cfg.TestName, ans)
	fmt.Println("sending email for generated offer letter", ans.Email)
	return mailer.SendMail(s.mg, MailCareer, "", nil)
}

func (s *Server) RevokePermission(args []byte) error {
	var ans TestAnswer
	err := json.Unmarshal(args, &ans)
	if err != nil {
		return err
	}
	return gdrive.RevokePermission(s.srvDrive, ans.DocId, ans.Email)
}

func (s *Server) RegisterQAAPI(m *macaron.Macaron) {
	m.Get("/_/qa/:configDocId/", func(ctx *macaron.Context, c cache.Cache, log *log.Logger) {
		configDocId := ctx.Params("configDocId")

		cfg, err := LoadConfig(s.srvSheets, c, configDocId)
		if err != nil {
			ctx.Data["Err"] = "Time passed for this test"
			log.Println(err)
		} else if time.Now().After(cfg.EndDate.Time) {
			ctx.Data["Err"] = "Time passed for this test"
			log.Println(err)
		} else {
			ctx.Data["TimeLeft"] = time.Until(cfg.EndDate.Time).Round(time.Minute)
			ctx.Data["Duration"] = int(cfg.Duration.Minutes())
		}
		if cfg != nil {
			ctx.Data["TestName"] = cfg.TestName
		}

		ctx.Data["ConfigDocId"] = configDocId
		ctx.HTML(200, "qa_form") // 200 is the response code.
	})

	m.Post("/_/qa/:configDocId/start", binding.Bind(RegisterRequest{}), func(ctx *macaron.Context, info RegisterRequest, c cache.Cache, log *log.Logger) {
		configDocId := ctx.Params("configDocId")

		go func() {
			err := s.startTest(c, GetIP(ctx.Req.Request), configDocId, info.Email)
			if err != nil {
				log.Println(err)
			}
		}()

		ctx.Data["ConfigDocId"] = configDocId
		ctx.Data["Email"] = info.Email
		ctx.Req.URL.Path = ctx.Req.URL.Path + "/load"
		ctx.HTML(200, "qa_start") // 200 is the response code.
	})

	m.Get("/_/qa/status", func(ctx *macaron.Context) {
		configDocId := ctx.QueryTrim("id")
		email := ctx.QueryTrim("email")

		ans, err := LoadTestAnswer(s.srvSheets, configDocId, email)
		if err == io.EOF {
			ctx.JSON(200, map[string]any{
				"wait": true,
			})
			return
		} else if err != nil {
			ctx.JSON(200, map[string]any{
				"err": err.Error(),
			})
			return
		}
		ctx.JSON(200, map[string]any{
			"docId":   ans.DocId,
			"endTime": ans.EndDate.Time.UTC().Format(time.RFC3339),
		})
	})
}
