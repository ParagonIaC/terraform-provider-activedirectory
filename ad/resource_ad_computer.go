package activedirectory

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// resourceADComputer is the main function for ad computer terraform resource
func resourceADComputer() *schema.Resource {
	return &schema.Resource{
		Create: resourceADComputerCreate,
		Read:   resourceADComputerRead,
		Update: resourceADComputerUpdate,
		Delete: resourceADComputerDelete,
		Schema: map[string]*schema.Schema{
			"computer_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ou_distinguished_name": {
				Type:     schema.TypeString,
				Required: true,
				// this is to ignore case in ldap distinguished name
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if strings.ToLower(old) == strings.ToLower(new) {
						return true
					}
					return false
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

// resourceADComputerCreate is 'create' part of terraform CRUD functions for AD provider
func resourceADComputerCreate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)

	desc := []string{}
	if d.Get("description").(string) != "" {
		desc = append(desc, d.Get("description").(string))
	}

	// create computer object
	computer := &Computer{
		Name: d.Get("computer_name").(string),
		Attributes: []*ldap.EntryAttribute{
			&ldap.EntryAttribute{
				Name:   "description",
				Values: desc,
			},
		},
	}

	// create ldap computer object
	if err := api.CreateComputer(computer, d.Get("ou_distinguished_name").(string)); err != nil {
		return err
	}

	d.SetId(computer.DN)

	// read current ad data to avoid drift
	return resourceADComputerRead(d, meta)
}

// resourceADComputerRead is 'read' part of terraform CRUD functions for AD provider
func resourceADComputerRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)

	computer, err := api.GetComputerByDN(d.Id(), "", []string{"description"})
	if err != nil {
		log.Errorf("failed to retrieve ad computer object")
		return err
	}

	// computer object no longer exists
	if computer == nil {
		d.SetId("")
		return nil
	}

	// set 'computer_name' field
	if err = d.Set("computer_name", computer.Name); err != nil {
		return err
	}

	// create ou name out of computer's distinguished name
	tmp := strings.Split(computer.DN, ",")
	ou := strings.Join(tmp[1:], ",")

	// set 'ou_distinguished_name' field
	if err = d.Set("ou_distinguished_name", ou); err != nil {
		return err
	}

	// set 'description' field
	description := ""
	for _, attr := range computer.Attributes {
		if attr.Name == "description" {
			if len(attr.Values[0]) != 0 {
				description = attr.Values[0]
			} else {
				description = ""
			}
		}
	}

	if err = d.Set("description", description); err != nil {
		return err
	}

	return nil
}

// resourceADComputerUpdate is 'update' part of terraform CRUD functions for ad provider
func resourceADComputerUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)

	//construct dn name with the "old" ou_distinguished_name, because it could have been changed
	old, _ := d.GetChange("ou_distinguished_name")

	computer := &Computer{
		Name: d.Get("computer_name").(string),
		DN:   fmt.Sprintf("cn=%s,%s", d.Get("computer_name").(string), old.(string)),
	}

	// let's try to update in parts
	d.Partial(true)

	// check description
	if d.HasChange("description") {
		desc := []string{}
		if d.Get("description").(string) != "" {
			desc = append(desc, d.Get("description").(string))
		}

		attr := []*ldap.EntryAttribute{
			&ldap.EntryAttribute{
				Name:   "description",
				Values: desc,
			},
		}

		// update attributes
		if err := api.UpdateComputerAttributes(computer, attr); err != nil {
			return err
		}

		d.SetPartial("description")
	}

	// check ou
	if d.HasChange("ou_distinguished_name") {
		// update ou
		if err := api.UpdateComputerOU(computer, d.Get("ou_distinguished_name").(string)); err != nil {
			return err
		}

		d.SetId(computer.DN)
	}

	d.Partial(false)

	// read current ad data to avoid drift
	return resourceADComputerRead(d, meta)
}

// resourceADComputerDelete is 'delete' part of terraform CRUD functions for ad provider
func resourceADComputerDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)

	// creating computer dn
	dn := fmt.Sprintf("cn=%s,%s", d.Get("computer_name").(string), d.Get("ou_distinguished_name").(string))

	// call ad to delete the computer object, no error means that object was deleted successfully
	return api.DeleteComputer(dn)
}
