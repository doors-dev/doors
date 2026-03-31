package action

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func gunzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestActionLogsAndInvocations(t *testing.T) {
	textPayload := NewText("hello")
	jsonPayload := NewJSON([]byte(`{"ok":true}`))
	cases := []struct {
		name            string
		action          Action
		log             string
		invocationName  string
		args            []any
		expectedPayload Payload
	}{
		{
			name:            "location reload",
			action:          LocationReload{},
			log:             "location_reload",
			invocationName:  "location_reload",
			args:            []any{},
			expectedPayload: NewNone(),
		},
		{
			name:            "location replace",
			action:          LocationReplace{URL: "/next", Origin: true},
			log:             "location_replace",
			invocationName:  "location_replace",
			args:            []any{"/next", true},
			expectedPayload: NewNone(),
		},
		{
			name:            "scroll",
			action:          Scroll{Selector: "#target", Options: map[string]any{"behavior": "smooth", "block": "center"}},
			log:             "scroll",
			invocationName:  "scroll",
			args:            []any{"#target", map[string]any{"behavior": "smooth", "block": "center"}},
			expectedPayload: NewNone(),
		},
		{
			name:            "location assign",
			action:          LocationAssign{URL: "/assign", Origin: false},
			log:             "location_assign",
			invocationName:  "location_assign",
			args:            []any{"/assign", false},
			expectedPayload: NewNone(),
		},
		{
			name:            "emit",
			action:          Emit{Name: "custom", DoorID: 9, Payload: textPayload},
			log:             "emit: custom",
			invocationName:  "emit",
			args:            []any{"custom", uint64(9)},
			expectedPayload: textPayload,
		},
		{
			name:            "dyna set",
			action:          DynaSet{ID: 7, Value: "value"},
			log:             "dyna_set",
			invocationName:  "dyna_set",
			args:            []any{uint64(7), "value"},
			expectedPayload: NewNone(),
		},
		{
			name:            "dyna remove",
			action:          DynaRemove{ID: 8},
			log:             "dyna_remove",
			invocationName:  "dyna_remove",
			args:            []any{uint64(8)},
			expectedPayload: NewNone(),
		},
		{
			name:            "set path",
			action:          SetPath{Path: "/path", Replace: true},
			log:             "set_path",
			invocationName:  "set_path",
			args:            []any{"/path", true},
			expectedPayload: NewNone(),
		},
		{
			name:            "door replace",
			action:          DoorReplace{ID: 10, Payload: jsonPayload},
			log:             "door_replace",
			invocationName:  "door_replace",
			args:            []any{uint64(10)},
			expectedPayload: jsonPayload,
		},
		{
			name:            "door update",
			action:          DoorUpdate{ID: 11, Payload: textPayload},
			log:             "door_update",
			invocationName:  "door_update",
			args:            []any{uint64(11)},
			expectedPayload: textPayload,
		},
		{
			name:            "indicate",
			action:          Indicate{Duration: 1500 * time.Millisecond, Indicate: map[string]any{"selector": "#id"}},
			log:             "indicate",
			invocationName:  "indicate",
			args:            []any{int64(1500), map[string]any{"selector": "#id"}},
			expectedPayload: NewNone(),
		},
		{
			name:            "report hook",
			action:          ReportHook{HookId: 12},
			log:             "report hook",
			invocationName:  "report_hook",
			args:            []any{uint64(12)},
			expectedPayload: NewNone(),
		},
		{
			name:            "update title",
			action:          UpdateTitle{Content: "Home", Attrs: map[string]string{"lang": "en"}},
			log:             "update_title",
			invocationName:  "update_title",
			args:            []any{"Home", map[string]string{"lang": "en"}},
			expectedPayload: NewNone(),
		},
		{
			name:            "update meta",
			action:          UpdateMeta{Name: "og:title", Property: true, Attrs: map[string]string{"content": "Doors"}},
			log:             "update_meta",
			invocationName:  "update_meta",
			args:            []any{"og:title", true, map[string]string{"content": "Doors"}},
			expectedPayload: NewNone(),
		},
		{
			name:            "test",
			action:          Test{Arg: []string{"a", "b"}},
			log:             "test",
			invocationName:  "test",
			args:            []any{[]string{"a", "b"}},
			expectedPayload: NewNone(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.action.Log(); got != tc.log {
				t.Fatalf("unexpected log: %q", got)
			}
			inv := tc.action.Invocation()
			if got := inv.Func(); !reflect.DeepEqual(got, []any{tc.invocationName, tc.args}) {
				t.Fatalf("unexpected invocation func: %#v", got)
			}
			if !reflect.DeepEqual(inv.Payload(), tc.expectedPayload) {
				t.Fatalf("unexpected payload: %#v", inv.Payload())
			}
		})
	}
}

func TestInvocationAndActionsJSON(t *testing.T) {
	withPayload := Emit{Name: "ready", DoorID: 3, Payload: NewText("ok")}.Invocation()
	payloadJSON, err := withPayload.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var payloadEncoded []any
	if err := json.Unmarshal(payloadJSON, &payloadEncoded); err != nil {
		t.Fatal(err)
	}
	if len(payloadEncoded) != 3 {
		t.Fatalf("expected invocation with payload to encode 3 fields, got %d", len(payloadEncoded))
	}

	withoutPayload := LocationReload{}.Invocation()
	plainJSON, err := withoutPayload.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var plainEncoded []any
	if err := json.Unmarshal(plainJSON, &plainEncoded); err != nil {
		t.Fatal(err)
	}
	if len(plainEncoded) != 2 {
		t.Fatalf("expected invocation without payload to encode 2 fields, got %d", len(plainEncoded))
	}

	actions := Actions{
		LocationReload{},
		Scroll{Selector: "#target", Options: map[string]any{"behavior": "smooth"}},
	}
	encoded, err := actions.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var decoded [][]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 2 {
		t.Fatalf("unexpected actions length: %d", len(decoded))
	}

	headers := http.Header{}
	if err := actions.Set(headers); err != nil {
		t.Fatal(err)
	}
	if headers.Get("D0-After") == "" {
		t.Fatal("expected D0-After header to be set")
	}
}

