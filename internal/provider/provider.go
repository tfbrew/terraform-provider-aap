// SPECIAL: Be sure to update any desriptions to match repo & run make generate
package provider

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/providervalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

// Ensure Provider satisfies various provider interfaces.
var _ provider.Provider = &theProvider{}
var _ provider.ProviderWithFunctions = &theProvider{}

// theProvider defines the provider implementation.
type theProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// theProviderModel describes the provider data model.
type theProviderModel struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	Token              types.String `tfsdk:"token"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	APIretry           types.Object `tfsdk:"api_retry"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

type apiRetryModel struct {
	APIretryCount        types.Int32 `tfsdk:"api_retry_count"`
	APIretryDelaySeconds types.Int32 `tfsdk:"api_retry_delay_seconds"`
}

func (p *theProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = configprefix.Prefix
	resp.Version = p.version
}

func (p *theProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This is a Terraform Provider for managing resources in Automation Controller such as AWX/Tower or Ansible Automation Platform (AAP).",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "URL for automation controller (i.e. https://aap.example.com)",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "Automation controller access token (instead of username/password). You can also set this using the AAP_OAUTH_TOKEN environment variable.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Automation controller username (instead of token). You can also set this using the AAP_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Automation controller password (instead of token). You can also set this using the AAP_PASSWORD environment variable.",
				Optional:    true,
			},
			"api_retry": schema.SingleNestedAttribute{
				Description: "An optional block to define if the provider should retry GET/read API requests that intitially fail.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"api_retry_count": schema.Int32Attribute{
						Description: "The number of times a GET/read API request should be reattempted should it not succeed on the first try. Can be useful when the number of Terraform objects in your plan creates many API calls and causes the AWX/AAP platform to bog down. Valid values are integers between 1 and 5. You can also set this using the AAP_API_RETRY_COUNT environment variable.",
						Required:    true,
						Validators: []validator.Int32{
							int32validator.Between(1, 5),
						},
					},
					"api_retry_delay_seconds": schema.Int32Attribute{
						Description: "The number of seconds this provider should wait before making a retry attempt. The value must be an integer value of 1 or greater. You can also set this using the AAP_API_RETRY_DELAY_SECONDS environment variable.",
						Required:    true,
						Validators: []validator.Int32{
							int32validator.AtLeast(1),
						},
					},
				},
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Description: "Set true to skip SSL certificate verification for the automation controller. You can also set this using the AAP_INSECURE_SKIP_VERIFY environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *theProvider) ConfigValidators(ctx context.Context) []provider.ConfigValidator {
	return []provider.ConfigValidator{
		providervalidator.Conflicting(
			path.MatchRoot("token"),
			path.MatchRoot("username"),
		),
		providervalidator.Conflicting(
			path.MatchRoot("token"),
			path.MatchRoot("password"),
		),
		providervalidator.RequiredTogether(
			path.MatchRoot("username"),
			path.MatchRoot("password"),
		),
	}
}

