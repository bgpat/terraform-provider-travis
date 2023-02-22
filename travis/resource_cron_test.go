package travis_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/shuheiktgw/go-travis"

	tptravis "github.com/bgpat/terraform-provider-travis/travis"
)

func TestAccResourceCron_basic(t *testing.T) {
	var cron travis.Cron

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCronResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCronResource(testBranch, "daily", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCronResourceExists(&cron),
					resource.TestCheckResourceAttr("travis_cron.foo", "repository_slug", testRepoSlug),
					resource.TestCheckResourceAttr("travis_cron.foo", "branch", testBranch),
					resource.TestCheckResourceAttr("travis_cron.foo", "interval", "daily"),
					resource.TestCheckResourceAttr("travis_cron.foo", "dont_run_if_recent_build_exists", "false"),
				),
			},
			{
				Config: testAccCronResource(testBranch, "weekly", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCronResourceExists(&cron),
					resource.TestCheckResourceAttr("travis_cron.foo", "repository_slug", testRepoSlug),
					resource.TestCheckResourceAttr("travis_cron.foo", "branch", testBranch),
					resource.TestCheckResourceAttr("travis_cron.foo", "interval", "weekly"),
					resource.TestCheckResourceAttr("travis_cron.foo", "dont_run_if_recent_build_exists", "false"),
				),
			},
			{
				Config: testAccCronResource(testBranch, "monthly", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCronResourceExists(&cron),
					resource.TestCheckResourceAttr("travis_cron.foo", "repository_slug", testRepoSlug),
					resource.TestCheckResourceAttr("travis_cron.foo", "branch", testBranch),
					resource.TestCheckResourceAttr("travis_cron.foo", "interval", "monthly"),
					resource.TestCheckResourceAttr("travis_cron.foo", "dont_run_if_recent_build_exists", "false"),
				),
			},
			{
				Config: testAccCronResource(testBranch, "daily", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCronResourceExists(&cron),
					resource.TestCheckResourceAttr("travis_cron.foo", "repository_slug", testRepoSlug),
					resource.TestCheckResourceAttr("travis_cron.foo", "branch", testBranch),
					resource.TestCheckResourceAttr("travis_cron.foo", "interval", "daily"),
					resource.TestCheckResourceAttr("travis_cron.foo", "dont_run_if_recent_build_exists", "true"),
				),
			},
		},
	})
}

func testAccCheckCronResourceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*tptravis.Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "travis_cron" {
			continue
		}
		id, err := strconv.ParseUint(rs.Primary.ID, 10, 64)
		if err != nil {
			return err
		}
		cron, _, err := client.Crons.Find(context.Background(), uint(id), &travis.CronOption{})
		if err == nil && cron != nil {
			return fmt.Errorf("cron %v still exists", rs.Primary.ID)
		}
		if err != nil && !tptravis.IsNotFound(err) {
			return err
		}
		return nil
	}
	return nil
}

func testAccCheckCronResourceExists(cron *travis.Cron) resource.TestCheckFunc {
	resourceName := "travis_cron.foo"
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("cron ID is not set")
		}
		client := testAccProvider.Meta().(*travis.Client)
		id, err := strconv.ParseUint(rs.Primary.ID, 10, 64)
		if err != nil {
			return err
		}
		result, _, err := client.Crons.Find(context.Background(), uint(id), &travis.CronOption{})
		if err != nil {
			return err
		}
		*cron = *result
		return nil
	}
}

func testAccCronResource(branch, interval string, drirbe bool) string {
	return fmt.Sprintf(`
resource "travis_cron" "foo" {
	repository_slug                 = %q
	branch                          = %q
	interval                        = %q
	dont_run_if_recent_build_exists = %t
}
`, testRepoSlug, branch, interval, drirbe)
}
