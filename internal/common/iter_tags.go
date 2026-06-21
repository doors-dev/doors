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
	"iter"
	"reflect"
	"strconv"
	"strings"
)

func IterTags(tag reflect.StructTag) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for tags := string(tag); tags != ""; {
			tags = strings.TrimLeft(tags, " ")
			if tags == "" {
				return
			}
			name, rest, ok := cutTagName(tags)
			if !ok || rest == "" || rest[0] != ':' {
				return
			}
			value, rest, ok := cutQuotedTagPart(rest[1:])
			if !ok {
				return
			}
			if !yield(name, value) {
				return
			}
			tags = rest
		}
	}
}

func cutTagName(tag string) (string, string, bool) {
	if tag[0] == '"' {
		return cutQuotedTagPart(tag)
	}
	end := 0
	for end < len(tag) && isTagNameByte(tag[end]) {
		end++
	}
	if end == 0 || end >= len(tag) || tag[end] != ':' {
		return "", "", false
	}
	return tag[:end], tag[end:], true
}

func isTagNameByte(b byte) bool {
	return b > ' ' && b != ':' && b != '"' && b != 0x7f
}

func cutQuotedTagPart(tag string) (string, string, bool) {
	if tag == "" || tag[0] != '"' {
		return "", "", false
	}
	quoted, err := strconv.QuotedPrefix(tag)
	if err != nil {
		return "", "", false
	}
	value, err := strconv.Unquote(quoted)
	if err != nil {
		return "", "", false
	}
	return value, tag[len(quoted):], true
}
