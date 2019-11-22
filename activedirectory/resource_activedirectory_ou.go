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
	dn := fmt.Sprintf("ou=%s,%s", d.Get("name").(string), d.Get("base_ou").(string))

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
	dn := fmt.Sprintf("ou=%s,%s", d.Get("name").(string), d.Get("base_ou").(string))

	ou, err := api.getOU(dn)
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

	baseOU := strings.Replace(ou.dn, "ou="+ou.name, "", 1)
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
	dn := fmt.Sprintf("ou=%s,%s", d.Get("name").(string), d.Get("base_ou").(string))

	// check ou
	if d.HasChange("base_ou") {
		old, _ := d.GetChange("base_ou")
		dn = fmt.Sprintf("ou=%s,%s", d.Get("name").(string), old.(string))
	}

	// let's try to update in parts
	d.Partial(true)

	// check description
	if d.HasChange("description") {
		if err := api.updateOUDescription(dn, d.Get("description").(string)); err != nil {
			return err
		}

		d.SetPartial("description")
	}

	// check ou
	if d.HasChange("base_ou") {
		// update ou
		if err := api.moveOU(dn, d.Get("name").(string), d.Get("base_ou").(string)); err != nil {
			return err
		}
	}

	d.Partial(false)
	d.SetId(dn)

	// read current ad data to avoid drift
	return resourceADOUObjectRead(d, meta)
}

// resourceADOUObjectDelete is 'delete' part of terraform CRUD functions for ad provider
func resourceADOUObjectDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)

	// creating computer dn
	dn := fmt.Sprintf("ou=%s,%s", d.Get("name").(string), d.Get("base_ou").(string))

	// call ad to delete the ou object, no error means that object was deleted successfully
	return api.deleteOU(dn)
}
