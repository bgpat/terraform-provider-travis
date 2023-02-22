package travis

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shuheiktgw/go-travis"
)

func resourceKeyPair() *schema.Resource {
	return &schema.Resource{
		Description:   "The `travis_key_pair` resource manages an RSA key pair for a repo.",
		UpdateContext: resourceKeyPairUpdate,
		CreateContext: resourceKeyPairCreate,
		ReadContext:   resourceKeyPairRead,
		DeleteContext: resourceKeyPairDelete,

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
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "A text description of this key pair.",
			},
			"value": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The private key",
				Required:    true,
				Sensitive:   true,
			},
			"fingerprint": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Fingerprint of the RSA key",
				Computed:    true,
			},
			"public_key": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The public key.",
				Computed:    true,
			},
		},
	}
}

func resourceKeyPairCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		client  = m.(*Client)
		keyPair *travis.KeyPair
		err     error
	)
	if repoID, ok := d.GetOk("repository_id"); ok {
		keyPair, _, err = client.KeyPair.CreateByRepoId(ctx, repoID.(uint), generateKeyPairBody(d))
		if err != nil {
			return diag.Errorf("error creating key pair by repo ID (%d): %s", repoID, err)
		}
	} else if repoSlug, ok := d.GetOk("repository_slug"); ok {
		keyPair, _, err = client.KeyPair.CreateByRepoSlug(ctx, repoSlug.(string), generateKeyPairBody(d))
		if err != nil {
			return diag.Errorf("error creating key pair by repo slug (%s): %s", repoSlug, err)
		}
	} else {
		return diag.Errorf("one of repository_id or repository_slug must be specified")
	}
	if err := assignKeyPair(keyPair, d); err != nil {
		return diag.Errorf("failed to assign key_pair: %v", err)
	}
	return nil
}

func resourceKeyPairRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		client  = m.(*Client)
		keyPair *travis.KeyPair
		err     error
	)
	if repoID, ok := d.GetOk("repository_id"); ok {
		keyPair, _, err = client.KeyPair.FindByRepoId(ctx, repoID.(uint))
		if err != nil {
			if isNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error reading key pair by repo ID (%d): %s", repoID, err)
		}
	} else if repoSlug, ok := d.GetOk("repository_slug"); ok {
		keyPair, _, err = client.KeyPair.FindByRepoSlug(ctx, repoSlug.(string))
		if err != nil {
			if isNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error reading key pair by repo slug (%s): %s", repoSlug, err)
		}
	} else {
		return diag.Errorf("one of repository_id or repository_slug must be specified")
	}
	if err := assignKeyPair(keyPair, d); err != nil {
		return diag.Errorf("failed to assign key_pair: %v", err)
	}
	return nil
}

func resourceKeyPairUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		client = m.(*Client)
		err    error
		update = &travis.KeyPairBody{}
	)

	if d.HasChange("value") {
		update.Value = d.Get("value").(string)
	}

	if d.HasChange("description") {
		update.Description = d.Get("description").(string)
	}

	if update.Value != "" || update.Description != "" {
		if repoID, ok := d.GetOk("repository_id"); ok {
			_, _, err = client.KeyPair.UpdateByRepoId(ctx, repoID.(uint), update)
			if err != nil {
				return diag.Errorf("error updating key pair by repo ID (%s).", repoID)
			}
		} else if repoSlug, ok := d.GetOk("repository_slug"); ok {
			_, _, err = client.KeyPair.UpdateByRepoSlug(ctx, repoSlug.(string), update)
			if err != nil {
				return diag.Errorf("error updating key pair by repo slug (%s).", repoSlug)
			}
		}
	}

	return resourceKeyPairRead(ctx, d, m)
}

func resourceKeyPairDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	if repoID, ok := d.GetOk("repository_id"); ok {
		_, err := client.KeyPair.DeleteByRepoId(ctx, repoID.(uint))
		if err != nil {
			return diag.Errorf("error deleting key pair by repo ID (%d): %s", repoID, err)
		}
	} else if repoSlug, ok := d.GetOk("repository_slug"); ok {
		_, err := client.KeyPair.DeleteByRepoSlug(ctx, repoSlug.(string))
		if err != nil {
			return diag.Errorf("error deleting key pair by repo slug (%s): %s", repoSlug, err)
		}
	} else {
		return diag.Errorf("one of repository_id or repository_slug must be specified")
	}
	d.SetId("")
	return nil
}

func generateKeyPairBody(d *schema.ResourceData) *travis.KeyPairBody {
	return &travis.KeyPairBody{
		Description: d.Get("description").(string),
		Value:       d.Get("value").(string),
	}
}

func assignKeyPair(keyPair *travis.KeyPair, d *schema.ResourceData) error {
	if val, ok := d.GetOk("repository_id"); ok {
		d.SetId(val.(string))
	} else if val, ok := d.GetOk("repository_slug"); ok {
		d.SetId(val.(string))
	}
	if err := d.Set("public_key", keyPair.PublicKey); err != nil {
		return err
	}
	if err := d.Set("fingerprint", keyPair.Fingerprint); err != nil {
		return err
	}
	return nil
}
