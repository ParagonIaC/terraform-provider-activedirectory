resource "activedirectory_group" "test_group" {
  name = "TerraformGroup"                                    # can be update
  base_ou = "OU=TerraformGroupOU,OU=Test,DC=example,DC=org"  # can be update
  user_base = "OU=Users,DC=example,DC=org"                   # can be update
  description = "Some description"                           # can be update
  member = [                                                 # can be update users by sAMAccountName
    "testuser",
    "othertestuser"
  ]
  ignore_members_unknown_by_terraform = false                # can be update ( if false remove from group users unknown by terraform )
}