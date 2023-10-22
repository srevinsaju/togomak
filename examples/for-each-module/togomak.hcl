togomak {
  version = 2
}


module "parallel" {
  for_each = toset(["alpha", "beta", "gamma"])
  source = "github.com/srevinsaju/togomak-first-module"
}
