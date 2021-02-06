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

package cmds

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/appscodelabs/offline-license-server/pkg/server"
	"github.com/oschwald/geoip2-golang"
	"github.com/spf13/cobra"
	"gocloud.dev/blob"
	. "gomodules.xyz/email-providers"
)

func NewCmdGenerateAccessLogCSV() *cobra.Command {
	LicenseBucket := server.LicenseBucket
	GeoCityDatabase := "/home/tamal/Downloads/1a/geoip/GeoLite2-Country_20201124/GeoLite2-City.mmdb"
	AccessLogFile := "/home/tamal/Downloads/1a/license-access-log.csv"
	cmd := &cobra.Command{
		Use:               "generate-access-log",
		Short:             `Generate Access Log in CSV format`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.OpenFile(AccessLogFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer func() {
				f.Close()
			}()
			w := csv.NewWriter(f)

			db, err := geoip2.Open(GeoCityDatabase)
			if err != nil {
				return err
			}
			defer func() {
				db.Close()
			}()

			bucket, err := blob.OpenBucket(context.TODO(), "gs://"+LicenseBucket)
			if err != nil {
				return err
			}
			defer func() {
				bucket.Close()
			}()

			iter := bucket.List(&blob.ListOptions{
				Prefix:     "domains",
				Delimiter:  "",
				BeforeList: nil,
			})

			records := make([]*server.LogEntry, 0)
			for {
				obj, err := iter.Next(context.TODO())
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
				if obj.IsDir {
					continue
				}
				parse := strings.Contains(obj.Key, "/products/") && (strings.Contains(obj.Key, "/accesslog/") ||
					strings.Contains(obj.Key, "/full-license-issued/"))
				if !parse {
					continue
				}

				data, err := bucket.ReadAll(context.TODO(), obj.Key)
				if err != nil {
					return err
				}
				accesslog := struct {
					server.LicenseForm
					IP string
				}{}
				err = json.Unmarshal(data, &accesslog)
				if err != nil {
					return err
				}

				if Domain(accesslog.Email) == "appscode.com" {
					continue
				}

				// fmt.Printf("%s = %+v\n", obj.Key, accesslog)

				parts := strings.Split(obj.Key, "/")

				entry := server.LogEntry{
					LicenseForm: accesslog.LicenseForm,
					Timestamp:   parts[len(parts)-1],
					GeoLocation: server.GeoLocation{
						IP: accesslog.IP,
					},
				}
				server.DecorateGeoData(db, &entry.GeoLocation)
				records = append(records, &entry)
			}

			sort.Slice(records, func(i, j int) bool {
				ti, ei := time.Parse(time.RFC3339, records[i].Timestamp)
				tj, ej := time.Parse(time.RFC3339, records[j].Timestamp)
				if ei != nil || ej != nil {
					return false
				}
				return ti.Before(tj)
			})

			for _, entry := range records {
				if err := w.Write(entry.Data()); err != nil {
					return err
				}
			}
			// Write any buffered data to the underlying writer (standard output).
			w.Flush()
			return w.Error()
		},
	}
	cmd.Flags().StringVar(&LicenseBucket, "bucket", LicenseBucket, "Name of GCS bucket used to store licenses")
	cmd.Flags().StringVar(&GeoCityDatabase, "geo-city-database-file", GeoCityDatabase, "Path to GeoLite2-City.mmdb")
	cmd.Flags().StringVar(&AccessLogFile, "access-log-file", AccessLogFile, "Path to access log csv file")

	return cmd
}
