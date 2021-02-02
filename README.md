# Terraform Provider for Travis CI

![release](https://github.com/bgpat/terraform-provider-travis/workflows/release/badge.svg)

https://registry.terraform.io/providers/bgpat/travis/latest

## Example

```terraform
variable "travis_api_token" {}

provider "travis" {
  com   = true
  token = var.travis_api_token
}

resource "travis_env_var" "foo" {
  repository_slug = "bgpat/test"
  name            = "FOO"
  value           = "foo"
}

resource "travis_env_var" "bar" {
  repository_slug = "bgpat/test"
  name            = "BAR"
  secure_value    = "foo"
}
```

- `org`, `com`, `url` - API endopint for Travis CI
  - `org` (bool) - open source projects on [travis-ci.org](https://travis-ci.org/)
  - `com` (bool) - private projects on [travis-ci.com](https://travis-ci.com/)
  - `url` (string) - enterprise projects on [a custom domain](https://enterprise.travis-ci.com/)
- `token` - API token [generated by the Travis CI client](https://developer.travis-ci.com/authentication).

## Resources

- `travis_env_var` - https://docs.travis-ci.com/user/environment-variables/

Check details of schema with `terraform providers schema`.
