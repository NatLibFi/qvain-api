package main

import (
	"bytes"
	stdlib "encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/buger/jsonparser"
	"github.com/francoispqt/gojay"
	"github.com/tidwall/gjson"
	"github.com/valyala/fastjson"
)

/*
  BenchmarkGetIdentifier/jsonparser-4             10000000               179 ns/op              64 B/op          1 allocs/op
  BenchmarkGetIdentifier/gjson-4                  10000000               180 ns/op              64 B/op          1 allocs/op
  BenchmarkGetIdentifier/fastjson-4                 300000              4685 ns/op              64 B/op          1 allocs/op
  BenchmarkGetIdentifier/fastjsonReuse-4            300000              4673 ns/op              64 B/op          1 allocs/op
  BenchmarkGetIdentifier/stdlib-4                    50000             35246 ns/op             360 B/op          7 allocs/op
  BenchmarkGetIdentifier/gojay-decoder-4            200000              6809 ns/op            7761 B/op          7 allocs/op
  BenchmarkGetIdentifier/gojay-unmarshal-4         1000000              1147 ns/op            4112 B/op          2 allocs/op
  BenchmarkGetIdentifiers/jsonparser-4              200000              6272 ns/op             336 B/op          8 allocs/op
  BenchmarkGetIdentifiers/gjson-4                   200000              7922 ns/op             352 B/op          4 allocs/op
  BenchmarkGetIdentifiers/fastjson-4                100000             22257 ns/op           34736 B/op        102 allocs/op
  BenchmarkGetIdentifiers/fastjsonReuse-4           300000              4865 ns/op             160 B/op          3 allocs/op
  BenchmarkGetIdentifiers/stdlib-4                   50000             36650 ns/op             488 B/op          9 allocs/op
  BenchmarkGetIdentifiers/gojay-decoder-4           200000              7598 ns/op            7841 B/op         11 allocs/op
  BenchmarkGetIdentifiers/gojay-unmarshal-4         200000              5447 ns/op            4144 B/op          2 allocs/op
*/

const testDatasetFile = "new_created.json"

// fastjsonParser re-uses an existing parser instead of creating a new instance every time.
var fastjsonParser fastjson.Parser

func readFile(t *testing.T, fn string) []byte {
	path := filepath.Join("..", "..", "pkg", "metax", "testdata", fn)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func getIdentifierJsonParser(json []byte) string {
	res, err := jsonparser.GetString(json, "identifier")
	if err != nil {
		panic(err)
	}
	return res
}

func getIdentifierGjson(json []byte) string {
	return gjson.GetBytes(json, "identifier").String()
}

func getIdentifierFastjson(json []byte) string {
	return fastjson.GetString(json, "identifier")
}

// getIdentifierFastjsonReuse: for one value, this is really the same code (and running time) as the shortcut GetString function above
func getIdentifierFastjsonReuse(json []byte) string {
	v, err := fastjsonParser.ParseBytes(json)
	if err != nil {
		// return ""
		panic(err)
	}
	return string(v.GetStringBytes("identifier"))
}

// getIdentifierGojayDecoder: this one uses an inline function instead of a pre-defined method
func getIdentifierGojayDecoder(json []byte) string {
	var id string

	dec := gojay.BorrowDecoder(bytes.NewReader(json))
	defer dec.Release()

	err := dec.DecodeObject(gojay.DecodeObjectFunc(func(dec *gojay.Decoder, key string) error {
		switch key {
		case "identifier":
			return dec.String(&id)
		}
		return nil
	}))
	if err != nil {
		panic(err)
	}

	return id
}

type gojayDataset struct {
	Identifier string
}

func (dataset *gojayDataset) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "identifier":
		return dec.String(&dataset.Identifier)
	}
	return nil
}

func (dataset *gojayDataset) NKeys() int {
	return 1
}

// getIdentifierGojayUnmarshal: here we use a pre-defined struct with unmarshal method
func getIdentifierGojayUnmarshal(json []byte) string {
	dataset := &gojayDataset{}
	err := gojay.UnmarshalJSONObject(json, dataset)
	if err != nil {
		panic(err)
	}
	return dataset.Identifier
}

