package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Implmenet the compare.ValueComparer interface in order to compare two vals.
//
//	Compare two values of different types as if they were both strings. Used mostly to
//	 compare IDs.
type compareTwoValuesAsStrings struct{}

func (mc *compareTwoValuesAsStrings) CompareValues(values ...any) error {
	if len(values) != 2 {
		return errors.New("expected exactly two values to compare")
	}

	v1 := fmt.Sprint(values[0])
	v2 := fmt.Sprint(values[1])

	if reflect.TypeOf(v1) != reflect.TypeOf(v2) {
		return fmt.Errorf("type mismatch: %T vs %T", v1, v2)
	}

	if v1 != v2 {
		return fmt.Errorf("value mismatch: %v vs %v", v1, v2)
	}

	return nil
}

type compareStringInList struct{}

func (mc *compareStringInList) CompareValues(values ...any) error {
	if len(values) != 2 {
		return errors.New("expected exactly two values: a string and a list of integers")
	}

	v1, ok1 := values[0].(string)
	v2, ok2 := values[1].([]interface{})

	if !ok1 {
		return fmt.Errorf("type mismatch: expected first value to be a string, got %T", values[0])
	}

	if !ok2 {
		return fmt.Errorf("type mismatch: expected second value to be a list of integers, got %T", values[1])
	}

	for _, num := range v2 {
		switch n := num.(type) {
		case int:
			if fmt.Sprint(n) == v1 {
				return nil // Match found
			}
		case json.Number:
			if n.String() == v1 {
				return nil // Match found
			}
		default:
			return fmt.Errorf("type mismatch: expected list elements to be integers or json.Number, got %T", n)
		}
	}

	return fmt.Errorf("value mismatch: %v not found in %v", v1, v2)
}

// Implmenet the compare.ValueComparer interface in order to compare Slack-specific Notification Configurations.
type compareTwoSlackConfigs struct {
	InitialValue string
}

func (mc *compareTwoSlackConfigs) CompareValues(values ...any) error {
	if len(values) != 1 {
		return errors.New("expected exactly 1 values to compare with InitialValue")
	}

	v1 := fmt.Sprint(values[0])

	var initialSlack, v1Slack SlackConfiguration

	err := json.Unmarshal([]byte(v1), &v1Slack)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(mc.InitialValue), &initialSlack)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(initialSlack, v1Slack) {
		return fmt.Errorf("value mismatch: %v vs %v", mc.InitialValue, v1)
	}

	return nil
}

// Implmenet the compare.ValueComparer interface in order to compare Webhook-specific Notification Configurations.
type compareTwoWebhookConfigs struct {
	InitialValue string
}

func (mc *compareTwoWebhookConfigs) CompareValues(values ...any) error {
	if len(values) != 1 {
		return errors.New("expected exactly 1 values to compare with InitialValue")
	}

	v1 := fmt.Sprint(values[0])

	var initialWebhook, v1Webhook WebhookConfiguration

	err := json.Unmarshal([]byte(v1), &v1Webhook)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(mc.InitialValue), &initialWebhook)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(initialWebhook, v1Webhook) {
		return fmt.Errorf("value mismatch: %v vs %v", mc.InitialValue, v1)
	}

	return nil
}

// ImportStateIdFunc to fetch job_template_id from state for resources that don't have an ID.
func importStateJobTemplateID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		jobTemplateID, exists := rs.Primary.Attributes["job_template_id"]
		if !exists {
			return "", fmt.Errorf("job_template_id not found in state")
		}

		return jobTemplateID, nil
	}
}

func importStateWorkflowJobTemplateID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		jobTemplateID, exists := rs.Primary.Attributes["workflow_job_template_id"]
		if !exists {
			return "", fmt.Errorf("workflow_job_template_id not found in state")
		}

		return jobTemplateID, nil
	}
}

// panic if can't convert to string.
func mustMarshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("marshal failed: %v", err))
	}
	return string(b)
}

func mustString(v any) string {
	s, ok := v.(string)
	if !ok {
		panic("value is not a string")
	}
	return s
}

func mustFloat64(v any) float64 {
	s, ok := v.(float64)
	if !ok {
		panic("value is not a float64")
	}
	return s
}

func mustBool(v any) bool {
	s, ok := v.(bool)
	if !ok {
		panic("value is not a bool")
	}
	return s
}
