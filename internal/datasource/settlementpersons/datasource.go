// Package settlementpersons implements the familio_settlement_persons data
// source: a read-only listing of every person linked to a familio.org
// settlement, backed by the public GET /api/v2/persons?settlement=<uuid>
// endpoint. This is the provider's fully-working read path.
package settlementpersons

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
	resp.TypeName = req.ProviderTypeName + "_settlement_persons"
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
