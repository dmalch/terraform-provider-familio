# A tag («метка») in the account's tag catalogue — one entry of the
# /profile/my-tags page. Tags are coloured labels used to group persons; they
# exist independently of any person and are attached via the `tags` attribute on
# familio_person.
#
# NOTE: tags are a Familio Plus feature. Without a subscription only one tag is
# usable — the one familio flags as free (`is_free`).

resource "familio_tag" "to_verify" {
  name        = "Проверить в архиве"
  color       = "mint-mist"
  description = "Нужен запрос в ЦГА — метрики не найдены онлайн."
}

resource "familio_tag" "dispossessed" {
  name  = "Раскулачены"
  color = "rose-mist"
}

# Attach tags to a person by id. The `tags` attribute is authoritative: the
# provider makes familio match it exactly. Omit it to leave a person's tags
# unmanaged; use [] to remove them all.
resource "familio_person" "ivan" {
  first_name = "Иван"
  last_name  = "Иванов"
  gender     = "male"

  birth {
    date = { year = 1890 }
  }

  tags = [
    familio_tag.to_verify.id,
    familio_tag.dispossessed.id,
  ]
}

# familio stores a palette code, not a hex. `hex` exposes the fill the web UI
# paints the tag with, for configs that render tags themselves.
output "to_verify_hex" {
  value = familio_tag.to_verify.hex # -> "#EBFFEB"
}
