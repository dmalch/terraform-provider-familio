package settlement

import "github.com/hashicorp/terraform-plugin-framework/types"

// Model is the familio_settlement data source state. uuid is the only input; the
// rest is read from familio. region/district/as_of_year are the settlement's
// administrative requisites («реквизиты»); latitude/longitude come from the
// GeoJSON coordinate.
type Model struct {
	UUID            types.String  `tfsdk:"uuid"`
	Name            types.String  `tfsdk:"name"`
	AdditionalNames types.List    `tfsdk:"additional_names"`
	Region          types.String  `tfsdk:"region"`
	District        types.String  `tfsdk:"district"`
	AsOfYear        types.Int64   `tfsdk:"as_of_year"`
	Type            types.String  `tfsdk:"type"`
	Status          types.String  `tfsdk:"status"`
	Latitude        types.Float64 `tfsdk:"latitude"`
	Longitude       types.Float64 `tfsdk:"longitude"`
}
