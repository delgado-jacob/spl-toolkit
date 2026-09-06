package mapper

import (
	"fmt"
	"reflect"
	"strings"
)

type configCollection struct {
	typeOf  reflect.Type
	pointer uintptr
}

type configCloneTracker map[configCollection]struct{}

func cloneAndValidateMappingConfig(config *MappingConfig) (*MappingConfig, error) {
	owned, err := cloneMappingConfig(config, make(configCloneTracker))
	if err != nil {
		return nil, fmt.Errorf("invalid mapping config: %w", err)
	}
	if result := owned.Validate(); !result.Valid {
		return nil, fmt.Errorf("invalid mapping config: %s", strings.Join(result.Errors, "; "))
	}
	return owned, nil
}

func cloneMappingConfig(config *MappingConfig, tracker configCloneTracker) (*MappingConfig, error) {
	owned := &MappingConfig{
		Version:     config.Version,
		Name:        config.Name,
		Description: config.Description,
		Mappings:    append([]FieldMapping(nil), config.Mappings...),
	}

	var err error
	owned.Rules, err = cloneConditionalRules(config.Rules, tracker, "rules")
	if err != nil {
		return nil, err
	}
	owned.DataModels, err = cloneDataModelMappings(config.DataModels, tracker)
	if err != nil {
		return nil, err
	}
	if config.Metadata != nil {
		metadata, err := cloneConfigValue(config.Metadata, tracker, "metadata")
		if err != nil {
			return nil, err
		}
		owned.Metadata = metadata.(map[string]interface{})
	}
	return owned, nil
}

func cloneConditionalRules(rules []ConditionalRule, tracker configCloneTracker, path string) ([]ConditionalRule, error) {
	if rules == nil {
		return nil, nil
	}
	leave, err := tracker.enter(rules, path)
	if err != nil {
		return nil, err
	}
	defer leave()

	owned := make([]ConditionalRule, len(rules))
	for i, rule := range rules {
		owned[i] = rule
		owned[i].Mappings = append([]FieldMapping(nil), rule.Mappings...)
		owned[i].Conditions, err = cloneConditions(rule.Conditions, tracker, fmt.Sprintf("%s[%d].conditions", path, i))
		if err != nil {
			return nil, err
		}
	}
	return owned, nil
}

func cloneConditions(conditions []Condition, tracker configCloneTracker, path string) ([]Condition, error) {
	if conditions == nil {
		return nil, nil
	}
	leave, err := tracker.enter(conditions, path)
	if err != nil {
		return nil, err
	}
	defer leave()

	owned := make([]Condition, len(conditions))
	for i, condition := range conditions {
		owned[i] = condition
		owned[i].Value, err = cloneConfigValue(condition.Value, tracker, fmt.Sprintf("%s[%d].value", path, i))
		if err != nil {
			return nil, err
		}
		owned[i].Children, err = cloneConditions(condition.Children, tracker, fmt.Sprintf("%s[%d].children", path, i))
		if err != nil {
			return nil, err
		}
	}
	return owned, nil
}

func cloneDataModelMappings(mappings []DataModelMapping, tracker configCloneTracker) ([]DataModelMapping, error) {
	if mappings == nil {
		return nil, nil
	}
	leave, err := tracker.enter(mappings, "datamodels")
	if err != nil {
		return nil, err
	}
	defer leave()

	owned := make([]DataModelMapping, len(mappings))
	for i, mapping := range mappings {
		owned[i] = mapping
		owned[i].FieldMappings = append([]DataModelFieldMapping(nil), mapping.FieldMappings...)
		owned[i].ConditionalMappings, err = cloneConditionalRules(mapping.ConditionalMappings, tracker, fmt.Sprintf("datamodels[%d].conditional_mappings", i))
		if err != nil {
			return nil, err
		}
	}
	return owned, nil
}

func cloneConfigValue(value interface{}, tracker configCloneTracker, path string) (interface{}, error) {
	switch value := value.(type) {
	case nil,
		bool, string,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return value, nil
	case []interface{}:
		leave, err := tracker.enter(value, path)
		if err != nil {
			return nil, err
		}
		defer leave()
		owned := make([]interface{}, len(value))
		for i, item := range value {
			owned[i], err = cloneConfigValue(item, tracker, fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return nil, err
			}
		}
		return owned, nil
	case map[string]interface{}:
		leave, err := tracker.enter(value, path)
		if err != nil {
			return nil, err
		}
		defer leave()
		owned := make(map[string]interface{}, len(value))
		for key, item := range value {
			owned[key], err = cloneConfigValue(item, tracker, path+"."+key)
			if err != nil {
				return nil, err
			}
		}
		return owned, nil
	default:
		if isScalar(value) {
			return value, nil
		}
		return nil, fmt.Errorf("%s contains unsupported value of type %T", path, value)
	}
}

func (tracker configCloneTracker) enter(value interface{}, path string) (func(), error) {
	reflected := reflect.ValueOf(value)
	if reflected.IsNil() {
		return func() {}, nil
	}
	collection := configCollection{typeOf: reflected.Type(), pointer: reflected.Pointer()}
	if _, exists := tracker[collection]; exists {
		return nil, fmt.Errorf("%s contains a cycle", path)
	}
	tracker[collection] = struct{}{}
	return func() { delete(tracker, collection) }, nil
}
