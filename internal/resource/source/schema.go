package source

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A source citation («Источник») on a familio.org person — an entry of the " +
			"person's «Источники» tab. A source references a catalogued entity (an archival " +
			"document/`case`, or a `catalog_person` index record) and carries an editable comment; " +
			"the reference is immutable (changing it forces replacement) while the comment edits in " +
			"place. The same sources can instead be managed inline via the `sources` block on " +
			"familio_person — manage a given person's sources through one surface or the other, not both.",
		Attributes: map[string]schema.Attribute{
			"person": schema.StringAttribute{
				Description:   "UUID of the person this source belongs to. Must already exist.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"reference_uuid": schema.StringAttribute{
				Description: "UUID of the catalogued entity this source cites: an archive case (дело) " +
					"uuid for type `case`, or a catalog-person record uuid for type `catalog_person`. " +
					"Immutable — changing it replaces the source.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"type": schema.StringAttribute{
				Description: "Source kind: `case` (a digitised archival document — дело — from the " +
					"organization → fund → register → case catalog) or `catalog_person` (a record from " +
					"a people index). Immutable.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(familio.SourceTypeCase, familio.SourceTypeCatalogPerson),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"catalog_key": schema.StringAttribute{
				Description: "For `catalog_person` sources, the id of the source catalog the record " +
					"comes from (e.g. `gwarmil`); a catalog-person uuid is only unique within its " +
					"catalog. Omit for `case` sources. Write-only at the familio API (not returned on " +
					"reads), so it is not recovered on import.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"comment": schema.StringAttribute{
				Description: "Free-text note on the citation (примечание). Edited in place.",
				Optional:    true,
			},

			"name":       computedString("Server-derived label for the cited source (e.g. «Ревизские сказки»)."),
			"requisites": computedString("Server-derived archive coordinates / реквизиты (e.g. «… ф. 145 оп. 1 д. 431»)."),
			"years":      computedString("Server-derived year span of the cited source."),
			"catalog":    computedString("Server-derived catalog descriptor, when applicable."),
			"created_at": computedString("Creation timestamp."),
			"updated_at": schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}

// computedString is a read-only string attribute derived from the immutable
// reference, so it carries its state value through a comment-only update.
func computedString(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		Computed:      true,
		Description:   desc,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}
