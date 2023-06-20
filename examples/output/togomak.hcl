togomak {
  version = 1
}

stage "agent" {
  script = <<-EOT
  set -u
  echo "AGENT=Ryoji Kaji" >> $TOGOMAK_OUTPUTS
  EOT
}

stage "seele" {
  depends_on = [stage.agent]
  name   = "seele"
  script = "echo The agent from Seele reporting! ${output.AGENT}"
}
