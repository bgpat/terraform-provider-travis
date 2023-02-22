package travis

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shuheiktgw/go-travis"
)

func resourceEnvVar() *schema.Resource {
	return &schema.Resource{
		Description: "The `travis_env_var` resource can create an environment variable.",

		CreateContext: resourceEnvVarCreate,
		ReadContext:   resourceEnvVarRead,
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
				ForceNew:    true,
			},
			"public_value": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The environment variable's value, e.g. bar.",
				ExactlyOneOf: []string{"value"},
				ForceNew:     true,
			},
			"value": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The environment variable's value, e.g. bar.",
				Sensitive:    true,
				ExactlyOneOf: []string{"public_value"},
				ForceNew:     true,
			},
			"public": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Whether this environment variable should be publicly visible or not.",
				Computed:    true,
				ForceNew:    true,
			},
			"branch": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The env_var's branch.",
				ForceNew:    true,
			},
		},

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			publicValue := d.Get("public_value").(string)
			value := d.Get("value").(string)
			switch {
			case publicValue != "" && value == "": // public: true
				if err := d.SetNew("public", true); err != nil {
					return err
				}
				if err := d.SetNew("value", nil); err != nil {
					return err
				}
			case value != "" && publicValue == "": // public: false
				if err := d.SetNew("public", false); err != nil {
					return err
				}
				if err := d.SetNew("public_value", nil); err != nil {
					return err
				}
			}
			return nil
		},

		Importer: &schema.ResourceImporter{
			StateContext: importEnvVar,
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
	if err := assignEnvVar(envVar, d); err != nil {
		return diag.Errorf("failed to assign env_var: %v", err)
	}
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
			if isNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error reading env var by repo ID (%d) and ID (%s): %s", repoID, d.Id(), err)
		}
	} else if repoSlug := d.Get("repository_slug").(string); repoSlug != "" {
		envVar, _, err = client.EnvVars.FindByRepoSlug(ctx, repoSlug, d.Id())
		if err != nil {
			if isNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error reading env var by repo slug (%s) and ID (%s): %s", repoSlug, d.Id(), err)
		}
	} else {
		return diag.Errorf("one of repository_id or repository_slug must be specified")
	}
	if err := assignEnvVar(envVar, d); err != nil {
		return diag.Errorf("failed to assign env_var: %v", err)
	}
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

	if value == "" {
		public = true
	}

	return &travis.EnvVarBody{
		Name:   d.Get("name").(string),
		Value:  value,
		Public: public,
		Branch: d.Get("branch").(string),
	}
}

func assignEnvVar(envVar *travis.EnvVar, d *schema.ResourceData) error {
	d.SetId(*envVar.Id)
	if err := d.Set("name", envVar.Name); err != nil {
		return err
	}
	if *envVar.Public {
		if err := d.Set("public_value", envVar.Value); err != nil {
			return err
		}
		if err := d.Set("value", nil); err != nil {
			return err
		}
	} else {
		if err := d.Set("public_value", nil); err != nil {
			return err
		}
	}
	if err := d.Set("public", envVar.Public); err != nil {
		return err
	}
	if err := d.Set("branch", envVar.Branch); err != nil {
		return err
	}
	return nil
}

func importEnvVar(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*Client)

	args := strings.Split(d.Id(), "/")
	if len(args) <= 1 {
		return nil, fmt.Errorf("expected format is \"<repository>/<name>\", but got invalid: %q", d.Id())
	}
	repo := strings.Join(args[:len(args)-1], "/")
	name := args[len(args)-1]

	envVars, _, err := client.EnvVars.ListByRepoSlug(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("error listing env vars of repo (%q): %w", repo, err)
	}

	for _, envVar := range envVars {
		if *envVar.Name == name {
			if err := assignEnvVar(envVar, d); err != nil {
				return nil, err
			}
			if repoID, err := strconv.Atoi(repo); err == nil {
				if err := d.Set("repository_id", repoID); err != nil {
					return nil, err
				}
			} else {
				if err := d.Set("repository_slug", repo); err != nil {
					return nil, err
				}
			}
			return []*schema.ResourceData{d}, nil
		}
	}
	return nil, fmt.Errorf("not found env var %q from repo %q", name, repo)
}
