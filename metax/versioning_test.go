package metax

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func readTestFile(t *testing.T, fn string) []byte {
	path := filepath.Join("testdata", fn)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func TestCreatedNew(t *testing.T) {
	tests := []struct {
		fn    string
		isNew bool
		id    string
	}{
		{
			fn:    "unpublished.json",
			isNew: false,
			id:    "",
		},
		{
			fn:    "published.json",
			isNew: false,
			id:    "",
		},
		{
			fn:    "new_created.json",
			isNew: true,
			id:    "urn:nbn:fi:att:cee7033b-0199-4ac8-be8f-2092a7a650f2", // 7733
		},
		{
			fn:    "error-new_created_not_an_object.json",
			isNew: true,
			id:    "",
		},
		{
			fn:    "error-new_created_id_missing.json",
			isNew: true,
			id:    "",
		},
	}

	for _, test := range tests {
		dataset := readTestFile(t, test.fn)

		t.Run(test.fn+"(bool)", func(t *testing.T) {
			result := CreatedNewVersion(dataset)

			if result != test.isNew {
				t.Errorf("isNew: expected %t, got %t", test.isNew, result)
			}
		})

		t.Run(test.fn+"(id)", func(t *testing.T) {
			result := MaybeNewVersionId(dataset)

			if result != test.id {
				t.Errorf("isNew: expected %s, got %v", test.id, result)
			}
		})
	}
}

func TestGetIdentifier(t *testing.T) {
	tests := []struct {
		fn    string
		hasId bool
		id    string
	}{
		{
			fn:    "unpublished.json",
			hasId: false,
			id:    "",
		},
		{
			fn:    "published.json",
			hasId: true,
			id:    "urn:nbn:fi:att:bfe2d120-6ceb-4949-9755-882ab54c45b2",
		},
	}

	for _, test := range tests {
		dataset := readTestFile(t, test.fn)

		t.Run(test.fn+"(bool)", func(t *testing.T) {
			result := IsPublished(dataset)

			if result != test.hasId {
				t.Errorf("IsPublished: expected %v, got %v", test.hasId, result)
			}
		})

		t.Run(test.fn+"(id)", func(t *testing.T) {
			result := GetIdentifier(dataset)

			if (result == "") == test.hasId {
				t.Errorf("GetIdentifier: expected %v, got %v", test.hasId, result == "")
			}
			if result != test.id {
				t.Errorf("GetIdentifier: expected %q, got %q", test.id, result)
			}
		})
	}
}

func TestEditor(t *testing.T) {
	tests := []struct {
		fn            string
		hasEditor     bool
		hasIdentifier bool
		id            string
	}{
		{
			fn:            "editor-id_012345678901234567890123456789012.json",
			hasEditor:     true,
			hasIdentifier: true,
			id:            "12345678901234567890123456789012",
		},
		{
			fn:            "editor-no_qvain_identifier.json",
			hasEditor:     true,
			hasIdentifier: false,
			id:            "12345678901234567890123456789012",
		},
		{
			fn:            "editor-no_editor.json",
			hasEditor:     false,
			hasIdentifier: false,
			id:            "12345678901234567890123456789012",
		},
	}

	for _, test := range tests {
		dataset := readTestFile(t, test.fn)

		t.Run(test.fn, func(t *testing.T) {
			result := GetQvainId(dataset)

			if (result == "") == test.hasIdentifier {
				t.Errorf("GetIdentifier: expected %v, got %v", test.hasIdentifier, result != "")
			}
			if test.hasIdentifier && result != test.id {
				t.Errorf("GetIdentifier: expected %q, got %q", test.id, result)
			}
		})
	}
}

func TestGetIdentifiers(t *testing.T) {
	tests := []struct {
		fn  string
		vId string
		nId string
		qId string
	}{
		{
			fn:  "editor-id_012345678901234567890123456789012.json",
			vId: "",
			nId: "",
			qId: "12345678901234567890123456789012",
		},
		{
			fn:  "editor-no_qvain_identifier.json",
			vId: "",
			nId: "",
			qId: "",
		},
		{
			fn:  "editor-no_editor.json",
			vId: "",
			nId: "",
			qId: "",
		},
		{
			fn:  "new_created.json",
			vId: "urn:nbn:fi:att:bfe2d120-6ceb-4949-9755-882ab54c45b2",
			nId: "urn:nbn:fi:att:cee7033b-0199-4ac8-be8f-2092a7a650f2",
			qId: "",
		},
		{
			fn:  "published.json",
			vId: "urn:nbn:fi:att:bfe2d120-6ceb-4949-9755-882ab54c45b2",
			nId: "",
			qId: "05747d5fbede2b67e2f076b322171d30",
		},
		{
			fn:  "error-new_created_id_missing.json",
			vId: "urn:nbn:fi:att:bfe2d120-6ceb-4949-9755-882ab54c45b2",
			nId: "",
			qId: "",
		},
		{
			fn:  "error-new_created_not_an_object.json",
			vId: "urn:nbn:fi:att:bfe2d120-6ceb-4949-9755-882ab54c45b2",
			nId: "",
			qId: "",
		},
		{
			fn:  "editor-error-id_012345678901234567890123456789012_no_ident.json",
			vId: "",
			nId: "",
			qId: "",
		},
		{
			fn:  "editor-error-id_012345678901234567890123456789012_not_qvain.json",
			vId: "",
			nId: "",
			qId: "",
		},
	}

	for _, test := range tests {
		dataset := readTestFile(t, test.fn)

		t.Run(test.fn, func(t *testing.T) {
			vId, nId, qId := GetIdentifiers(dataset)

			if vId != test.vId || nId != test.nId || qId != test.qId {
				t.Errorf("GetIdentifiers: expected %q %q %q, got %v %v %v",
					test.vId, test.nId, test.qId,
					vId, nId, qId,
				)
			}
		})
	}
}
