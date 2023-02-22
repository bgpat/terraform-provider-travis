package travis

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/shuheiktgw/go-travis"
)

func resourceCron() *schema.Resource {
	return &schema.Resource{
		Description: "The `travis_cron` resource creates a cron job for a branch.",

		CreateContext: resourceCronCreate,
		ReadContext:   resourceCronRead,
		DeleteContext: resourceCronDelete,

		Schema: map[string]*schema.Schema{
			"repository_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Value uniquely identifying the repository.",
				ForceNew:     true,
				ExactlyOneOf: []string{"repository_slug"},
			},
			"repository_slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Same as {repository.owner.name}/{repository.name}.",
				ForceNew:     true,
				ExactlyOneOf: []string{"repository_id"},
			},
			"branch": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The branch to which this cron job belongs.",
				ForceNew:    true,
			},
			"interval": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Interval at which this cron runs. Can be daily, weekly, or monthly.",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"daily", "weekly", "monthly"}, false),
			},
			"dont_run_if_recent_build_exists": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether a cron build should run if there has been a build on this branch in the last 24 hours.",
				ForceNew:    true,
			},
			"last_run": {
				Type:        schema.TypeString,
				Description: "When the cron ran last.",
				Computed:    true,
				ForceNew:    true,
			},
			"next_run": {
				Type:        schema.TypeString,
				Description: "When the cron is scheduled to run next.",
				Computed:    true,
				ForceNew:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "When the cron was created.",
				Computed:    true,
				ForceNew:    true,
			},
			"active": {
				Type:        schema.TypeBool,
				Description: "Whether the cron is active.",
				Computed:    true,
				ForceNew:    true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: importCron,
		},
	}
}

func resourceCronCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		client = m.(*Client)
		cron   *travis.Cron
		err    error
	)
	if repoID := d.Get("repository_id").(int); repoID > 0 {
		cron, _, err = client.Crons.CreateByRepoId(ctx, uint(repoID), d.Get("branch").(string), generateCronBody(d))
		if err != nil {
			return diag.Errorf("error creating cron by repo ID (%d): %s", repoID, err)
		}
	} else if repoSlug, ok := d.GetOk("repository_slug"); ok {
		cron, _, err = client.Crons.CreateByRepoSlug(ctx, repoSlug.(string), d.Get("branch").(string), generateCronBody(d))
		if err != nil {
			return diag.Errorf("error creating cron by repo slug (%s): %s", repoSlug, err)
		}
	} else {
		return diag.Errorf("one of repository_id or repository_slug must be specified")
	}
	if err := assignCron(cron, d); err != nil {
		return diag.Errorf("failed to assign cron: %v", err)
	}
	return nil
}

func resourceCronRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		client = m.(*Client)
		cron   *travis.Cron
		err    error
	)
	if repoID := d.Get("repository_id").(int); repoID > 0 {
		cron, _, err = client.Crons.FindByRepoId(ctx, uint(repoID), d.Get("branch").(string), nil)
		if err != nil {
			if isNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error reading cron by repo ID (%d) and ID (%s): %s", repoID, d.Id(), err)
		}
	} else if repoSlug, ok := d.GetOk("repository_slug"); ok {
		cron, _, err = client.Crons.FindByRepoSlug(ctx, repoSlug.(string), d.Get("branch").(string), nil)
		if err != nil {
			if isNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("error reading cron by repo slug (%s) and ID (%s): %s", repoSlug, d.Id(), err)
		}
	} else {
		return diag.Errorf("one of repository_id or repository_slug must be specified")
	}
	if err := assignCron(cron, d); err != nil {
		return diag.Errorf("failed to assign cron: %v", err)
	}
	return nil
}

func resourceCronDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	cronID, err := strconv.ParseUint(d.Id(), 10, 64)
	if err != nil {
		return diag.Errorf("failed to convert cron ID to uint: %s", err)
	}
	_, err = client.Crons.Delete(ctx, uint(cronID))
	if err != nil {
		return diag.Errorf("error deleting cron by ID (%s): %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}

func generateCronBody(d *schema.ResourceData) *travis.CronBody {
	return &travis.CronBody{
		DontRunIfRecentBuildExists: d.Get("dont_run_if_recent_build_exists").(bool),
		Interval:                   d.Get("interval").(string),
	}
}

func assignCron(cron *travis.Cron, d *schema.ResourceData) error {
	d.SetId(strconv.FormatUint(uint64(*cron.Id), 10))
	if err := d.Set("branch", cron.Branch.Name); err != nil {
		return err
	}
	if err := d.Set("interval", cron.Interval); err != nil {
		return err
	}
	if err := d.Set("dont_run_if_recent_build_exists", cron.DontRunIfRecentBuildExists); err != nil {
		return err
	}
	if err := d.Set("last_run", cron.LastRun); err != nil {
		return err
	}
	if err := d.Set("next_run", cron.NextRun); err != nil {
		return err
	}
	if err := d.Set("created_at", cron.CreatedAt); err != nil {
		return err
	}
	if err := d.Set("active", cron.Active); err != nil {
		return err
	}
	return nil
}

func importCron(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*Client)

	args := strings.Split(d.Id(), "/")
	if len(args) <= 1 {
		return nil, fmt.Errorf("expected format is \"<repository>/<branch>\", but got invalid: %q", d.Id())
	}
	repo := strings.Join(args[:len(args)-1], "/")
	branch := args[len(args)-1]

	var cron *travis.Cron
	if repoID, err := strconv.Atoi(repo); err == nil {
		cron, _, err = client.Crons.FindByRepoId(ctx, uint(repoID), branch, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting cron in branch (%q) of repo id (%d): %w", branch, repoID, err)
		}
		if err := d.Set("repository_id", repoID); err != nil {
			return nil, err
		}
	} else {
		cron, _, err = client.Crons.FindByRepoSlug(ctx, repo, branch, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting cron in branch (%q) of repo slug (%q): %w", branch, repo, err)
		}
		if err := d.Set("repository_slug", repo); err != nil {
			return nil, err
		}
	}

	if err := assignCron(cron, d); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
