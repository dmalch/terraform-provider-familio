// Package tfdate bridges familio's complex-date model and a Terraform nested
// {year, month, day} attribute, shared by the person and marriage resources.
package tfdate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	Year  types.Int64 `tfsdk:"year"`
	Month types.Int64 `tfsdk:"month"`
	Day   types.Int64 `tfsdk:"day"`
}

// AttrTypes is the attr-type map for the nested date object.
var AttrTypes = map[string]attr.Type{
	"year":  types.Int64Type,
	"month": types.Int64Type,
	"day":   types.Int64Type,
}

// Block builds a nested {year, month, day} date attribute. When requiresReplace
// is true, changing the date forces replacement of the owning resource (used
// where the underlying event cannot be edited in place — e.g. the marriage
// resource). When false, the date is edited in place by the owning resource's
// Update (the person resource rebuilds its birth/death event).
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
		},
	}
}

// PartFromObject converts a nested {year,month,day} object into a DatePart
// (nil when the object is null/unknown).
func PartFromObject(ctx context.Context, obj types.Object) (*familio.DatePart, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}
	var dm model
	diags.Append(obj.As(ctx, &dm, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	part := &familio.DatePart{Year: int(dm.Year.ValueInt64())}
	if !dm.Month.IsNull() && !dm.Month.IsUnknown() {
		mm := int(dm.Month.ValueInt64())
		part.Month = &mm
	}
	if !dm.Day.IsNull() && !dm.Day.IsUnknown() {
		dd := int(dm.Day.ValueInt64())
		part.Day = &dd
	}
	return part, diags
}

// Object builds a nested date object from a DatePart (null when nil).
func Object(part *familio.DatePart) types.Object {
	if part == nil {
		return types.ObjectNull(AttrTypes)
	}
	obj, _ := types.ObjectValue(AttrTypes, map[string]attr.Value{
		"year":  types.Int64Value(int64(part.Year)),
		"month": int64PtrValue(part.Month),
		"day":   int64PtrValue(part.Day),
	})
	return obj
}

func int64PtrValue(p *int) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*p))
}
