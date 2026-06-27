# Look up a single person by UUID: its owning account and relationships.
data "familio_person" "ancestor" {
  uuid = "85781e3b-0000-0000-0000-000000000000"
}

# Parent UUIDs you can adopt into your tree with `terraform import`.
output "ancestor_parents" {
  value = data.familio_person.ancestor.parents
}

# Tell your own tree from other researchers' profiles by owner.
output "is_mine" {
  value = data.familio_person.ancestor.owner_id == "894dc7d5"
}
