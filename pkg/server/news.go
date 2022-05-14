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
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/go-macaron/cache"
	"github.com/gocarina/gocsv"
	gdrive "gomodules.xyz/gdrive-utils"
	"gopkg.in/macaron.v1"
)

type NewsSnippet struct {
	Content   string `json:"content" csv:"Content"`
	StartDate Date   `json:"startDate" csv:"Start Date"`
	EndDate   Date   `json:"endDate" csv:"End Date"`
}

func (s *Server) RegisterNewsAPI(m *macaron.Macaron) {
	m.Get("/_/news", func(ctx *macaron.Context, c cache.Cache, log *log.Logger) {
		key := ctx.Req.URL.Path
		out := c.Get(key)
		if out == nil {
			schedule, err := s.NextNewsSnippet()
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			out = schedule
			_ = c.Put(key, out, 60) // cache for 60 seconds
		} else {
			log.Println(key, "found")
		}
		ctx.JSON(http.StatusOK, out)
	})
}

func (s *Server) NextNewsSnippet() (*NewsSnippet, error) {
	now := time.Now()

	reader, err := gdrive.NewRowReader(s.srvSheets, NewsSnippetSpreadsheetId, NewsSnippetSheet, &gdrive.Predicate{
		Header: "End Date",
		By: func(column []interface{}) (int, error) {
			pos := math.MaxInt
			for i, v := range column {
				var d Date
				err := d.UnmarshalCSV(v.(string))
				if err != nil {
					return -1, err
				}
				if d.Time.After(now) {
					pos = i
					break
				}
			}
			if pos == math.MaxInt {
				return -1, io.EOF
			}
			return pos, nil
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
	for i, s := range snippets {
		if s.EndDate.After(now) {
			snippets = snippets[i:]
			break
		}
	}
	if now.After(snippets[0].StartDate.Time) {
		return snippets[0], nil
	}
	return &NewsSnippet{}, nil
}
