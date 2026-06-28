package source

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/gomega"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

func TestRefFromModel(t *testing.T) {
	RegisterTestingT(t)

	caseModel := &ResourceModel{
		ReferenceUUID: types.StringValue("case-uuid"),
		Type:          types.StringValue(familio.SourceTypeCase),
		CatalogKey:    types.StringNull(),
	}
	ref := refFromModel(caseModel)
	Expect(ref.UUID).To(Equal("case-uuid"))
	Expect(ref.Type).To(Equal(familio.SourceTypeCase))
	Expect(ref.CatalogKey).To(BeNil())

	catModel := &ResourceModel{
		ReferenceUUID: types.StringValue("cp-uuid"),
		Type:          types.StringValue(familio.SourceTypeCatalogPerson),
		CatalogKey:    types.StringValue("gwarmil"),
	}
	ref = refFromModel(catModel)
	Expect(*ref.CatalogKey).To(Equal("gwarmil"))
}

func TestApplySourceToState(t *testing.T) {
	RegisterTestingT(t)
	// catalog_key is write-only at the API, so it must survive a read untouched.
	m := &ResourceModel{
		Person:        types.StringValue("p1"),
		ReferenceUUID: types.StringValue("u1"),
		Type:          types.StringValue(familio.SourceTypeCatalogPerson),
		CatalogKey:    types.StringValue("gwarmil"),
	}
	applySourceToState(&familio.Source{
		UUID: "u1", Type: familio.SourceTypeCatalogPerson, Comment: "note",
		Name: "Списки", Requisites: "Иванов", Years: "", Catalog: "",
		CreatedAt: "t1", UpdatedAt: "t2",
	}, m)

	Expect(m.Person.ValueString()).To(Equal("p1"), "person is preserved across read")
	Expect(m.CatalogKey.ValueString()).To(Equal("gwarmil"), "write-only catalog_key preserved")
	Expect(m.Comment.ValueString()).To(Equal("note"))
	Expect(m.Name.ValueString()).To(Equal("Списки"))
	Expect(m.Years.IsNull()).To(BeTrue(), "server-empty year span ⇒ null")
	Expect(m.UpdatedAt.ValueString()).To(Equal("t2"))
}
