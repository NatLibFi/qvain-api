package metax

import (
	"encoding/json"
)

// nil slice for error use
var noRecords []MetaxRecord

// top-level, as of 2017-11-21: id research_dataset preservation_state created_by_api modified_by_api alternate_record_set contract data_catalog preservation_state_modified mets_object_identifier dataset_group_edit
type MetaxRecord struct {
	Id               int64   `json:"id"`
	
	CreatedByUserId  *string `json:"created_by_user_id"`
	CreatedByApi     *string `json:"created_by_api"`
	ModifiedByUserId *string `json:"modified_by_user_id"`
	ModifiedByApi    *string `json:"modified_by_api"`
	
	//Editor           *string `json:"editor"`
	Editor           *Editor `json:"editor"`
	
	ResearchDataset  json.RawMessage `json:"research_dataset"`
	Contract         json.RawMessage `json:"contract"`
}


type Editor struct {
	OwnerId    *string `json:"owner_id"`
	CreatorId  *string `json:"creator_id"`
	Identifier *string `json:"identifier"`
}
