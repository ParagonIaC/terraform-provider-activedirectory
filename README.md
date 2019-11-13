# Terraform Provider - Active Directory/LDAP

[![GolangCI](https://golangci.com/badges/github.com/golangci/golangci-lint.svg)](https://golangci.com)

This is a Terraform  Provider to work with Active Directory/LDAP.

This provider currently supports only computer objects, but more active directory resources are planned. Please feel free to contribute.

For general information about Terraform, visit the [official website][3] and the [GitHub project page][4].

[3]: https://terraform.io/
[4]: https://github.com/hashicorp/terraform

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12+
- [Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)

## Developing the Provider
TODO

## Using the Provider
The provider is useful for adding and managing computer objects in Active Directory.
### Example
```hcl
# Configure the Active Directory Provider
provider "activedirectory" {
  domain         = "${var.ad_server_domain}"
  user           = "${var.ad_server_user}"
  password       = "${var.ad_server_password}"
  server_host    = "${var.ad_server_host}"
  server_port    = "${var.ad_server_host}" # 389 is the default value
}

# Add computer to Active Directory
resource "activedirectory_computer" "foo" {
  computer_name           = "TestComputerTF"                       # update will force destroy and new
  ou_distinguished_name   = "CN=Computers,DC=mycompany,DC=local"   # can be updated
  description             = "terraform sample server"              # can be updated
}
```

### Updating Dependencies
```console
$ go get URL
$ go mod tidy
$ go mod vendor
```

## Testing the Provider
TODO

## Contributing
TODO
