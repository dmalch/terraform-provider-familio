package person

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	. "github.com/onsi/gomega"
)

// personConfig builds a tfsdk.Config from the resource schema with every
// attribute null except the names provided.
func personConfig(t *testing.T, first, last *string) tfsdk.Config {
	t.Helper()
	ctx := context.Background()
	r := &Resource{}
	sresp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, sresp)

	objType, ok := sresp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("person schema type is not a tftypes.Object")
	}
	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, typ := range objType.AttributeTypes {
		vals[name] = tftypes.NewValue(typ, nil)
	}
	if first != nil {
		vals["first_name"] = tftypes.NewValue(tftypes.String, *first)
	}
	if last != nil {
		vals["last_name"] = tftypes.NewValue(tftypes.String, *last)
	}
	return tfsdk.Config{Schema: sresp.Schema, Raw: tftypes.NewValue(objType, vals)}
}

func TestPersonValidateConfig(t *testing.T) {
	ctx := context.Background()
	r := &Resource{}
	name := "Иван"

	t.Run("no name errors", func(t *testing.T) {
		RegisterTestingT(t)
		resp := &resource.ValidateConfigResponse{}
		r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: personConfig(t, nil, nil)}, resp)
		Expect(resp.Diagnostics.HasError()).To(BeTrue(), "expected an error when neither first_name nor last_name is set")
	})

	t.Run("first name ok", func(t *testing.T) {
		RegisterTestingT(t)
		resp := &resource.ValidateConfigResponse{}
		r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: personConfig(t, &name, nil)}, resp)
		Expect(resp.Diagnostics.HasError()).To(BeFalse())
	})

	t.Run("last name ok", func(t *testing.T) {
		RegisterTestingT(t)
		resp := &resource.ValidateConfigResponse{}
		r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: personConfig(t, nil, &name)}, resp)
		Expect(resp.Diagnostics.HasError()).To(BeFalse())
	})
}
