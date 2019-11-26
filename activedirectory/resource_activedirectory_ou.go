package activedirectory

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	log "github.com/sirupsen/logrus"
)

// resourceADOUObject is the main function for ad ou terraform resource
func resourceADOUObject() *schema.Resource {
	return &schema.Resource{
		Create: resourceADOUObjectCreate,
		Read:   resourceADOUObjectRead,
		Update: resourceADOUObjectUpdate,
		Delete: resourceADOUObjectDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				// this is to ignore case in ad distinguished name
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"base_ou": {
				Type:     schema.TypeString,
				Required: true,
				// this is to ignore case in ad distinguished name
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
		},
	}
}

// resourceADOUObjectCreate is 'create' part of terraform CRUD functions for AD provider
func resourceADOUObjectCreate(d *schema.ResourceData, meta interface{}) error {
	log.Infof("Creating AD OU object")

	api := meta.(APIInterface)

	if err := api.createOU(d.Get("name").(string), d.Get("base_ou").(string), d.Get("description").(string)); err != nil {
		return fmt.Errorf("resourceADOUObjectCreate - %s", err)
	}

	d.SetId(strings.ToLower(fmt.Sprintf("ou=%s,%s", d.Get("name").(string), d.Get("base_ou").(string))))
	return resourceADOUObjectRead(d, meta)
}

// resourceADOUObjectRead is 'read' part of terraform CRUD functions for AD provider
func resourceADOUObjectRead(d *schema.ResourceData, meta interface{}) error {
	log.Infof("Reading AD OU object")

	api := meta.(APIInterface)

	ou, err := api.getOU(d.Get("name").(string), d.Get("base_ou").(string))
	if err != nil {
		return fmt.Errorf("resourceADOUObjectRead - %s", err)
	}

	if ou == nil {
		log.Infof("OU object %s no longer exists under %s", d.Get("name").(string), d.Get("base_ou").(string))

		d.SetId("")
		return nil
	}

	if err := d.Set("name", ou.name); err != nil {
		return fmt.Errorf("resourceADOUObjectRead - failed to set ou name to %s: %s", ou.name, err)
	}

	baseOU := strings.ToLower(ou.dn[(len(ou.name) + 1 + 3):]) // remove 'ou=' and ',' and ou name
	if err := d.Set("base_ou", baseOU); err != nil {
		return fmt.Errorf("resourceADOUObjectRead - failed to set ou base_ou to %s: %s", baseOU, err)
	}

	if err := d.Set("description", ou.description); err != nil {
		return fmt.Errorf("resourceADOUObjectRead - failed to set ou description to %s: %s", ou.description, err)
	}

	d.SetId(strings.ToLower(ou.dn))

	return nil
}

// resourceADOUObjectUpdate is 'update' part of terraform CRUD functions for ad provider
func resourceADOUObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Infof("Updating AD OU object")

	api := meta.(APIInterface)

	oldOU, newOU := d.GetChange("base_ou")
	oldName, newName := d.GetChange("name")

	// let's try to update in parts
	d.Partial(true)

	// check description
	if d.HasChange("description") {
		if err := api.updateOUDescription(oldName.(string), oldOU.(string), d.Get("description").(string)); err != nil {
			return fmt.Errorf("resourceADOUObjectUpdate - %s", err)
		}

		d.SetPartial("description")
	}

	// check description
	if d.HasChange("name") {
		if err := api.updateOUName(oldName.(string), oldOU.(string), newName.(string)); err != nil {
			return fmt.Errorf("resourceADOUObjectUpdate - %s", err)
		}

		d.SetPartial("name")
	}

	// check ou
	if d.HasChange("base_ou") {
		if err := api.moveOU(newName.(string), oldOU.(string), newOU.(string)); err != nil {
			return fmt.Errorf("resourceADOUObjectUpdate - %s", err)
		}
	}

	d.Partial(false)
	d.SetId(strings.ToLower(fmt.Sprintf("ou=%s,%s", newName.(string), newOU.(string))))

	// read current ad data to avoid drift
	return resourceADOUObjectRead(d, meta)
}

// resourceADOUObjectDelete is 'delete' part of terraform CRUD functions for ad provider
func resourceADOUObjectDelete(d *schema.ResourceData, meta interface{}) error {
	log.Infof("Deleting AD OU object")

	api := meta.(APIInterface)

	// call ad to delete the ou object, no error means that object was deleted successfully
	return api.deleteOU(strings.ToLower(fmt.Sprintf("ou=%s,%s", d.Get("name").(string), d.Get("base_ou").(string))))
}
