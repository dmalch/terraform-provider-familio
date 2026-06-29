package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/dmalch/go-familio"
	"github.com/dmalch/terraform-provider-familio/internal/tfdate"
)

// A life event (birth/death/christening) is exposed as a nested block grouping
// its date, place (settlement uuid) and free-text comment; the birth block also
// carries the person's parents. These attribute-type maps must match the schema
// blocks below exactly.
var (
	lifeEventAttrTypes = map[string]attr.Type{
		"date":    types.ObjectType{AttrTypes: tfdate.AttrTypes},
		"place":   types.StringType,
		"comment": types.StringType,
	}
	birthAttrTypes = map[string]attr.Type{
		"date":    types.ObjectType{AttrTypes: tfdate.AttrTypes},
		"place":   types.StringType,
		"comment": types.StringType,
		"parents": types.SetType{ElemType: types.StringType},
	}
)

// lifeEventModel mirrors a death/christening block.
type lifeEventModel struct {
	Date    types.Object `tfsdk:"date"`
	Place   types.String `tfsdk:"place"`
	Comment types.String `tfsdk:"comment"`
}

// birthModel mirrors the birth block (a life event plus parents).
type birthModel struct {
	Date    types.Object `tfsdk:"date"`
	Place   types.String `tfsdk:"place"`
	Comment types.String `tfsdk:"comment"`
	Parents types.Set    `tfsdk:"parents"`
}

// birthBlock is the schema for the birth event: date, place, comment, parents.
func birthBlock() schema.SingleNestedAttribute {
	attrs := lifeEventAttributes(
		"Birth date.",
		"Birth place — familio's «Место рождения». The UUID of a familio settlement (the same id "+
			"familio_settlement_persons / the familio_person data source speak).",
		"Free-text comment (примечание) on the birth event.",
	)
	attrs["parents"] = schema.SetAttribute{
		Description: "UUIDs of this person's parents (0–2). familio stores them as gender-agnostic " +
			"participants on this person's birth event, so order does not matter and a parent's " +
			"father/mother role is inferred from their own gender. Each parent must already exist.",
		Optional:    true,
		ElementType: types.StringType,
		Validators:  []validator.Set{setvalidator.SizeBetween(0, 2)},
	}
	return schema.SingleNestedAttribute{
		Description: "Birth event — date, place, comment and the person's parents. Edited in place.",
		Optional:    true,
		Attributes:  attrs,
	}
}

// deathBlock is the schema for the death event: date, place, comment.
func deathBlock() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Death event — date, place (familio's «Место смерти»), comment. Setting any field " +
			"records the event; clearing the whole block deletes it. Edited in place.",
		Optional: true,
		Attributes: lifeEventAttributes(
			"Death date.",
			"Death place — a familio settlement UUID.",
			"Free-text comment on the death event.",
		),
	}
}

// christeningBlock is the schema for the christening («Крещение») event.
func christeningBlock() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Christening (baptism) event — familio's «Крещение»: date, place, comment. " +
			"Setting any field records the event; clearing the whole block deletes it.",
		Optional: true,
		Attributes: lifeEventAttributes(
			"Christening date.",
			"Christening place — a familio settlement UUID.",
			"Free-text comment on the christening event.",
		),
	}
}

// lifeEventAttributes builds the shared date/place/comment attributes.
func lifeEventAttributes(dateDesc, placeDesc, commentDesc string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"date":    tfdate.Block(dateDesc, false),
		"place":   schema.StringAttribute{Description: placeDesc, Optional: true},
		"comment": schema.StringAttribute{Description: commentDesc, Optional: true},
	}
}

// birthParts extracts the birth event's facets from the birth block (zero values
// when the block is null/unknown).
func birthParts(ctx context.Context, block types.Object) (date *familio.DateRange, place, comment string, parents []string, diags diag.Diagnostics) {
	if block.IsNull() || block.IsUnknown() {
		return nil, "", "", nil, diags
	}
	var b birthModel
	diags.Append(block.As(ctx, &b, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, "", "", nil, diags
	}
	date, d := tfdate.RangeFromObject(ctx, b.Date)
	diags.Append(d...)
	parents, dp := parentList(ctx, b.Parents)
	diags.Append(dp...)
	return date, strValue(b.Place), strValue(b.Comment), parents, diags
}

// lifeEventParts extracts date/place/comment from a death/christening block.
func lifeEventParts(ctx context.Context, block types.Object) (date *familio.DateRange, place, comment string, diags diag.Diagnostics) {
	if block.IsNull() || block.IsUnknown() {
		return nil, "", "", diags
	}
	var e lifeEventModel
	diags.Append(block.As(ctx, &e, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, "", "", diags
	}
	date, d := tfdate.RangeFromObject(ctx, e.Date)
	diags.Append(d...)
	return date, strValue(e.Place), strValue(e.Comment), diags
}

// hasInfo reports whether a life event carries anything worth recording.
func hasInfo(date *familio.DateRange, place, comment string) bool {
	return date != nil || place != "" || comment != ""
}

// birthBlockValue builds the birth block from a read-back event, or null when it
// carries no information (so an omitted block does not perpetually diff).
func birthBlockValue(ctx context.Context, date *familio.DateRange, place, comment string, parents []string) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !hasInfo(date, place, comment) && len(parents) == 0 {
		return types.ObjectNull(birthAttrTypes), diags
	}
	parentsVal := types.SetNull(types.StringType)
	if len(parents) > 0 {
		s, d := types.SetValueFrom(ctx, types.StringType, parents)
		diags.Append(d...)
		parentsVal = s
	}
	obj, d := types.ObjectValue(birthAttrTypes, map[string]attr.Value{
		"date":    tfdate.ObjectFromRange(date),
		"place":   strOrNull(place),
		"comment": strOrNull(comment),
		"parents": parentsVal,
	})
	diags.Append(d...)
	return obj, diags
}

// lifeEventBlockValue builds a death/christening block from a read-back event, or
// null when the event is absent/empty.
func lifeEventBlockValue(date *familio.DateRange, place, comment string) types.Object {
	if !hasInfo(date, place, comment) {
		return types.ObjectNull(lifeEventAttrTypes)
	}
	obj, _ := types.ObjectValue(lifeEventAttrTypes, map[string]attr.Value{
		"date":    tfdate.ObjectFromRange(date),
		"place":   strOrNull(place),
		"comment": strOrNull(comment),
	})
	return obj
}
