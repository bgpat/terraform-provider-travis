package travis_test

import (
	"os"
	"testing"

	"github.com/bgpat/terraform-provider-travis/travis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceUser_current(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `data "travis_user" "current" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.travis_user.current", "id", testUserID),
					resource.TestCheckResourceAttr("data.travis_user.current", "login", testUserLogin),
				),
			},
		},
	})
}

func TestAccDataSourceUser_byUserID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `data "travis_user" "by_user_id" {
					user_id = ` + testUserID + `
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.travis_user.by_user_id", "id", testUserID),
					resource.TestCheckResourceAttr("data.travis_user.by_user_id", "login", testUserLogin),
				),
			},
		},
	})
}

func TestAccDataSourceUser_sync(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC to run acceptance tests")
	}
	testAccPreCheck(t)

	ctx := t.Context()
	provider := travis.Provider()
	if diagnostics := provider.Configure(ctx, terraform.NewResourceConfigRaw(nil)); diagnostics.HasError() {
		t.Fatalf("configure Travis provider: %v", diagnostics)
	}

	dataSource := provider.DataSourcesMap["travis_user"]
	data := schema.TestResourceDataRaw(t, dataSource.Schema, map[string]interface{}{
		"wait_sync": true,
	})
	// Terraform reads the data source multiple times during a full test step.
	// Read it once here so the test requests only one live account sync.
	if diagnostics := dataSource.ReadContext(ctx, data, provider.Meta()); diagnostics.HasError() {
		t.Fatalf("read Travis user after sync: %v", diagnostics)
	}
	if got := data.Id(); got != testUserID {
		t.Errorf("user ID = %q, want %q", got, testUserID)
	}
	if got := data.Get("login"); got != testUserLogin {
		t.Errorf("login = %q, want %q", got, testUserLogin)
	}
	if got := data.Get("is_syncing"); got != false {
		t.Errorf("is_syncing = %v, want false", got)
	}
}

func TestAccDataSourceUser_withRepos(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `data "travis_user" "with_repos" {
					include = ["user.repositories"]
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.travis_user.with_repos", "id", testUserID),
					resource.TestCheckResourceAttr("data.travis_user.with_repos", "login", testUserLogin),
					resource.TestCheckTypeSetElemNestedAttrs("data.travis_user.with_repos", "repositories.*", map[string]string{
						"slug": testRepoSlug,
					}),
				),
			},
		},
	})
}
