#! /bin/bash

curl --location "${AAP_HOST}api/controller/v2/credential_types/?page_size=50" \
     --header "Authorization: Bearer ${AAP_OAUTH_TOKEN}" | \
jq -r '.results[].inputs.fields? // []'
# jq -r '.results[].inputs.fields? // [] | .[].id' | sort | uniq    # This one says if fields is null use empty [] instead. 

# curl --location "${AAP_HOST}api/controller/v2/credential_types/?page_size=50" \
#   --header "Authorization: Bearer $AAP_OAUTH_TOKEN" \
# | jq -r '.results[] | .id as $id | (.inputs.fields? // [])[] | "\($id) \"\(.type)\""'

# curl --location "${AAP_HOST}api/controller/v2/credentials/?page_size=50" \
#   --header "Authorization: Bearer $AAP_OAUTH_TOKEN" \
# | jq -r '.results[] | .id as $id | .inputs , $id '

# Has input of type boolean
# curl --location "${AAP_HOST}api/controller/v2/credentials/?page_size=50&id=1774" \
#   --header "Authorization: Bearer $AAP_OAUTH_TOKEN" \
# | jq -r '.results[] | .id as $id | .inputs , $id '