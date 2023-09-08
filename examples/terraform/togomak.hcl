togomak {
  version = 2
}

data "tf" "this" {
  source = "."
  allow_apply = true
}


stage "hello" {
  script = "echo hello world"
}

stage "delayed" {
  script = "echo Here is a random pet name: ${data.tf.this.random_pet.pet.id}"
}
