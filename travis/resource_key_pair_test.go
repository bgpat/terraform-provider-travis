package travis

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/shuheiktgw/go-travis"
	"golang.org/x/crypto/ssh"
)

func TestAccResourceKeyPair_basic(t *testing.T) {
	var keyPair travis.KeyPair
	testAccPrivateKey, testAccPublicKey, testAccFingerprint := makeKeyPair(t)
	desc := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKeyPairResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairResource(desc, testAccPrivateKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPairResourceExists("travis_key_pair.foo", &keyPair),
					resource.TestCheckResourceAttr("travis_key_pair.foo", "repository_slug", testRepoSlug),
					resource.TestCheckResourceAttr("travis_key_pair.foo", "description", desc),
					resource.TestCheckResourceAttr("travis_key_pair.foo", "fingerprint", testAccFingerprint),
					resource.TestCheckResourceAttr("travis_key_pair.foo", "public_key", testAccPublicKey),
				),
			},
		},
	})
}

func testAccCheckKeyPairResourceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "travis_key_pair" {
			continue
		}
		slug := rs.Primary.Attributes["repository_slug"]
		keyPair, _, err := client.KeyPair.FindByRepoSlug(context.Background(), slug)
		if err == nil && keyPair != nil {
			return fmt.Errorf("key pair %q still exists", slug)
		}
		if err != nil && !isNotFound(err) {
			return err
		}
		return nil
	}
	return nil
}

func testAccCheckKeyPairResourceExists(resourceName string, keyPair *travis.KeyPair) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("key pair is not set")
		}
		client := testAccProvider.Meta().(*Client)
		result, _, err := client.KeyPair.FindByRepoSlug(context.Background(), testRepoSlug)
		if err != nil {
			return err
		}
		*keyPair = *result
		return nil
	}
}

func testAccKeyPairResource(desc, testAccPrivateKey string) string {
	return fmt.Sprintf(`
resource "travis_key_pair" "foo" {
	repository_slug = %q
	description     = %q
	value           = %q
}
`, testRepoSlug, desc, testAccPrivateKey)
}

func makeKeyPair(t *testing.T) (string, string, string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(fmt.Errorf("Cannot generate RSA key\n"))
	}
	publicKey := &privateKey.PublicKey
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		t.Fatal(err)
	}

	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	return string(pem.EncodeToMemory(privateKeyBlock)), string(pem.EncodeToMemory(pubKeyBlock)), ssh.FingerprintLegacyMD5(sshPublicKey)
}
