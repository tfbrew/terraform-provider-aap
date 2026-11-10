package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

type providerClient struct {
	client   *http.Client
	endpoint string
	auth     string
	// AAP-AWX-DIFF-BEGIN: aap routes requests to the gateway, controller, EDA and galaxy APIs; awx always uses /api/v2/.
	aapVersion float32
	// AAP-AWX-DIFF-END
	apiRetryCount        int32
	apiRetryDelaySeconds int32
}

// A wrapper for http.NewRequestWithContext() that prepends tower endpoint to URL & sets authorization
// headers and then makes the actual http request.
// AAP-AWX-DIFF-BEGIN: aap routes requests to the gateway, controller, EDA and galaxy APIs; awx always uses /api/v2/.
func (c *providerClient) GenericAPIRequest(ctx context.Context, method, url string, requestBody any, successCodes []int, api_endpoint string) (responseBody []byte, statusCode int, errorMessage error) {

	url = c.buildAPIUrl(url, api_endpoint)
	// AAP-AWX-DIFF-END

	var body io.Reader

	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			errorMessage = fmt.Errorf("unable to marshal requestBody into json: %s", err.Error())
			return
		}

		body = strings.NewReader(string(jsonData))
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		errorMessage = fmt.Errorf("error generating http request: %v", err)
		return
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", c.auth)

	var success bool
	var httpResp *http.Response
	var maxAttempts int

	if method == http.MethodGet {
		maxAttempts = 1 + int(c.apiRetryCount)
	} else {
		maxAttempts = 1
	}

	for i := 0; i < maxAttempts; i++ {
		httpResp, err = c.client.Do(httpReq)
		if err != nil {
			errorMessage = fmt.Errorf("error doing http request: %v", err)
			return
		}

		for _, successCode := range successCodes {
			if httpResp.StatusCode == successCode {
				success = true
				break
			}
		}

		if success {
			break
		}

		// if not success & have remaining attempts
		if !success && ((maxAttempts - i) > 1) {
			SleepWithContext(ctx, time.Duration(c.apiRetryDelaySeconds)*time.Second)
		}

	}

	responseBody, err = io.ReadAll(httpResp.Body)
	statusCode = httpResp.StatusCode

	if err != nil {
		errorMessage = fmt.Errorf("unable to read the http response data body. body: %v", responseBody)
		return
	}
	defer httpResp.Body.Close()

	if !success {
		errorMessage = fmt.Errorf("expected %v http response code for API call, got %d with message %s", successCodes, statusCode, responseBody)
		return
	}

	return
}

func SleepWithContext(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
	case <-timer.C:
	}
}

// WaitForDeletion polls url until the API returns 404. Some objects, such as inventories, are deleted
// asynchronously (the DELETE returns 202), and deleting a parent object like the organization while the
// background delete is still running can fail with a database deadlock.
func (c *providerClient) WaitForDeletion(ctx context.Context, url string, aap25_api_endpoint_hint string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		_, statusCode, err := c.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, aap25_api_endpoint_hint)
		if err != nil {
			return err
		}
		if statusCode == http.StatusNotFound {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s still exists %s after the delete request", url, timeout)
		}

		SleepWithContext(ctx, 2*time.Second)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// AAP-AWX-DIFF-BEGIN: aap routes requests to the gateway, controller, EDA and galaxy APIs; awx always uses /api/v2/.
func (c *providerClient) CreateUpdateAPIRequest(ctx context.Context, method, url string, requestBody any, successCodes []int, api_endpoint string) (returnedData map[string]any, statusCode int, errorMessage error) {

	url = c.buildAPIUrl(url, api_endpoint)
	// AAP-AWX-DIFF-END

	var body io.Reader

	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			errorMessage = fmt.Errorf("unable to marshal requestBody into json: %s", err.Error())
			return
		}

		body = strings.NewReader(string(jsonData))
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		errorMessage = fmt.Errorf("error generating http request: %v", err)
		return
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", c.auth)

	var success bool
	var httpResp *http.Response
	var maxAttempts int

	if method == http.MethodGet {
		maxAttempts = 1 + int(c.apiRetryCount)
	} else {
		maxAttempts = 1
	}

	for i := 0; i < maxAttempts; i++ {
		httpResp, err = c.client.Do(httpReq)
		if err != nil {
			errorMessage = fmt.Errorf("error doing http request: %v", err)
			return
		}

		for _, successCode := range successCodes {
			if httpResp.StatusCode == successCode {
				success = true
				break
			}
		}
		if success {
			break
		}

		// if not success & have remaining attempts
		if !success && ((maxAttempts - i) > 1) {
			SleepWithContext(ctx, time.Duration(c.apiRetryDelaySeconds)*time.Second)
		}

	}

	if !success {
		body, err := io.ReadAll(httpResp.Body)
		defer httpResp.Body.Close()
		if err != nil {
			errorMessage = errors.New("unable to read http request response body to retrieve error message")
			return
		}
		errorMessage = fmt.Errorf("expected %v http response code for API call, got %d with message %s", successCodes, httpResp.StatusCode, body)
		return
	}

	statusCode = httpResp.StatusCode
	httpRespBodyData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		errorMessage = errors.New("unable to read http request response body to retrieve id")
		return
	}
	err = json.Unmarshal(httpRespBodyData, &returnedData)
	if err != nil {
		errorMessage = errors.New("unable to unmarshal http request response body to retrieve returnedData")
		return
	}
	return
}

// AAP-AWX-DIFF-BEGIN: aap routes requests to the gateway, controller, EDA and galaxy APIs; awx always uses /api/v2/.
func (c *providerClient) buildAPIUrl(resourceUrl, api_endpoint string) (url string) {

	if api_endpoint == "gateway" && configprefix.Prefix == "aap" {
		url = c.endpoint + "/api/gateway/v1/" + resourceUrl
	} else if api_endpoint == "gateway26" && configprefix.Prefix == "aap" && c.aapVersion >= float32(2.6) {
		url = c.endpoint + "/api/gateway/v1/" + resourceUrl
	} else if api_endpoint == "eda" && configprefix.Prefix == "aap" {
		url = c.endpoint + "/api/eda/v1/" + resourceUrl
	} else if api_endpoint == "galaxy" && configprefix.Prefix == "aap" {
		url = c.endpoint + "/api/galaxy/v3/" + resourceUrl
	} else if (api_endpoint == "controller" || api_endpoint == "") && configprefix.Prefix == "aap" {
		url = c.endpoint + "/api/controller/v2/" + resourceUrl
	} else if configprefix.Prefix == "awx" {
		url = c.endpoint + "/api/v2/" + resourceUrl // handle awx
	}
	// AAP-AWX-DIFF-END

	return
}
