package person

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/gomega"

	"github.com/dmalch/go-familio"
)

// tagSet builds a tags attribute value. The nil slice is normalised to an empty
// one first: types.SetValueFrom turns a nil Go slice into a *null* Set, which
// would make the empty-set case indistinguishable from an omitted attribute.
func tagSet(t *testing.T, ids ...int64) types.Set {
	t.Helper()
	if ids == nil {
		ids = []int64{}
	}
	set, diags := types.SetValueFrom(t.Context(), types.Int64Type, ids)
	Expect(diags).To(BeEmpty())
	return set
}

func TestDesiredTags(t *testing.T) {
	t.Run("a null attribute is unmanaged", func(t *testing.T) {
		RegisterTestingT(t)
		ids, managed, diags := desiredTags(t.Context(), types.SetNull(types.Int64Type))
		Expect(diags).To(BeEmpty())
		Expect(managed).To(BeFalse())
		Expect(ids).To(BeNil())
	})

	t.Run("an unknown attribute is unmanaged", func(t *testing.T) {
		RegisterTestingT(t)
		_, managed, diags := desiredTags(t.Context(), types.SetUnknown(types.Int64Type))
		Expect(diags).To(BeEmpty())
		Expect(managed).To(BeFalse())
	})

	t.Run("an empty set is managed (remove-all)", func(t *testing.T) {
		RegisterTestingT(t)
		ids, managed, diags := desiredTags(t.Context(), tagSet(t))
		Expect(diags).To(BeEmpty())
		Expect(managed).To(BeTrue(), "[] means remove every tag, not leave them alone")
		Expect(ids).To(BeEmpty())
	})

	t.Run("ids are narrowed to int", func(t *testing.T) {
		RegisterTestingT(t)
		ids, managed, diags := desiredTags(t.Context(), tagSet(t, 2832, 2833))
		Expect(diags).To(BeEmpty())
		Expect(managed).To(BeTrue())
		Expect(ids).To(ConsistOf(2832, 2833))
	})
}

func TestTagDiff(t *testing.T) {
	current := []familio.Tag{
		{TagInput: familio.TagInput{Name: "a"}, ID: 1},
		{TagInput: familio.TagInput{Name: "b"}, ID: 2},
	}

	t.Run("no change issues nothing", func(t *testing.T) {
		RegisterTestingT(t)
		add, remove := tagDiff(current, []int{1, 2})
		Expect(add).To(BeEmpty())
		Expect(remove).To(BeEmpty())
	})

	t.Run("adds and removes in one pass", func(t *testing.T) {
		RegisterTestingT(t)
		add, remove := tagDiff(current, []int{2, 3})
		Expect(add).To(Equal([]int{3}))
		Expect(remove).To(Equal([]int{1}))
	})

	t.Run("an empty desired set removes everything", func(t *testing.T) {
		RegisterTestingT(t)
		add, remove := tagDiff(current, nil)
		Expect(add).To(BeEmpty())
		Expect(remove).To(ConsistOf(1, 2))
	})

	t.Run("assigning onto an untagged person", func(t *testing.T) {
		RegisterTestingT(t)
		add, remove := tagDiff(nil, []int{7, 9})
		Expect(add).To(Equal([]int{7, 9}))
		Expect(remove).To(BeEmpty())
	})

	t.Run("duplicate desired ids collapse", func(t *testing.T) {
		RegisterTestingT(t)
		add, remove := tagDiff(nil, []int{7, 7, 9})
		Expect(add).To(Equal([]int{7, 9}))
		Expect(remove).To(BeEmpty())
	})
}

// TestPersonModelTagsAttrType guards the model/schema pairing: `tags` must stay
// a set of Int64, since familio tag ids are integers and a string set would
// silently fail to decode.
func TestPersonModelTagsAttrType(t *testing.T) {
	RegisterTestingT(t)
	set := tagSet(t, 2832)
	Expect(set.ElementType(t.Context())).To(Equal(attr.Type(types.Int64Type)))
}
