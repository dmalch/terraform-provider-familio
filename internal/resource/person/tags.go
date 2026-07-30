package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/go-familio"
)

// tagsAttribute is the person's «метки» as an authoritative set of tag ids.
// When set, the provider makes the person's tag set exactly match it (assigning
// the missing, unassigning the extra); when omitted (null) the provider does not
// manage the person's tags at all. An empty set means "remove all tags".
//
// The elements are integers, not uuids: familio keys tags by a small sequential
// id. Reference them as familio_tag.<name>.id, or look an existing tag up with
// the familio_tags data source.
func tagsAttribute() schema.SetAttribute {
	return schema.SetAttribute{
		Description: "Tag ids («метки») attached to this person, managed as an authoritative set: " +
			"the provider makes familio match it exactly. Omit the attribute to leave the person's " +
			"tags unmanaged; use `[]` to remove them all. Elements are the **integer** ids of " +
			"familio_tag resources (`familio_tag.x.id`) or of existing tags found via the " +
			"familio_tags data source. Attaching a tag never deletes it — removing an id here only " +
			"unassigns it from this person.",
		Optional:    true,
		ElementType: types.Int64Type,
		Validators: []validator.Set{
			setvalidator.ValueInt64sAre(int64validator.AtLeast(1)),
		},
	}
}

// desiredTags decodes the tags attribute into tag ids. A null/unknown attribute
// (the person's tags are unmanaged) yields nil, false.
func desiredTags(ctx context.Context, set types.Set) (ids []int, managed bool, diags diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, false, diags
	}
	var raw []int64
	diags = set.ElementsAs(ctx, &raw, false)
	if diags.HasError() {
		return nil, true, diags
	}
	ids = make([]int, 0, len(raw))
	for _, id := range raw {
		ids = append(ids, int(id))
	}
	return ids, true, diags
}

// tagDiff computes the two batches needed to move a person from its current
// tags to the desired id set: which ids to assign and which to unassign. Both
// preserve their input order, and a no-op reconcile yields two empty slices so
// the caller issues no request at all. Duplicate desired ids collapse, since
// familio treats a re-assignment as a no-op anyway.
func tagDiff(current []familio.Tag, desired []int) (add, remove []int) {
	have := make(map[int]bool, len(current))
	for _, t := range current {
		have[t.ID] = true
	}
	want := make(map[int]bool, len(desired))

	for _, id := range desired {
		if want[id] {
			continue
		}
		want[id] = true
		if !have[id] {
			add = append(add, id)
		}
	}
	for _, t := range current {
		if !want[t.ID] {
			remove = append(remove, t.ID)
		}
	}
	return add, remove
}

// writeTags makes the person's familio tag set match the desired ids
// (authoritative): assign the missing, unassign the extra. Both calls take
// batches, so this is at most two requests plus the current-state read.
func (r *Resource) writeTags(ctx context.Context, uuid string, desired []int) diag.Diagnostics {
	var diags diag.Diagnostics

	current, err := r.client.GetPersonTags(ctx, uuid)
	if err != nil {
		diags.AddError("Cannot read familio_person tags before reconciling", err.Error())
		return diags
	}

	add, remove := tagDiff(current, desired)

	if len(add) > 0 {
		if _, err := r.client.AssignPersonTags(ctx, uuid, add); err != nil {
			diags.AddError("Cannot assign familio_person tags", err.Error())
			return diags
		}
	}
	if len(remove) > 0 {
		if err := r.client.UnassignPersonTags(ctx, uuid, remove); err != nil {
			diags.AddError("Cannot unassign familio_person tags", err.Error())
			return diags
		}
	}
	return diags
}

// readTags rebuilds the tags set from familio. A null prior attribute stays null
// (unmanaged) and no read is performed. The result is a set, so element order
// does not matter to Terraform and no prior-order reconstruction is needed —
// unlike the sources list.
func (r *Resource) readTags(ctx context.Context, uuid string, prior types.Set) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	if prior.IsNull() || prior.IsUnknown() {
		return types.SetNull(types.Int64Type), diags
	}

	tags, err := r.client.GetPersonTags(ctx, uuid)
	if err != nil {
		diags.AddError("Error reading familio_person tags", err.Error())
		return types.SetNull(types.Int64Type), diags
	}

	ids := make([]int64, 0, len(tags))
	for _, t := range tags {
		ids = append(ids, int64(t.ID))
	}

	set, d := types.SetValueFrom(ctx, types.Int64Type, ids)
	diags.Append(d...)
	return set, diags
}
