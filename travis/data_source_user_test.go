package travis

import (
	_ "embed"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `data "travis_user" "sync" {
					wait_sync = true
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.travis_user.sync", "id", testUserID),
					resource.TestCheckResourceAttr("data.travis_user.sync", "login", testUserLogin),
					resource.TestCheckResourceAttr("data.travis_user.sync", "is_syncing", "false"),
				),
			},
		},
	})
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
