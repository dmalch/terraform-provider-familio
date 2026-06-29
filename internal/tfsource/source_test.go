package tfsource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/gomega"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

func TestRefCatalogKey(t *testing.T) {
	RegisterTestingT(t)
	// A case source: no catalog_key ⇒ nil (marshals to null).
	ref := Ref("u1", familio.SourceTypeCase, types.StringNull())
	Expect(ref.UUID).To(Equal("u1"))
	Expect(ref.Type).To(Equal(familio.SourceTypeCase))
	Expect(ref.CatalogKey).To(BeNil())

	// A catalog_person source carries its catalog id.
	ref = Ref("u2", familio.SourceTypeCatalogPerson, types.StringValue("gwarmil"))
	Expect(ref.CatalogKey).ToNot(BeNil())
	Expect(*ref.CatalogKey).To(Equal("gwarmil"))
}

func TestModelFromSourcePreservesCatalogKey(t *testing.T) {
	RegisterTestingT(t)
	s := familio.Source{
		UUID: "u2", Type: familio.SourceTypeCatalogPerson, Comment: "",
		Name: "Списки", Requisites: "Иванов", Years: "",
		Catalog:   &familio.SourceCatalog{Key: "gwarmil"},
		CreatedAt: "t1", UpdatedAt: "t2",
	}
	// catalog_key is write-only at the API, so it's carried from prior state.
	m := ModelFromSource(s, types.StringValue("gwarmil"))
	Expect(m.CatalogKey.ValueString()).To(Equal("gwarmil"))
	// the server's catalog OBJECT maps to its key on the computed attribute
	Expect(m.Catalog.ValueString()).To(Equal("gwarmil"))
	Expect(m.Name.ValueString()).To(Equal("Списки"))
	// Server-empty strings become null (no permadiff on omitted optionals).
	Expect(m.Comment.IsNull()).To(BeTrue())
	Expect(m.Years.IsNull()).To(BeTrue())
	Expect(m.CreatedAt.ValueString()).To(Equal("t1"))
}

func TestObjectModelRoundTrip(t *testing.T) {
	RegisterTestingT(t)
	in := ModelFromSource(
		familio.Source{UUID: "u1", Type: familio.SourceTypeCase, Comment: "note", Name: "Ревизские сказки"},
		types.StringNull(),
	)
	obj, diags := ObjectFromModel(t.Context(), in)
	Expect(diags).To(BeEmpty())

	out, diags := ModelFromObject(t.Context(), obj)
	Expect(diags).To(BeEmpty())
	Expect(out).To(Equal(in))
}
