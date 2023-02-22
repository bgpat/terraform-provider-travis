package travis_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/bgpat/terraform-provider-travis/travis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	testAccProviders map[string]*schema.Provider
	testAccProvider  *schema.Provider

	testRepoSlug  = os.Getenv("TRAVIS_REPO_SLUG")
	testBranch    = os.Getenv("TRAVIS_BRANCH")
	testUserID    = os.Getenv("TRAVIS_USER_ID")
	testUserLogin = os.Getenv("TRAVIS_USER_LOGIN")

	uuidPattern = regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
)

func init() {
	testAccProvider = travis.Provider()
	testAccProviders = map[string]*schema.Provider{
		"travis": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := travis.Provider().InternalValidate(); err != nil {
		t.Fatal(err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = travis.Provider()
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if v := os.Getenv("TRAVIS_TOKEN"); v == "" {
		t.Fatal("TRAVIS_TOKEN must be set for acceptance tests")
	}
	if testRepoSlug == "" {
		t.Fatal("TRAVIS_REPO_SLUG must be set for acceptance tests")
	}
	if testBranch == "" {
		t.Fatal("TRAVIS_BRANCH must be set for acceptance tests")
	}
	if testUserID == "" {
		t.Fatal("TRAVIS_USER_ID must be set for acceptance tests")
	}
	if testUserLogin == "" {
		t.Fatal("TRAVIS_USER_LOGIN must be set for acceptance tests")
	}
}
