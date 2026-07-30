package tags

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model is the familio_tags data source state. It takes no input: the catalogue
// belongs to the authenticated account. ByName is the lookup the data source
// exists for; Tags carries the full records.
type Model struct {
	ByName types.Map  `tfsdk:"by_name"`
	IDs    types.Set  `tfsdk:"ids"`
	Tags   types.List `tfsdk:"tags"`
}

// TagModel is one element of the tags list.
type TagModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Color       types.String `tfsdk:"color"`
	Description types.String `tfsdk:"description"`
	Hex         types.String `tfsdk:"hex"`
	IsFree      types.Bool   `tfsdk:"is_free"`
}

// tagAttrTypes mirrors TagModel's tfsdk tags; it must stay in step with both
// TagModel and the nested object in schema.go.
var tagAttrTypes = map[string]attr.Type{
	"id":          types.Int64Type,
	"name":        types.StringType,
	"color":       types.StringType,
	"description": types.StringType,
	"hex":         types.StringType,
	"is_free":     types.BoolType,
}

// tagElemType is the object type of one element of the tags list.
var tagElemType = types.ObjectType{AttrTypes: tagAttrTypes}
