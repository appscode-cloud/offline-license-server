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
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/sheets/v4"
)

type Spreadsheet struct {
	srv            *sheets.Service
	SpreadSheetId  string
	CurrentSheetID int64
}

func main2() {
	si, err := NewSpreadsheet("1evwv2ON94R38M-Lkrw8b6dpVSkRYHUWsNOuI7X0_-zA") // Share this sheet with the service account email
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	info := LogEntry{
		LicenseForm: LicenseForm{
			Name:    "Fahim Abrar",
			Email:   "fahimabrar@appscode.com",
			Product: "Kubeform Community",
			Cluster: "bad94a42-0210-4c81-b07a-99bae529ec14",
		},
		IP:        "",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	err = si.Append(info)
	if err != nil {
		log.Fatal(err)
	}
}

func NewSpreadsheet(spreadsheetId string) (*Spreadsheet, error) {
	// Set env GOOGLE_APPLICATION_CREDENTIALS to service account json path
	srv, err := sheets.NewService(context.TODO())
	if err != nil {
		return nil, err
	}

	return &Spreadsheet{
		srv:           srv,
		SpreadSheetId: spreadsheetId,
	}, nil
}

func (si *Spreadsheet) getCellData(row, column int64) (string, error) {
	resp, err := si.srv.Spreadsheets.GetByDataFilter(si.SpreadSheetId, &sheets.GetSpreadsheetByDataFilterRequest{
		IncludeGridData: true,
	}).Do()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	var val string

	for _, s := range resp.Sheets {
		if s.Properties.SheetId == si.CurrentSheetID {
			val = s.Data[0].RowData[row].Values[column].FormattedValue
		}
	}

	return val, nil
}

// ref: https://developers.google.com/sheets/api/guides/batchupdate
func (si *Spreadsheet) updateRowData(row int64, data []string, formatCell bool) error {
	var format *sheets.CellFormat

	if formatCell {
		// for updating header color and making it bold
		format = &sheets.CellFormat{
			TextFormat: &sheets.TextFormat{
				Bold: true,
			},
			BackgroundColor: &sheets.Color{
				Alpha: 1,
				Blue:  149.0 / 255.0,
				Green: 226.0 / 255.0,
				Red:   239.0 / 255.0,
			},
		}
	}

	vals := make([]*sheets.CellData, 0, len(data))
	for i := range data {
		vals = append(vals, &sheets.CellData{
			UserEnteredFormat: format,
			UserEnteredValue: &sheets.ExtendedValue{
				StringValue: data[i],
			},
		})
	}

	req := []*sheets.Request{
		{
			UpdateCells: &sheets.UpdateCellsRequest{
				Fields: "*",
				Start: &sheets.GridCoordinate{
					ColumnIndex: 0,
					RowIndex:    row,
					SheetId:     si.CurrentSheetID,
				},
				Rows: []*sheets.RowData{
					{
						Values: vals,
					},
				},
			},
		},
	}
	_, err := si.srv.Spreadsheets.BatchUpdate(si.SpreadSheetId, &sheets.BatchUpdateSpreadsheetRequest{
		IncludeSpreadsheetInResponse: false,
		Requests:                     req,
		ResponseIncludeGridData:      false,
	}).Do()
	if err != nil {
		return fmt.Errorf("unable to update: %v", err)
	}

	return nil
}

// ref: https://developers.google.com/sheets/api/guides/batchupdate
func (si *Spreadsheet) appendRowData(data []string, formatCell bool) error {
	var format *sheets.CellFormat

	if formatCell {
		// for updating header color and making it bold
		format = &sheets.CellFormat{
			TextFormat: &sheets.TextFormat{
				Bold: true,
			},
			BackgroundColor: &sheets.Color{
				Alpha: 1,
				Blue:  149.0 / 255.0,
				Green: 226.0 / 255.0,
				Red:   239.0 / 255.0,
			},
		}
	}

	vals := make([]*sheets.CellData, 0, len(data))
	for i := range data {
		vals = append(vals, &sheets.CellData{
			UserEnteredFormat: format,
			UserEnteredValue: &sheets.ExtendedValue{
				StringValue: data[i],
			},
		})
	}

	req := []*sheets.Request{
		{
			AppendCells: &sheets.AppendCellsRequest{
				SheetId: si.CurrentSheetID,
				Fields:  "*",
				Rows: []*sheets.RowData{
					{
						Values: vals,
					},
				},
			},
		},
	}
	_, err := si.srv.Spreadsheets.BatchUpdate(si.SpreadSheetId, &sheets.BatchUpdateSpreadsheetRequest{
		IncludeSpreadsheetInResponse: false,
		Requests:                     req,
		ResponseIncludeGridData:      false,
	}).Do()
	if err != nil {
		return fmt.Errorf("unable to update: %v", err)
	}

	return nil
}

func (si *Spreadsheet) getSheetId(name string) (int64, error) {
	resp, err := si.srv.Spreadsheets.Get(si.SpreadSheetId).Do()
	if err != nil {
		return -1, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	var id int64
	for _, sheet := range resp.Sheets {
		if sheet.Properties.Title == name {
			id = sheet.Properties.SheetId
		}

	}

	return id, nil
}

func (si *Spreadsheet) addNewSheet(name string) error {
	req := sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: name,
			},
		},
	}

	rbb := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{&req},
	}

	_, err := si.srv.Spreadsheets.BatchUpdate(si.SpreadSheetId, rbb).Context(context.Background()).Do()
	if err != nil {
		return err
	}

	return nil
}

func (si *Spreadsheet) ensureSheet(name string) (int64, error) {
	id, err := si.getSheetId(name)
	if err != nil {
		return 0, err
	}

	if id == 0 {
		err = si.addNewSheet(name)
		if err != nil {
			return 0, err
		}

		id, err = si.getSheetId(name)
		if err != nil {
			return 0, err
		}

		si.CurrentSheetID = id

		err = si.ensureHeader()
		if err != nil {
			return 0, err
		}

		return id, nil
	}

	si.CurrentSheetID = id
	return id, nil
}

func (si *Spreadsheet) ensureHeader() error {
	return si.updateRowData(0, LogEntry{}.Headers(), true)
}

func (si *Spreadsheet) findEmptyCell() (int64, error) {
	resp, err := si.srv.Spreadsheets.GetByDataFilter(si.SpreadSheetId, &sheets.GetSpreadsheetByDataFilterRequest{
		IncludeGridData: true,
	}).Do()
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	for _, s := range resp.Sheets {
		if s.Properties.SheetId == si.CurrentSheetID {
			return int64(len(s.Data[0].RowData)), nil
		}
	}

	return 0, errors.New("no empty cell found")
}

func (si *Spreadsheet) Append(info LogEntry) error {
	_, err := si.ensureSheet("License Issue Log")
	if err != nil {
		return err
	}
	return si.appendRowData(info.Data(), false)
}
