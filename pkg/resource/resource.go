package resource

import (
	"reflect"
	"time"

	awsls "github.com/jckuester/awsls/aws"
	"github.com/pkg/errors"
)

// Resources converts given raw resources for a given resource type
// into a format that can be deleted by the Terraform API.
func DeletableResources(resType string, resources interface{}, client awsls.Client) ([]awsls.Resource, error) {
	var deletableResources []awsls.Resource

	reflectResources := reflect.ValueOf(resources)
	for i := 0; i < reflectResources.Len(); i++ {
		deleteID, err := getDeleteID(resType)
		if err != nil {
			return nil, err
		}

		deleteIDField, err := getField(deleteID, reflect.Indirect(reflectResources.Index(i)))
		if err != nil {
			return nil, errors.Wrapf(err, "Field with delete ID required for deleting resource")
		}

		var creationTime *time.Time
		creationTimeField, err := findField(creationTimeFieldNames, reflect.Indirect(reflectResources.Index(i)))
		if err == nil {
			creationTimeCastTime, ok := creationTimeField.Interface().(*time.Time)
			if ok {
				creationTime = creationTimeCastTime
			} else {
				creationTimeCastString, ok := creationTimeField.Interface().(*string)
				if ok {
					parsedCreationTime, err := time.Parse("2006-01-02T15:04:05.000Z0700", *creationTimeCastString)
					if err == nil {
						creationTime = &parsedCreationTime
					}
				}
			}
		}

		deletableResources = append(deletableResources, awsls.Resource{
			Type:      resType,
			ID:        deleteIDField.Elem().String(),
			CreatedAt: creationTime,
			Region:    client.Region,
			Profile:   client.Profile,
			AccountID: client.AccountID,
		})
	}

	return deletableResources, nil
}

func getField(name string, v reflect.Value) (reflect.Value, error) {
	field := v.FieldByName(name)

	if !field.IsValid() {
		return reflect.Value{}, errors.Errorf("Field not found: %s", name)
	}
	return field, nil
}

func findField(names []string, v reflect.Value) (reflect.Value, error) {
	for _, name := range names {
		field, err := getField(name, v)
		if err == nil {
			return field, nil
		}
	}
	return reflect.Value{}, errors.Errorf("Fields not found: %s", names)
}
