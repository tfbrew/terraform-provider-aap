package provider

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlParser "net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

// This function below verifies that the authentication actually works after using the new write-only passowrd
//
//	functionality. It includes a test where the password is intentionally changed OUTSIDE of the normal
//	testing apparatus to guarantee is should NOT be able to authenticate. The step will fail if it authentication
//	succeeds.
func TestAccUserResource_passwordWoAuthentication(t *testing.T) {
	username := "test-user-wo-" + acctest.RandString(8)
	passwordV1 := acctest.RandString(20)
	passwordV2 := acctest.RandString(20)
	passwordV3 := acctest.RandString(20)
	passwordOutOfBand := acctest.RandString(20)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourcePasswordWoAuthConfig(username, passwordV1, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccUserCanAuthenticate(username, passwordV1),
				),
			},
			{
				Config: testAccUserResourcePasswordWoAuthConfig(username, passwordV2, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccUserCanAuthenticate(username, passwordV2),
					testAccUserCannotAuthenticate(username, passwordV1),
				),
			},
			{
				Config: testAccUserResourcePasswordWoAuthConfig(username, passwordV3, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccUserCanAuthenticate(username, passwordV3),
					testAccSetUserPasswordOutOfBand(username, passwordOutOfBand),
					testAccUserCannotAuthenticate(username, passwordV3),
					testAccUserCanAuthenticate(username, passwordOutOfBand),
				),
			},
		},
	})
}

func TestAccUserResource(t *testing.T) {
	if configprefix.Prefix == "awx" {
		resource1 := UserAPIModel{
			Username:  "test-user-" + acctest.RandString(5),
			FirstName: "test-firstname",
			LastName:  "test-lastname",
			Email:     "test@example.com",
			Password:  acctest.RandString(20),
		}
		resource2 := UserAPIModel{
			Username:    "test-user-" + acctest.RandString(5),
			FirstName:   "test-firstname2",
			LastName:    "test-lastname2",
			Email:       "test2@example.com",
			Password:    acctest.RandString(20),
			IsSuperuser: true,
		}
		resource3 := UserAPIModel{
			Username:    "test-user-" + acctest.RandString(5),
			FirstName:   "test-firstname",
			LastName:    "test-lastname",
			Email:       "test@example.com",
			Password:    acctest.RandString(20),
			IsSuperuser: true,
		}
		resource4 := UserAPIModel{
			Username:        "test-user-" + acctest.RandString(5),
			FirstName:       "test-firstname",
			LastName:        "test-lastname",
			Email:           "test@example.com",
			Password:        acctest.RandString(20),
			IsSystemAuditor: true,
		}
		resource.Test(t, resource.TestCase{
			PreCheck: func() { testAccPreCheck(t) },
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
			},
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccUserResource1Config(resource1),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource1.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource1.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource1.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource1.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(false),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("is_system_auditor"),
							knownvalue.Bool(false),
						),
					},
				},
				// ImportState testing
				{
					ResourceName:            fmt.Sprintf("%s_user.test", configprefix.Prefix),
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"password"},
				},
				// Update and Read testing
				{
					Config: testAccUserResource2Config(resource2),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource2.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource2.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource2.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource2.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(resource2.IsSuperuser),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("is_system_auditor"),
							knownvalue.Bool(false),
						),
					},
				},
				// Test superuser
				{
					Config: testAccUserResource3Config(resource3),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource3.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource3.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource3.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource3.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("password"),
							knownvalue.StringExact(resource3.Password),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(resource3.IsSuperuser),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("is_system_auditor"),
							knownvalue.Bool(resource3.IsSystemAuditor),
						),
					},
				},
				// Test system auditor
				{
					Config: testAccUserResource4Config(resource4),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource4.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource4.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource4.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource4.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("password"),
							knownvalue.StringExact(resource4.Password),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(resource4.IsSuperuser),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("is_system_auditor"),
							knownvalue.Bool(resource4.IsSystemAuditor),
						),
					},
				},
			},
		})
	} else {
		resource1 := UserAPIModel{
			Username:  "test-user-" + acctest.RandString(5),
			FirstName: "test-firstname",
			LastName:  "test-lastname",
			Email:     "test@example.com",
			Password:  acctest.RandString(20),
		}
		resource2 := UserAPIModel{
			Username:    "test-user-" + acctest.RandString(5),
			FirstName:   "test-firstname2",
			LastName:    "test-lastname2",
			Email:       "test2@example.com",
			Password:    acctest.RandString(20),
			IsSuperuser: true,
		}
		resource3 := UserAPIModel{
			Username:    "test-user-" + acctest.RandString(5),
			FirstName:   "test-firstname",
			LastName:    "test-lastname",
			Email:       "test@example.com",
			Password:    acctest.RandString(20),
			IsSuperuser: true,
		}
		resource4 := UserAPIModel{
			Username:  "test-user-" + acctest.RandString(5),
			FirstName: "test-firstname",
			LastName:  "test-lastname",
			Email:     "test@example.com",
			Password:  acctest.RandString(20),
		}
		resource.Test(t, resource.TestCase{
			PreCheck: func() { testAccPreCheck(t) },
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
			},
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccUserResource1Config(resource1),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource1.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource1.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource1.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource1.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(false),
						),
					},
				},
				// ImportState testing
				{
					ResourceName:            fmt.Sprintf("%s_user.test", configprefix.Prefix),
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"password"},
				},
				// Update and Read testing
				{
					Config: testAccUserResource2Config(resource2),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource2.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource2.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource2.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource2.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(resource2.IsSuperuser),
						),
					},
				},
				// Test superuser
				{
					Config: testAccUserResource3Config(resource3),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource3.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource3.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource3.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource3.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("password"),
							knownvalue.StringExact(resource3.Password),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-3", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(resource3.IsSuperuser),
						),
					},
				},
				// Test system auditor
				{
					Config: testAccUserResource4Config(resource4),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("username"),
							knownvalue.StringExact(resource4.Username),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("first_name"),
							knownvalue.StringExact(resource4.FirstName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("last_name"),
							knownvalue.StringExact(resource4.LastName),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("email"),
							knownvalue.StringExact(resource4.Email),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("password"),
							knownvalue.StringExact(resource4.Password),
						),
						statecheck.ExpectKnownValue(
							fmt.Sprintf("%s_user.test-4", configprefix.Prefix),
							tfjsonpath.New("is_superuser"),
							knownvalue.Bool(resource4.IsSuperuser),
						),
					},
				},
			},
		})
	}
}

