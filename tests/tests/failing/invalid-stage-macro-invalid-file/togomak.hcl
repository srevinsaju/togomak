togomak {
  version = 2
}
macro "pen_pen" {
  source = "pen_pen.hcl"
}
stage "example" {
  use {
    macro = macro.pen_pen
  }
}
