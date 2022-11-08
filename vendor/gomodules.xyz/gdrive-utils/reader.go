package gdrive_utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gocarina/gocsv"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/sheets/v4"
)

type Predicate struct {
	Header string
	By     func(v []interface{}) (int, error)
}

type SheetReader struct {
	srv           *sheets.Service
	spreadsheetId string
	sheetName     string
	columnStart   string
	columnEnd     string
	rowStart      int

	idx    int
	header bool

	ValueRenderOption    string
	DateTimeRenderOption string
}

var _ gocsv.CSVReader = &SheetReader{}

func NewReader(srv *sheets.Service, spreadsheetId, sheetName string, rowStart int) (*SheetReader, error) {
	r := &SheetReader{
		srv:                  srv,
		spreadsheetId:        spreadsheetId,
		sheetName:            sheetName,
		columnStart:          "A",
		rowStart:             rowStart,
		idx:                  rowStart,
		ValueRenderOption:    "FORMATTED_VALUE",
		DateTimeRenderOption: "SERIAL_NUMBER",
	}

	values, err := r.readHeader()
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	sb.WriteRune(rune('A'+len(values[0])) - 1)
	r.columnEnd = sb.String()
	return r, nil
}

func NewLastRowReader(srv *sheets.Service, spreadsheetId, sheetName string) (*SheetReader, error) {
	readRange := fmt.Sprintf("%s!A:A", sheetName)
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).
		ValueRenderOption("FORMATTED_VALUE").
		DateTimeRenderOption("SERIAL_NUMBER").
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	if len(resp.Values) == 0 {
		return nil, io.EOF
	}
	return NewReader(srv, spreadsheetId, sheetName, len(resp.Values))
}

func NewColumnReader(srv *sheets.Service, spreadsheetId, sheetName, header string) (*SheetReader, error) {
	r := &SheetReader{
		srv:                  srv,
		spreadsheetId:        spreadsheetId,
		sheetName:            sheetName,
		columnStart:          "A",
		rowStart:             -1,
		idx:                  -1,
		ValueRenderOption:    "FORMATTED_VALUE",
		DateTimeRenderOption: "SERIAL_NUMBER",
	}

	values, err := r.readHeader()
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	sb.WriteRune(rune('A' + len(values[0]) - 1))
	r.columnEnd = sb.String()

	sb.Reset()
	for i, v := range values[0] {
		if v.(string) == header {
			sb.WriteRune(rune('A' + i))
			break
		}
	}
	if sb.Len() == 0 {
		return nil, fmt.Errorf("missing header %s", header)
	}

	r.columnStart = sb.String()
	r.columnEnd = sb.String()
	r.rowStart = 1
	r.idx = 1
	return r, nil
}

func NewRowReader(srv *sheets.Service, spreadsheetId, sheetName string, predicate *Predicate) (*SheetReader, error) {
	if predicate == nil {
		return NewReader(srv, spreadsheetId, sheetName, 1)
	}

	r := &SheetReader{
		srv:                  srv,
		spreadsheetId:        spreadsheetId,
		sheetName:            sheetName,
		columnStart:          "A",
		rowStart:             -1,
		idx:                  -1,
		ValueRenderOption:    "FORMATTED_VALUE",
		DateTimeRenderOption: "SERIAL_NUMBER",
	}

	values, err := r.readHeader()
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	sb.WriteRune(rune('A' + len(values[0]) - 1))
	r.columnEnd = sb.String()

	sb.Reset()
	for i, v := range values[0] {
		if v.(string) == predicate.Header {
			sb.WriteRune(rune('A' + i))
			break
		}
	}
	if sb.Len() == 0 {
		return nil, fmt.Errorf("missing header %s", predicate.Header)
	}

	// read column
	readRange := fmt.Sprintf("%s!%s2:%s", sheetName, sb.String(), sb.String())
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).
		MajorDimension("COLUMNS").
		ValueRenderOption("FORMATTED_VALUE").
		DateTimeRenderOption("SERIAL_NUMBER").
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	if len(resp.Values) == 0 {
		// column only has header row
		return nil, io.EOF
	}

	idx, err := predicate.By(resp.Values[0])
	if err != nil {
		return nil, err
	}
	if idx == -1 {
		return nil, io.EOF
	}

	r.rowStart = idx + 2 // starts from 1, also includes header row
	r.idx = idx + 2
	return r, nil
}

