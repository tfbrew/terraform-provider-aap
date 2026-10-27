<!-- AAP-AWX-DIFF-BEGIN: This readme should be altered to match each repo to which it is included. -->
# Terraform Provider for Ansible Automation Platform (AAP) version 2.6 & 2.7

Please see our other provider for [AWX/Tower and AAP verison 2.4.](https://github.com/tfbrew/terraform-provider-aap)

This is a terraform provider for AAP >= 2.6 built with the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) and based on the [Terraform Provider Scafolding Framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework). While this provider will also 99% work with AAP2.5, now that both AAP2.6 & AAP2.7 have been released, priority will be given to new features and supporting API changes instead of testing functionality with 2.5.
<!-- AAP-AWX-DIFF-END -->

If you find any bugs or have a feature request, please open a GitHub issue.

## Use the GNUmakefile

This repo has modified the GNUmake file inherited from the Terraform scaffold repo to ensure commands work with the internal/configprefix necessecity to define one of two build tags: repoAAP or repoAWX.

## Code sharing

This code is used for two different providers:

- tfbrew/aap: Supports aap2.6 and greater (along with limited support for 2.5).
- tfbrew/awx: Supports awx and aap2.4.

**This requires manual intervention & using Go build tags to make the code as simliar as possible accross all three repositories.**

## Build Tags

This repo has 2 build tags: repoAWX and repoAAP. This is so that this code can be used for 2 differently named repositories & Terraform Providers.

The scaffold template's GNUmakefile has been altered to include refencing these tags. Therefore, use the `make` commands to self-compile instead of just using `go` raw. For example, run `make install` instead fo `go install`.

For **VS Code** this repo includes the workspace files `.vscode/settings.json` and `.vscode/launch.json` that set repo-specific flags necessary for linters and debugging with the appropriate build tag set.

## Code Handling for Each Repo

Apart from the Go module path, every file is kept identical between this repository and its sibling (terraform-provider-aap / terraform-provider-awx). Lines that must differ are wrapped in `AAP-AWX-DIFF` begin/end comments, and files that exist in only one repository carry an `AAP-AWX-DIFF` file comment. Search for the phrase `AAP-AWX-DIFF` to find them. The maintainers check for unmarked drift with the `drift-check.sh` script in the tfbrew helpers repository.

## Writing Acceptance Tests

When writing acceptance test, you often have to write Terraform HCL code. Make sure to write your embedded HCL such that it will use the configprefix.Prefix to prefix the resource ID properly.

For example:

```go
return fmt.Sprintf(`
resource "%[1]s_credential_type" "test-name" {
  name         = "%[2]s"
  description  = "%[2]s"
}
data "%[1]s_credential_type" "test-name" {
  name = %[1]s_credential_type.test-name.name
  kind = %[1]s_credential_type.test-name.kind
}
  `,configprefix.Prefix, resource.Name, resource.Description)
}
```
