# The account's tag catalogue («Мои метки»). Use it to reference tags that
# already exist on familio without importing them as familio_tag resources:
# familio keys tags by an opaque integer, so `by_name` is how a config names one.
data "familio_tags" "all" {}

# Attach an existing tag to a person by name rather than by hardcoding its id.
resource "familio_person" "ivan" {
  first_name = "Иван"
  last_name  = "Иванов"
  gender     = "male"

  birth {
    date = { year = 1890 }
  }

  tags = [data.familio_tags.all.by_name["Проверить в архиве"]]
}

# Everything familio knows about each tag, for auditing the catalogue.
output "tag_catalogue" {
  value = {
    for t in data.familio_tags.all.tags : t.name => {
      id      = t.id
      color   = t.color
      hex     = t.hex
      is_free = t.is_free
    }
  }
}

# On a free account only one tag is usable — the one flagged free.
output "usable_without_plus" {
  value = [for t in data.familio_tags.all.tags : t.name if t.is_free]
}
