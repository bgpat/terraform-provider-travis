# ${repository_id}/${name}
terraform import travis_env_var.public_value bgpat/test/PUBLIC_VALUE

# ${repository_slug}/${name}
terraform import 'travis_env_var.secret_values["foo"]' 2562785/SECRET_VALUE_FOO
