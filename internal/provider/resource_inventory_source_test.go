package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestAccInventorySourceResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	rName2 := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	IdCompare := &compareTwoValuesAsStrings{}
	inventory_source1 := InventorySourceAPIModel{
		Name:                 "test-inventory-source-" + acctest.RandString(5),
		Description:          "Example description 1",
		Source:               "scm",
		SourcePath:           "test",
		ExecutionEnvironment: 1,
		Overwrite:            true,
		OverwriteVars:        true,
		UpdateOnLaunch:       true,
	}

	inventory_source2 := InventorySourceAPIModel{
		Name:        "test-inventory-source-" + acctest.RandString(5),
		Description: "Example description 1",
		Source:      "scm",
		SourcePath:  "test",
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInventorySourceResource1Config(inventory_source1, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(inventory_source1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact(inventory_source1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName),
						tfjsonpath.New("source"),
						knownvalue.StringExact(inventory_source1.Source),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName),
						tfjsonpath.New("source_path"),
						knownvalue.StringExact(inventory_source1.SourcePath),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName),
						tfjsonpath.New("overwrite"),
						knownvalue.Bool(inventory_source1.Overwrite),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName),
						tfjsonpath.New("overwrite_vars"),
						knownvalue.Bool(inventory_source1.OverwriteVars),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName),
						tfjsonpath.New("update_on_launch"),
						knownvalue.Bool(inventory_source1.UpdateOnLaunch),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName),
						tfjsonpath.New("source_project"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInventorySourceResource2Config(inventory_source2, rName2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("name"),
						knownvalue.StringExact(inventory_source2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("description"),
						knownvalue.StringExact(inventory_source2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("source"),
						knownvalue.StringExact(inventory_source2.Source),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("source_path"),
						knownvalue.StringExact(inventory_source2.SourcePath),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("overwrite"),
						knownvalue.Bool(inventory_source2.Overwrite),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("overwrite_vars"),
						knownvalue.Bool(inventory_source2.OverwriteVars),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("update_on_launch"),
						knownvalue.Bool(inventory_source2.UpdateOnLaunch),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_inventory_source.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("source_project"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccInventorySourceResource1Config(resource InventorySourceAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[11]s" {
  name        = "%[2]s"
}

resource "%[1]s_project" "%[11]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[11]s.id
  scm_type     = "git"
  scm_url      = "git@github.com:user/repo.git"
}

resource "%[1]s_inventory" "%[11]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[11]s.id
}

resource "%[1]s_inventory_source" "%[11]s" {
  name             		= "%[3]s"
  description	   		= "%[4]s"
  inventory        		= %[1]s_inventory.%[11]s.id
  source           		= "%[5]s"
  source_project   		= %[1]s_project.%[11]s.id
  source_path      		= "%[6]s"
  overwrite        		= %[7]v
  overwrite_vars   		= %[8]v
  update_on_launch 		= %[9]v
  execution_environment = %[10]v
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Source, resource.SourcePath, resource.Overwrite, resource.OverwriteVars, resource.UpdateOnLaunch, resource.ExecutionEnvironment, rName)
}

func testAccInventorySourceResource2Config(resource InventorySourceAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[10]s" {
  name        = "%[2]s"
}

resource "%[1]s_project" "%[10]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[10]s.id
  scm_type     = "git"
  scm_url      = "git@github.com:user/repo.git"
}

resource "%[1]s_inventory" "%[10]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[10]s.id
}

resource "%[1]s_inventory_source" "%[10]s" {
  name             		= "%[3]s"
  description	   		= "%[4]s"
  inventory        		= %[1]s_inventory.%[10]s.id
  source           		= "%[5]s"
  source_project   		= %[1]s_project.%[10]s.id
  source_path      		= "%[6]s"
  overwrite        		= %[7]v
  overwrite_vars   		= %[8]v
  update_on_launch 		= %[9]v
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Source, resource.SourcePath, resource.Overwrite, resource.OverwriteVars, resource.UpdateOnLaunch, rName)
}
