resource "travis_env_var" "public_value" {
  repository_slug = "bgpat/test"
  name            = "PUBLIC_VALUE"
  public_value    = "public"
}

resource "travis_env_var" "secret_values" {
  for_each        = set("foo", "bar", "buzz")
  repository_slug = "bgpat/test"
  name            = "SECRET_VALUE_${upper(each.key)}"
  value           = each.value
}
