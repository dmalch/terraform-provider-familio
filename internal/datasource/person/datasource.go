// Package person implements the familio_person data source: a single-person
// lookup by uuid that surfaces the owning account (owner_id) and the person's
// relationships (parents/spouses/children), so importable tree nodes can be
// discovered declaratively without dropping to the raw familio API.
package person

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
	resp.TypeName = req.ProviderTypeName + "_person"
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
