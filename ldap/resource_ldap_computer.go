package ldap

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	log "github.com/sirupsen/logrus"
)

// resourceLDAPComputerObject is the main function for ad computer terraform resource
func resourceLDAPComputerObject() *schema.Resource {
	return &schema.Resource{
		Create: resourceLDAPComputerObjectCreate,
		Read:   resourceLDAPComputerObjectRead,
		Update: resourceLDAPComputerObjectUpdate,
		Delete: resourceLDAPComputerObjectDelete,
		Exists: resourceLDAPComputerObjectExists,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// this is to ignore case in ldap distinguished name
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"ou": {
				Type:     schema.TypeString,
				Required: true,
				// this is to ignore case in ldap distinguished name
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

// resourceLDAPComputerObjectCreate is 'create' part of terraform CRUD functions for AD provider
func resourceLDAPComputerObjectCreate(d *schema.ResourceData, meta interface{}) error {
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
		log.Errorf("Error while creating ldap computer object %s: %s", dn, err)
		return err
	}

	d.SetId(dn)
	return resourceLDAPComputerObjectRead(d, meta)
}

// resourceLDAPComputerObjectRead is 'read' part of terraform CRUD functions for AD provider
func resourceLDAPComputerObjectRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)
	dn := fmt.Sprintf("cn=%s,%s", d.Get("name").(string), d.Get("ou").(string))

	computer, err := api.getComputer(dn, []string{"description"})
	if err != nil {
		log.Errorf("Error while reading ldap computer object %s: %s", dn, err)
		return err
	}

	if computer == nil {
		log.Debugf("Computer object %s no longer exists", dn)

		d.SetId("")
		return nil
	}

	d.SetId(dn)

	if err := d.Set("description", computer.attributes["description"][0]); err != nil {
		log.Errorf("Error while setting ldap object's %s description: %s", dn, err)
		return err
	}

	return nil
}

// resourceLDAPComputerObjectUpdate is 'update' part of terraform CRUD functions for ad provider
func resourceLDAPComputerObjectUpdate(d *schema.ResourceData, meta interface{}) error {
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

		api.updateComputerAttributes(dn, nil, attributes, nil)

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
	return resourceLDAPComputerObjectRead(d, meta)
}

// resourceLDAPComputerObjectDelete is 'delete' part of terraform CRUD functions for ad provider
func resourceLDAPComputerObjectDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)

	// creating computer dn
	dn := fmt.Sprintf("cn=%s,%s", d.Get("name").(string), d.Get("ou").(string))

	// call ad to delete the computer object, no error means that object was deleted successfully
	return api.deleteComputer(dn)
}

func resourceLDAPComputerObjectExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	api := meta.(APIInterface)
	dn := fmt.Sprintf("cn=%s,%s", d.Get("name").(string), d.Get("ou").(string))

	objects, err := api.searchObject("(objectClass=*)", dn, nil)
	if err != nil {
		log.Errorf("Error while searching for ldap computer object %s: %s", dn, err)
		return false, err
	}

	if len(objects) == 0 {
		log.Infof("Computer object %d not found", dn)
		return false, nil
	}

	log.Infof("Computer object %d found", dn)

	return true, nil
}
