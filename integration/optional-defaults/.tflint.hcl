plugin "terraform" {
  enabled = false
}

plugin "redeploy" {
  enabled = true
}

rule "terraform_standard_module_structure" {
  enabled = false
}
