// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"reflect"
	"testing"
)

type tagPair struct {
	name  string
	value string
}

func TestIterTags(t *testing.T) {
	tests := []struct {
		name string
		tag  reflect.StructTag
		want []tagPair
	}{
		{
			name: "empty",
		},
		{
			name: "standard tags",
			tag:  `  path:"/" query:"q" empty:""`,
			want: []tagPair{
				{name: "path", value: "/"},
				{name: "query", value: "q"},
				{name: "empty", value: ""},
			},
		},
		{
			name: "standard key chars",
			tag:  `json-name:"name,omitempty" data_id:"42"`,
			want: []tagPair{
				{name: "json-name", value: "name,omitempty"},
				{name: "data_id", value: "42"},
			},
		},
		{
			name: "standard escaped value",
			tag:  `json:"a\"b"`,
			want: []tagPair{{name: "json", value: `a"b`}},
		},
		{
			name: "standard invalid key stops",
			tag:  `bad key:"value" path:"/"`,
		},
		{
			name: "quoted complex keys",
			tag:  `  "/content/:slug":"value" "key:mod":"value2" "key with spaces":"value3" "":"empty"`,
			want: []tagPair{
				{name: "/content/:slug", value: "value"},
				{name: "key:mod", value: "value2"},
				{name: "key with spaces", value: "value3"},
				{name: "", value: "empty"},
			},
		},
		{
			name: "mixed standard and quoted keys",
			tag:  `path:"/" "/content/:slug":"value" query:"mode"`,
			want: []tagPair{
				{name: "path", value: "/"},
				{name: "/content/:slug", value: "value"},
				{name: "query", value: "mode"},
			},
		},
		{
			name: "quoted escaped key",
			tag:  `"key\"quote":"value"`,
			want: []tagPair{{name: `key"quote`, value: "value"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []tagPair
			for name, value := range IterTags(tt.tag) {
				got = append(got, tagPair{name: name, value: value})
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected tags: got %#v, want %#v", got, tt.want)
			}
		})
	}
}
