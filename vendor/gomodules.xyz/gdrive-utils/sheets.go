package gdrive_utils

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/api/sheets/v4"
)

type Spreadsheet struct {
	srv           *sheets.Service
	SpreadSheetId string
}

func NewSpreadsheet(srv *sheets.Service, spreadsheetId string) (*Spreadsheet, error) {
	return &Spreadsheet{
		srv:           srv,
		SpreadSheetId: spreadsheetId,
	}, nil
}

// ref: https://developers.google.com/sheets/api/guides/batchupdate
func (si *Spreadsheet) updateRowData(sheetId, row int64, data []string, formatCell bool) error {
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
				StringValue: &data[i],
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
					SheetId:     sheetId,
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
func (si *Spreadsheet) AppendRowData(sheetId int64, data []string, formatCell bool) error {
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
				StringValue: &data[i],
			},
		})
	}

	req := []*sheets.Request{
		{
			AppendCells: &sheets.AppendCellsRequest{
				SheetId: sheetId,
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
	var id int64 = -1
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

func (si *Spreadsheet) EnsureSheet(name string, headers []string) (int64, error) {
	sheetId, err := si.getSheetId(name)
	if err != nil {
		return -1, err
	}
	if sheetId >= 0 {
		return sheetId, nil
	}

	// create worksheet

	err = si.addNewSheet(name)
	if err != nil {
		return 0, err
	}

	sheetId, err = si.getSheetId(name)
	if err != nil {
		return -1, err
	}

	if len(headers) > 0 {
		err = si.ensureHeader(sheetId, headers)
		if err != nil {
			return -1, err
		}
	}

	return sheetId, nil
}

func (si *Spreadsheet) ensureHeader(sheetId int64, headers []string) error {
	return si.updateRowData(sheetId, 0, headers, true)
}

func (si *Spreadsheet) FindEmptyCell(sheetName string) (string, error) {
	resp, err := si.srv.Spreadsheets.GetByDataFilter(si.SpreadSheetId, &sheets.GetSpreadsheetByDataFilterRequest{
		DataFilters: []*sheets.DataFilter{
			{
				A1Range:                 sheetName + "!A1:A",
				DeveloperMetadataLookup: nil,
				GridRange:               nil,
				ForceSendFields:         nil,
				NullFields:              nil,
			},
		},
		IncludeGridData: true,
	}).Do()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	for _, s := range resp.Sheets {
		if s.Properties.Title == sheetName {
			n := len(s.Data[0].RowData)
			return s.Data[0].RowData[n-1].Values[0].FormattedValue, nil
		}
	}

	return "", errors.New("no empty cell found")
}
