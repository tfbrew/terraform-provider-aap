package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	urlParser "net/url"
	"time"
)

// Organizations are created via the gateway API, which syncs them to the controller and EDA APIs
// asynchronously. These control how long lookups wait for that sync before giving up.
var (
	orgSyncPollInterval = 2 * time.Second
	orgSyncTimeout      = 30 * time.Second
)

// lookupOrgIdByName queries apiEndpoint for the organization with the given name and returns its ID.
// found is false (with a nil error) if the endpoint returned 404, e.g. when EDA isn't installed.
// If waitForSync is true, a result count of 0 is retried until orgSyncTimeout elapses, to allow
// for an org recently created in gateway that hasn't been synced to apiEndpoint yet.
func lookupOrgIdByName(ctx context.Context, client *providerClient, name, apiEndpoint string, waitForSync bool) (int, bool, error) {
	url := fmt.Sprintf("organizations/?name=%s", urlParser.QueryEscape(name))
	deadline := time.Now().Add(orgSyncTimeout)

	for {
		body, statusCode, err := client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, apiEndpoint)
		if err != nil {
			return 0, false, fmt.Errorf("error making API http request: %s", err.Error())
		}
		if statusCode == 404 {
			return 0, false, nil
		}

		var result JTChildAPIRead
		err = json.Unmarshal(body, &result)
		if err != nil {
			return 0, false, fmt.Errorf("unable to unmarshal response body into result object: %s", err.Error())
		}
		if result.Count == 1 {
			return result.Results[0].Id, true, nil
		}

		// only a count of 0 can be resolved by waiting; more than 1 match is never going to change
		if result.Count != 0 || !waitForSync || ctx.Err() != nil || time.Now().Add(orgSyncPollInterval).After(deadline) {
			return 0, false, fmt.Errorf("querying for org %q by name against %s endpoint resulted in result count of %d instead of 1", name, apiEndpoint, result.Count)
		}

		SleepWithContext(ctx, orgSyncPollInterval)
	}
}
