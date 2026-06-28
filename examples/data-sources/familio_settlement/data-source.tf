# Look up a settlement (place) by UUID — the same UUID that a person's
# birth/death/christening place and familio_source reference.
data "familio_settlement" "birthplace" {
  uuid = "40d1b180-b739-4ecb-9ee5-ced6fefcd0d8"
}

# Resolve a UUID to a human-readable place, e.g. to document or label a config.
output "birthplace_name" {
  # -> "Нижняя Верея, город Выкса, Нижегородская область"
  value = join(", ", compact([
    data.familio_settlement.birthplace.name,
    data.familio_settlement.birthplace.district,
    data.familio_settlement.birthplace.region,
  ]))
}

output "birthplace_coords" {
  value = {
    lat = data.familio_settlement.birthplace.latitude
    lon = data.familio_settlement.birthplace.longitude
  }
}
