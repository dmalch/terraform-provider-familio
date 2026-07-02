// Package tree implements the familio_tree data source: a breadth-first crawl of
// the persons connected to a root person, returning each with normalized
// relations (parents/spouses/children) and — on each spouse — the underlying
// wedding-event (marriage) uuid. It folds the hand-written BFS crawl every tree
// onboarding needed (to harvest the uuids to import against) into the terraform
// graph. Backed by go-familio's Client.CrawlTree.
package tree

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
	resp.TypeName = req.ProviderTypeName + "_tree"
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
