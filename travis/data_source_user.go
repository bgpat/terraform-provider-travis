package travis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shuheiktgw/go-travis"
)

const syncInitialInterval = 10 * time.Second

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get the user resource.",

		ReadContext: dataSourceTravisRead,

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Value uniquely identifying the user. If not set, get the current user.",
			},
			"include": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Include additional attributes. To know available values, see https://developer.travis-ci.com/resource/user.",
			},
			"wait_sync": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If true, exec sync user API and wait.",
			},

			"login": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Login set on GitHub.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name set on GitHub.",
			},
			"github_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID set on GitHub.",
			},
			"avatar_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Avatar URL set on GitHub.",
			},
			"education": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the user has an education account.",
			},
			"is_syncing": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the user is currently being synced with Github.",
			},
			"synced_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last time the user was synced with GitHub.",
			},
			"repositories": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Repositories belonging to this user.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slug": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"emails": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The user's emails.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceTravisRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		client = m.(*Client)
		opt    = &travis.UserOption{}
		userID uint
	)

	if v, ok := d.GetOk("include"); ok {
		for _, i := range v.(*schema.Set).List() {
			opt.Include = append(opt.Include, i.(string))
		}
	}

	waitSync := d.Get("wait_sync").(bool)
	ctx = tflog.SetField(ctx, "waitSync", waitSync)

	userIDValue, hasUserID := d.GetOk("user_id")
	ctx = tflog.SetField(ctx, "hasUserID", hasUserID)
	if hasUserID {
		userID = uint(userIDValue.(int))
	} else {
		user, _, err := client.User.Current(ctx, opt)
		if err != nil {
			return diag.Errorf("failed to get current user: %v", err)
		}
		if !waitSync {
			if err := assignUser(user, d); err != nil {
				return diag.Errorf("failed to set user: %v", err)
			}
			return nil
		}
		if user.Id == nil {
			return diag.Errorf("id is nil")
		}
		userID = *user.Id
	}
	ctx = tflog.SetField(ctx, "userID", userID)

	if waitSync {
		_, _, err := client.User.Sync(ctx, userID)
		if err != nil {
			return diag.Errorf("failed to sync user %v: %v", userID, err)
		}
	}

	eb := backoff.NewExponentialBackOff()
	eb.InitialInterval = syncInitialInterval
	if err := backoff.RetryNotify(func() error {
		user, _, err := client.User.Find(ctx, userID, opt)
		if err != nil {
			return backoff.Permanent(err)
		}
		if user == nil {
			return errors.New("user is nil")
		}
		if waitSync {
			if user.IsSyncing != nil && *user.IsSyncing {
				return errors.New("syncing user")
			}
		}
		return assignUser(user, d)
	}, backoff.WithContext(eb, ctx), func(err error, d time.Duration) {
		tflog.Debug(ctx, "retry to get user", map[string]interface{}{
			"reason": err,
			"sleep":  d,
		})
	}); err != nil {
		return diag.Errorf("failed to get user %v: %v", userID, err)
	}
	return nil
}

func assignUser(user *travis.User, d *schema.ResourceData) error {
	if user.Id != nil {
		d.SetId(strconv.FormatUint(uint64(*user.Id), 10))
	}
	if user.Login != nil {
		if err := d.Set("login", *user.Login); err != nil {
			return err
		}
	}
	if user.Name != nil {
		if err := d.Set("name", *user.Name); err != nil {
			return err
		}
	}
	if user.GithubId != nil {
		if err := d.Set("github_id", *user.GithubId); err != nil {
			return err
		}
	}
	if user.AvatarUrl != nil {
		if err := d.Set("avatar_url", *user.AvatarUrl); err != nil {
			return err
		}
	}
	if user.Education != nil {
		if err := d.Set("education", *user.Education); err != nil {
			return err
		}
	}
	if user.IsSyncing != nil {
		if err := d.Set("is_syncing", *user.IsSyncing); err != nil {
			return err
		}
	}
	if user.Repositories != nil {
		repos := make([]map[string]interface{}, 0, len(user.Repositories))
		for _, repo := range user.Repositories {
			repos = append(repos, map[string]interface{}{
				"id":   float64(*repo.Id),
				"name": *repo.Name,
				"slug": *repo.Slug,
			})
		}
		if err := d.Set("repositories", repos); err != nil {
			return err
		}
	}
	if user.Emails != nil {
		if err := d.Set("emails", user.Emails); err != nil {
			return err
		}
	}
	return nil
}
