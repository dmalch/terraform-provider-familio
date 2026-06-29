package tfdate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/dmalch/go-familio"
)

// dateRangeValidator enforces familio complex-date rules a static schema can't:
// approximation (circa/end_circa) is a whole-date "about" type and cannot combine
// with a range; a "between" range needs end_year; before/after are single bounds
// that take no second date. Rejecting these is deliberate — silently dropping an
// approximation or a second endpoint would publish a subtly wrong date.
type dateRangeValidator struct{}

func (dateRangeValidator) Description(context.Context) string {
	return "validates familio complex-date field combinations (circa vs range, between endpoints)"
}

func (v dateRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (dateRangeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var m model
	resp.Diagnostics.Append(req.ConfigValue.As(ctx, &m, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	rangeKind := m.Range.ValueString()
	approximate := m.Circa.ValueBool() || m.EndCirca.ValueBool()

	if rangeKind != "" && approximate {
		resp.Diagnostics.AddAttributeError(req.Path, "Approximate range not supported",
			"familio cannot mark a range (before/after/between) as approximate. Use circa for "+
				"a single \"about\" date, or a range, but not both (and not end_circa).")
	}

	hasEnd := !m.EndYear.IsNull() || !m.EndMonth.IsNull() || !m.EndDay.IsNull()
	switch rangeKind {
	case familio.RangeBetween:
		if m.EndYear.IsNull() || m.EndYear.IsUnknown() {
			resp.Diagnostics.AddAttributeError(req.Path, "Missing end_year",
				"range = \"between\" needs end_year (the second endpoint).")
		}
	case familio.RangeBefore, familio.RangeAfter:
		if hasEnd {
			resp.Diagnostics.AddAttributeError(req.Path, "Unexpected end date",
				"range = \""+rangeKind+"\" is a single bound; end_year/end_month/end_day are only "+
					"valid with range = \"between\".")
		}
	}
}
