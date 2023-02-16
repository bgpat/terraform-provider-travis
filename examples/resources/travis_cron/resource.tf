resource "travis_cron" "main_daily" {
  repository_slug = "bgpat/test"
  branch          = "main"
  interval        = "daily"
}
