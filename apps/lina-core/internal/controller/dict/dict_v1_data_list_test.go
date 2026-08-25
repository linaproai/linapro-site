// This file verifies dictionary data API projections keep the host core
// contract to value and label and omit workbench tag-style fields.

package dict

import (
	"encoding/json"
	"strings"
	"testing"

	"lina-core/internal/model/entity"
	dictsvc "lina-core/internal/service/dict"
)

// TestDictDataItemOmitsWorkbenchTagStyle drives the shipped dictionary DTO
// mapper and asserts tagStyle/cssClass are not part of the host contract.
func TestDictDataItemOmitsWorkbenchTagStyle(t *testing.T) {
	item := dictDataItem(&dictsvc.DictDataProjection{
		SysDictData: &entity.SysDictData{
			Id:       1,
			DictType: "sys_user_sex",
			Label:    "Male",
			Value:    "1",
			TagStyle: "primary",
			CssClass: "text-green",
		},
	})
	if item.Label != "Male" || item.Value != "1" {
		t.Fatalf("expected value+label core fields, got label=%q value=%q", item.Label, item.Value)
	}
	encoded, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal dict data item: %v", err)
	}
	body := string(encoded)
	for _, token := range []string{"tagStyle", "cssClass"} {
		if strings.Contains(body, token) {
			t.Fatalf("dictionary core JSON must not include %s, got %s", token, body)
		}
	}
}
