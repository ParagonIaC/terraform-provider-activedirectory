---
page_title: "activedirectory_ou Resource - terraform-provider-activedirectory"
subcategory: ""
description: |-
  
---

# Resource `activedirectory_ou`



## Example Usage

```terraform
resource "activedirectory_ou" "test_ou" {
  name           = "TerraformOU"                            # can be updated
  base_ou        = "OU=Test,CN=Computers,DC=example,DC=org" # can be updated
  description    = "terraform sample ou"                    # can be updated
}
```

## Schema

### Required

- **base_ou** (String)
- **name** (String)

### Optional

- **description** (String)
- **id** (String) The ID of this resource.


