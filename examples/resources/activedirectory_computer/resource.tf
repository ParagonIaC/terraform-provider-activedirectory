resource "activedirectory_computer" "test_computer" {
  name           = "TerraformComputer"                                  # update will force destroy and new
  ou             = "OU=TerraformComputerOU,OU=Test,DC=example,DC=org"   # can be updated
  description    = "terraform sample server"                            # can be updated
}