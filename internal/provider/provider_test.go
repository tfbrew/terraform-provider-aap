// SPECIAL: Environment variables
package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	configprefix.Prefix: providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("AAP_HOST"); v == "" {
		t.Fatal("AAP_HOST must be set for acceptance tests")
	}
	if os.Getenv("AAP_OAUTH_TOKEN") == "" && os.Getenv("AAP_TOKEN") == "" && (os.Getenv("AAP_USERNAME") == "" || os.Getenv("AAP_PASSWORD") == "") {
		t.Fatal("AAP_OAUTH_TOKEN, AAP_TOKEN, or AAP_USERNAME/AAP_PASSWORD must be set for acceptance tests")
	}
}

func testAccProviderClient() (*providerClient, error) {
	endpoint := os.Getenv("AAP_HOST")
	if endpoint == "" {
		return nil, fmt.Errorf("AAP_HOST must be set")
	}

	if endpoint[len(endpoint)-1] == '/' {
		endpoint = endpoint[:len(endpoint)-1]
	}

	var auth string
	if token := os.Getenv("AAP_TOKEN"); token != "" {
		auth = "Bearer " + token
	} else if token := os.Getenv("AAP_OAUTH_TOKEN"); token != "" {
		auth = "Bearer " + token
	} else if username := os.Getenv("AAP_USERNAME"); username != "" {
		if password := os.Getenv("AAP_PASSWORD"); password != "" {
			auth = "Basic " + username + ":" + password
		}
	}

	if auth == "" {
		return nil, fmt.Errorf("AAP_TOKEN, AAP_OAUTH_TOKEN, or AAP_USERNAME/AAP_PASSWORD must be set")
	}

	httpclient := &http.Client{
		Timeout: 30 * time.Second,
	}

	if envInsecure := os.Getenv("AAP_INSECURE_SKIP_VERIFY"); envInsecure == "true" || envInsecure == "1" {
		httpclient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	client := &providerClient{
		client:   httpclient,
		endpoint: endpoint,
		auth:     auth,
	}

	switch configprefix.Prefix {
	case "awx":
		client.urlPrefix = "/api/v2/"
	case "aap":
		client.urlPrefix = "/api/controller/v2/"
	}

	return client, nil
}

// testAccCheckResourceDisappears is a helper function for disappears tests.
// It deletes the resource using the API and is used to verify that Terraform
// properly handles resources that are deleted outside of Terraform.
//
// Parameters:
//   - resourceName: The terraform resource name (e.g., "aap_eda_project.test")
//   - deleteFunc: A function that performs the actual deletion via API
//
// Example usage:
//
//	func testAccCheckEdaProjectDisappears(resourceName string) resource.TestCheckFunc {
//	    return testAccCheckResourceDisappears(resourceName, func(ctx context.Context, client *providerClient, id string) error {
//	        url := fmt.Sprintf("api/eda/v1/projects/%s/", id)
//	        _, _, err := client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204, 404}, "eda")
//	        return err
//	    })
//	}
func testAccCheckResourceDisappears(resourceName string, deleteFunc func(context.Context, *providerClient, string) error) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		client, err := testAccProviderClient()
		if err != nil {
			return fmt.Errorf("Error creating test client: %w", err)
		}

		ctx := context.Background()
		return deleteFunc(ctx, client, rs.Primary.ID)
	}
}
