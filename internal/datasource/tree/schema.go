package tree

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/dmalch/go-familio"
)

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	refAttributes := map[string]schema.Attribute{
		"uuid": schema.StringAttribute{Computed: true, Description: "The related person's UUID."},
		"name": schema.StringAttribute{Computed: true, Description: "The related person's display name (when the event carried one)."},
	}
	resp.Schema = schema.Schema{
		Description: "Breadth-first crawl of the persons connected to a root person, each returned " +
			"with normalized relations (parents, spouses, children) and the wedding-event (marriage) " +
			"uuid on every spouse. Folds the BFS crawl a tree onboarding would otherwise run out-of-band " +
			"(to harvest the UUIDs to import against) into the terraform graph. Use each spouse's " +
			"marriage_uuid to build a familio_marriage import id (\"<person_uuid>:<marriage_uuid>\").",
		Attributes: map[string]schema.Attribute{
			"root": schema.StringAttribute{
				Description: "UUID of the person to crawl outward from (crawl distance 0).",
				Required:    true,
			},
			"direction": schema.StringAttribute{
				Description: "Which edges to follow: up (ancestors only), down (descendants only), " +
					"or component (parents, spouses and children — the whole connected component). " +
					"Defaults to component.",
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf(familio.TreeUp, familio.TreeDown, familio.TreeComponent)},
			},
			"surname": schema.StringAttribute{
				Description: "When set, the crawl does not expand through people whose surname (or " +
					"maiden surname) does not match — the way to keep married-in branches out of an " +
					"on-surname component. Non-matching people are still returned (they are referenced), " +
					"just not expanded. The root is always expanded. Case-insensitive.",
				Optional: true,
			},
			"depth": schema.Int64Attribute{
				Description: "Cap on crawl distance from the root (root = 0). Omit or set 0 for unlimited.",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(0)},
			},
			"nodes": schema.ListNestedAttribute{
				Description: "The crawled persons, in breadth-first discovery order.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{Computed: true, Description: "The person's UUID."},
						"name": schema.StringAttribute{Computed: true, Description: "The person's display name."},
						"year": schema.Int64Attribute{Computed: true, Description: "Birth year (null when unknown)."},
						"parents": schema.ListNestedAttribute{
							Computed:     true,
							Description:  "This person's parents.",
							NestedObject: schema.NestedAttributeObject{Attributes: refAttributes},
						},
						"spouses": schema.ListNestedAttribute{
							Computed:    true,
							Description: "This person's spouses, each with the marriage_uuid identifying the underlying union.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"uuid": schema.StringAttribute{Computed: true, Description: "The spouse's UUID."},
									"name": schema.StringAttribute{Computed: true, Description: "The spouse's display name."},
									"marriage_uuid": schema.StringAttribute{
										Computed: true,
										Description: "UUID of the wedding event that is this marriage — pair it with either " +
											"spouse's UUID to import a familio_marriage.",
									},
								},
							},
						},
						"children": schema.ListNestedAttribute{
							Computed:     true,
							Description:  "This person's children.",
							NestedObject: schema.NestedAttributeObject{Attributes: refAttributes},
						},
					},
				},
			},
		},
	}
}
