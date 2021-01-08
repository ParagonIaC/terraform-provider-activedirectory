package activedirectory

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	log "github.com/sirupsen/logrus"
)

// resourceADComputerObject is the main function for ad computer terraform resource
func resourceADComputerObject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADComputerObjectCreate,
		ReadContext:   resourceADComputerObjectRead,
		UpdateContext: resourceADComputerObjectUpdate,
		DeleteContext: resourceADComputerObjectDelete,

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
func resourceADComputerObjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Infof("Creating AD computer object")

	api := meta.(APIInterface)

	var diags diag.Diagnostics

	if err := api.createComputer(d.Get("name").(string), d.Get("ou").(string), d.Get("description").(string)); err != nil {
		return diag.Errorf("resourceADComputerObjectCreate - create - %s", err)
	}

	d.SetId(strings.ToLower(fmt.Sprintf("cn=%s,%s", d.Get("name").(string), d.Get("ou").(string))))
	resourceADComputerObjectRead(ctx, d, meta)
	return diags
}

// resourceADComputerObjectRead is 'read' part of terraform CRUD functions for AD provider
func resourceADComputerObjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Infof("Reading AD computer object")

	api := meta.(APIInterface)

	var diags diag.Diagnostics

	computer, err := api.getComputer(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("resourceADComputerObjectRead - getComputer - %s", err)
	}

	if computer == nil {
		log.Infof("Computer object %s no longer exists", d.Get("name").(string))

		d.SetId("")
		return diags
	}

	d.SetId(strings.ToLower(computer.dn))

	if err := d.Set("description", computer.description); err != nil {
		return diag.Errorf("resourceADComputerObjectRead - set description - failed to set description: %s", err)
	}

	return diags
}

// resourceADComputerObjectUpdate is 'update' part of terraform CRUD functions for ad provider
func resourceADComputerObjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Infof("Updating AD computer object")

	api := meta.(APIInterface)

	var diags diag.Diagnostics

	oldOU, newOU := d.GetChange("ou")

	// let's try to update in parts
	d.Partial(true)

	// check description
	if d.HasChange("description") {
		if err := api.updateComputerDescription(d.Get("name").(string), oldOU.(string), d.Get("description").(string)); err != nil {
			return diag.Errorf("resourceADComputerObjectUpdate - update description - %s", err)
		}
	}

	// check ou
	if d.HasChange("ou") {
		if err := api.updateComputerOU(d.Get("name").(string), oldOU.(string), newOU.(string)); err != nil {
			return diag.Errorf("resourceADComputerObjectUpdate - update ou - %s", err)
		}
	}

	d.Partial(false)
	d.SetId(strings.ToLower(fmt.Sprintf("cn=%s,%s", d.Get("name").(string), d.Get("ou").(string))))

	// read current ad data to avoid drift
	resourceADComputerObjectRead(ctx, d, meta)
	return diags
}

// resourceADComputerObjectDelete is 'delete' part of terraform CRUD functions for ad provider
func resourceADComputerObjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Infof("Deleting AD computer object")

	api := meta.(APIInterface)

	var diags diag.Diagnostics

	// call ad to delete the computer object, no error means that object was deleted successfully
	err := api.deleteComputer(d.Get("name").(string), d.Get("ou").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
