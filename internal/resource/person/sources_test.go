package person

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/gomega"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
	"github.com/dmalch/terraform-provider-familio/internal/tfsource"
)

func sourceList(t *testing.T, models ...tfsource.Model) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(t.Context(), sourceElemType, models)
	Expect(diags).To(BeEmpty())
	return list
}

func TestDesiredSources(t *testing.T) {
	t.Run("a null block is unmanaged", func(t *testing.T) {
		RegisterTestingT(t)
		models, managed, diags := desiredSources(t.Context(), types.ListNull(sourceElemType))
		Expect(diags).To(BeEmpty())
		Expect(managed).To(BeFalse())
		Expect(models).To(BeNil())
	})

	t.Run("an empty block is managed (delete-all)", func(t *testing.T) {
		RegisterTestingT(t)
		// `sources = []` is a non-null, empty list (≠ an omitted/null block).
		empty := types.ListValueMust(sourceElemType, []attr.Value{})
		models, managed, diags := desiredSources(t.Context(), empty)
		Expect(diags).To(BeEmpty())
		Expect(managed).To(BeTrue())
		Expect(models).To(BeEmpty())
	})

	t.Run("decodes the configured sources", func(t *testing.T) {
		RegisterTestingT(t)
		in := sourceList(t,
			tfsource.Model{
				ReferenceUUID: types.StringValue("u1"),
				Type:          types.StringValue(familio.SourceTypeCase),
				CatalogKey:    types.StringNull(),
				Comment:       types.StringValue("note"),
			},
			tfsource.Model{
				ReferenceUUID: types.StringValue("u2"),
				Type:          types.StringValue(familio.SourceTypeCatalogPerson),
				CatalogKey:    types.StringValue("gwarmil"),
			},
		)
		models, managed, diags := desiredSources(t.Context(), in)
		Expect(diags).To(BeEmpty())
		Expect(managed).To(BeTrue())
		Expect(models).To(HaveLen(2))
		Expect(models[0].ReferenceUUID.ValueString()).To(Equal("u1"))
		Expect(models[0].Comment.ValueString()).To(Equal("note"))
		Expect(models[1].CatalogKey.ValueString()).To(Equal("gwarmil"))

		// And those decode straight into create bodies.
		ref := tfsource.Ref(models[1].ReferenceUUID.ValueString(), models[1].Type.ValueString(), models[1].CatalogKey)
		Expect(*ref.CatalogKey).To(Equal("gwarmil"))
	})
}
