package travis

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shuheiktgw/go-travis"
)

func resourceEnvVar() *schema.Resource {
	return &schema.Resource{
		Description: "The `travis_env_var` resource can create an environment variable.",

		CreateContext: resourceEnvVarCreate,
		ReadContext:   resourceEnvVarRead,
		UpdateContext: resourceEnvVarUpdate,
		DeleteContext: resourceEnvVarDelete,

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
			"public_value": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The environment variable's value, e.g. bar.",
				ExactlyOneOf: []string{"value"},
			},
			"value": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The environment variable's value, e.g. bar.",
				Sensitive:    true,
				ExactlyOneOf: []string{"public_value"},
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

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			publicValue := d.Get("public_value").(string)
			value := d.Get("value").(string)
			switch {
			case publicValue != "" && value == "": // public: true
				d.SetNew("public", true)
				d.SetNew("value", "")
			case publicValue != "" && value == "": // public: false
				d.SetNew("public", false)
				d.SetNew("value", "")
			case publicValue == "" && value == "": // If both public_value and value are empty, public is true
				d.SetNew("public", true)
				d.SetNew("public_value", "")
				d.ForceNew("value")
			}
			return nil
		},
	}
}

func resourceEnvVarCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		client = m.(*Client)
		envVar *travis.EnvVar
		err    error
	)
	if repoID := d.Get("repository_id").(int); repoID > 0 {
		envVar, _, err = client.EnvVars.CreateByRepoId(ctx, uint(repoID), generateEnvVarBody(d))
		if err != nil {
			return diag.Errorf("error creating env var by repo ID (%d): %s", repoID, err)
		}
	} else if repoSlug := d.Get("repository_slug").(string); repoSlug != "" {
		envVar, _, err = client.EnvVars.CreateByRepoSlug(ctx, repoSlug, generateEnvVarBody(d))
		if err != nil {
			return diag.Errorf("error creating env var by repo slug (%s): %s", repoSlug, err)
		}
	} else {
		return diag.Errorf("one of repository_id or repository_slug must be specified")
	}
	assignEnvVar(envVar, d)
	return nil
}

func resourceEnvVarRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		client = m.(*Client)
		envVar *travis.EnvVar
		err    error
	)
	if repoID := d.Get("repository_id").(int); repoID > 0 {
		envVar, _, err = client.EnvVars.FindByRepoId(ctx, uint(repoID), d.Id())
		if err != nil {
			if IsNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error reading env var by repo ID (%d) and ID (%s): %s", repoID, d.Id(), err)
		}
	} else if repoSlug := d.Get("repository_slug").(string); repoSlug != "" {
		envVar, _, err = client.EnvVars.FindByRepoSlug(ctx, repoSlug, d.Id())
		if err != nil {
			if IsNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error reading env var by repo slug (%s) and ID (%s): %s", repoSlug, d.Id(), err)
		}
	} else {
		return diag.Errorf("one of repository_id or repository_slug must be specified")
	}
	assignEnvVar(envVar, d)
	return nil
}

func resourceEnvVarUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		client = m.(*Client)
		envVar *travis.EnvVar
		err    error
	)
	if repoID := d.Get("repository_id").(int); repoID > 0 {
		envVar, _, err = client.EnvVars.UpdateByRepoId(ctx, uint(repoID), d.Id(), generateEnvVarBody(d))
		if err != nil {
			return diag.Errorf("error updating env var by repo ID (%d) and ID (%s): %s", repoID, d.Id(), err)
		}
	} else if repoSlug := d.Get("repository_slug").(string); repoSlug != "" {
		envVar, _, err = client.EnvVars.UpdateByRepoSlug(ctx, repoSlug, d.Id(), generateEnvVarBody(d))
		if err != nil {
			return diag.Errorf("error updating env var by repo slug (%s) and ID (%s): %s", repoSlug, d.Id(), err)
		}
	} else {
		return diag.Errorf("one of repository_id or repository_slug must be specified")
	}
	assignEnvVar(envVar, d)
	return nil
}

func resourceEnvVarDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	if repoID := d.Get("repository_id").(int); repoID > 0 {
		_, err := client.EnvVars.DeleteByRepoId(ctx, uint(repoID), d.Id())
		if err != nil {
			return diag.Errorf("error deleting env var by repo ID (%d) and ID (%s): %s", repoID, d.Id(), err)
		}
	} else if repoSlug := d.Get("repository_slug").(string); repoSlug != "" {
		_, err := client.EnvVars.DeleteByRepoSlug(ctx, repoSlug, d.Id())
		if err != nil {
			return diag.Errorf("error deleting env var by repo slug (%s) and ID (%s): %s", repoSlug, d.Id(), err)
		}
	} else {
		return diag.Errorf("one of repository_id or repository_slug must be specified")
	}
	d.SetId("")
	return nil
}

func generateEnvVarBody(d *schema.ResourceData) *travis.EnvVarBody {
	public := d.Get("public").(bool)
	value := d.Get("value").(string)
	if public {
		value = d.Get("public_value").(string)
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
		d.Set("public_value", envVar.Value)
		d.Set("value", nil)
	} else {
		d.Set("public_value", nil)
		d.Set("value", envVar.Value)
	}
	d.Set("public", envVar.Public)
	d.Set("branch", envVar.Branch)
}
