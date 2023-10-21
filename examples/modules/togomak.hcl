togomak {
  version = 2
}


module "something" {
  source = "./module"
}


stage "example" {
  script = "echo hello world from root"
}
