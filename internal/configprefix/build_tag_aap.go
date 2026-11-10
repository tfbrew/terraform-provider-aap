//go:build repoAAP

package configprefix

const Prefix = "aap"
const OrgDataSourceIdDescription = "Organization ID. Be sure this ID is the controller ID, not the gateway ID."
const TeamResourceOrgIdDescription = "Organization ID of the team. This should be the gateway ID of the organization, not the controller ID."

// GenericEndpointApiEndpoints are the accepted values of the generic_endpoint resource's api_endpoint attribute.
var GenericEndpointApiEndpoints = []string{"eda", "galaxy", "gateway", "controller"}

const GenericEndpointApiEndpointDescription = "API endpoint for the GenericEndpoint. If the URL is /api/gateway/v1/authenticators/, the api_endpoint is gateway. Must be one of `eda`, `galaxy`, `gateway`, `controller`."

// EnvPrefix is the prefix of the environment variables the provider reads, e.g. AAP_HOST.
const EnvPrefix = "AAP_"
