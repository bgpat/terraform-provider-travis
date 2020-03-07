package travis

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/shuheiktgw/go-travis"
)

func resourceEnvVar() *schema.Resource {
	return &schema.Resource{
		Create: resourceEnvVarCreate,
		Read:   resourceEnvVarRead,
		Update: resourceEnvVarUpdate,
		Delete: resourceEnvVarDelete,

		Schema: map[string]*schema.Schema{
			"repository_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Value uniquely identifying the repository.",
				ForceNew:     true,
				ExactlyOneOf: []string{"repository_slug"},
			},
			"repository_slug": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Same as {repository.owner.name}/{repository.name}.",
				ForceNew:     true,
				ExactlyOneOf: []string{"repository_id"},
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment variable name, e.g. FOO.",
			},
			"value": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The environment variable's value, e.g. bar.",
				ExactlyOneOf: []string{"secure_value"},
			},
			"secure_value": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The environment variable's value, e.g. bar.",
				Sensitive:    true,
				ExactlyOneOf: []string{"value"},
			},
			"public": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Whether this environment variable should be publicly visible or not.",
				Computed:    true,
			},
			"branch": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The env_var's branch.",
			},
		},

		CustomizeDiff: func(d *schema.ResourceDiff, meta interface{}) error {
			value := d.Get("value").(string)
			secure := d.Get("secure_value").(string)
			switch {
			case value != "" && secure == "": // public: true
				d.SetNew("public", true)
				d.SetNew("secure_value", "")
			case secure != "" && value == "": // public: false
				d.SetNew("public", false)
				d.SetNew("value", "")
			case value == "" && secure == "": // If both value and secure_value are empty, public is true
				d.SetNew("public", true)
				d.ForceNew("value")
				d.SetNew("secure_value", "")
			}
			return nil
		},
	}
}

func resourceEnvVarCreate(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*Client)
	var envVar *travis.EnvVar
	if repoID := d.Get("repository_id").(int); repoID > 0 {
		envVar, _, err = client.EnvVars.CreateByRepoId(context.Background(), uint(repoID), generateEnvVarBody(d))
		if err != nil {
			return
		}
	} else if repoSlug := d.Get("repository_slug").(string); repoSlug != "" {
		envVar, _, err = client.EnvVars.CreateByRepoSlug(context.Background(), repoSlug, generateEnvVarBody(d))
		if err != nil {
			return
		}
	} else {
		return errors.New("one of repository_id or repository_slug must be specified.")
	}
	assignEnvVar(envVar, d)
	return nil
}

func resourceEnvVarRead(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*Client)
	var envVar *travis.EnvVar
	if repoID := d.Get("repository_id").(int); repoID > 0 {
		envVar, _, err = client.EnvVars.FindByRepoId(context.Background(), uint(repoID), d.Id())
		if err != nil {
			if IsNotFound(err) {
				d.SetId("")
				return nil
			}
			return
		}
	} else if repoSlug := d.Get("repository_slug").(string); repoSlug != "" {
		envVar, _, err = client.EnvVars.FindByRepoSlug(context.Background(), repoSlug, d.Id())
		if err != nil {
			if IsNotFound(err) {
				d.SetId("")
				return nil
			}
			return
		}
	} else {
		return errors.New("one of repository_id or repository_slug must be specified.")
	}
	assignEnvVar(envVar, d)
	return
}

func resourceEnvVarUpdate(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*Client)
	var envVar *travis.EnvVar
	if repoID := d.Get("repository_id").(int); repoID > 0 {
		envVar, _, err = client.EnvVars.UpdateByRepoId(context.Background(), uint(repoID), d.Id(), generateEnvVarBody(d))
		if err != nil {
			return
		}
	} else if repoSlug := d.Get("repository_slug").(string); repoSlug != "" {
		envVar, _, err = client.EnvVars.UpdateByRepoSlug(context.Background(), repoSlug, d.Id(), generateEnvVarBody(d))
		if err != nil {
			return
		}
	} else {
		return errors.New("one of repository_id or repository_slug must be specified.")
	}
	assignEnvVar(envVar, d)
	return
}

func resourceEnvVarDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	if repoID := d.Get("repository_id").(int); repoID > 0 {
		_, err := client.EnvVars.DeleteByRepoId(context.Background(), uint(repoID), d.Id())
		if err != nil {
			return err
		}
	} else if repoSlug := d.Get("repository_slug").(string); repoSlug != "" {
		_, err := client.EnvVars.DeleteByRepoSlug(context.Background(), repoSlug, d.Id())
		if err != nil {
			return err
		}
	} else {
		return errors.New("one of repository_id or repository_slug must be specified.")
	}
	d.SetId("")
	return nil
}

func generateEnvVarBody(d *schema.ResourceData) *travis.EnvVarBody {
	public := d.Get("public").(bool)
	value := d.Get("value").(string)
	if !public {
		value = d.Get("secure_value").(string)
	}
	return &travis.EnvVarBody{
		Name:   d.Get("name").(string),
		Value:  value,
		Public: public,
		Branch: d.Get("branch").(string),
	}
}

func assignEnvVar(envVar *travis.EnvVar, d *schema.ResourceData) {
	d.SetId(*envVar.Id)
	d.Set("name", envVar.Name)
	if *envVar.Public {
		d.Set("value", envVar.Value)
		d.Set("secure_value", nil)
	} else {
		d.Set("value", nil)
	}
	d.Set("public", envVar.Public)
	d.Set("branch", envVar.Branch)
}
