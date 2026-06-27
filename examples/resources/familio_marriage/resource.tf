# A marriage between two existing persons (a "wedding" event under the hood).
resource "familio_marriage" "marriage" {
  partners = [familio_person.ivan.uuid, familio_person.maria.uuid]

  marriage_date = {
    year  = 1875
    month = 5
    day   = 12
  }

  comment = "Венчание в Спасо-Преображенской церкви."
}