func testAccUserResource1Config(resource UserAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_user" "test" {
  username              = "%[2]s"
  first_name    	    = "%[3]s"
  last_name         	= "%[4]s"
  email		        	= "%[5]s"
  password_wo 		    = "%[6]s"
  password_wo_version   = 1
}
  `, configprefix.Prefix, resource.Username, resource.FirstName, resource.LastName, resource.Email, resource.Password)
}

func testAccUserResource2Config(resource UserAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_user" "test" {
  username              = "%[2]s"
  first_name        	= "%[3]s"
  last_name         	= "%[4]s"
  email		         	= "%[5]s"
  password_wo 		    = "%[6]s"
  password_wo_version   = 1
  is_superuser          = %[7]v
}
  `, configprefix.Prefix, resource.Username, resource.FirstName, resource.LastName, resource.Email, resource.Password, resource.IsSuperuser)
}

func testAccUserResource3Config(resource UserAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_user" "test-3" {
  username      = "%[2]s"
  first_name 	= "%[3]s"
  last_name 	= "%[4]s"
  email			= "%[5]s"
  password 		= "%[6]s"
  is_superuser  = %[7]v
}
  `, configprefix.Prefix, resource.Username, resource.FirstName, resource.LastName, resource.Email, resource.Password, resource.IsSuperuser)
}

func testAccUserResource4Config(resource UserAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_user" "test-4" {
  username      	= "%[2]s"
  first_name 		= "%[3]s"
  last_name 		= "%[4]s"
  email				= "%[5]s"
  password 			= "%[6]s"
  is_system_auditor = %[7]v
}
  `, configprefix.Prefix, resource.Username, resource.FirstName, resource.LastName, resource.Email, resource.Password, resource.IsSystemAuditor)
}

func testAccUserResourcePasswordWoAuthConfig(username, password string, passwordVersion int) string {
	return fmt.Sprintf(`
resource "%[1]s_user" "password_wo_auth" {
  username            = "%[2]s"
  password_wo         = "%[3]s"
  password_wo_version = %[4]d
}
  `, configprefix.Prefix, username, password, passwordVersion)
}