func getIdentifierStdlib(json []byte) string {
	var dataset = struct {
		Identifier string `json:"Identifier"`
	}{}
	err := stdlib.Unmarshal(json, &dataset)
	if err != nil {
		panic(err)
	}
	return dataset.Identifier
}

func getIdentifiersJsonParser(json []byte) (string, string, string) {
	paths := [][]string{
		{"identifier"},
		{"next_dataset_version", "identifier"},
		{"editor", "owner_id"},
	}
	var s1, s2, s3 string

	jsonparser.EachKey(json, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			s1 = string(value)
		case 1:
			s2 = string(value)
		case 2:
			s3 = string(value)
		}
	}, paths...)

	return s1, s2, s3
}

func getIdentifiersGjson(json []byte) (string, string, string) {
	results := gjson.GetManyBytes(json, "identifier", "next_dataset_version.identifier", "editor.owner_id")
	return results[0].String(), results[1].String(), results[2].String()
}

func getIdentifiersFastjson(json []byte) (string, string, string) {
	var parser fastjson.Parser
	v, err := parser.ParseBytes(json)
	if err != nil {
		panic(err)
	}
	return string(v.GetStringBytes("identifier")), string(v.GetStringBytes("next_dataset_version", "identifier")), string(v.GetStringBytes("editor", "owner_id"))
}

func getIdentifiersFastjsonReuse(json []byte) (string, string, string) {
	v, err := fastjsonParser.ParseBytes(json)
	if err != nil {
		panic(err)
	}
	return string(v.GetStringBytes("identifier")), string(v.GetStringBytes("next_dataset_version", "identifier")), string(v.GetStringBytes("editor", "owner_id"))
}

func getIdentifiersStdlib(json []byte) (string, string, string) {
	var dataset = struct {
		Identifier         string `json:"identifier"`
		NextDatasetVersion struct {
			Identifier string `json:"identifier"`
		} `json:"next_dataset_version"`
		Editor struct {
			OwnerId string `json:"owner_id"`
		} `json:"editor"`
	}{}
	err := stdlib.Unmarshal(json, &dataset)
	if err != nil {
		panic(err)
	}
	return dataset.Identifier, dataset.NextDatasetVersion.Identifier, dataset.Editor.OwnerId
}

// getIdentifierGojayDecoder: this one uses an inline function instead of a pre-defined method
func getIdentifiersGojayDecoder(json []byte) (string, string, string) {
	var id1, id2, id3 string

	dec := gojay.BorrowDecoder(bytes.NewReader(json))
	defer dec.Release()

	err := dec.DecodeObject(gojay.DecodeObjectFunc(func(dec *gojay.Decoder, key string) error {
		switch key {
		case "identifier":
			return dec.String(&id1)
		case "next_dataset_version":
			return dec.Object(gojay.DecodeObjectFunc(func(dec *gojay.Decoder, key string) error {
				switch key {
				case "identifier":
					return dec.String(&id2)
				}
				return nil
			}))
		case "editor":
			return dec.Object(gojay.DecodeObjectFunc(func(dec *gojay.Decoder, key string) error {
				switch key {
				case "owner_id":
					return dec.String(&id3)
				}
				return nil
			}))
		}
		return nil
	}))
	if err != nil {
		panic(err)
	}

	return id1, id2, id3
}

type gojayDatasetMultiple struct {
	Identifier         string
	NextDatasetVersion gojayNextDatasetVersion
	Editor             gojayEditor
}

type gojayNextDatasetVersion struct {
	Identifier string
}

func (ndv *gojayNextDatasetVersion) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "identifier":
		return dec.String(&ndv.Identifier)
	}
	return nil
}

func (ndv *gojayNextDatasetVersion) NKeys() int {
	return 1
}

type gojayEditor struct {
	OwnerId string
}

func (editor *gojayEditor) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "owner_id":
		return dec.String(&editor.OwnerId)
	}
	return nil
}

func (dataset *gojayEditor) NKeys() int {
	return 1
}

func (dataset *gojayDatasetMultiple) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "identifier":
		return dec.String(&dataset.Identifier)
	case "next_dataset_version":
		return dec.Object(&dataset.NextDatasetVersion)
	case "editor":
		return dec.Object(&dataset.Editor)
	}
	return nil
}

