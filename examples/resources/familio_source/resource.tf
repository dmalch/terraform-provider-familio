# A source citation («Источник») attached to a person. A source references a
# catalogued entity by UUID; the reference is immutable (changing it replaces the
# source) while the comment edits in place.
#
# NOTE: a person's sources can be managed EITHER through standalone familio_source
# resources (below) OR through the `sources` block on familio_person — never both
# for the same person.

# An archival document (a digitised дело from the organization → fund → register →
# case catalog). type = "case"; no catalog_key.
resource "familio_source" "ivan_revision" {
  person         = familio_person.ivan.uuid
  reference_uuid = "58e68fa4-9e58-4f11-84bd-510a2dc015eb" # the archive case (дело) UUID
  type           = "case"
  comment        = "Ревизская сказка 1811 г."
}

# A record from a people index (e.g. the «Памяти героев Великой войны» project).
# type = "catalog_person"; catalog_key names the source catalog the record is in.
resource "familio_source" "ivan_ww1" {
  person         = familio_person.ivan.uuid
  reference_uuid = "0123e5fb-e298-46e7-8779-a9bfa793ca5a"
  type           = "catalog_person"
  catalog_key    = "gwarmil"
}
