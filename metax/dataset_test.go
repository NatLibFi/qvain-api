package metax

import (
	"encoding/json"
	"testing"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	// testBaseDataset is a (published) Metax dataset with identifier, owner and date fields set.
	testBaseDataset = "published.json"

	// testPathToId is the dotted path to the application id in the test dataset.
	testPathToId = "editor.record_id"

	// testPathToAppIdent is the dotted path to the application ident(ifier) in the test dataset.
	testPathToAppIdent = "editor.identifier"

	// testPathToUser is the dotted path to the user identity in the test dataset.
	testPathToUser = "metadata_provider_user"

	// testNoIdString is the string representation of an unset uuid.
	testNoIdString = "00000000000000000000000000000000"
)

func TestMetaxDatasetParsing(t *testing.T) {
	// note: readTestFile is a helper function defined in another test file
	datasetBytes := readTestFile(t, testBaseDataset)

	qid := gjson.GetBytes(datasetBytes, testPathToId).Str
	if qid == "" {
		t.Fatalf("can't find a valid Qvain id from test dataset; expected at path: %q", testPathToId)
	}

	if app := gjson.GetBytes(datasetBytes, testPathToAppIdent).Str; app != appIdent {
		t.Fatalf("wrong application identifier in test dataset; expected: %q, got: %q", appIdent, app)
	}

	identity := gjson.GetBytes(datasetBytes, testPathToUser).Str
	if identity == "" {
		t.Fatal("can't find a valid Qvain id from test dataset")
	}

	tests := []struct {
		// name of test and function to modify the dataset before test
		name   string
		before func([]byte) []byte

		// result fields to check against
		isNew bool
		id    string
		err   error
	}{
		{
			name:   "existing dataset",
			before: nil,
			isNew:  false,
			id:     qid,
			err:    nil,
		},
		{
			name:   "new dataset without editor",
			before: func(data []byte) []byte { data, _ = sjson.DeleteBytes(data, "editor"); return data },
			isNew:  true,
			id:     testNoIdString,
			err:    nil,
		},
		{
			name:   "existing dataset but missing id",
			before: func(data []byte) []byte { data, _ = sjson.DeleteBytes(data, testPathToId); return data },
			isNew:  true,
			id:     testNoIdString,
			err:    nil,
		},
		{
			name:   "existing dataset but invalid id (not uuid)",
			before: func(data []byte) []byte { data, _ = sjson.SetBytes(data, testPathToId, "invalid-id-666"); return data },
			isNew:  false,
			id:     testNoIdString,
			err:    ErrInvalidId,
		},
		{
			name:   "new dataset because of wrong application identifier",
			before: func(data []byte) []byte { data, _ = sjson.SetBytes(data, testPathToAppIdent, "not_qvain"); return data },
			isNew:  true,
			id:     testNoIdString,
			err:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := datasetBytes
			if test.before != nil {
				// no need to copy, sjson will allocate a new slice for us
				data = test.before(data)
			}

			unparsed := MetaxRawRecord{json.RawMessage(data)}

			// parse
			dataset, isNew, err := unparsed.ToQvain()
			if err != nil {
				// make sure these error condition tests do not continue but return
				if test.err == nil {
					t.Fatalf("expected no error, got: %v", err)
				} else if err != test.err {
					t.Fatalf("unexpected error, expected: %q, got: %q", test.err, err)
				} else {
					// expected error matched, pass
					return
				}
			}

			// check `new` detection
			if isNew != test.isNew {
				t.Errorf("isNew: expected %t, got %t", test.isNew, isNew)
			}

			// check id
			if dataset.Id.String() != test.id {
				t.Errorf("id doesn't match: expected %v, got %v", test.id, dataset.Id)
			}

			// only check creation time for new datasets; we don't care about existing ones
			if test.isNew && dataset.Created.IsZero() {
				t.Error("dataset.Created should be set to upstream creation time but is zero")
			}
		})
	}
}
