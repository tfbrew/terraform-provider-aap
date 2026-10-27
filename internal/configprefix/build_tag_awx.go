//go:build repoAWX

package configprefix

const Prefix = "awx"
const OrgDataSourceIdDescription = "Organization ID."
const TeamResourceOrgIdDescription = OrgDataSourceIdDescription

// GenericEndpointApiEndpoints are the accepted values of the generic_endpoint resource's api_endpoint attribute.
var GenericEndpointApiEndpoints = []string{"controller"}

const GenericEndpointApiEndpointDescription = "API endpoint for the GenericEndpoint. Must be `controller`, as all AWX/AAP2.4 endpoints are served from /api/v2/."

// EnvPrefix is the prefix of the environment variables the provider reads, e.g. TOWER_HOST.
const EnvPrefix = "TOWER_"
