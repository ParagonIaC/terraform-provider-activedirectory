provider "activedirectory" {
  port=636
  use_tls=true
  no_cert_verify=false
  domain="example.org"
  host="ad.example.org"
  user = "example-terrafrom-user"
  password = "example-terrafrom-password"
}