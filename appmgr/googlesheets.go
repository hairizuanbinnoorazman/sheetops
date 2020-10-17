package appmgr

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hairizuanbinnoorazman/sheetops/logger"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheets struct {
	logger logger.Logger
	sheets *sheets.Service
}

func NewGoogleSheets(logger logger.Logger, sheets *sheets.Service) GoogleSheets {
	return GoogleSheets{
		logger: logger,
		sheets: sheets,
	}
}

// GetValues is a function that would retrieve the values from defined spreadsheetID, sheetID etc
// First row will be ignored - assumed header row
func (g *GoogleSheets) GetValues(spreadsheetID string, cellRef string) ([]AppSetting, error) {
	if spreadsheetID == "" || cellRef == "" {
		return []AppSetting{}, fmt.Errorf("Empty inputs passed to function")
	}
	zz := g.sheets.Spreadsheets.Values.Get(spreadsheetID, cellRef)
	zz = zz.Context(context.TODO())
	resp, err := zz.Do()
	if err != nil {
		return []AppSetting{}, fmt.Errorf("Unable to retrieve values from spreadsheet. Err: %v", err)
	}
	if len(resp.Values) <= 1 {
		return []AppSetting{}, fmt.Errorf("No data will be available. First row is headers so there should be at least 2 rows of data")
	}
	var settings []AppSetting
	for idx, vals := range resp.Values {
		if len(vals) != 3 {
			return []AppSetting{}, fmt.Errorf("Unexpected only 3 columns")
		}
		// Check headers
		if idx == 0 {
			if vals[0].(string) != "App Name" || vals[1].(string) != "Image" || vals[2].(string) != "Replicas" {
				return []AppSetting{}, fmt.Errorf("This is unexpected dataset. We need the following columns: 'App Name', 'Image' and 'Replicas'")
			}
		}

		if idx >= 1 {
			replicas := vals[2].(string)
			replicasInt, err := strconv.Atoi(replicas)
			if err != nil {
				g.logger.Errorf("Unable to parse. Assume minimum of 1")
				replicasInt = 1
			}

			a := AppSetting{
				Name:     vals[0].(string),
				Image:    vals[1].(string),
				Replicas: replicasInt,
			}
			settings = append(settings, a)
		}
	}
	return settings, nil
}
