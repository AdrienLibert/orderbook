package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var folderPath string = "../../schemas/"

type Contract struct {
	Type       string                 `json:"type"`
	Version    string                 `json:"version"`
	Metadata   map[string]interface{} `json:"metadata"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}

// LoadContracts reads all JSON schema files from the contracts directory
func LoadContracts(folder string) (map[string]Contract, error) {
	contracts := make(map[string]Contract)
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".json" {
			file, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			var contract Contract
			if err := json.Unmarshal(file, &contract); err != nil {
				return err
			}
			contracts[info.Name()] = contract
		}
		return nil
	})
	return contracts, err
}

func matchType(jsonType, goType string) bool {
	switch jsonType {
	case "integer":
		return goType == "int64"
	case "string":
		return goType == "string"
	case "boolean":
		return goType == "bool"
	case "array":
		return goType == "slice"
	case "object":
		return goType == "struct"
	case "number":
		return goType == "float64"
	default:
		return false
	}
}

func ValidateStructAgainstContract(schema Contract, obj interface{}) bool {
	objType := reflect.TypeOf(obj)
	//objValue := reflect.ValueOf(obj)

	for jsonFieldName, fieldSchema := range schema.Properties {
		var field reflect.StructField
		found := false
		for i := 0; i < objType.NumField(); i++ {
			if objType.Field(i).Tag.Get("json") == jsonFieldName {
				field = objType.Field(i)
				found = true
				break
			}
		}
		if !found {
			return false
		}
		// Simple type check (expand for complex types)
		jsonType := fieldSchema.(map[string]interface{})["type"].(string)
		goType := field.Type.Kind().String()
		if !matchType(jsonType, goType) {
			return false
		}
	}
	return true
}

func TestOrderAgainstContract(t *testing.T) {
	contractName := "order.json"
	contracts, err := LoadContracts(folderPath)
	if err != nil {
		t.Fatalf("Failed to load contracts: %v", err)
	}

	orderContract, ok := contracts[contractName]
	if !ok {
		t.Errorf("Contr load {%s}", contractName)
	}

	t.Run(contractName, func(t *testing.T) {
		valid := ValidateStructAgainstContract(orderContract, Order{})
		if !valid {
			t.Errorf("Struct does not match contract: %s", contractName)
		}
	})
}

func TestTradeAgainstContract(t *testing.T) {
	contractName := "trade.json"
	contracts, _ := LoadContracts(folderPath)
	orderContract, ok := contracts[contractName]
	if !ok {
		t.Errorf("Contr load {%s}", contractName)
	}

	t.Run(contractName, func(t *testing.T) {
		valid := ValidateStructAgainstContract(orderContract, Trade{})
		if !valid {
			t.Errorf("Struct does not match contract: %s", contractName)
		}
	})

}
func TestPricePointAgainstContract(t *testing.T) {
	contractName := "pricepoint.json"
	contracts, _ := LoadContracts(folderPath)
	orderContract, ok := contracts[contractName]
	if !ok {
		t.Errorf("Contr load {%s}", contractName)
	}

	t.Run(contractName, func(t *testing.T) {
		valid := ValidateStructAgainstContract(orderContract, PricePoint{})
		if !valid {
			t.Errorf("Struct does not match contract: %s", contractName)
		}
	})

}
