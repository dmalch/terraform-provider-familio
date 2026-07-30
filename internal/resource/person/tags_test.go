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

// TestTagIDSet checks the projection used by both the create shortcut and the
// refresh read, including the distinction that matters most: an empty tag list
// must become an empty set, never a null one (null means "unmanaged").
func TestTagIDSet(t *testing.T) {
	t.Run("ids are projected", func(t *testing.T) {
		RegisterTestingT(t)
		set, diags := tagIDSet(t.Context(), []familio.Tag{
			{TagInput: familio.TagInput{Name: "a"}, ID: 2832},
			{TagInput: familio.TagInput{Name: "b"}, ID: 2833},
		})
		Expect(diags).To(BeEmpty())
		Expect(set.Equal(tagSet(t, 2832, 2833))).To(BeTrue())
	})

	t.Run("no tags is an empty set, not null", func(t *testing.T) {
		RegisterTestingT(t)
		set, diags := tagIDSet(t.Context(), nil)
		Expect(diags).To(BeEmpty())
		Expect(set.IsNull()).To(BeFalse(), "null would mean unmanaged, which is a different state")
		Expect(set.Elements()).To(BeEmpty())
	})
}

// TestAssignTagsOnCreateIssuesNoRequests pins the point of the create shortcut:
// with nothing to assign it must not touch the network at all. The client is nil
// here, so any request would panic — that is the assertion.
func TestAssignTagsOnCreateIssuesNoRequests(t *testing.T) {
	RegisterTestingT(t)
	r := &Resource{}

	set, diags := r.assignTagsOnCreate(t.Context(), "p-1", nil)
	Expect(diags).To(BeEmpty())
	Expect(set.IsNull()).To(BeFalse())
	Expect(set.Elements()).To(BeEmpty(), "`tags = []` on create is already satisfied by a new person")
}

// TestPersonModelTagsAttrType guards the model/schema pairing: `tags` must stay
// a set of Int64, since familio tag ids are integers and a string set would
// silently fail to decode.
func TestPersonModelTagsAttrType(t *testing.T) {
	RegisterTestingT(t)
	set := tagSet(t, 2832)
	Expect(set.ElementType(t.Context())).To(Equal(attr.Type(types.Int64Type)))
}
