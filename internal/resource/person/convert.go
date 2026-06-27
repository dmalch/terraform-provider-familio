package person

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

// applyToState copies the computed fields returned by familio.org onto the
// model. Config-only fields (first/last/patronymic/gender/parents) are left
// untouched — the public read endpoint does not return them.
func applyToState(p *familio.Person, m *ResourceModel) {
	m.UUID = types.StringValue(p.UUID)
	m.DisplayName = types.StringValue(p.DisplayName)
	m.ShortDisplayName = types.StringValue(p.ShortDisplayName)
	m.BirthDate = flexDateValue(p.BirthDate)
	m.DeathDate = flexDateValue(p.DeathDate)
	m.HasDeathEvent = types.BoolValue(p.HasDeathEvent)
	m.CatalogKey = stringPtrValue(p.CatalogKey)
	m.CatalogName = types.StringValue(p.CatalogName)
	m.Type = types.StringValue(p.Type)
	m.UpdatedAt = types.StringValue(p.UpdatedAt)
}

// inputFromModel builds the create/update request body from the plan.
func inputFromModel(m *ResourceModel) familio.PersonInput {
	return familio.PersonInput{
		FirstName:       m.FirstName.ValueString(),
		LastName:        m.LastName.ValueString(),
		Patronymic:      m.Patronymic.ValueString(),
		Gender:          m.Gender.ValueString(),
		BirthSettlement: m.BirthSettlement.ValueString(),
		FatherUUID:      m.FatherUUID.ValueString(),
		MotherUUID:      m.MotherUUID.ValueString(),
	}
}

func stringPtrValue(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func flexDateValue(d familio.FlexDate) types.String {
	if v, ok := d.Value(); ok {
		return types.StringValue(v)
	}
	return types.StringNull()
}
