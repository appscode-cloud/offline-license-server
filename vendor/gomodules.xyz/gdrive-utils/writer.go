package gdrive_utils

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gocarina/gocsv"
	"google.golang.org/api/sheets/v4"
)

type SheetWriter struct {
	srv           *sheets.Service
	spreadsheetId string
	sheetName     string

	ValueRenderOption    string
	DateTimeRenderOption string

	data [][]string
	e    error

	filter *Predicate
}

var _ gocsv.CSVWriter = &SheetWriter{}

func NewWriter(srv *sheets.Service, spreadsheetId, sheetName string) *SheetWriter {
	return &SheetWriter{
		srv:                  srv,
		spreadsheetId:        spreadsheetId,
		sheetName:            sheetName,
		ValueRenderOption:    "FORMATTED_VALUE",
		DateTimeRenderOption: "SERIAL_NUMBER",
	}
}

func NewRowWriter(srv *sheets.Service, spreadsheetId, sheetName string, predicate *Predicate) *SheetWriter {
	return &SheetWriter{
		srv:                  srv,
		spreadsheetId:        spreadsheetId,
		sheetName:            sheetName,
		ValueRenderOption:    "FORMATTED_VALUE",
		DateTimeRenderOption: "SERIAL_NUMBER",
		filter:               predicate,
	}
}

func (w *SheetWriter) Write(row []string) error {
	out := make([]string, len(row))
	copy(out, row)
	w.data = append(w.data, out)
	return nil
}