func testAccUserCanAuthenticate(username, password string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		statusCode, body, err := testAccUserAuthenticationRequest(username, password)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK {
			return fmt.Errorf("expected auth status 200 for user %q, got %d with response: %s", username, statusCode, body)
		}

		return nil
	}
}

func testAccUserCannotAuthenticate(username, password string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		statusCode, body, err := testAccUserAuthenticationRequest(username, password)
		if err != nil {
			return err
		}

		if statusCode != http.StatusUnauthorized && statusCode != http.StatusForbidden {
			return fmt.Errorf("expected auth status 401 or 403 for user %q, got %d with response: %s", username, statusCode, body)
		}

		return nil
	}
}

func testAccUserAuthenticationRequest(username, password string) (int, string, error) {
	endpoint := os.Getenv("AAP_HOST")

	if endpoint == "" {
		return 0, "", fmt.Errorf("AAP_HOST must be set for authentication checks")
	}

	endpoint = strings.TrimRight(endpoint, "/")
	authPath := "/api/controller/v2/me/"
	if configprefix.Prefix == "awx" {
		authPath = "/api/v2/me/"
	}

	client := &http.Client{Timeout: 30 * time.Second}
	envInsecure := strings.ToLower(os.Getenv("AAP_INSECURE_SKIP_VERIFY"))
	if envInsecure == "1" || envInsecure == "true" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	url := endpoint + authPath

	var lastStatusCode int
	var lastBody string

	for range 5 {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return 0, "", fmt.Errorf("unable to create authentication request: %w", err)
		}

		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return 0, "", fmt.Errorf("authentication request failed: %w", err)
		}

		responseBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return 0, "", fmt.Errorf("unable to read authentication response body: %w", readErr)
		}

		lastStatusCode = resp.StatusCode
		lastBody = strings.TrimSpace(string(responseBody))

		if resp.StatusCode == http.StatusOK {
			return resp.StatusCode, lastBody, nil
		}

		if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
			return resp.StatusCode, lastBody, nil
		}

		time.Sleep(2 * time.Second)
	}

	return lastStatusCode, lastBody, nil
}

func testAccSetUserPasswordOutOfBand(username, newPassword string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		endpoint := os.Getenv("AAP_HOST")
		if endpoint == "" {
			return fmt.Errorf("AAP_HOST must be set for out-of-band password update")
		}

		token := os.Getenv("AAP_OAUTH_TOKEN")
		if token == "" {
			return fmt.Errorf("AAP_OAUTH_TOKEN must be set for out-of-band password update")
		}

		client := &providerClient{
			client:   &http.Client{Timeout: 30 * time.Second},
			endpoint: strings.TrimRight(endpoint, "/"),
			auth:     "Bearer " + token,
		}

		envInsecure := strings.ToLower(os.Getenv("AAP_INSECURE_SKIP_VERIFY"))
		if envInsecure == "1" || envInsecure == "true" {
			client.client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		lookupURL := fmt.Sprintf("users/?username=%s", urlParser.QueryEscape(username))
		body, _, err := client.GenericAPIRequest(context.Background(), http.MethodGet, lookupURL, nil, []int{200}, "gateway")
		if err != nil {
			return fmt.Errorf("unable to look up user %q for out-of-band password update: %w", username, err)
		}

		result := struct {
			Count   int            `json:"count"`
			Results []UserAPIModel `json:"results"`
		}{}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("unable to parse user lookup response for %q: %w", username, err)
		}

		if result.Count != 1 || len(result.Results) != 1 {
			return fmt.Errorf("expected exactly 1 user for %q, got count=%d", username, result.Count)
		}

		user := result.Results[0]
		updateBody := UserAPIModel{
			Username:        user.Username,
			FirstName:       user.FirstName,
			LastName:        user.LastName,
			Email:           user.Email,
			IsSuperuser:     user.IsSuperuser,
			IsSystemAuditor: user.IsSystemAuditor,
			Password:        newPassword,
		}

		updateURL := fmt.Sprintf("users/%d/", user.Id)
		_, _, err = client.CreateUpdateAPIRequest(context.Background(), http.MethodPut, updateURL, updateBody, []int{200}, "gateway")
		if err != nil {
			return fmt.Errorf("unable to set out-of-band password for %q: %w", username, err)
		}

		return nil
	}
}
