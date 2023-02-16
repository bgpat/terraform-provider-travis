package travis

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	testAccProviders map[string]*schema.Provider
	testAccProvider  *schema.Provider

	testRepoSlug = os.Getenv("TRAVIS_REPO_SLUG")
	testBranch   = os.Getenv("TRAVIS_BRANCH")

	uuidPattern = regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
)

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"travis": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatal(err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("TRAVIS_TOKEN"); v == "" {
		t.Fatal("TRAVIS_TOKEN must be set for acceptance tests")
	}
	if testRepoSlug == "" {
		t.Fatal("TRAVIS_REPO_SLUG must be set for acceptance tests")
	}
	if testBranch == "" {
		t.Fatal("TRAVIS_BRANCH must be set for acceptance tests")
	}
}
