package ldap

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	log "github.com/sirupsen/logrus"
)

func resourceLDAPObject() *schema.Resource {
	return &schema.Resource{
		Create: resourceLDAPObjectCreate,
		Read:   resourceLDAPObjectRead,
		Update: resourceLDAPObjectUpdate,
		Delete: resourceLDAPObjectDelete,
		Exists: resourceLDAPObjectExists,

		Schema: map[string]*schema.Schema{
			"dn": {
				Type:        schema.TypeString,
				Description: "The Distinguished Name (DN) of the object.",
				Required:    true,
				ForceNew:    true,
				// this is to ignore case in ldap distinguished name
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"object_classes": {
				Type:        schema.TypeSet,
				Description: "The set of classes this object conforms to (e.g. computer).",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Required:    true,
			},
			"attributes": {
				Type:        schema.TypeMap,
				Description: "The map of attributes of this object; each attribute can be multi-valued.",

				Elem: &schema.Schema{
					Type:        schema.TypeList,
					Description: "The list of values for a given attribute.",
					MinItems:    1,
					Elem: &schema.Schema{
						Type:        schema.TypeString,
						Description: "The individual value for the given attribute.",
					},
				},
				Optional: true,
			},
		},
	}
}

func resourceLDAPObjectCreate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)
	dn := d.Get("dn").(string)

	objectClasses := []string{}
	for _, oc := range (d.Get("object_classes").(*schema.Set)).List() {
		objectClasses = append(objectClasses, oc.(string))
	}

	attributes := map[string][]string{}
	if v, ok := d.GetOk("attributes"); ok {
		for key, values := range v.(map[string]interface{}) {
			attributes[key] = values.([]string)
		}
	}

	if err := api.createObject(dn, objectClasses, attributes); err != nil {
		log.Errorf("Error while creating ldap object %s: %s", dn, err)
		return err
	}

	d.SetId(dn)
	return resourceLDAPObjectRead(d, meta)
}

func resourceLDAPObjectRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)
	dn := d.Get("dn").(string)

	object, err := api.getObject(dn, nil)
	if err != nil {
		log.Errorf("Error while reading ldap object %s: %s", dn, err)
		return err
	}

	if object == nil {
		log.Debugf("Object %s no longer exists", dn)

		d.SetId("")
		return nil
	}

	d.SetId(dn)

	attributes := object.attributes

	if err := d.Set("object_classes", attributes["object_classes"]); err != nil {
		log.Errorf("Error while setting ldap object's %s object classes: %s", dn, err)
		return err
	}

	delete(attributes, "object_classes")

	if err := d.Set("attributes", attributes); err != nil {
		log.Errorf("Error while setting ldap object's %s attributes: %s", dn, err)
		return err
	}

	return nil
}

func resourceLDAPObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)
	dn := d.Get("dn").(string)

	var classes []string
	if d.HasChange("object_classes") {
		for _, oc := range (d.Get("object_classes").(*schema.Set)).List() {
			classes = append(classes, oc.(string))
		}
	}

	var added, changed, removed map[string][]string
	if d.HasChange("attributes") {
		o, n := d.GetChange("attributes")
		added, changed, removed = computeDeltas(o.(map[string]interface{}), n.(map[string]interface{}))
	}

	if err := api.updateObject(dn, classes, added, changed, removed); err != nil {
		log.Errorf("Error while updating ldap object's %s attributes: %s", dn, err)
		return err
	}

	return resourceLDAPObjectRead(d, meta)
}

func resourceLDAPObjectDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(APIInterface)
	dn := d.Get("dn").(string)

	return api.deleteObject(dn)
}

func resourceLDAPObjectExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	api := meta.(APIInterface)
	dn := d.Get("dn").(string)

	objects, err := api.searchObject("(objectClass=*)", dn, nil)
	if err != nil {
		log.Errorf("Error while searching for ldap object %s: %s", dn, err)
		return false, err
	}

	if len(objects) == 0 {
		log.Infof("Object %d not found", dn)
		return false, nil
	}

	log.Infof("Object %d found", dn)

	return true, nil
}

func computeDeltas(oldSet, newSet map[string]interface{}) (added, changed, removed map[string][]string) {
	for key, value := range oldSet {
		if _, ok := newSet[key]; ok {
			changed[key] = value.([]string)
		} else {
			removed[key] = []string{}
		}
	}

	for key, value := range newSet {
		if _, ok := oldSet[key]; !ok {
			added[key] = value.([]string)
		}
	}

	return
}
