//go:build repoAAP
// +build repoAAP

package configprefix

import "strings"

const Prefix = "aap"

func ReplaceText(input string) string {
	return strings.ReplaceAll(input, "awx_", "aap_")
}

const OrgDataSourceIdDescription = "Organization ID. Be sure this ID is the controller ID, not the gateway ID."
const TeamResourceOrgIdDescription = "Organization ID of the team. This should be the gateway ID of the organization, not the controller ID."
