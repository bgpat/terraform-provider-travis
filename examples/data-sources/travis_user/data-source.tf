# get current user
data "travis_user" "current" {}

# get user by user_id
data "travis_user" "by_user_id" {
  user_id = 190625
}

# sync and get user
data "travis_user" "sync" {
  wait_sync = true
}

# get user with repositories
data "travis_user" "with_repos" {
  include = ["user.repositories"]
}