func (w *SheetWriter) Flush() {
	err := w.ensureSheet(w.sheetName)
	if err != nil {
		w.e = err
		return
	}

	// read first column
	readRange := fmt.Sprintf("%s!A:A", w.sheetName)
	resp, err := w.srv.Spreadsheets.Values.Get(w.spreadsheetId, readRange).
		MajorDimension("COLUMNS").
		ValueRenderOption(w.ValueRenderOption).
		DateTimeRenderOption(w.DateTimeRenderOption).
		Do()
	if err != nil {
		w.e = fmt.Errorf("unable to retrieve data from sheet: %v", err)
		return
	}

	var vals sheets.ValueRange

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		vals = sheets.ValueRange{
			MajorDimension: "ROWS",
			Range:          fmt.Sprintf("%s!A%d", w.sheetName, 1),
			Values:         make([][]interface{}, len(w.data)),
		}
		for i := range w.data {
			vals.Values[i] = make([]interface{}, len(w.data[i]))
			for j := range w.data[i] {
				vals.Values[i][j] = w.data[i][j]
			}
		}
	} else {
		// read first row == header row
		readRange := fmt.Sprintf("%s!1:1", w.sheetName)
		headerResp, err := w.srv.Spreadsheets.Values.Get(w.spreadsheetId, readRange).
			ValueRenderOption(w.ValueRenderOption).
			DateTimeRenderOption(w.DateTimeRenderOption).
			Do()
		if err != nil {
			w.e = fmt.Errorf("unable to retrieve data from sheet: %v", err)
			return
		}

		type Index struct {
			Before int
			After  int
		}

		headerMap := map[string]*Index{}
		headerLength := 0
		for i, header := range headerResp.Values[0] {
			headerMap[header.(string)] = &Index{
				Before: i,
				After:  -1,
			}
			headerLength++
		}
		newHeaderStart := headerLength
		var newHeaders []interface{}

		for i, header := range w.data[0] {
			if _, ok := headerMap[header]; ok {
				headerMap[header].After = i
			} else {
				headerMap[header] = &Index{
					Before: headerLength,
					After:  i,
				}
				newHeaders = append(newHeaders, header)
				headerLength++
			}
		}

		idmap := map[int]int{}
		for _, index := range headerMap {
			if index.After != -1 {
				idmap[index.After] = index.Before
			}
		}

		if len(newHeaders) > 0 {
			// add new headers

			var sb strings.Builder
			sb.WriteRune(rune('A' + newHeaderStart))
			headerVals := sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          fmt.Sprintf("%s!%s%d", w.sheetName, sb.String(), 1),
				Values: [][]interface{}{
					newHeaders,
				},
			}
			_, err = w.srv.Spreadsheets.Values.Append(w.spreadsheetId, headerVals.Range, &headerVals).
				IncludeValuesInResponse(false).
				InsertDataOption("OVERWRITE").
				ValueInputOption("USER_ENTERED").
				Do()
			if err != nil {
				w.e = fmt.Errorf("unable to write new headers to sheet: %v", err)
				return
			}
		}

		if w.filter != nil {
			// read column
			// detect index

			index, ok := headerMap[w.filter.Header]
			if !ok {
				w.e = fmt.Errorf("missing header %s", w.filter.Header)
				return
			}

			var sb strings.Builder
			sb.WriteRune(rune('A' + index.Before))

			// read column
			readRange := fmt.Sprintf("%s!%s2:%s", w.sheetName, sb.String(), sb.String())
			resp, err := w.srv.Spreadsheets.Values.Get(w.spreadsheetId, readRange).
				MajorDimension("COLUMNS").
				ValueRenderOption("FORMATTED_VALUE").
				DateTimeRenderOption("SERIAL_NUMBER").
				Do()
			if err != nil {
				w.e = fmt.Errorf("unable to retrieve data from sheet: %v", err)
				return
			}
			if len(resp.Values) == 0 {
				// column only has header row
				w.e = io.EOF
				return
			}

			idx, err := w.filter.By(resp.Values[0])
			if err != nil {
				w.e = err
				return
			}
			if idx == -1 {
				w.e = io.EOF
				return
			}

			vals = sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          fmt.Sprintf("%s!A%d", w.sheetName, idx+2),
				Values:         make([][]interface{}, len(w.data)-1), // skip header
			}
			// reorder values as idmap
			d22 := w.data[1:]
			for i := range d22 {
				vals.Values[i] = make([]interface{}, headerLength) // header length
				for j := range d22[i] {
					vals.Values[i][idmap[j]] = d22[i][j]
				}
			}
			// update row in place
			_, err = w.srv.Spreadsheets.Values.Update(w.spreadsheetId, vals.Range, &vals).
				IncludeValuesInResponse(false).
				ValueInputOption("USER_ENTERED").
				Do()
			if err != nil {
				w.e = fmt.Errorf("unable to write data to sheet: %v", err)
				return
			}
			return // Done
		} else {
			vals = sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          fmt.Sprintf("%s!A%d", w.sheetName, 1+len(resp.Values[0])),
				Values:         make([][]interface{}, len(w.data)-1), // skip header
			}
		}
		// reorder values as idmap
		d22 := w.data[1:]
		for i := range d22 {
			vals.Values[i] = make([]interface{}, headerLength) // header length
			for j := range d22[i] {
				vals.Values[i][idmap[j]] = d22[i][j]
			}
		}
	}

	_, err = w.srv.Spreadsheets.Values.Append(w.spreadsheetId, vals.Range, &vals).
		IncludeValuesInResponse(false).
		InsertDataOption("OVERWRITE"). // INSERT_ROWS
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		w.e = fmt.Errorf("unable to write data to sheet: %v", err)
		return
	}
}

func (w *SheetWriter) Error() error {
	return w.e
}

func (w *SheetWriter) getSheetId(name string) (int64, error) {
	resp, err := w.srv.Spreadsheets.Get(w.spreadsheetId).Do()
	if err != nil {
		return -1, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	var id int64 = -1
	for _, sheet := range resp.Sheets {
		if sheet.Properties.Title == name {
			id = sheet.Properties.SheetId
		}
	}

	return id, nil
}

func (w *SheetWriter) addNewSheet(name string) error {
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

	_, err := w.srv.Spreadsheets.BatchUpdate(w.spreadsheetId, rbb).Context(context.Background()).Do()
	if err != nil {
		return err
	}

	return nil
}

func (w *SheetWriter) ensureSheet(name string) error {
	sheetId, err := w.getSheetId(name)
	if err != nil {
		return err
	}
	if sheetId >= 0 {
		return nil
	}
	return w.addNewSheet(name)
}
