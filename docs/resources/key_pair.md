---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "travis_key_pair Resource - terraform-provider-travis"
subcategory: ""
description: |-
  The travis_key_pair resource manages an RSA key pair for a repo.
---

# travis_key_pair (Resource)

The `travis_key_pair` resource manages an RSA key pair for a repo.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **description** (String) A text description of this key pair.
- **value** (String, Sensitive) The private key

### Optional

- **id** (String) The ID of this resource.
- **repository_id** (Number) Value uniquely identifying the repository.
- **repository_slug** (String) Same as {repository.owner.name}/{repository.name}.

### Read-Only

- **fingerprint** (String) Fingerprint of the RSA key
- **public_key** (String) The public key.

