# A person in your familio.org family tree.
resource "familio_person" "ivan" {
  first_name = "Иван"
  last_name  = "Иванов"
  patronymic = "Петрович"
  gender     = "male"
  privacy    = "visible_for_all"

  birth_date = {
    year  = 1850
    month = 3
    day   = 14
  }

  death_date = {
    year = 1911
  }
}

# A person with only a year of birth and a maiden surname.
resource "familio_person" "maria" {
  first_name      = "Мария"
  last_name       = "Иванова"
  birth_last_name = "Петрова"
  patronymic      = "Сергеевна"
  gender          = "female"

  birth_date = {
    year = 1855
  }
}
