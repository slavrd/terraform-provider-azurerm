package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
)

func TestAccAzureRMDatacatalog_basic(t *testing.T) {
	rn := "azurerm_datacatalog.test"
	ri := tf.AccRandTimeInt()
	pw := "p@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMDatacatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMDatacatalog_basic(ri, testLocation(), pw),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDatacatalogExists(rn),
				),
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureRMDatacatalog_requiresImport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	rn := "azurerm_datacatalog.test"
	ri := tf.AccRandTimeInt()
	pw := "p@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMDatacatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMDatacatalog_basic(ri, testLocation(), pw),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDatacatalogExists(rn),
				),
			},
			{
				Config:      testAccAzureRMDatacatalog_requiresImport(ri, testLocation(), pw),
				ExpectError: testRequiresImportError("azurerm_datacatalog"),
			},
		},
	})
}

func testCheckAzureRMDatacatalogExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Bad: Not found: %s", resourceName)
		}

		catalogName := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]
		client := testAccProvider.Meta().(*ArmClient).dataCatalog.CatalogsClient(catalogName)

		resp, err := client.Get(ctx, resourceGroup)
		if err != nil {
			return fmt.Errorf("Bad: Getting Data Catalog: %s (resource group: %s): %v", catalogName, resourceGroup, err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Data Catalog %s (resource group: %s) does not exist", catalogName, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureRMDatacatalogDestroy(s *terraform.State) error {
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_databricks_workspace" {
			continue
		}

		catalogName := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]
		client := testAccProvider.Meta().(*ArmClient).dataCatalog.CatalogsClient(catalogName)

		resp, err := client.Get(ctx, resourceGroup)
		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Bad: Data Catalog still exists:\n%#v", resp.ID)
		}
	}

	return nil
}

func testAccAzureRMDatacatalog_basic(rInt int, location, pw string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

data "azuread_domains" "tenant_domain" {
  only_initial = true
}

resource "azuread_user" "test" {
  user_principal_name = "acctest-DCUser.%[1]d@${data.azuread_domains.tenant_domain.domains.0.domain_name}"
  display_name        = "acctestDCUser-%[1]d"
  password            = "%[3]s"
}

resource "azurerm_datacatalog" "test" {
  name                = "acctest-DC-%[1]d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  sku                 = "Free"

  admin {
    upn = "${azuread_user.test.user_principal_name}"
  }
}
`, rInt, location, pw)
}

func testAccAzureRMDatacatalog_requiresImport(rInt int, location, pw string) string {
	template := testAccAzureRMDatacatalog_basic(rInt, location, pw)
	return fmt.Sprintf(`
%s

resource "azurerm_datacatalog" "import" {
  name                = "$[azurerm_datacatalog.test.name}"
  resource_group_name = "${azurerm_datacatalog.test.resource_group_name}"
  location            = "${azurerm_datacatalog.test.location}"
  sku                 = "${azurerm_datacatalog.test.sku}"
}
`, template)
}
