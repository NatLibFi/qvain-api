package metax

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/CSCfi/qvain-api/pkg/models"
	"github.com/wvh/uuid"
)

const (
	// MetaxDatasetFamily is the dataset type for Fairdata datasets.
	MetaxDatasetFamily = 2

	// appIdent is the ident used to recognise our application's Editor metadata.
	appIdent = "qvain"

	// allowCreation either allows a remote service to create a new record or not.
	allowCreation = true
)

// nil slice for error use
var noRecords []MetaxRecord

func init() {
	models.RegisterFamily(MetaxDatasetFamily, "metax", NewMetaxDataset, LoadMetaxDataset, []string{"research_dataset", "contracts"})
}

// MetaxDataset wraps a models.Dataset.
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
func (dataset *MetaxDataset) CreateData(family int, schema string, blob []byte, extra map[string]string) error {
	if family == 0 {
		return errors.New("need schema family")
	}

	if _, ok := parsedTemplates[schema]; !ok {
		return errors.New("unknown schema")
	}

	template := parsedTemplates[schema]

	// don't set Creator and Owner since we don't update the json if they change
	editor := &Editor{
		Identifier: strptr(appIdent),
		RecordId:   strptr(dataset.Dataset.Id.String()),
	}

	editorJson, err := json.Marshal(editor)
	if err != nil {
		fmt.Println("can't serialise editor", err)
	}
	template["research_dataset"] = (*json.RawMessage)(&blob)
	template["editor"] = (*json.RawMessage)(&editorJson)

	//user, _ := json.Marshal(dataset.Dataset.Creator.String())
	//template["metadata_provider_user"] = (*json.RawMessage)(&user)
	if extra != nil {
		if extid, ok := extra["identity"]; ok && extid != "" {
			extidJson, _ := json.Marshal(extid)
			template["metadata_provider_user"] = (*json.RawMessage)(&extidJson)
		}
		if org, ok := extra["org"]; ok && org != "" {
			orgJson, _ := json.Marshal(org)
			template["metadata_provider_org"] = (*json.RawMessage)(&orgJson)
		}
	}

	newBlob, err := json.MarshalIndent(template, "", "\t")
	if err != nil {
		return err
	}

	dataset.Dataset.SetData(family, schema, newBlob)
	return nil
}

// UpdateData creates a partial dataset JSON blob to patch an existing one with.
func (dataset *MetaxDataset) UpdateData(family int, schema string, blob []byte, extra map[string]string) error {
	if family == 0 {
		return errors.New("need schema family")
	}

	if _, ok := parsedTemplates[schema]; !ok {
		return errors.New("unknown schema")
	}

	// don't set Creator and Owner since we don't update the json if they change
	editor := &Editor{
		Identifier: strptr(appIdent),
		RecordId:   strptr(dataset.Dataset.Id.String()),
	}

	var extid string
	if extra != nil {
		extid, _ = extra["identity"]
	}

	patchedFields := &struct {
		ResearchDataset *json.RawMessage `json:"research_dataset"`
		Editor          *Editor          `json:"editor"`
		Extid           string           `json:"metadata_provider_user,omitempty"`
	}{
		ResearchDataset: (*json.RawMessage)(&blob),
		Editor:          editor,
		Extid:           extid,
	}

	newBlob, err := json.MarshalIndent(patchedFields, "", "\t")
	if err != nil {
		return err
	}

	dataset.Dataset.SetData(family, schema, newBlob)

	return nil
}

// MetaxRecord is a helper struct to parse the fields we need from a Metax dataset.
type MetaxRecord struct {
	Id         int64  `json:"id"`
	Identifier string `json:"identifier"`

	// deprecated
	/*
		CreatedByUserId  *string `json:"created_by_user_id"`
		CreatedByApi     *string `json:"created_by_api"`
		ModifiedByUserId *string `json:"modified_by_user_id"`
		ModifiedByApi    *string `json:"modified_by_api"`
	*/

	DataCatalog *DataCatalog `json:"data_catalog"`

	MetadataProviderUser *string `json:"metadata_provider_user"`

	DateCreated  *time.Time `json:"date_created"`
	DateModified *time.Time `json:"date_modified"`

	Editor *Editor `json:"editor"`

	ResearchDataset json.RawMessage `json:"research_dataset"`
	Contract        json.RawMessage `json:"contract"`
}

