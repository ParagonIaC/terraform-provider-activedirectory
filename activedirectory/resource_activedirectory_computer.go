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
	api := meta.(APIInterface)
	dn := fmt.Sprintf("cn=%s,%s", d.Get("name").(string), d.Get("ou").(string))

	desc := []string{}
	if d.Get("description").(string) != "" {
		desc = append(desc, d.Get("description").(string))
	}

	attributes := map[string][]string{
		"description": desc,
	}

	if err := api.createComputer(dn, d.Get("name").(string), attributes); err != nil {
		log.Errorf("Error while creating ad computer object %s: %s", dn, err)
		return err
	}

	d.SetId(dn)
	return resourceADComputerObjectRead(d, meta)
}

// resourceADComputerObjectRead is 'read' part of terraform CRUD functions for AD provider
func resourceADComputerObjectRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)
	dn := fmt.Sprintf("cn=%s,%s", d.Get("name").(string), d.Get("ou").(string))

	computer, err := api.getComputer(dn, []string{"description"})
	if err != nil {
		log.Errorf("Error while reading ad computer object %s: %s", dn, err)
		return err
	}

	if computer == nil {
		log.Debugf("Computer object %s no longer exists", dn)

		d.SetId("")
		return nil
	}

	d.SetId(dn)

	if err := d.Set("description", computer.attributes["description"][0]); err != nil {
		log.Errorf("Error while setting ad object's %s description: %s", dn, err)
		return err
	}

	return nil
}

// resourceADComputerObjectUpdate is 'update' part of terraform CRUD functions for ad provider
func resourceADComputerObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)
	dn := fmt.Sprintf("cn=%s,%s", d.Get("name").(string), d.Get("ou").(string))

	// check ou
	if d.HasChange("ou") {
		old, _ := d.GetChange("ou")
		dn = fmt.Sprintf("cn=%s,%s", d.Get("name").(string), old.(string))
	}

	// let's try to update in parts
	d.Partial(true)

	// check description
	if d.HasChange("description") {
		desc := []string{}
		if d.Get("description").(string) != "" {
			desc = append(desc, d.Get("description").(string))
		}

		attributes := map[string][]string{
			"description": desc,
		}

		if err := api.updateComputerAttributes(dn, nil, attributes, nil); err != nil {
			return err
		}

		d.SetPartial("description")
	}

	// check ou
	if d.HasChange("ou") {
		// update ou
		if err := api.updateComputerOU(dn, d.Get("name").(string), d.Get("ou").(string)); err != nil {
			return err
		}
	}

	d.Partial(false)
	d.SetId(dn)

	// read current ad data to avoid drift
	return resourceADComputerObjectRead(d, meta)
}

// resourceADComputerObjectDelete is 'delete' part of terraform CRUD functions for ad provider
func resourceADComputerObjectDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)

	// creating computer dn
	dn := fmt.Sprintf("cn=%s,%s", d.Get("name").(string), d.Get("ou").(string))

	// call ad to delete the computer object, no error means that object was deleted successfully
	return api.deleteComputer(dn)
}
