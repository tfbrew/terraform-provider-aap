//go:build repoAWX
// +build repoAWX

package configprefix

import "strings"

const Prefix = "awx"

func ReplaceText(input string) string {
	return strings.ReplaceAll(input, "aap_", "awx_")
}

const OrgDataSourceIdDescription = "Organization ID."
const TeamResourceOrgIdDescription = OrgDataSourceIdDescription
