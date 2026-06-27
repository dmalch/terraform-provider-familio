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

  # Christening / baptism (familio's «Крещение») event.
  christening_date = {
    year  = 1850
    month = 3
    day   = 21
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

# A child linked to both parents. The parents set (0–2 person UUIDs) is stored on
# the child's birth event; order does not matter and a parent's father/mother
# role is inferred from their own gender. Parents (and the birth date) can be
# changed in place — editing them does not recreate the person.
resource "familio_person" "pyotr" {
  first_name = "Пётр"
  last_name  = "Иванов"
  patronymic = "Иванович"
  gender     = "male"

  birth_date = {
    year = 1878
  }

  parents = [
    familio_person.ivan.uuid,
    familio_person.maria.uuid,
  ]
}
