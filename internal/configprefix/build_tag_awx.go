//go:build repoAWX

package configprefix

const Prefix = "awx"
const OrgDataSourceIdDescription = "Organization ID."
const TeamResourceOrgIdDescription = OrgDataSourceIdDescription

// EnvPrefix is the prefix of the environment variables the provider reads, e.g. TOWER_HOST.
const EnvPrefix = "TOWER_"