// Read reads one record (a slice of fields) from r.
// If the record has an unexpected number of fields,
// Read returns the record along with the error ErrFieldCount.
// Except for that case, Read always returns either a non-nil
// record or a non-nil error, but not both.
// If there is no data left to be read, Read returns nil, io.EOF.
// If ReuseRecord is true, the returned slice may be shared
// between multiple calls to Read.
func (r *SheetReader) Read() (record []string, err error) {
	if !r.header && r.idx > 1 {
		record, err = r.read(1)
		if err != nil {
			return nil, err
		}
		r.header = true
		return record, err
	}

	record, err = r.read(r.idx)
	if err != nil {
		return nil, err
	}
	r.idx++
	return record, nil
}

func (r *SheetReader) read(idx int) (record []string, err error) {
	readRange := fmt.Sprintf("%s!%s%d:%s%d", r.sheetName, r.columnStart, idx, r.columnEnd, idx)
	resp, err := r.srv.Spreadsheets.Values.Get(r.spreadsheetId, readRange).
		ValueRenderOption(r.ValueRenderOption).
		DateTimeRenderOption(r.DateTimeRenderOption).
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	if resp.Values == nil {
		return nil, io.EOF
	}
	if len(resp.Values) > 1 {
		return nil, fmt.Errorf("multiple rows returned")
	}

	record = make([]string, len(resp.Values[0]))
	for i := range resp.Values[0] {
		record[i] = fmt.Sprintf("%v", resp.Values[0][i])
	}
	return record, nil
}

func (r *SheetReader) readHeader() ([][]interface{}, error) {
	// read first row
	readRange := fmt.Sprintf("%s!1:1", r.sheetName)
	resp, err := r.srv.Spreadsheets.Values.Get(r.spreadsheetId, readRange).
		ValueRenderOption(r.ValueRenderOption).
		DateTimeRenderOption(r.DateTimeRenderOption).
		Do()
	if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusBadRequest {
		return nil, io.EOF
	} else if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	if len(resp.Values) == 0 {
		return nil, io.EOF
	}
	return resp.Values, nil
}

// ReadAll reads all the remaining records from r.
// Each record is a slice of fields.
// A successful call returns err == nil, not err == io.EOF. Because ReadAll is
// defined to read until EOF, it does not treat end of file as an error to be
// reported.
func (r *SheetReader) ReadAll() (records [][]string, err error) {
	readRange := fmt.Sprintf("%s!%s%d:%s", r.sheetName, r.columnStart, r.idx, r.columnEnd)
	resp, err := r.srv.Spreadsheets.Values.Get(r.spreadsheetId, readRange).
		ValueRenderOption(r.ValueRenderOption).
		DateTimeRenderOption(r.DateTimeRenderOption).
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	if resp.Values == nil {
		return nil, io.EOF
	}

	offset := 0
	if !r.header && r.idx > 1 {
		records = make([][]string, len(resp.Values)+1)
		records[0], err = r.read(1)
		if err != nil {
			return nil, err
		}
		r.header = true
		offset = 1
	} else {
		records = make([][]string, len(resp.Values))
		if r.idx == 1 {
			r.header = true
		}
	}

	for i, row := range resp.Values {
		records[i+offset] = make([]string, len(row))
		for j := range row {
			records[i+offset][j] = fmt.Sprintf("%v", row[j])
		}
	}
	r.idx += len(resp.Values)
	return records, nil
}
