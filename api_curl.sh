#! /bin/bash

# curl --location 'https://aap25-provider-test.morford.dev/api/controller/v2/credential_types/?page_size=50' \
#      --header   "Authorization: Bearer $TOWER_OAUTH_TOKEN" | \
# jq -r '.results[].inputs.fields? // [] '
# jq -r '.results[].inputs.fields? // [] | .[].id' | sort | uniq    # This one says if fields is null use empty [] instead. 




# curl --location 'https://aap25-provider-test.morford.dev/api/controller/v2/credential_types/?page_size=50' \
#   --header "Authorization: Bearer $TOWER_OAUTH_TOKEN" \
# | jq -r '.results[] | .id as $id | (.inputs.fields? // [])[] | "\($id) \"\(.type)\""'

curl --location 'https://aap25-provider-test.morford.dev/api/controller/v2/credentials/?page_size=50' \
  --header "Authorization: Bearer $TOWER_OAUTH_TOKEN" \
| jq -r '.results[] | .id as $id | .inputs , $id '

# Has input of type boolean
# curl --location 'https://aap25-provider-test.morford.dev/api/controller/v2/credentials/?page_size=50&id=1774' \
#   --header "Authorization: Bearer $TOWER_OAUTH_TOKEN" \
# | jq -r '.results[] | .id as $id | .inputs , $id '