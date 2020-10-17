package appmgr

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/hairizuanbinnoorazman/sheetops/logger"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func sheetsClientHelper() *sheets.Service {
	credJSON, _ := ioutil.ReadFile("../sheetops-auth.json")
	xClient, _ := sheets.NewService(context.Background(), option.WithCredentialsJSON(credJSON))
	return xClient
}

func TestGoogleSheets_GetValues(t *testing.T) {
	type fields struct {
		logger logger.Logger
		sheets *sheets.Service
	}
	type args struct {
		spreadsheetID string
		cellRef       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []AppSetting
		wantErr bool
	}{
		{
			name: "successful case",
			fields: fields{
				logger: logger.LoggerForTests{Tester: t},
				sheets: sheetsClientHelper(),
			},
			args: args{
				spreadsheetID: "",
				cellRef:       "Sheet2!A1:C",
			},
			want: []AppSetting{
				AppSetting{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GoogleSheets{
				logger: tt.fields.logger,
				sheets: tt.fields.sheets,
			}
			got, err := g.GetValues(tt.args.spreadsheetID, tt.args.cellRef)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoogleSheets.GetValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GoogleSheets.GetValues() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
