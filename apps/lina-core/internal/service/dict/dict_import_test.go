// This file verifies dictionary import/export keep value+label as the core
// contract and do not treat tagStyle or cssClass as shipped columns.

package dict

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	_ "lina-core/pkg/dbdriver"
)

// TestDictImportTemplatesOmitTagStyleColumns verifies download templates expose
// value, label, sort, status, and remark without workbench tag-style columns.
func TestDictImportTemplatesOmitTagStyleColumns(t *testing.T) {
	ctx := context.Background()
	svc := New(nil)

	combined, err := svc.CombinedImportTemplate(ctx)
	if err != nil {
		t.Fatalf("combined import template: %v", err)
	}
	assertExcelHeadersOmitTagStyle(t, combined, "Dictionary Data")

	dataTemplate, err := svc.GenerateDataImportTemplate(ctx)
	if err != nil {
		t.Fatalf("data import template: %v", err)
	}
	assertExcelHeadersOmitTagStyle(t, dataTemplate, "Sheet1")
}

// TestDictExportOmitsTagStyleColumns verifies export workbooks omit tag-style
// columns even when storage rows still have leftover style fields.
func TestDictExportOmitsTagStyleColumns(t *testing.T) {
	ctx := context.Background()
	svc := New(nil)

	exported, err := svc.DataExport(ctx, DataExportInput{
		DictType: fmt.Sprintf("missing_export_%d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("data export: %v", err)
	}
	assertExcelHeadersOmitTagStyle(t, exported, "Sheet1")

	combined, err := svc.CombinedExport(ctx, CombinedExportInput{
		Type: fmt.Sprintf("missing_export_%d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("combined export: %v", err)
	}
	assertExcelHeadersOmitTagStyle(t, combined, "Dictionary Data")
}

// TestDictDataImportUsesValueLabelCoreColumns verifies the shipped import parser
// reads status and remark after sort and does not write tagStyle from that
// column.
func TestDictDataImportUsesValueLabelCoreColumns(t *testing.T) {
	var (
		ctx      = context.Background()
		dictType = insertDictTypeForDeleteGuard(t, ctx, false)
		value    = fmt.Sprintf("core_%d", time.Now().UnixNano())
		svc      = New(nil)
	)
	importData := buildDictDataImportFile(t, []string{
		dictType.Type, "Core Label", value, "7", "0", "core remark",
	})

	result, err := svc.DataImport(ctx, bytes.NewReader(importData), false)
	if err != nil {
		t.Fatalf("import dictionary data: %v", err)
	}
	if result.Success != 1 || result.Fail != 0 {
		t.Fatalf("expected one successful data import, got success=%d fail=%d failures=%#v",
			result.Success, result.Fail, result.FailList)
	}

	var row *entity.SysDictData
	if err = dao.SysDictData.Ctx(ctx).
		Where(do.SysDictData{DictType: dictType.Type, Value: value}).
		Scan(&row); err != nil {
		t.Fatalf("load imported dictionary data: %v", err)
	}
	if row == nil {
		t.Fatal("imported dictionary data row not found")
	}
	if row.Label != "Core Label" || row.Sort != 7 || row.Status != 0 || row.Remark != "core remark" {
		t.Fatalf("unexpected imported core fields: label=%q sort=%d status=%d remark=%q",
			row.Label, row.Sort, row.Status, row.Remark)
	}
	if row.TagStyle != "" || row.CssClass != "" {
		t.Fatalf("import must not write tagStyle/cssClass as core columns, got tagStyle=%q cssClass=%q",
			row.TagStyle, row.CssClass)
	}
}

// assertExcelHeadersOmitTagStyle reads one sheet header row and rejects leftover
// workbench tag-style column names.
func assertExcelHeadersOmitTagStyle(t *testing.T, data []byte, sheet string) {
	t.Helper()

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("open workbook: %v", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			t.Fatalf("close workbook: %v", closeErr)
		}
	}()

	rows, err := f.GetRows(sheet)
	if err != nil {
		t.Fatalf("read sheet %s: %v", sheet, err)
	}
	if len(rows) == 0 {
		t.Fatalf("sheet %s has no header row", sheet)
	}
	joined := strings.Join(rows[0], "|")
	for _, token := range []string{"Tag Style", "CSS Class", "tagStyle", "cssClass"} {
		if strings.Contains(joined, token) {
			t.Fatalf("sheet %s headers must not include %s, got %s", sheet, token, joined)
		}
	}
}
