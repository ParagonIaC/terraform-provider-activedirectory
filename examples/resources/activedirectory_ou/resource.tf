resource "activedirectory_ou" "test_ou_computers" {
  name           = "TerraformComputerOU"                    # can be updated
  base_ou        = "OU=Test,DC=example,DC=org"              # can be updated
  description    = "terraform sample ou"                    # can be updated
}

resource "activedirectory_ou" "test_ou_groups" {
  name           = "TerraformGroupOU"                       # can be updated
  base_ou        = "OU=Test,DC=example,DC=org"              # can be updated
  description    = "terraform sample ou"                    # can be updated
}

resource "activedirectory_ou" "test_ou_users" {
  name           = "TerraformUsersOU"                       # can be updated
  base_ou        = "OU=Test,DC=example,DC=org"              # can be updated
  description    = "terraform sample ou"                    # can be updated
}