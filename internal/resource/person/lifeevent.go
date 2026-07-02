package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
			"father/mother role is inferred from their own gender. Each parent must already exist. " +
			"Preserve-on-omit: within a managed birth block, omitting this keeps the current parents " +
			"(set to [] to clear them).",
		Optional:      true,
		Computed:      true,
		ElementType:   types.StringType,
		PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
		Validators:    []validator.Set{setvalidator.SizeBetween(0, 2)},
	}
	return schema.SingleNestedAttribute{
		Description: "Birth event — date, place, comment and the person's parents. Edited in place. " +
			blockPreserveDoc,
		Optional:   true,
		Attributes: attrs,
	}
}

// deathBlock is the schema for the death event: date, place, comment.
func deathBlock() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Death event — date, place (familio's «Место смерти»), comment. Edited in place. " +
			blockPreserveDoc,
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
			blockPreserveDoc,
		Optional: true,
		Attributes: lifeEventAttributes(
			"Christening date.",
			"Christening place — a familio settlement UUID.",
			"Free-text comment on the christening event.",
		),
	}
}

// blockPreserveDoc documents the preserve-on-omit contract shared by the three
// life-event blocks (see issue #22 and read.go/update.go for the mechanism).
const blockPreserveDoc = "Preserve-on-omit: omitting the whole block leaves the person's existing " +
	"event untouched (it is treated as unmanaged, like the sources block), so importing a person " +
	"and enriching it never clobbers events the config does not carry. Within a block you do declare, " +
	"omitted fields are likewise preserved. To remove an event, delete it in the familio UI."

// lifeEventAttributes builds the shared date/place/comment attributes. place and
// comment are preserve-on-omit (Optional + Computed + UseStateForUnknown) so that
// setting one facet of a managed event does not null the others — e.g. updating a
// date keeps that event's comment (issue #22). date stays plain Optional: a
// Computed nested-object attribute triggers a perpetual "known after apply" plan
// in terraform-plugin-framework, so whole-block/date preservation is handled in
// Read/Update instead (an omitted block is left unmanaged).
func lifeEventAttributes(dateDesc, placeDesc, commentDesc string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"date": tfdate.Block(dateDesc, false),
		"place": schema.StringAttribute{
			Description:   placeDesc + " Preserve-on-omit within a managed block.",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"comment": schema.StringAttribute{
			Description:   commentDesc + " Preserve-on-omit within a managed block.",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
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

// birthPartsMerged is birthParts for the Update write path: within a managed
// birth block, an omitted comment/place/parents (unknown — e.g. right after
// import, before UseStateForUnknown has a prior state to reuse) is filled from
// the person's CURRENT birth event rather than cleared, so setting a birth date
// never strips the existing comment or parents (issue #22). date is authoritative
// (taken from the block): managing a birth means declaring its date.
func birthPartsMerged(ctx context.Context, block types.Object, server *familio.Event) (date *familio.DateRange, place, comment string, parents []string, diags diag.Diagnostics) {
	var b birthModel
	diags.Append(block.As(ctx, &b, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, "", "", nil, diags
	}
	date, d := tfdate.RangeFromObject(ctx, b.Date)
	diags.Append(d...)
	parents, dp := mergeParents(ctx, b.Parents, serverParents(server))
	diags.Append(dp...)
	return date, mergeString(b.Place, serverPlace(server)), mergeString(b.Comment, serverComment(server)), parents, diags
}

// lifeEventPartsMerged is lifeEventParts with the same server-fallback merge for
// a death/christening block's place and comment.
func lifeEventPartsMerged(ctx context.Context, block types.Object, server *familio.Event) (date *familio.DateRange, place, comment string, diags diag.Diagnostics) {
	var e lifeEventModel
	diags.Append(block.As(ctx, &e, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, "", "", diags
	}
	date, d := tfdate.RangeFromObject(ctx, e.Date)
	diags.Append(d...)
	return date, mergeString(e.Place, serverPlace(server)), mergeString(e.Comment, serverComment(server)), diags
}

// mergeString returns the planned string when the config set it (known — an
// explicit "" clears the field), else the server's current value (preserve).
func mergeString(planned types.String, server string) string {
	if planned.IsUnknown() || planned.IsNull() {
		return server
	}
	return planned.ValueString()
}

// mergeParents returns the planned parents when the config set them (known — an
// explicit [] clears them), else the server's current parents (preserve).
func mergeParents(ctx context.Context, planned types.Set, server []string) ([]string, diag.Diagnostics) {
	if planned.IsUnknown() || planned.IsNull() {
		return server, nil
	}
	return parentList(ctx, planned)
}

func serverComment(ev *familio.Event) string {
	if ev == nil {
		return ""
	}
	return ev.Comment
}

func serverPlace(ev *familio.Event) string {
	if ev == nil {
		return ""
	}
	return ev.SettlementUUID()
}

func serverParents(ev *familio.Event) []string {
	if ev == nil {
		return nil
	}
	return ev.ParentUUIDs()
}

// hasInfo reports whether a life event carries anything worth recording.
func hasInfo(date *familio.DateRange, place, comment string) bool {
	return date != nil || place != "" || comment != ""
}

// firstEventOfType returns the first event of the given type, or nil.
func firstEventOfType(events []familio.Event, typ string) *familio.Event {
	for i := range events {
		if events[i].Type == typ {
			return &events[i]
		}
	}
	return nil
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
