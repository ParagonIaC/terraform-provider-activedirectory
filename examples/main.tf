provider "activedirectory" {
  port = 389
  use_tls = false
  no_cert_verify = true
  host = var.ad_host
  domain = var.ad_domain
  user = var.ad_user
  password = var.ad_password
}

variable "ad_user" {
  type = string
}
variable "ad_password" {
  type = string
}
variable "ad_host" {
  type = string
}
variable "ad_domain" {
  type = string
}
variable "ad_base_ou" {
  type = string
}

resource "activedirectory_ou" "test" {
  name = "test"
  base_ou = var.ad_base_ou
  description = "Cos"
}