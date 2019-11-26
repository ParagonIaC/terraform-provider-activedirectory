package activedirectory

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	log "github.com/sirupsen/logrus"
)

// resourceADComputerObject is the main function for ad computer terraform resource
func resourceADComputerObject() *schema.Resource {
	return &schema.Resource{
		Create: resourceADComputerObjectCreate,
		Read:   resourceADComputerObjectRead,
		Update: resourceADComputerObjectUpdate,
		Delete: resourceADComputerObjectDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// this is to ignore case in ad distinguished name
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"ou": {
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

// resourceADComputerObjectCreate is 'create' part of terraform CRUD functions for AD provider
func resourceADComputerObjectCreate(d *schema.ResourceData, meta interface{}) error {
	log.Infof("Creating AD computer object")

	api := meta.(APIInterface)

	if err := api.createComputer(d.Get("name").(string), d.Get("ou").(string), d.Get("description").(string)); err != nil {
		return fmt.Errorf("resourceADComputerObjectCreate - create - %s", err)
	}

	d.SetId(strings.ToLower(fmt.Sprintf("cn=%s,%s", d.Get("name").(string), d.Get("ou").(string))))
	return resourceADComputerObjectRead(d, meta)
}

// resourceADComputerObjectRead is 'read' part of terraform CRUD functions for AD provider
func resourceADComputerObjectRead(d *schema.ResourceData, meta interface{}) error {
	log.Infof("Reading AD computer object")

	api := meta.(APIInterface)

	computer, err := api.getComputer(d.Get("name").(string))
	if err != nil {
		return fmt.Errorf("resourceADComputerObjectRead - getComputer - %s", err)
	}

	if computer == nil {
		log.Infof("Computer object %s no longer exists", d.Get("name").(string))

		d.SetId("")
		return nil
	}

	d.SetId(strings.ToLower(computer.dn))

	if err := d.Set("description", computer.description); err != nil {
		return fmt.Errorf("resourceADComputerObjectRead - set description - failed to set description: %s", err)
	}

	return nil
}

// resourceADComputerObjectUpdate is 'update' part of terraform CRUD functions for ad provider
func resourceADComputerObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Infof("Updating AD computer object")

	api := meta.(APIInterface)

	oldOU, newOU := d.GetChange("ou")

	// let's try to update in parts
	d.Partial(true)

	// check description
	if d.HasChange("description") {
		if err := api.updateComputerDescription(d.Get("name").(string), oldOU.(string), d.Get("description").(string)); err != nil {
			return fmt.Errorf("resourceADComputerObjectUpdate - update description - %s", err)
		}

		d.SetPartial("description")
	}

	// check ou
	if d.HasChange("ou") {
		if err := api.updateComputerOU(d.Get("name").(string), oldOU.(string), newOU.(string)); err != nil {
			return fmt.Errorf("resourceADComputerObjectUpdate - update ou - %s", err)
		}
	}

	d.Partial(false)
	d.SetId(strings.ToLower(fmt.Sprintf("cn=%s,%s", d.Get("name").(string), d.Get("ou").(string))))

	// read current ad data to avoid drift
	return resourceADComputerObjectRead(d, meta)
}

// resourceADComputerObjectDelete is 'delete' part of terraform CRUD functions for ad provider
func resourceADComputerObjectDelete(d *schema.ResourceData, meta interface{}) error {
	log.Infof("Deleting AD computer object")

	api := meta.(APIInterface)

	// call ad to delete the computer object, no error means that object was deleted successfully
	return api.deleteComputer(d.Get("name").(string), d.Get("ou").(string))
}
