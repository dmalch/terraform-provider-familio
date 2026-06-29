// Package settlement implements the familio_settlement data source: a single
// settlement (place) lookup by uuid, surfacing its canonical name, administrative
// requisites («реквизиты»), classification and coordinates — to resolve/validate
// the settlement UUIDs that familio_person's birth/death/christening places and
// familio_source speak. Backed by GET /api/v2/settlements/<uuid>.
package settlement

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/dmalch/go-familio"
	"github.com/dmalch/terraform-provider-familio/internal/config"
)

type DataSource struct {
	client *familio.Client
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_settlement"
}

func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*config.ClientData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *config.ClientData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = data.Client
}
