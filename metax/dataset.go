package metax

import (
	"encoding/json"
	"fmt"
	"time"
	"errors"

	"github.com/NatLibFi/qvain-api/models"
	"github.com/wvh/uuid"
)

const MetaxDatasetFamily = 2

// nil slice for error use
var noRecords []MetaxRecord

func init() {
	models.RegisterFamily(2, "metax", NewMetaxDataset, LoadMetaxDataset, []string{"research_metadata", "contracts"})
}


type MetaxDataset struct {
	*models.Dataset
}

// NewMetaxDataset creates a Metax dataset.
func NewMetaxDataset(creator uuid.UUID) (models.TypedDataset, error) {
	ds, err := models.NewDataset(creator)
	if err != nil {
		return nil, err
	}

	return &MetaxDataset{ds}, nil
}

// LoadMetaxDataset constructs an existing MetaxDataset from an existing base dataset.
func LoadMetaxDataset(ds *models.Dataset) models.TypedDataset {
	return &MetaxDataset{Dataset: ds}
}

// CreateData creates a dataset from template and merges set fields.
func (dataset *MetaxDataset) CreateData(family int, schema string, blob []byte) error {
	if family == 0 {
		return errors.New("need schema family")
	}

	if _, ok := parsedTemplates[schema]; !ok {
		return errors.New("unknown schema")
	}

	template := parsedTemplates[schema]

	// don't set Creator and Owner since we don't update the json if they change
	editor := &Editor{
		Identifier: static("qvain"),
		RecordId:   static(dataset.Dataset.Id.String()),
	}

	editorJson, err := json.Marshal(editor)
	if err != nil {
		fmt.Println("can't serialise editor", err)
	}
	template["research_dataset"] = (*json.RawMessage)(&blob)
	template["editor"] = (*json.RawMessage)(&editorJson)

	user, _ := json.Marshal(dataset.Dataset.Creator.String())
	template["metadata_provider_user"] = (*json.RawMessage)(&user)

	newBlob, err := json.MarshalIndent(template, "", "\t")
	if err != nil {
		return err
	}

	dataset.Dataset.SetData(family, schema, newBlob)
	return nil
}

// UpdateData creates a partial dataset JSON blob to patch an existing one with.
func (dataset *MetaxDataset) UpdateData(family int, schema string, blob []byte) error {
	if family == 0 {
		return errors.New("need schema family")
	}

	if _, ok := parsedTemplates[schema]; !ok {
		return errors.New("unknown schema")
	}

	// don't set Creator and Owner since we don't update the json if they change
	editor := &Editor{
		Identifier: static("qvain"),
		RecordId:   static(dataset.Dataset.Id.String()),
	}

	patchedFields := &struct {
		ResearchDataset *json.RawMessage `json:"research_dataset"`
		Editor          *Editor          `json:"editor"`
	}{
		ResearchDataset: (*json.RawMessage)(&blob),
		Editor:          editor,
	}

	newBlob, err := json.MarshalIndent(patchedFields, "", "\t")
	if err != nil {
		return err
	}

	dataset.Dataset.SetData(family, schema, newBlob)

	return nil
}



type MetaxRecord struct {
	Id int64 `json:"id"`

	CreatedByUserId  *string `json:"created_by_user_id"`
	CreatedByApi     *string `json:"created_by_api"`
	ModifiedByUserId *string `json:"modified_by_user_id"`
	ModifiedByApi    *string `json:"modified_by_api"`

	//Editor           *string `json:"editor"`
	Editor *Editor `json:"editor"`

	ResearchDataset json.RawMessage `json:"research_dataset"`
	Contract        json.RawMessage `json:"contract"`
}


// Editor is the Go representation of the Editor object in a Metax record.
type Editor struct {
	Identifier *string `json:"identifier"`
	RecordId   *string `json:"record_id"`
	CreatorId  *string `json:"creator_id,omitempty"`
	OwnerId    *string `json:"owner_id,omitempty"`
	ExtId      *string `json:"fd_id,omitempty"`
}


// MetaxRawRecord is an alias for json.RawMessage that contains an unparsed JSON []byte slice.
type MetaxRawRecord struct {
	json.RawMessage
}

// Record unmarshals the raw JSON and validates it, returning either a partially parsed MetaxRecord or an error.
func (raw MetaxRawRecord) Record() (*MetaxRecord, error) {
	var record MetaxRecord
	err := json.Unmarshal(raw.RawMessage, &record)
	if err != nil {
		return nil, err
	}

	if err := record.Validate(); err != nil {
		return nil, err
	}

	return &record, nil
}

// Validate checks if the Metax record contains the fields we need to identify the record.
func (record *MetaxRecord) Validate() error {
	if record.Editor == nil {
		return NewLinkingError()
	}

	if record.Editor.Identifier == nil {
		return NewLinkingError("identifier")
	}

	/*
	if record.Editor.RecordId == nil {
		return NewLinkingError("record_id")
	}
	*/

	if record.Editor.CreatorId == nil {
		return NewLinkingError("creator_id")
	}

	if record.Editor.OwnerId == nil {
		return NewLinkingError("owner_id")
	}

	return nil
}

// ToQvain converts a Metax record in raw JSON to a Qvain record using the values in the Editor object.
func (raw MetaxRawRecord) ToQvain() (*models.Dataset, error) {
	var mrec MetaxRecord
	var err error

	err = json.Unmarshal(raw.RawMessage, &mrec)
	if err != nil {
		return nil, err
	}

	if err = mrec.Validate(); err != nil {
		return nil, err
	}

	var id, creator, owner uuid.UUID
	/*
	if id, err = uuid.FromString(*mrec.Editor.Identifier); err != nil {
		fmt.Printf("%#v // %v\n", *mrec.Editor.Identifier, err)
		return nil, NewLinkingError("identifier-uuid")
	}
	*/
	if id, err = uuid.FromString(fmt.Sprintf("056bffbcc41edad4853bea91%08d", mrec.Id)); err != nil {
		fmt.Printf("%8d // %v\n", mrec.Id, err)
		return nil, NewLinkingError("id-uuid")
	}

	if creator, err = uuid.FromString(*mrec.Editor.CreatorId); err != nil {
		return nil, NewLinkingError("creator_id")
	}

	if owner, err = uuid.FromString(*mrec.Editor.OwnerId); err != nil {
		return nil, NewLinkingError("owner_id")
	}

	now := time.Now()

	qrec := models.Dataset{
		Id: id,
		Creator: creator,
		Owner: owner,

		Created: now,
		Modified: now,
	}
	qrec.SetData(2, "metax", raw.RawMessage)
	return &qrec, nil
}

func static(s string) *string {
	return &s
}
