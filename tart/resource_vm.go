package tart

import (
    //"context"
    //"fmt"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVM() *schema.Resource {
    return &schema.Resource{
        Create: resourceVMCreate,
        Read:   resourceVMRead,
        Delete: resourceVMDelete,
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
        Schema: map[string]*schema.Schema{
            "name":  {Type: schema.TypeString, Required: true, ForceNew: true},
            "image": {Type: schema.TypeString, Required: true, ForceNew: true},
            "status": {Type: schema.TypeString, Computed: true},
        },
    }
}

func resourceVMCreate(d *schema.ResourceData, m interface{}) error {
	conf := m.(*config)
	id, status, err := createVM(conf, d.Get("name").(string), d.Get("image").(string))
	if err != nil {
		return err
	}
	d.SetId(id)
	d.Set("status", status)
	return resourceVMRead(d, m)
}

func resourceVMRead(d *schema.ResourceData, m interface{}) error {
	conf := m.(*config)
	id := d.Id()
	name, image, status, err := getVM(conf, id)
	if err != nil {
		d.SetId("")
		return nil
	}
	d.Set("name", name)
	d.Set("image", image)
	d.Set("status", status)
	return nil
}

func resourceVMDelete(d *schema.ResourceData, m interface{}) error {
	conf := m.(*config)
	id := d.Id()
	if err := deleteVM(conf, id); err != nil {
		return err
	}
	d.SetId("")
	return nil
}
