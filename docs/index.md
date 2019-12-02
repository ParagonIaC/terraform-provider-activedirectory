---
layout: default
---
[![CircleCI](https://img.shields.io/circleci/build/github/adlerrobert/terraform-provider-activedirectory?style=for-the-badge&label=BUILDING)](https://circleci.com/gh/adlerrobert/terraform-provider-activedirectory)
[![Codecov](https://img.shields.io/codecov/c/gh/adlerrobert/terraform-provider-activedirectory?style=for-the-badge)](https://codecov.io/gh/adlerrobert/terraform-provider-activedirectory)
[![GitHub license](https://img.shields.io/github/license/adlerrobert/terraform-provider-activedirectory.svg?style=for-the-badge)](https://github.com/adlerrobert/terraform-provider-activedirectory/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/adlerrobert/terraform-provider-activedirectory.svg?style=for-the-badge)](https://GitHub.com/adlerrobert/terraform-provider-activedirectory/releases/)

<div markdown="1" style="float: right; display: flex; flex-flow: column wrap; width: 8rem;">

![HashiCorp Terraform](/terraform-provider-activedirectory/assets/img/terraform.png "HashiCorp Terraform"){:width="100%"}
![Microsoft Active Directory](/terraform-provider-activedirectory/assets/img/active-directory.png "Microsoft Active Directory"){:width="100%"}

</div>

# Table of Content

* [Overview](#overview)
* [Using the Provider](#using-the-provider)
* [Examples](#examples)
* [Provider Development](#provider-development)
  * [Requirements](#requirements)
  * [Environment](#environment)
  * [Testing the Provider](#testing-the-provider)
* [Contributing](#contributing)

# Overview

This is a community-driven Terraform provider for Microsoft Active Directory. The following Active Directory object types are currently supported:
* computer
* organizational unit

More Active Directory resources are planned. Please feel free to contribute.

<sup>[back to top](#top)</sup>

# Using the Provider

To instead use a custom-built provider in your Terraform environment (e.g. the provider binary from the build instructions below), follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory, run `terraform init` to initialize it.

The Terraform Active Directory Provider is used to interact with Microsoft Active Directory. Thus, the provider needs to be configured with the proper credentials before it can be used.

<sup>[back to top](#top)</sup>

# Example
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
```
<sup>[back to top](#top)</sup>

# Provider Development

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (please check the [requirements](#requirements) before proceeding).

_**Note:**_ This project uses [Go Modules](https://blog.golang.org/using-go-modules) making it safe to work with it outside of your existing [GOPATH](http://golang.org/doc/code.html#GOPATH). The instructions that follow assume a directory in your home directory outside of the standard GOPATH (i.e `$HOME/development/terraform-providers/`).

<sup>[back to top](#top)</sup>

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12+
- [Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)

<sup>[back to top](#top)</sup>

## Environment

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
<sup>[back to top](#top)</sup>

## Testing the Provider
In order to test the provider, you can run `make test`. This will run so-called unit tests.
```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`. Please make sure that a working Domain Controller is reachable and you have the needed permissions

_**Note:**_ Acceptance tests create real resources! Please read [Running an Acceptance Test](/terraform-provider-activedirectory/CONTRIBUTING#running-an-acceptance-test) in the contribution guidelines for more information on usage.

```sh
$ make testacc
```

 For `make testacc` you have to set the following environment variables:

 | Variable | Description | Example | Default | Required |
 | -------- | ----------- | ------- | ------- | :------: |
 | AD_HOST | Domain Controller | dc.example.org | - | yes |
 | AD_PORT | LDAP Port | 389 | 389 | no |
 | AD_DOMAIN | Domain | eample.org | - | yes |
 | AD_USE_TLS | Use secure connection | false | true | no |
 | AD_USER | Admin user name or DN | admin | - | yes |
 | AD_PASSWORD | Password of the admin user | secret | - | yes |
 | AD_TEST_BASE_OU | OU for the test cases | ou=Tests,dc=example,dc=org | - | yes (tests only) |

<sup>[back to top](#top)</sup>

# Contributing
Terraform is the work of thousands of contributors. We appreciate your help!

To contribute, please read the contribution guidelines: [Contributing to Terraform - Active Directory Provider](/terraform-provider-activedirectory/CONTRIBUTING)

Issues on GitHub are intended to be related to bugs or feature requests with provider codebase. See [https://www.terraform.io/docs/extend/community/index.html](https://www.terraform.io/docs/extend/community/index.html) for a list of community resources to ask questions about Terraform.

<sup>[back to top](#top)</sup>
