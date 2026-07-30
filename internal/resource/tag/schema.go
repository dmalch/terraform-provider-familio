package tag

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/dmalch/go-familio"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A tag («метка») in the account's tag catalogue — an entry of the " +
			"/profile/my-tags page. Tags are coloured labels used to group persons; attach them " +
			"to a person with the `tags` attribute on familio_person. Everything about a tag " +
			"(name, colour, description) edits in place.\n\n" +
			"Tags are a **Familio Plus** feature. Without a subscription only one tag is usable " +
			"— the one familio flags as free (see the computed `is_free` attribute) — though the " +
			"API returns any others and Terraform can still manage them.\n\n" +
			"~> Deleting a familio_tag removes the tag from **every** person it was attached to, " +
			"not just the ones this configuration manages.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The tag's familio id. Unlike other familio resources a tag is keyed by " +
					"a small sequential **integer**, not a uuid — this is also the value the `tags` " +
					"attribute on familio_person takes, and the import ID.",
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "The tag text («Название»), familio's `tag` field. Up to " +
					strconv.Itoa(familio.TagNameMaxLen) + " characters. familio requires it to be " +
					"unique (case-insensitively) among the account's tags.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, familio.TagNameMaxLen),
				},
			},
			"color": schema.StringAttribute{
				Description: "The tag's colour. familio stores a palette **code**, not a hex value; " +
					"one of: `" + strings.Join(familio.TagColors, "`, `") + "`.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(familio.TagColors...),
				},
			},
			"description": schema.StringAttribute{
				Description: "Free-text description («Описание»). Up to " +
					strconv.Itoa(familio.TagDescriptionMaxLen) + " characters.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(familio.TagDescriptionMaxLen),
				},
			},

			"hex": schema.StringAttribute{
				Computed: true,
				Description: "The fill the familio web UI paints `color` with (e.g. `#EBFFEB`). " +
					"Derived locally from the palette, not returned by the API.",
			},
			// No Default: familio decides this, and a default would make the value
			// known-false at plan time and then mismatch the applied truth.
			"is_free": schema.BoolAttribute{
				Computed: true,
				Description: "Whether familio considers this tag usable without a Familio Plus " +
					"subscription. Server-computed; not settable.",
			},
		},
	}
}