// DataCatalog contains the catalog identifier
type DataCatalog struct {
	Identifier *string `json:"identifier"`
}

// Editor is the Go representation of the Editor object in a Metax record.
type Editor struct {
	Identifier *string `json:"identifier"`
	RecordId   *string `json:"record_id"`
	CreatorId  *string `json:"creator_id,omitempty"`
	OwnerId    *string `json:"owner_id,omitempty"`
	ExtId      *string `json:"fd_id,omitempty"`
}

// MetaxRawRecord embeds a json.RawMessage containing an unparsed JSON []byte slice with the Metax dataset.
type MetaxRawRecord struct {
	json.RawMessage
}

// Record unmarshals the raw JSON and validates it, returning either a partially parsed MetaxRecord or an error.
//
// -wvh- NOTE: (2019-03-28) Validation disabled to allow creating new datasets.
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

// Validate checks if the Metax record contains the fields we need to identify the record (those below the `editor` key).
//
// -wvh- NOTE: (2019-03-28) Deprecated to allow creating new datasets if there is no existing application metadata in the dataset.
func (record *MetaxRecord) Validate() error {
	if record.Editor == nil {
		return NewLinkingError()
	}

	if record.Editor.Identifier == nil {
		return NewLinkingError("identifier")
	}

	if record.Editor.RecordId == nil {
		return NewLinkingError("record_id")
	}

	if record.Editor.CreatorId == nil {
		return NewLinkingError("creator_id")
	}

	if record.Editor.OwnerId == nil {
		return NewLinkingError("owner_id")
	}

	return nil
}

// IsNewDataset checks if the Metax record should be treated as a new dataset. A dataset is new if it doesn't have a Qvain id.
func (raw MetaxRawRecord) GetQvainId(mrec *MetaxRecord) (*uuid.UUID, error) {
	var qid uuid.UUID
	var err error

	// no editor, no editor.identifier, not ours, or no record_id: return nil;
	// in theory, we should return errors for different combinations of missing fields, but nobody will be there to do anything about it
	if mrec.Editor == nil || mrec.Editor.Identifier == nil || *mrec.Editor.Identifier != appIdent || mrec.Editor.RecordId == nil {
		return nil, nil
	}

	// have an id, try to parse
	if qid, err = uuid.FromString(*mrec.Editor.RecordId); err != nil {
		return nil, ErrInvalidId
	}

	return &qid, nil
}

// ToQvain converts a Metax record in raw JSON to a Qvain record using the values in the Editor object.
//
// If the Editor metadata contains valid data, consider the dataset ours and populate (all) the Dataset struct fields; boolean New is false.
// If the Editor metadata does not exist, consider the dataset new and let the caller handle creation; the ids and ownership fields are not set.
// If the Editor metadata is invalid, return an error.
func (raw MetaxRawRecord) ToQvain() (*models.Dataset, bool, error) {
	var mrec MetaxRecord
	var err error
	var isNew bool

	err = json.Unmarshal(raw.RawMessage, &mrec)
	if err != nil {
		return nil, isNew, err
	}

	qid, err := raw.GetQvainId(&mrec)
	if err != nil {
		return nil, isNew, err
	}

	// no id; new dataset
	if qid == nil {
		// not allowed to create
		if !allowCreation {
			return nil, isNew, ErrIdRequired
		}

		// can't be nil pointer; set to null array
		qid = new(uuid.UUID)
		isNew = true
	}

	qdataset := new(models.Dataset)

	if isNew {
		qdataset.Created = timeOrNow(mrec.DateCreated)
		qdataset.Modified = timeOrNow(mrec.DateModified)
	} else {
		qdataset.Id = *qid
	}
	schema, ok := CatalogIdentifiers[*mrec.DataCatalog.Identifier]
	if !ok {
		return nil, isNew, fmt.Errorf("Metax dataset schema unknown or missing: %s", *mrec.DataCatalog.Identifier)
	}
	qdataset.SetData(MetaxDatasetFamily, schema, raw.RawMessage)

	return qdataset, isNew, nil
}

// timeOrNow returns the current time if the given time is not set.
func timeOrNow(t *time.Time) time.Time {
	if t == nil {
		return time.Now()
	}
	return *t
}

func strptr(s string) *string {
	return &s
}

func timeptr(t time.Time) *time.Time {
	return &t
}