func (p *theProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var (
		token, endpoint, username, password, auth string
	)

	var data theProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	} else if aapEnv, exists := os.LookupEnv("AAP_HOST"); exists {
		endpoint = aapEnv
	} else {
		endpoint = os.Getenv("TOWER_HOST")
	}

	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing API Endpoint Configuration",
			"While configuring the provider, the API endpoint hostname was not found in "+
				"the AAP_HOST environment variable or provider "+
				"configuration block endpoint attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	} else {
		// strip trailing '/' character from endpoint if present
		endpointRunes := []rune(endpoint)
		lastChar := endpointRunes[len(endpointRunes)-1]
		if lastChar == '/' {
			endpoint = string(endpointRunes[:len(endpointRunes)-1])
		}
	}

	// Prefer AAP_ env vars, fallback to TOWER_ if not set
	envToken, tokenExists := os.LookupEnv("AAP_OAUTH_TOKEN")
	if !tokenExists {
		envToken, tokenExists = os.LookupEnv("TOWER_OAUTH_TOKEN")
	}
	envUsername, userExists := os.LookupEnv("AAP_USERNAME")
	if !userExists {
		envUsername, userExists = os.LookupEnv("TOWER_USERNAME")
	}
	envPassword, passwordExists := os.LookupEnv("AAP_PASSWORD")
	if !passwordExists {
		envPassword, passwordExists = os.LookupEnv("TOWER_PASSWORD")
	}

	// Get token if password/username not set
	if data.Token.IsNull() && data.Username.IsNull() && data.Password.IsNull() && tokenExists {
		token = envToken
	}

	if data.Token.IsNull() && data.Username.IsNull() && data.Password.IsNull() && !tokenExists && userExists && passwordExists {
		username = envUsername
		password = envPassword
	}

	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	if !data.Username.IsNull() {
		username = data.Username.ValueString()
	}

	if !data.Password.IsNull() {
		password = data.Password.ValueString()
	}

	if (token != "" && (username != "" || password != "")) || (token == "" && (username == "" || password == "")) {
		resp.Diagnostics.AddError(
			"Provider Configuration Error",
			"Specify a token (AAP_OAUTH_TOKEN) OR username/password (AAP_USERNAME/AAP_PASSWORD).")
		return
	}

	if token != "" {
		auth = "Bearer" + " " + token
	} else {
		authString := username + ":" + password
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))
		auth = "Basic" + " " + encodedAuth
	}

	httpclient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Check for insecure skip verify option
	var insecureSkipVerify bool

	if !data.InsecureSkipVerify.IsNull() {
		insecureSkipVerify = data.InsecureSkipVerify.ValueBool()
	} else if envInsecure, exists := os.LookupEnv("AAP_INSECURE_SKIP_VERIFY"); exists {
		if envInsecure == "true" || envInsecure == "1" {
			insecureSkipVerify = true
		}
	}

	if insecureSkipVerify {
		httpclient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	client := new(providerClient)

	client.client = httpclient
	client.endpoint = endpoint
	client.auth = auth

	if data.APIretry.IsNull() {
		envAPIRetryCount, envAPIRetryCountExists := os.LookupEnv("AAP_API_RETRY_COUNT")
		envAPIRetryDelaySeconds, envAPIRetryDelaySecondsExists := os.LookupEnv("AAP_API_RETRY_DELAY_SECONDS")

		if envAPIRetryCountExists != envAPIRetryDelaySecondsExists {
			resp.Diagnostics.AddError(
				"Provider Configuration Error",
				"Both AAP_API_RETRY_COUNT and AAP_API_RETRY_DELAY_SECONDS environment variables must be set together.",
			)
			return
		} else if envAPIRetryCountExists && envAPIRetryDelaySecondsExists {
			retryCountInt64, err := strconv.ParseInt(envAPIRetryCount, 10, 32)
			if err != nil {
				resp.Diagnostics.AddError(
					"Provider Configuration Error",
					fmt.Sprintf("AAP_API_RETRY_COUNT must be a 32-bit integer, got: %s", envAPIRetryCount),
				)
				return
			}
			retryDelayInt64, err := strconv.ParseInt(envAPIRetryDelaySeconds, 10, 32)
			if err != nil {
				resp.Diagnostics.AddError(
					"Provider Configuration Error",
					fmt.Sprintf("AAP_API_RETRY_DELAY_SECONDS must be a 32-bit integer, got: %s", envAPIRetryDelaySeconds),
				)
				return
			}
			client.apiRetryCount = int32(retryCountInt64)
			client.apiRetryDelaySeconds = int32(retryDelayInt64)
		}
	}

	if !data.APIretry.IsNull() {
		var retryBlock apiRetryModel

		resp.Diagnostics.Append(data.APIretry.As(ctx, &retryBlock, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		client.apiRetryCount = retryBlock.APIretryCount.ValueInt32()
		client.apiRetryDelaySeconds = retryBlock.APIretryDelaySeconds.ValueInt32()
	}

	url := "me/"

	_, _, err := client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"AAP authentication failure",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	body, _, err := client.GenericAPIRequest(ctx, http.MethodGet, "ping/", nil, []int{200, 404}, "gateway")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to determine AAP version",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	var responseData PingAPIModel
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	versionFloat, err := strconv.ParseFloat(responseData.Version, 32)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to convert AAP version to float",
			fmt.Sprintf("Version string: %+v.", responseData.Version))
		return
	}

	client.aapVersion = float32(versionFloat)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *theProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExecutionEnvironmentResource,
		NewConstructedInventoryResource,
		NewCredentialResource,
		NewCredentialInputSourcesResource,
		NewCredentialTypeResource,
		NewEdaCredentialResource,
		NewEdaCredentialTypeResource,
		NewEdaDecisionEnvironmentResource,
		NewEdaProjectResource,
		NewHostResource,
		NewGenericEndpointResource,
		NewGroupResource,
		NewGroupHostResource,
		NewInstanceGroupResource,
		NewInventoryInputResource,
		NewInventoryResource,
		NewInventorySourceResource,
		NewJobTemplateCredentialResource,
		NewJobTemplateInstanceGroupsResource,
		NewJobTemplateLabelsResource,
		NewJobTemplateNotifTemplErrResource,
		NewJobTemplateNotifTemplStartedResource,
		NewJobTemplateNotifTemplSuccessResource,
		NewJobTemplateResource,
		NewJobTemplateSurveyResource,
		NewLabelsResource,
		NewNotificationTemplatesResource,
		NewOrganizationResource,
		NewProjectResource,
		NewRoleDefinitionResource,
		NewRoleUserAssignmentResource,
		NewRoleTeamAssignmentResource,
		NewScheduleResource,
		NewTeamResource,
		NewUserResource,
		NewWorkflowJobTemplateApprovalNodeResource,
		NewWorkflowJobTemplateJobNodeResource,
		NewWorkflowJobTemplateNodeLabelResource,
		NewWorkflowJobTemplateNodeSuccessResource,
		NewWorkflowJobTemplateNodeFailureResource,
		NewWorkflowJobTemplateNodeAlwaysResource,
		NewWorkflowJobTemplateNotifTemplApprovalsResource,
		NewWorkflowJobTemplateNotifTemplErrorResource,
		NewWorkflowJobTemplateNotifTemplStartedResource,
		NewWorkflowJobTemplateNotifTemplSuccessResource,
		NewWorkflowJobTemplateResource,
	}
}

func (p *theProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCredentialDataSource,
		NewCredentialTypeDataSource,
		NewEdaCredentialTypeDataSource,
		NewEdaCredentialDataSource,
		NewEdaDecisionEnvironmentDataSource,
		NewEdaProjectDataSource,
		NewExecutionEnvironmentDataSource,
		NewGroupDataSource,
		NewHostDataSource,
		NewInventoryDataSource,
		NewInventorySourceDataSource,
		NewInstanceGroupDataSource,
		NewJobTemplateDataSource,
		NewNotificationTemplateDataSource,
		NewOrganizationDataSource,
		NewProjectDataSource,
		NewRoleDefinitionDataSource,
		NewScheduleDataSource,
		NewTeamDataSource,
		NewUserDataSource,
	}
}

func (p *theProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		//NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &theProvider{
			version: version,
		}
	}
}
