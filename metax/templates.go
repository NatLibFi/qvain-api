package metax

import (
	"encoding/json"
)

var templates = map[string]json.RawMessage{
	"metax-ida": json.RawMessage(`
		{
			"data_catalog": "urn:nbn:fi:att:data-catalog-ida",
			"metadata_provider_org": "trump corporation",
			"metadata_provider_user": "donald_trump",
			"research_dataset": {}
		}`),
	"metax-att": json.RawMessage(`
		{
		"data_catalog": "urn:nbn:fi:att:data-catalog-att",
		"metadata_provider_org": "trump corporation",
		"metadata_provider_user": "donald_trump",
		"research_dataset": {}
	}`),
}

var parsedTemplates = map[string]map[string]*json.RawMessage{}

func init() {
	for schema, template := range templates {
		var parsed map[string]*json.RawMessage
		err := json.Unmarshal(template, &parsed)
		if err != nil {
			panic(err)
		}
		parsedTemplates[schema] = parsed
	}
}
