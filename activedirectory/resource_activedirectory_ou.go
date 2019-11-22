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
	api := meta.(APIInterface)
	dn := strings.ToLower(fmt.Sprintf("OU=%s,%s", d.Get("name").(string), d.Get("base_ou").(string)))

	if err := api.createOU(dn, d.Get("name").(string), d.Get("description").(string)); err != nil {
		log.Errorf("Error while creating ad ou object %s: %s", dn, err)
		return err
	}

	d.SetId(dn)
	return resourceADOUObjectRead(d, meta)
}

// resourceADOUObjectRead is 'read' part of terraform CRUD functions for AD provider
func resourceADOUObjectRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)
	dn := strings.ToLower(fmt.Sprintf("OU=%s,%s", d.Get("name").(string), d.Get("base_ou").(string)))

	ou, err := api.getOU(d.Get("name").(string), d.Get("base_ou").(string))
	if err != nil {
		log.Errorf("Error while reading ad ou object %s: %s", dn, err)
		return err
	}

	if ou == nil {
		log.Debugf("Computer object %s no longer exists", dn)

		d.SetId("")
		return nil
	}

	d.SetId(dn)

	if err := d.Set("name", ou.name); err != nil {
		log.Errorf("Error while setting ad object's %s name: %s", dn, err)
		return err
	}

	baseOU := strings.ToLower(ou.dn[(len(ou.name) + 1 + 3):]) // remove 'ou=' and ','
	if err := d.Set("base_ou", baseOU); err != nil {
		log.Errorf("Error while setting ad object's %s base_ou: %s", dn, err)
		return err
	}

	if err := d.Set("description", ou.description); err != nil {
		log.Errorf("Error while setting ad object's %s description: %s", dn, err)
		return err
	}

	return nil
}

// resourceADOUObjectUpdate is 'update' part of terraform CRUD functions for ad provider
func resourceADOUObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)

	oldOU, newOU := d.GetChange("base_ou")
	oldName, newName := d.GetChange("name")

	// let's try to update in parts
	d.Partial(true)

	// check description
	if d.HasChange("description") {
		dn := strings.ToLower(fmt.Sprintf("OU=%s,%s", oldName.(string), oldOU.(string)))
		if err := api.updateOUDescription(dn, d.Get("description").(string)); err != nil {
			return err
		}

		d.SetPartial("description")
	}

	// check description
	if d.HasChange("name") {
		dn := strings.ToLower(fmt.Sprintf("OU=%s,%s", oldName.(string), oldOU.(string)))
		if err := api.updateOUName(dn, newName.(string)); err != nil {
			return err
		}

		d.SetPartial("name")
	}

	// check ou
	if d.HasChange("base_ou") {
		dn := strings.ToLower(fmt.Sprintf("OU=%s,%s", newName.(string), oldOU.(string)))
		if err := api.moveOU(dn, newName.(string), d.Get("base_ou").(string)); err != nil {
			return err
		}
	}

	d.Partial(false)
	d.SetId(strings.ToLower(fmt.Sprintf("OU=%s,%s", newName.(string), newOU.(string))))

	// read current ad data to avoid drift
	return resourceADOUObjectRead(d, meta)
}

// resourceADOUObjectDelete is 'delete' part of terraform CRUD functions for ad provider
func resourceADOUObjectDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)

	// creating computer dn
	dn := strings.ToLower(fmt.Sprintf("OU=%s,%s", d.Get("name").(string), d.Get("base_ou").(string)))

	// call ad to delete the ou object, no error means that object was deleted successfully
	return api.deleteOU(dn)
}
