package travis

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/shuheiktgw/go-travis"
)

func TestAccResourceEnvVar_basic(t *testing.T) {
	var envVar travis.EnvVar
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEnvVarResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvVarResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvVarResourceExists("travis_env_var.foo", &envVar),
					resource.TestCheckResourceAttr("travis_env_var.foo", "repository_slug", testRepoSlug),
					resource.TestCheckResourceAttr("travis_env_var.foo", "name", rName),
					resource.TestCheckResourceAttr("travis_env_var.foo", "value", "secret"),
					resource.TestCheckResourceAttr("travis_env_var.foo", "public_value", ""),
					resource.TestCheckResourceAttr("travis_env_var.foo", "public", "false"),
				),
			},
			{
				Config: testAccPublicEnvVarResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvVarResourceExists("travis_env_var.foo", &envVar),
					resource.TestCheckResourceAttr("travis_env_var.foo", "repository_slug", testRepoSlug),
					resource.TestCheckResourceAttr("travis_env_var.foo", "name", rName),
					resource.TestCheckResourceAttr("travis_env_var.foo", "public_value", "public"),
					resource.TestCheckResourceAttr("travis_env_var.foo", "value", ""),
					resource.TestCheckResourceAttr("travis_env_var.foo", "public", "true"),
				),
			},
		},
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("TRAVIS_TOKEN"); v == "" {
		t.Fatal("TRAVIS_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("TRAVIS_REPO_SLUG"); v == "" {
		t.Fatal("TRAVIS_REPO_SLUG must be set for acceptance tests")
	}
}

func testAccCheckEnvVarResourceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "travis_env_var" {
			continue
		}
		slug := rs.Primary.Attributes["repository_slug"]
		envVar, _, err := client.EnvVars.FindByRepoSlug(context.Background(), slug, rs.Primary.ID)
		if err == nil && envVar != nil {
			return fmt.Errorf("env var %q still exists", rs.Primary.Attributes["name"])
		}
		if err != nil && !isNotFound(err) {
			return err
		}
		return nil
	}
	return nil
}

func testAccCheckEnvVarResourceExists(resourceName string, envVar *travis.EnvVar) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("env var ID is not set")
		}
		client := testAccProvider.Meta().(*Client)
		result, _, err := client.EnvVars.FindByRepoSlug(context.Background(), testRepoSlug, rs.Primary.ID)
		if err != nil {
			return err
		}
		*envVar = *result
		return nil
	}
}

func testAccEnvVarResource(name string) string {
	return fmt.Sprintf(`
resource "travis_env_var" "foo" {
	repository_slug = %q
	name            = %q
	value           = "secret"
}
`, testRepoSlug, name)
}

func testAccPublicEnvVarResource(name string) string {
	return fmt.Sprintf(`
resource "travis_env_var" "foo" {
	repository_slug = %q
	name            = %q
	public_value    = "public"
}
`, testRepoSlug, name)
}
