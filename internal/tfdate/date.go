// Package tfdate bridges familio's complex-date model and a Terraform nested
// date attribute, shared by the person, marriage and event resources. The block
// mirrors the date model of the sibling terraform-provider-genealogy: a primary
// {year, month, day} with an optional approximation (circa), bound/range
// (range + end_*) and calendar. The familio wire shape lives behind
// familio.DateRange / familio.EventDateFromRange (see internal/familio/date.go).
package tfdate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

// model is one nested date block.
type model struct {
	Year     types.Int64  `tfsdk:"year"`
	Month    types.Int64  `tfsdk:"month"`
	Day      types.Int64  `tfsdk:"day"`
	Circa    types.Bool   `tfsdk:"circa"`
	Range    types.String `tfsdk:"range"`
	EndYear  types.Int64  `tfsdk:"end_year"`
	EndMonth types.Int64  `tfsdk:"end_month"`
	EndDay   types.Int64  `tfsdk:"end_day"`
	EndCirca types.Bool   `tfsdk:"end_circa"`
	Calendar types.String `tfsdk:"calendar"`
}

// AttrTypes is the attr-type map for the nested date object.
var AttrTypes = map[string]attr.Type{
	"year":      types.Int64Type,
	"month":     types.Int64Type,
	"day":       types.Int64Type,
	"circa":     types.BoolType,
	"range":     types.StringType,
	"end_year":  types.Int64Type,
	"end_month": types.Int64Type,
	"end_day":   types.Int64Type,
	"end_circa": types.BoolType,
	"calendar":  types.StringType,
}

// rangePath points at the sibling "range" attribute, used by the end_* fields'
// AlsoRequires validators.
var rangePath = path.MatchRelative().AtParent().AtName("range")

// Block builds the nested date attribute. When requiresReplace is true, changing
// the date forces replacement of the owning resource (used where the underlying
// event cannot be edited in place — marriage, event). When false, the date is
// edited in place by the owning resource's Update (the person resource rebuilds
// its birth/death event).
func Block(desc string, requiresReplace bool) schema.SingleNestedAttribute {
	var mods []planmodifier.Object
	if requiresReplace {
		desc += " Changing it forces a new resource (event editing is not yet supported)."
		mods = append(mods, objectplanmodifier.RequiresReplace())
	}
	return schema.SingleNestedAttribute{
		Description:   desc,
		Optional:      true,
		PlanModifiers: mods,
		Validators:    []validator.Object{dateRangeValidator{}},
		Attributes: map[string]schema.Attribute{
			"year": schema.Int64Attribute{
				Description: "Year (e.g. 1900).",
				Required:    true,
			},
			"month": schema.Int64Attribute{
				Description: "Month, 1-12.",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 12)},
			},
			"day": schema.Int64Attribute{
				Description: "Day of month, 1-31.",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 31)},
			},
			"circa": schema.BoolAttribute{
				Description: "Approximate (\"circa\") date — familio's \"about\" type. " +
					"Cannot be combined with range.",
				Optional: true,
			},
			"range": schema.StringAttribute{
				Description: "Open bound or range: before | after | between. Omit for a " +
					"single date. \"between\" needs end_year (the second endpoint).",
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf(familio.RangeBefore, familio.RangeAfter, familio.RangeBetween)},
			},
			"end_year": schema.Int64Attribute{
				Description: "Second endpoint year (only with range = \"between\").",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.AlsoRequires(rangePath)},
			},
			"end_month": schema.Int64Attribute{
				Description: "Second endpoint month, 1-12 (only with range = \"between\").",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 12), int64validator.AlsoRequires(rangePath)},
			},
			"end_day": schema.Int64Attribute{
				Description: "Second endpoint day, 1-31 (only with range = \"between\").",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 31), int64validator.AlsoRequires(rangePath)},
			},
			"end_circa": schema.BoolAttribute{
				Description: "Accepted for cross-provider config symmetry, but familio has no " +
					"per-endpoint approximation, so it cannot be combined with range.",
				Optional: true,
			},
			"calendar": schema.StringAttribute{
				Description: "Calendar: gregorian (default) | julian.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.OneOf(calendarGregorian, calendarJulian)},
			},
		},
	}
}

const (
	calendarGregorian = "gregorian"
	calendarJulian    = "julian"
)

// RangeFromObject converts a nested date object into a familio.DateRange (nil
// when the object is null/unknown).
func RangeFromObject(ctx context.Context, obj types.Object) (*familio.DateRange, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}
	var m model
	diags.Append(obj.As(ctx, &m, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	r := &familio.DateRange{
		Year:     int(m.Year.ValueInt64()),
		Month:    int64PtrToInt(m.Month),
		Day:      int64PtrToInt(m.Day),
		Circa:    m.Circa.ValueBool(),
		Range:    m.Range.ValueString(),
		EndYear:  int64PtrToInt(m.EndYear),
		EndMonth: int64PtrToInt(m.EndMonth),
		EndDay:   int64PtrToInt(m.EndDay),
		EndCirca: m.EndCirca.ValueBool(),
		Calendar: m.Calendar.ValueString(),
	}
	return r, diags
}

// ObjectFromRange builds a nested date object from a familio.DateRange (null when
// nil).
func ObjectFromRange(r *familio.DateRange) types.Object {
	if r == nil {
		return types.ObjectNull(AttrTypes)
	}
	obj, _ := types.ObjectValue(AttrTypes, map[string]attr.Value{
		"year":      types.Int64Value(int64(r.Year)),
		"month":     intPtrToInt64(r.Month),
		"day":       intPtrToInt64(r.Day),
		"circa":     boolOrNull(r.Circa),
		"range":     stringOrNull(r.Range),
		"end_year":  intPtrToInt64(r.EndYear),
		"end_month": intPtrToInt64(r.EndMonth),
		"end_day":   intPtrToInt64(r.EndDay),
		"end_circa": boolOrNull(r.EndCirca),
		"calendar":  stringOrNull(r.Calendar),
	})
	return obj
}

func int64PtrToInt(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}

func intPtrToInt64(p *int) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*p))
}

// boolOrNull keeps an unset bool null so it never shows a spurious diff against a
// config that omits it.
func boolOrNull(b bool) types.Bool {
	if !b {
		return types.BoolNull()
	}
	return types.BoolValue(true)
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
