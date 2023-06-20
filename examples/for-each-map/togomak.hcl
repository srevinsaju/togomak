togomak {
  version = 1
}

locals {
  months = {
    jan = "January"
    feb = "February"
    mar = "March"
    apr = "April"
    may = "May"
    jun = "June"
    jul = "July"
    aug = "August"
    sep = "September"
    oct = "October"
    nov = "November"
    dec = "December"
  }
}

stage "months" {
  for_each = local.months
  script = "echo ${each.key} is ${each.value}"
}
