package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestNewEdaProjectDataSource(t *testing.T) {
	testDataSource := NewEdaProjectDataSource()

	if testDataSource == nil {
		t.Error("NewEdaProjectDataSource() returned nil")
	}

	if _, ok := testDataSource.(*EdaProjectDataSource); !ok {
		t.Errorf("Incorrect datasource type returned. Got: %T, wanted: *EdaProjectDataSource", testDataSource)
	}
}

func TestAccEdaProjectDataSource(t *testing.T) {
	rName := "tf-test-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEdaProjectDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s_eda_project.test", configprefix.Prefix), "name", rName),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("data.%s_eda_project.test", configprefix.Prefix), "id"),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("data.%s_eda_project.test", configprefix.Prefix), "url"),
				),
			},
		},
	})
}

func testAccEdaProjectDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "%[2]s_eda_project" "test" {
  name            = "%[1]s"
  description     = "Test EDA project for data source"
  url             = "https://github.com/ansible/ansible-rulebook"
  organization_id = 1
}

data "%[2]s_eda_project" "test" {
  name = %[2]s_eda_project.test.name
}
`, name, configprefix.Prefix)
}
