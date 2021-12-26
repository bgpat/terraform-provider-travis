resource "travis_env_var" "public_value" {
  repository_slug = "bgpat/test"
  name            = "PUBLIC_VALUE"
  public_value    = "public"
}

resource "travis_env_var" "secret_values" {
  for_each      = toset(["foo", "bar", "buzz"])
  repository_id = 2562785
  name          = "SECRET_VALUE_${upper(each.key)}"
  value         = each.value
}
