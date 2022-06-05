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
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/go-macaron/cache"
	"github.com/gocarina/gocsv"
	csvtypes "gomodules.xyz/encoding/csv/types"
	gdrive "gomodules.xyz/gdrive-utils"
	"gomodules.xyz/sets"
	"gopkg.in/macaron.v1"
)

type NewsSnippet struct {
	Content   string               `json:"content" csv:"Content"`
	Link      string               `json:"link" csv:"Link"`
	StartDate csvtypes.Date        `json:"startDate" csv:"Start Date"`
	EndDate   csvtypes.Date        `json:"endDate" csv:"End Date"`
	Products  csvtypes.StringSlice `json:"products" csv:"Products"`
}

func (s *Server) RegisterNewsAPI(m *macaron.Macaron) {
	m.Get("/_/news", func(ctx *macaron.Context, c cache.Cache, log *log.Logger) {
		key := ctx.Req.URL.Path
		out := c.Get(key)
		if out == nil {
			p := ctx.Query("p")
			if p == "" {
				p = "AppsCode"
			}
			news, err := s.NextNewsSnippet(p)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			out = news
			_ = c.Put(key, out, 60) // cache for 60 seconds
		} else {
			log.Println(key, "found")
		}
		ctx.JSON(http.StatusOK, out)
	})
}

func (s *Server) NextNewsSnippet(p string) (*NewsSnippet, error) {
	now := time.Now()

	reader, err := gdrive.NewRowReader(s.srvSheets, NewsSnippetSpreadsheetId, NewsSnippetSheet, &gdrive.Predicate{
		Header: "End Date",
		By: func(column []interface{}) (int, error) {
			for i, v := range column {
				var d csvtypes.Date
				err := d.UnmarshalCSV(v.(string))
				if err != nil {
					return -1, err
				}
				if d.Time.After(now) {
					return i, nil
				}
			}
			return -1, io.EOF
		},
	})
	if err == io.EOF {
		return &NewsSnippet{}, nil
	} else if err != nil {
		return nil, err
	}

	snippets := []*NewsSnippet{}
	if err := gocsv.UnmarshalCSV(reader, &snippets); err != nil { // Load clients from file
		return nil, err
	}
	sort.Slice(snippets, func(i, j int) bool {
		return snippets[i].EndDate.Before(snippets[j].EndDate.Time)
	})
	// keep unexpired news
	for i, s := range snippets {
		if s.EndDate.After(now) {
			snippets = snippets[i:]
			break
		}
	}
	// only keep relevant product news
	for i := 0; i < len(snippets); i++ {
		if !sets.NewString(snippets[i].Products...).Has(p) {
			snippets = append(snippets[:i], snippets[i+1:]...)
		} else {
			i++
		}
	}

	var result NewsSnippet
	for _, s := range snippets {
		if now.After(s.StartDate.Time) {
			// take the most recently published news
			if result.Content == "" || s.StartDate.After(result.StartDate.Time) {
				result = *s
			}
		}
	}
	return &result, nil
}
