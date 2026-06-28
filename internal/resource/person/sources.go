package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
	"github.com/dmalch/terraform-provider-familio/internal/tfsource"
)

// sourceElemType is the object type of one element of the `sources` list.
var sourceElemType = types.ObjectType{AttrTypes: tfsource.AttrTypes}

// sourcesBlock is the person's «Источники» as an authoritative list. When set,
// the provider makes the person's source set exactly match it (creating,
// deleting and re-commenting as needed); when omitted (null) the provider does
// not manage the person's sources at all — leaving them to standalone
// familio_source resources. An empty list means "remove all sources". Manage a
// given person's sources through this block OR via familio_source, never both.
func sourcesBlock() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "Source citations («Источники») for this person, managed as an authoritative " +
			"set: the provider makes familio match this list exactly. Omit the block to leave the " +
			"person's sources unmanaged; use `[]` to remove them all. Mutually exclusive with " +
			"standalone familio_source resources for the same person.",
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"reference_uuid": schema.StringAttribute{
					Description: "UUID of the cited entity (an archive case/дело for type `case`, or a " +
						"catalog-person record for type `catalog_person`).",
					Required: true,
				},
				"type": schema.StringAttribute{
					Description: "Source kind: `case` or `catalog_person`.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.OneOf(familio.SourceTypeCase, familio.SourceTypeCatalogPerson),
					},
				},
				"catalog_key": schema.StringAttribute{
					Description: "For `catalog_person`, the source-catalog id (e.g. `gwarmil`); omit for " +
						"`case`. Write-only at the familio API.",
					Optional: true,
				},
				"comment": schema.StringAttribute{
					Description: "Free-text note on the citation. Edited in place.",
					Optional:    true,
				},
				"name":       schema.StringAttribute{Computed: true, Description: "Server-derived source label."},
				"requisites": schema.StringAttribute{Computed: true, Description: "Server-derived archive coordinates."},
				"years":      schema.StringAttribute{Computed: true, Description: "Server-derived year span."},
				"catalog":    schema.StringAttribute{Computed: true, Description: "Server-derived catalog descriptor."},
				"created_at": schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
				"updated_at": schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
			},
		},
	}
}

// desiredSources decodes the sources block into per-source models. A null/unknown
// block (the person's sources are unmanaged) yields nil, false.
func desiredSources(ctx context.Context, list types.List) (models []tfsource.Model, managed bool, diags diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, false, diags
	}
	diags = list.ElementsAs(ctx, &models, false)
	return models, true, diags
}

// writeSources makes the person's familio source set match the desired list
// (authoritative): create the missing, delete the extra, patch changed comments.
func (r *Resource) writeSources(ctx context.Context, uuid string, desired []tfsource.Model) diag.Diagnostics {
	var diags diag.Diagnostics

	current, err := r.client.GetPersonSources(ctx, uuid)
	if err != nil {
		diags.AddError("Cannot read familio_person sources before reconciling", err.Error())
		return diags
	}
	byID := make(map[string]familio.Source, len(current))
	for _, s := range current {
		byID[s.UUID] = s
	}

	keep := make(map[string]bool, len(desired))
	for _, d := range desired {
		ref := d.ReferenceUUID.ValueString()
		keep[ref] = true
		comment := d.Comment.ValueString()
		existing, ok := byID[ref]
		if !ok {
			created, err := r.client.CreateSource(ctx, uuid, tfsource.Ref(ref, d.Type.ValueString(), d.CatalogKey))
			if err != nil {
				diags.AddError("Cannot create familio_person source", err.Error())
				return diags
			}
			if comment != "" {
				if _, err := r.client.UpdateSourceComment(ctx, uuid, created.UUID, comment); err != nil {
					diags.AddError("Cannot set familio_person source comment", err.Error())
					return diags
				}
			}
			continue
		}
		if existing.Comment != comment {
			if _, err := r.client.UpdateSourceComment(ctx, uuid, ref, comment); err != nil {
				diags.AddError("Cannot update familio_person source comment", err.Error())
				return diags
			}
		}
	}

	for _, s := range current {
		if !keep[s.UUID] {
			if err := r.client.DeleteSource(ctx, uuid, s.UUID); err != nil {
				diags.AddError("Cannot delete familio_person source", err.Error())
				return diags
			}
		}
	}
	return diags
}

// readSources rebuilds the sources list from familio, ordered to match the prior
// list (by reference_uuid) so an unchanged set re-plans empty, then appending any
// sources present in familio but not in prior (out-of-band drift). The write-only
// catalog_key is carried from the matching prior element. A null prior block stays
// null (unmanaged) and no read is performed.
func (r *Resource) readSources(ctx context.Context, uuid string, prior types.List) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if prior.IsNull() || prior.IsUnknown() {
		return types.ListNull(sourceElemType), diags
	}

	priorModels, _, d := desiredSources(ctx, prior)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(sourceElemType), diags
	}
	catalogKeyByID := make(map[string]types.String, len(priorModels))
	for _, m := range priorModels {
		catalogKeyByID[m.ReferenceUUID.ValueString()] = m.CatalogKey
	}

	sources, err := r.client.GetPersonSources(ctx, uuid)
	if err != nil {
		diags.AddError("Error reading familio_person sources", err.Error())
		return types.ListNull(sourceElemType), diags
	}
	byID := make(map[string]familio.Source, len(sources))
	for _, s := range sources {
		byID[s.UUID] = s
	}

	ordered := make([]tfsource.Model, 0, len(sources))
	seen := make(map[string]bool, len(sources))
	for _, m := range priorModels { // prior order first
		ref := m.ReferenceUUID.ValueString()
		if s, ok := byID[ref]; ok {
			ordered = append(ordered, tfsource.ModelFromSource(s, m.CatalogKey))
			seen[ref] = true
		}
	}
	for _, s := range sources { // then any out-of-band additions
		if !seen[s.UUID] {
			ordered = append(ordered, tfsource.ModelFromSource(s, catalogKeyByID[s.UUID]))
		}
	}

	list, d := types.ListValueFrom(ctx, sourceElemType, ordered)
	diags.Append(d...)
	return list, diags
}
