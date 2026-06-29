package tfdate

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	. "github.com/onsi/gomega"

	"github.com/dmalch/go-familio"
)

func validateRange(t *testing.T, r *familio.DateRange) bool {
	t.Helper()
	req := validator.ObjectRequest{
		Path:        path.Root("date"),
		ConfigValue: ObjectFromRange(r),
	}
	resp := &validator.ObjectResponse{}
	dateRangeValidator{}.ValidateObject(context.Background(), req, resp)
	return resp.Diagnostics.HasError()
}

func TestDateRangeValidator(t *testing.T) {
	y := func(v int) *int { return &v }

	cases := []struct {
		name      string
		in        *familio.DateRange
		wantError bool
	}{
		{"exact", &familio.DateRange{Year: 1846}, false},
		{"circa alone", &familio.DateRange{Year: 1846, Circa: true}, false},
		{"before bound", &familio.DateRange{Year: 1910, Range: familio.RangeBefore}, false},
		{"between with end_year", &familio.DateRange{Year: 1846, Range: familio.RangeBetween, EndYear: y(1850)}, false},
		{"julian exact", &familio.DateRange{Year: 1846, Calendar: "julian"}, false},

		{"circa with range", &familio.DateRange{Year: 1846, Circa: true, Range: familio.RangeBetween, EndYear: y(1850)}, true},
		{"end_circa with range", &familio.DateRange{Year: 1846, EndCirca: true, Range: familio.RangeBetween, EndYear: y(1850)}, true},
		{"between without end_year", &familio.DateRange{Year: 1846, Range: familio.RangeBetween}, true},
		{"before with end_year", &familio.DateRange{Year: 1910, Range: familio.RangeBefore, EndYear: y(1920)}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			RegisterTestingT(t)
			Expect(validateRange(t, tc.in)).To(Equal(tc.wantError))
		})
	}
}
