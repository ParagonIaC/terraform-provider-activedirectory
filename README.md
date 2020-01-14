# Terraform Provider - Active Directory

[![Go Report Card](https://goreportcard.com/badge/github.com/ParagonIaC/terraform-provider-activedirectory?style=flat-square&label=building)](https://goreportcard.com/report/github.com/ParagonIaC/terraform-provider-activedirectory)
[![CircleCI](https://img.shields.io/circleci/build/github/ParagonIaC/terraform-provider-activedirectory?style=flat-square&label=building)](https://circleci.com/gh/ParagonIaC/terraform-provider-activedirectory)
[![Codecov](https://img.shields.io/codecov/c/gh/ParagonIaC/terraform-provider-activedirectory?style=flat-square)](https://codecov.io/gh/ParagonIaC/terraform-provider-activedirectory)
[![GitHub license](https://img.shields.io/github/license/ParagonIaC/terraform-provider-activedirectory.svg?style=flat-square)](https://github.com/ParagonIaC/terraform-provider-activedirectory/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/ParagonIaC/terraform-provider-activedirectory.svg?style=flat-square)](https://GitHub.com/ParagonIaC/terraform-provider-activedirectory/releases/)

<img alt="HashiCorp Terraform" src="/docs/assets/img/terraform.png" width="15%"><img alt="Microsoft Active Directory" src="/docs/assets/img/active-directory.png" width="15%">

This is a Terraform  Provider to work with Active Directory.

This provider currently supports only computer objects, but more active directory resources are planned. Please feel free to contribute.

For general information about Terraform, visit the [official website][3] and the [GitHub project page][4].

[3]: https://terraform.io/
[4]: https://github.com/hashicorp/terraform

More information can be found on the Active Directory Provider [GitHub pages](https://ParagonIaC.github.io/terraform-provider-activedirectory/).

## Simple Usage Example
```hcl
# Configure the AD Provider
provider "activedirectory" {
  host     = "ad.example.org"
  domain   = "example.org"
  use_tls  = false
  user     = "admin"
  password = "password"
}

# Add computer to Active Directory
resource "activedirectory_computer" "test_computer" {
  name           = "TerraformComputer"                      # update will force destroy and new
  ou             = "CN=Computers,DC=example,DC=org"         # can be updated
  description    = "terraform sample server"                # can be updated
}

# Add ou to Active Directory
resource "activedirectory_ou" "test_ou" {
  name           = "TerraformOU"                            # can be updated
  base_ou        = "OU=Test,CN=Computers,DC=example,DC=org" # can be updated
  description    = "terraform sample ou"                    # can be updated
}

# Add group to Active Directory
resource "activedirectory_group" "test_group" {
  name           = "TerraformGroup"                         # can be updated
  base_ou        = activedirectory_ou.test_ou.dn            # can be updated
  description    = "terraform sample group"                 # can be updated
  user_base = "CN=Users,DC=example,DC=org"                  # can be updated, where to search users, optional
                                                            # if not set default is set on AD domain.(in this example 'DC=example,DC=org' )
  member    = [ somebody ]                                  # can be updated, sAMAaccount as user/group id
  ignore_members_unknown_by_terraform = false               # can be updated, not remove user unknown by terraform
}
# Add group to Active Directory
resource "activedirectory_group" "test_group_two" {
  name           = "TerraformGroupTwo"                      # can be updated
  base_ou        = activedirectory_ou.test_ou.dn            # can be updated
  description    = "terraform sample group two"             # can be updated
  member    = [ activedirectory_group.test_group.name ]     # can be updated, group can be also member
  ignore_members_unknown_by_terraform = false               # can be updated, not remove user unknown by terraform
}
```