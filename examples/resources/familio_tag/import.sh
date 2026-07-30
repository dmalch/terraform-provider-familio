# familio_tag is imported by the tag's numeric id — NOT a uuid. Tags are the one
# familio resource keyed by a small sequential integer; `familio tags list` (from
# the go-familio CLI) prints the account's tag ids.
terraform import familio_tag.to_verify 2832