func TestPayloadConstructorsAndOutput(t *testing.T) {
	textGZ := []byte{0x1f, 0x8b, 0x08}
	jsonBytes := []byte(`{"a":1}`)
	jsonGZ := []byte{0x1f, 0x8b, 0x08, 0x00}
	binaryBytes := []byte{0x01, 0x02, 0x03}
	binaryGZ := []byte{0x1f, 0x8b, 0x08, 0x01}

	cases := []struct {
		name   string
		p      Payload
		typ    PayloadType
		length int
		out    []byte
		isNone bool
	}{
		{name: "none", p: NewNone(), typ: PayloadNone, length: 0, out: nil, isNone: true},
		{name: "empty", p: NewEmptyPayload(), typ: PayloadNone, length: 0, out: nil, isNone: true},
		{name: "text bytes", p: NewTextBytes([]byte("abc")), typ: PayloadText, length: 3, out: []byte("abc")},
		{name: "text", p: NewText("hello"), typ: PayloadText, length: 5, out: []byte("hello")},
		{name: "text gz", p: NewTextGZ(textGZ), typ: PayloadTextGZ, length: len(textGZ), out: textGZ},
		{name: "json", p: NewJSON(jsonBytes), typ: PayloadJSON, length: len(jsonBytes), out: jsonBytes},
		{name: "json gz", p: NewJSONGZ(jsonGZ), typ: PayloadJSONGZ, length: len(jsonGZ), out: jsonGZ},
		{name: "binary", p: NewBinary(binaryBytes), typ: PayloadBinary, length: len(binaryBytes), out: binaryBytes},
		{name: "binary gz", p: NewBinaryGZ(binaryGZ), typ: PayloadBinaryGZ, length: len(binaryGZ), out: binaryGZ},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.p.IsNone(); got != tc.isNone {
				t.Fatalf("unexpected none state: %v", got)
			}
			if got := tc.p.Type(); got != tc.typ {
				t.Fatalf("unexpected type: %v", got)
			}
			if got := tc.p.Len(); got != tc.length {
				t.Fatalf("unexpected length: %d", got)
			}
			var buf bytes.Buffer
			if err := tc.p.Output(&buf); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buf.Bytes(), tc.out) {
				t.Fatalf("unexpected payload output: %v", buf.Bytes())
			}
			if _, err := tc.p.MarshalJSON(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPayloadMarshalHelpersAndIntoPayload(t *testing.T) {
	textBytesJSON, err := TextBytes([]byte("hello")).MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(textBytesJSON) != `"hello"` {
		t.Fatalf("unexpected text bytes json: %s", textBytesJSON)
	}

	jsonValueJSON, err := JSON([]byte(`{"ok":true}`)).MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(jsonValueJSON) != `{"ok":true}` {
		t.Fatalf("unexpected json payload: %s", jsonValueJSON)
	}

	var raw []byte
	binaryPayload, err := IntoPayload(raw, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := binaryPayload.Type(); got != PayloadBinary {
		t.Fatalf("unexpected binary type: %v", got)
	}
	if got := binaryPayload.Len(); got != 0 {
		t.Fatalf("unexpected binary length: %d", got)
	}

	textPayload, err := IntoPayload("hello", false)
	if err != nil {
		t.Fatal(err)
	}
	if got := textPayload.Type(); got != PayloadText {
		t.Fatalf("unexpected text type: %v", got)
	}

	textGZPayload, err := IntoPayload("hello", true)
	if err != nil {
		t.Fatal(err)
	}
	if got := textGZPayload.Type(); got != PayloadTextGZ {
		t.Fatalf("unexpected text gz type: %v", got)
	}
	var textGZBuf bytes.Buffer
	if err := textGZPayload.Output(&textGZBuf); err != nil {
		t.Fatal(err)
	}
	if got := string(gunzipBytes(t, textGZBuf.Bytes())); got != "hello" {
		t.Fatalf("unexpected unzipped text: %q", got)
	}

	jsonPayload, err := IntoPayload(map[string]any{"ok": true}, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := jsonPayload.Type(); got != PayloadJSON {
		t.Fatalf("unexpected json type: %v", got)
	}
	var jsonBuf bytes.Buffer
	if err := jsonPayload.Output(&jsonBuf); err != nil {
		t.Fatal(err)
	}
	var decoded map[string]bool
	if err := json.Unmarshal(jsonBuf.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded["ok"] {
		t.Fatal("expected json payload to round-trip")
	}

	jsonGZPayload, err := IntoPayload(map[string]any{"ok": true}, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := jsonGZPayload.Type(); got != PayloadJSONGZ {
		t.Fatalf("unexpected json gz type: %v", got)
	}
	var jsonGZBuf bytes.Buffer
	if err := jsonGZPayload.Output(&jsonGZBuf); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(gunzipBytes(t, jsonGZBuf.Bytes()), &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded["ok"] {
		t.Fatal("expected gz json payload to round-trip")
	}

	if _, err := IntoPayload(make(chan int), false); err == nil {
		t.Fatal("expected unsupported payload type error")
	}
}
