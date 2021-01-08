---
page_title: "activedirectory_computer Resource - terraform-provider-activedirectory"
subcategory: ""
description: |-
  
---

# Resource `activedirectory_computer`


```terraform
resource "activedirectory_computer" "test_computer" {
name           = "TerraformComputer"                      # update will force destroy and new
ou             = "CN=Computers,DC=example,DC=org"         # can be updated
description    = "terraform sample server"                # can be updated
}
```
## Schema

## Example Usage

### Required

- **name** (String)
- **ou** (String)

### Optional

- **description** (String)
- **id** (String) The ID of this resource.


