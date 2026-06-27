# A residence (familio "location") event on a person, spanning a date range.
resource "familio_event" "ivan_residence" {
  person   = familio_person.ivan.uuid
  type     = "location"
  date     = { year = 1878 }
  end_date = { year = 1890 }
  comment  = "Москва, Тверская улица"
}

# An occupation, at a point in time.
resource "familio_event" "ivan_job" {
  person  = familio_person.ivan.uuid
  type    = "profession"
  date    = { year = 1882, month = 9 }
  comment = "Кузнец"
}

# Military service.
resource "familio_event" "ivan_army" {
  person   = familio_person.ivan.uuid
  type     = "militaryService"
  date     = { year = 1900 }
  end_date = { year = 1903 }
}
