# Crawl the connected persons around a root, bounded to one surname, so a module
# can harvest the UUIDs to import against without an out-of-band BFS.
data "familio_tree" "vereya" {
  root      = "85781e3b-0000-0000-0000-000000000000"
  direction = "component" # up | down | component (default)
  surname   = "Персонов"  # do not expand through married-in branches
  depth     = 4           # 0 = unlimited
}

# All discovered person UUIDs (feed terraform import or other resources).
output "person_uuids" {
  value = [for n in data.familio_tree.vereya.nodes : n.uuid]
}

# Every marriage in the component, as familio_marriage import ids
# ("<person_uuid>:<marriage_uuid>"). De-dup on marriage_uuid, since both spouses
# report the same union.
output "marriage_import_ids" {
  value = {
    for m in distinct(flatten([
      for n in data.familio_tree.vereya.nodes : [
        for s in n.spouses : {
          id            = "${n.uuid}:${s.marriage_uuid}"
          marriage_uuid = s.marriage_uuid
        }
      ]
    ])) : m.marriage_uuid => m.id
  }
}
