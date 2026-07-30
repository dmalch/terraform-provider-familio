package tag

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/gomega"

	"github.com/dmalch/go-familio"
)

func TestInputFromModel(t *testing.T) {
	RegisterTestingT(t)

	in := inputFromModel(&ResourceModel{
		Name:        types.StringValue("Проверить в архиве"),
		Color:       types.StringValue(familio.TagColorMintMist),
		Description: types.StringValue("Нужен запрос в ЦГА"),
	})
	Expect(in).To(Equal(familio.TagInput{
		Name:        "Проверить в архиве",
		Color:       familio.TagColorMintMist,
		Description: "Нужен запрос в ЦГА",
	}))

	// familio has no null description; an omitted one is sent as "".
	in = inputFromModel(&ResourceModel{
		Name:        types.StringValue("Раскулачены"),
		Color:       types.StringValue(familio.TagColorRoseMist),
		Description: types.StringNull(),
	})
	Expect(in.Description).To(BeEmpty())

	// Whatever the model holds must pass the client's own validation, so the
	// two layers cannot disagree about what is acceptable.
	Expect(in.Validate()).To(Succeed())
}

func TestApplyTagToState(t *testing.T) {
	RegisterTestingT(t)

	m := &ResourceModel{}
	applyTagToState(&familio.Tag{
		TagInput: familio.TagInput{
			Name:        "Проверить в архиве",
			Color:       familio.TagColorMintMist,
			Description: "Нужен запрос в ЦГА",
		},
		ID:     2832,
		IsFree: true,
	}, m)

	Expect(m.ID.ValueInt64()).To(Equal(int64(2832)), "the familio id is an integer, not a uuid")
	Expect(m.Name.ValueString()).To(Equal("Проверить в архиве"))
	Expect(m.Color.ValueString()).To(Equal(familio.TagColorMintMist))
	Expect(m.Description.ValueString()).To(Equal("Нужен запрос в ЦГА"))
	Expect(m.Hex.ValueString()).To(Equal("#EBFFEB"), "hex is derived locally from the palette")
	Expect(m.IsFree.ValueBool()).To(BeTrue())
}

func TestApplyTagToStateEmptyDescription(t *testing.T) {
	RegisterTestingT(t)

	// A server-empty description must read back as null, or an omitted optional
	// value would permadiff against "".
	m := &ResourceModel{Description: types.StringNull()}
	applyTagToState(&familio.Tag{
		TagInput: familio.TagInput{Name: "Раскулачены", Color: familio.TagColorRoseMist},
		ID:       2833,
	}, m)

	Expect(m.Description.IsNull()).To(BeTrue())
	Expect(m.IsFree.ValueBool()).To(BeFalse())
}

func TestApplyTagToStateUnknownColour(t *testing.T) {
	RegisterTestingT(t)

	// If familio ever adds a palette entry the provider doesn't know, the tag
	// must still round-trip — only the derived hex goes empty.
	m := &ResourceModel{}
	applyTagToState(&familio.Tag{
		TagInput: familio.TagInput{Name: "Новая", Color: "chartreuse-dream"},
		ID:       1,
	}, m)

	Expect(m.Color.ValueString()).To(Equal("chartreuse-dream"))
	Expect(m.Hex.ValueString()).To(BeEmpty())
}

func TestFindByID(t *testing.T) {
	RegisterTestingT(t)

	catalogue := []familio.Tag{
		{TagInput: familio.TagInput{Name: "a"}, ID: 1},
		{TagInput: familio.TagInput{Name: "b"}, ID: 2832},
	}
	Expect(findByID(catalogue, 2832).Name).To(Equal("b"))
	Expect(findByID(catalogue, 1).Name).To(Equal("a"))
	Expect(findByID(catalogue, 99)).To(BeNil(), "a deleted tag reads as gone, not as an error")
	Expect(findByID(nil, 1)).To(BeNil())
}
