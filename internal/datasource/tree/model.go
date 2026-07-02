package tree

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model is the familio_tree data source state. root/direction/surname/depth are
// the inputs (they bound the crawl); nodes is the crawled result.
type Model struct {
	Root      types.String `tfsdk:"root"`
	Direction types.String `tfsdk:"direction"`
	Surname   types.String `tfsdk:"surname"`
	Depth     types.Int64  `tfsdk:"depth"`
	Nodes     types.List   `tfsdk:"nodes"`
}

// NodeModel is one crawled person with their normalized relations.
type NodeModel struct {
	UUID     types.String `tfsdk:"uuid"`
	Name     types.String `tfsdk:"name"`
	Year     types.Int64  `tfsdk:"year"`
	Parents  types.List   `tfsdk:"parents"`
	Spouses  types.List   `tfsdk:"spouses"`
	Children types.List   `tfsdk:"children"`
}

// RefModel is a minimal reference to a related person (parents/children).
type RefModel struct {
	UUID types.String `tfsdk:"uuid"`
	Name types.String `tfsdk:"name"`
}

// SpouseModel is a spouse reference plus the wedding-event uuid that identifies
// the marriage — the id needed to import a familio_marriage.
type SpouseModel struct {
	UUID         types.String `tfsdk:"uuid"`
	Name         types.String `tfsdk:"name"`
	MarriageUUID types.String `tfsdk:"marriage_uuid"`
}

// Object-type maps for the nested lists. They must mirror the schema exactly.
var (
	refAttrTypes = map[string]attr.Type{
		"uuid": types.StringType,
		"name": types.StringType,
	}
	spouseAttrTypes = map[string]attr.Type{
		"uuid":          types.StringType,
		"name":          types.StringType,
		"marriage_uuid": types.StringType,
	}
	nodeAttrTypes = map[string]attr.Type{
		"uuid":     types.StringType,
		"name":     types.StringType,
		"year":     types.Int64Type,
		"parents":  types.ListType{ElemType: types.ObjectType{AttrTypes: refAttrTypes}},
		"spouses":  types.ListType{ElemType: types.ObjectType{AttrTypes: spouseAttrTypes}},
		"children": types.ListType{ElemType: types.ObjectType{AttrTypes: refAttrTypes}},
	}
)

func refObjectType() types.ObjectType    { return types.ObjectType{AttrTypes: refAttrTypes} }
func spouseObjectType() types.ObjectType { return types.ObjectType{AttrTypes: spouseAttrTypes} }
func nodeObjectType() types.ObjectType   { return types.ObjectType{AttrTypes: nodeAttrTypes} }
