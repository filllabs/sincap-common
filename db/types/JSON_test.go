package types_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/filllabs/sincap-common/db/types"
)

var jArr = types.JSON(`["John","Jane"]`)
var jMap = types.JSON(`{"Name":"John","Surname":"Doe"}`)

func TestJSON_UnmarshalArray(t *testing.T) {

	var arr []interface{}
	json.Unmarshal(jArr, &arr)

	tests := []struct {
		name    string
		j       *types.JSON
		want    []interface{}
		wantErr bool
	}{
		{name: "nil", j: nil, wantErr: true},
		{name: "positive", j: &jArr, want: arr},
		{name: "negative", j: &jMap, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.j.UnmarshalArray()
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON.UnmarshalArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JSON.UnmarshalArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSON_UnmarshalMap(t *testing.T) {
	var m map[string]interface{}
	json.Unmarshal(jMap, &m)

	tests := []struct {
		name    string
		j       *types.JSON
		want    map[string]interface{}
		wantErr bool
	}{
		{name: "nil", j: nil, wantErr: true},
		{name: "positive", j: &jMap, want: m},
		{name: "negative", j: &jArr, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.j.UnmarshalMap()
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON.UnmarshalMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JSON.UnmarshalMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