func (dataset *gojayDatasetMultiple) NKeys() int {
	return 3
}

func getIdentifiersGojayUnmarshal(json []byte) (string, string, string) {
	dataset := &gojayDatasetMultiple{}
	err := gojay.UnmarshalJSONObject(json, dataset)
	if err != nil {
		panic(err)
	}
	return dataset.Identifier, dataset.NextDatasetVersion.Identifier, dataset.Editor.OwnerId
}

func TestGetIdentifier(t *testing.T) {
	expectedId := "urn:nbn:fi:att:bfe2d120-6ceb-4949-9755-882ab54c45b2"
	tests := []struct {
		name string
		fn   func([]byte) string
	}{
		{name: "jsonparser", fn: getIdentifierJsonParser},
		{name: "gjson", fn: getIdentifierGjson},
		{name: "fastjson", fn: getIdentifierFastjson},
		{name: "stdlib", fn: getIdentifierStdlib},
		{name: "gojay-decoder", fn: getIdentifierGojayDecoder},
		{name: "gojay-unmarshal", fn: getIdentifierGojayUnmarshal},
	}
	dataset := readFile(t, testDatasetFile)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.fn(dataset)

			if result != expectedId {
				t.Errorf("expected %q, got %q", expectedId, result)
			}
		})
	}
}

func TestGetIdentifiers(t *testing.T) {
	expectedIds := []string{
		"urn:nbn:fi:att:bfe2d120-6ceb-4949-9755-882ab54c45b2",
		"urn:nbn:fi:att:cee7033b-0199-4ac8-be8f-2092a7a650f2",
		"053bffbcc41edad4853bea91fc42ea18",
	}
	tests := []struct {
		name string
		fn   func([]byte) (string, string, string)
	}{
		{name: "jsonparser", fn: getIdentifiersJsonParser},
		{name: "gjson", fn: getIdentifiersGjson},
		{name: "fastjson", fn: getIdentifiersFastjson},
		{name: "stdlib", fn: getIdentifiersStdlib},
		{name: "gojay-decoder", fn: getIdentifiersGojayDecoder},
		{name: "gojay-unmarshal", fn: getIdentifiersGojayUnmarshal},
	}
	dataset := readFile(t, testDatasetFile)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results := make([]string, 3)
			results[0], results[1], results[2] = test.fn(dataset)

			if !reflect.DeepEqual(results, expectedIds) {
				t.Errorf("expected %q, got %q", expectedIds, results)
			}
		})
	}
}

func BenchmarkGetIdentifier(b *testing.B) {
	tests := []struct {
		name string
		fn   func([]byte) string
	}{
		{name: "jsonparser", fn: getIdentifierJsonParser},
		{name: "gjson", fn: getIdentifierGjson},
		{name: "fastjson", fn: getIdentifierFastjson},
		{name: "fastjsonReuse", fn: getIdentifierFastjsonReuse},
		{name: "stdlib", fn: getIdentifierStdlib},
		{name: "gojay-decoder", fn: getIdentifierGojayDecoder},
		{name: "gojay-unmarshal", fn: getIdentifierGojayUnmarshal},
	}
	dataset := readFile(nil, testDatasetFile)

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = test.fn(dataset)
			}
		})
	}
}

func BenchmarkGetIdentifiers(b *testing.B) {
	tests := []struct {
		name string
		fn   func([]byte) (string, string, string)
	}{
		{name: "jsonparser", fn: getIdentifiersJsonParser},
		{name: "gjson", fn: getIdentifiersGjson},
		{name: "fastjson", fn: getIdentifiersFastjson},
		{name: "fastjsonReuse", fn: getIdentifiersFastjsonReuse},
		{name: "stdlib", fn: getIdentifiersStdlib},
		{name: "gojay-decoder", fn: getIdentifiersGojayDecoder},
		{name: "gojay-unmarshal", fn: getIdentifiersGojayUnmarshal},
	}
	dataset := readFile(nil, testDatasetFile)

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _, _ = test.fn(dataset)
			}
		})
	}
}
