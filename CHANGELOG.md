## 0.0.6 (November 28, 2019)

BUG FIXES:
* Connection failed when TLS was enabled since no server name was defined and InsecureSkipVerify was set to true. We are now setting to server name in the TLS config and also allowing InsecureSkipVerify to be set to false via provider setup, which results in a TLS connection without checking the certificate.

```hcl
# Configure the AD Provider
provider "activedirectory" {
  ...
  use_tls  = true
  no_cert_verify = true
  ...
}

## 0.0.6 (November 28, 2019)

NOTES:
* We added GoReleaser to our releasing pipeline.

## 0.0.6 (November 28, 2019)

NOTES:
* We added GoReleaser to our releasing pipeline.

## 0.0.5 (November 27, 2019)

NOTES:
* Provider and resource attributes were adjusted to conform more with Microsoft Active Directory terminology:

      ldap_host     -> host
      ldap_port     -> port
      bind_user     -> user
      bind_password -> password

* Provider attribute `domain` was added to be the base domain for `ldap.search` requests

FEATURES:
* **New Resource:** `activedirectory_ou`

ENHANCEMENTS:
* The AD API is now checking the existence of an AD object before trying to perform create/update/delete operations
* `getOU` and `getComputer` are now using go `ldap.v3` directly to get AD objects instead of search in a more generic way via `getObject`
* Error logs are now giving more information about the context of the code where the error happend
* If no base dn is provided Object.search will now use `api.domain` as its base dn
* User authentication can now be done in MS AD style `admin@domain` as well as LDAP style `uid=admin,ou=system`
** AD style: `user` must be set to just the username, `domain` will be attached by the API
** LDAP style: `user` must be set to the DN of the user, like `uid=admin,ou=system`

BUG FIXES:
* Provider has attribute `password` set to `sensitive=true` to prevent printing and saving
* Set resource ID to `strings.toLower(dn)` to prevent ID mismatch due to case sensitivity
* `searchObject` return an error when nothing was found, now it returns nil

## 0.0.4 (November 20, 2019)

This is the first changelog entry.

FEATURES:

* **New Resource:** `activedirectory_computer`
