# A marriage between two existing persons (a "wedding" event under the hood).
resource "familio_marriage" "marriage" {
  partners = [familio_person.ivan.uuid, familio_person.maria.uuid]

  marriage_date = {
    year  = 1875
    month = 5
    day   = 12
  }

  # Wedding place — a familio settlement UUID (the same id a person's birth
  # place takes), typically the parish where the marriage was recorded.
  place = "40d1b180-b739-4ecb-9ee5-ced6fefcd0d8" # Нижняя Верея

  comment = "Венчание в Спасо-Преображенской церкви."
}
