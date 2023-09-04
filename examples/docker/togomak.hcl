togomak {
  version = 2
}

stage "example" {
  container {
    image = "ubuntu"
    volume {
      source      = "${cwd}/diary"
      destination = "/newdiary"
    }
  }
  script = <<-EOT
  #!/usr/bin/env bash
  ls -al
  for i in $(seq 1 10); do
    sleep 1
    echo "Loading $i..."
  done
  cat rei.diary.txt
  ls -al /newdiary
  EOT
}
