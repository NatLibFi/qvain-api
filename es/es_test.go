package es

import (
	"testing"
)

func BenchmarkFilters(b *testing.B) {
	var testdata = []byte(`[{"id":"funder_type_tekes","code":"tekes","type":"funder_type","uri":"http://purl.org/att/es/reference_data/funder_type/funder_type_tekes","wkt":"","label":{"fi":"Tekes","en":"Tekes","und":"Tekes"},"parent_ids":[],"child_ids":[],"has_children":false,"same_as":[]},{"id":"funder_type_tekes-shok","code":"tekes-shok","type":"funder_type","uri":"http://purl.org/att/es/reference_data/funder_type/funder_type_tekes-shok","wkt":"","label":{"fi":"Tekes SHOK","en":"Tekes SHOK","und":"Tekes SHOK"},"parent_ids":[],"child_ids":[],"has_children":false,"same_as":[]},{"id":"funder_type_eu-esr","code":"eu-esr","type":"funder_type","uri":"http://purl.org/att/es/reference_data/funder_type/funder_type_eu-esr","wkt":"","label":{"fi":"EU Euroopan sosiaalirahasto ESR","en":"EU European Social Fund ESR","und":"EU Euroopan sosiaalirahasto ESR"},"parent_ids":[],"child_ids":[],"has_children":false,"same_as":[]},{"id":"funder_type_eu-other","code":"eu-other","type":"funder_type","uri":"http://purl.org/att/es/reference_data/funder_type/funder_type_eu-other","wkt":"","label":{"fi":"EU muu rahoitus","en":"EU other funding","und":"EU muu rahoitus"},"parent_ids":[],"child_ids":[],"has_children":false,"same_as":[]},{"id":"funder_type_commercial","code":"commercial","type":"funder_type","uri":"http://purl.org/att/es/reference_data/funder_type/funder_type_commercial","wkt":"","label":{"fi":"Yritys","en":"Commercial","und":"Yritys"},"parent_ids":[],"child_ids":[],"has_children":false,"same_as":[]},{"id":"funder_type_academy-of-finland","code":"academy-of-finland","type":"funder_type","uri":"http://purl.org/att/es/reference_data/funder_type/funder_type_academy-of-finland","wkt":"","label":{"fi":"Suomen Akatemia","en":"Academy of Finland","und":"Suomen Akatemia"},"parent_ids":[],"child_ids":[],"has_children":false,"same_as":[]},{"id":"funder_type_eu-framework-programme","code":"eu-framework-programme","type":"funder_type","uri":"http://purl.org/att/es/reference_data/funder_type/funder_type_eu-framework-programme","wkt":"","label":{"fi":"EU puiteohjelmat","en":"EU Framework Programme","und":"EU puiteohjelmat"},"parent_ids":[],"child_ids":[],"has_children":false,"same_as":[]},{"id":"funder_type_eu-eakr","code":"eu-eakr","type":"funder_type","uri":"http://purl.org/att/es/reference_data/funder_type/funder_type_eu-eakr","wkt":"","label":{"fi":"EU Euroopan aluekehitysrahasto EAKR","en":"EU Regional Development Fund EAKR","und":"EU Euroopan aluekehitysrahasto EAKR"},"parent_ids":[],"child_ids":[],"has_children":false,"same_as":[]},{"id":"funder_type_finnish-fof","code":"finnish-fof","type":"funder_type","uri":"http://purl.org/att/es/reference_data/funder_type/funder_type_finnish-fof","wkt":"","label":{"fi":"Kotimainen rahasto tai säätiö","en":"Finnish fund or foundation","und":"Kotimainen rahasto tai säätiö"},"parent_ids":[],"child_ids":[],"has_children":false,"same_as":[]},{"id":"funder_type_foreign-fof","code":"foreign-fof","type":"funder_type","uri":"http://purl.org/att/es/reference_data/funder_type/funder_type_foreign-fof","wkt":"","label":{"fi":"Ulkomainen rahasto tai säätiö","en":"Foreign fund or foundation","und":"Ulkomainen rahasto tai säätiö"},"parent_ids":[],"child_ids":[],"has_children":false,"same_as":[]},{"id":"funder_type_other-public","code":"other-public","type":"funder_type","uri":"http://purl.org/att/es/reference_data/funder_type/funder_type_other-public","wkt":"","label":{"fi":"Muu julkinen rahoitus","en":"Other public funding","und":"Muu julkinen rahoitus"},"parent_ids":[],"child_ids":[],"has_children":false,"same_as":[]}]`)

	b.Run("Filter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Filter(testdata)
		}
	})
	b.Run("FilterAndCount", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = FilterAndCount(testdata)
		}
	})
}
