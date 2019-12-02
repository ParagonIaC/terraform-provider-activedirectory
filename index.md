---
layout: default
---
[![CircleCI](https://img.shields.io/circleci/build/github/adlerrobert/terraform-provider-activedirectory?style=for-the-badge&label=BUILDING)](https://circleci.com/gh/adlerrobert/terraform-provider-activedirectory)
[![Codecov](https://img.shields.io/codecov/c/gh/adlerrobert/terraform-provider-activedirectory?style=for-the-badge)](https://codecov.io/gh/adlerrobert/terraform-provider-activedirectory)
[![GitHub license](https://img.shields.io/github/license/adlerrobert/terraform-provider-activedirectory.svg?style=for-the-badge)](https://github.com/adlerrobert/terraform-provider-activedirectory/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/adlerrobert/terraform-provider-activedirectory.svg?style=for-the-badge)](https://GitHub.com/adlerrobert/terraform-provider-activedirectory/releases/)

The following Active Directory object types are supported:
* computer
* organizational unit

More Active Directory resources are planned. Please feel free to contribute.

For general information about Terraform, visit the [official website][1] and the [GitHub project page][3]. More information about Terraform Providers can be found on the [official provider website][2].

[1]: https://terraform.io/
[2]: https://www.terraform.io/docs/providers/index.html
[3]: https://github.com/hashicorp/terraform

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12+
- [Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)

## Developing the Provider
If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (please check the [requirements](https://github.com/adlerrobert/terraform-provider-activedirectory#requirements) before proceeding).

*Note:* This project uses [Go Modules](https://blog.golang.org/using-go-modules) making it safe to work with it outside of your existing [GOPATH](http://golang.org/doc/code.html#GOPATH). The instructions that follow assume a directory in your home directory outside of the standard GOPATH (i.e `$HOME/development/terraform-providers/`).

Clone repository to: `$HOME/development/terraform-providers/`

```sh
$ mkdir -p $HOME/development/terraform-providers/; cd $HOME/development/terraform-providers/
$ git clone git@github.com:adlerrobert/terraform-provider-activedirectory
...
```

Enter the provider directory and run `make tools`. This will install the needed tools for the provider.

```sh
$ make tools
```

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-activedirectory
...
```

## Using the Provider
To use a released provider in your Terraform environment, run [`terraform init`](https://www.terraform.io/docs/commands/init.html) and Terraform will automatically install the provider. To specify a particular provider version when installing released providers, see the [Terraform documentation on provider versioning](https://www.terraform.io/docs/configuration/providers.html#version-provider-versions).

To instead use a custom-built provider in your Terraform environment (e.g. the provider binary from the build instructions above), follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory, run `terraform init` to initialize it.

The Active Directory provider is use to interact with Microsoft Active Directory. The provider needs to be configured with the proper credentials before it can be used.

Currently the provider only supports Active Directory Computer objects.

### Example
```hcl
# Configure the AD Provider
provider "activedirectory" {
  host     = "ad.example.org"
  domain   = "example.org"
  use_tls  = true
  user     = "admin"
  password = "admin"
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
```

## Testing the Provider
In order to test the provider, you can run `make test`. This will run so-called unit tests.
```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`. Please make sure that a working Domain Controller is reachable and you have the needed permissions
*Note:* Acceptance tests create real resources! Please read [Running an Acceptance Test](https://github.com/adlerrobert/terraform-provider-axctivedirectory/blob/master/.github/CONTRIBUTING.md#running-an-acceptance-test) in the contribution guidelines for more information on usage.

```sh
$ make testacc
```

 For `make testacc` you have to set the following environment variables:

 | Variable | Description | Example | Default | Required |
 | -------- | ----------- | ------- | ------- | :------: |
 | AD_HOST | Domain Controller | dc.example.org | - | yes |
 | AD_PORT | LDAP Port - 389 TCP | 389 | 389 | no |
 | AD_DOMAIN | Domain | eample.org | - | yes |
 | AD_USE_TLS | Use secure connection | false | true | no |
 | AD_USER | Admin user DN | admin | - | yes |
 | AD_PASSWORD | Password of the admin user | secret | - | yes |
 | AD_TEST_BASE_OU | OU for the test cases | ou=Tests,dc=example,dc=org | - | yes (tests only) |

## Contributing
Terraform is the work of thousands of contributors. We appreciate your help!

To contribute, please read the contribution guidelines: [Contributing to Terraform - Active Directory Provider](CONTRIBUTING.md)

Issues on GitHub are intended to be related to bugs or feature requests with provider codebase. See https://www.terraform.io/docs/extend/community/index.html for a list of community resources to ask questions about Terraform.
