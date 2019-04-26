package metax

import (
	"encoding/json"
)

const (
	SchemaIda = "metax-ida"
	SchemaAtt = "metax-att"
)

var (
	// CatalogIdentifiers are used to identify schema in datasets imported from Metax.
	CatalogIdentifiers = map[string]string{
		"urn:nbn:fi:att:data-catalog-ida": SchemaIda,
		"urn:nbn:fi:att:data-catalog-att": SchemaAtt,
	}

	// templates variable contains the base templates for empty metax datasets.
	templates = map[string]json.RawMessage{
		SchemaIda: json.RawMessage(`
		{
			"data_catalog": "urn:nbn:fi:att:data-catalog-ida",
			"metadata_provider_org": "",
			"metadata_provider_user": "",
			"research_dataset": {}
		}`),
		SchemaAtt: json.RawMessage(`
		{
			"data_catalog": "urn:nbn:fi:att:data-catalog-att",
			"metadata_provider_org": "",
			"metadata_provider_user": "",
			"research_dataset": {}
		}`),
	}

	// parsedTemplates contains the pre-parsed unserialised json for the templates.
	parsedTemplates = map[string]map[string]*json.RawMessage{}
)

func init() {
	// preparse on start-up
	for schema, template := range templates {
		var parsed map[string]*json.RawMessage
		err := json.Unmarshal(template, &parsed)
		if err != nil {
			panic(err)
		}
		parsedTemplates[schema] = parsed
	}
}
