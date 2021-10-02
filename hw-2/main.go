package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"time"
)

type JSONCompany struct {
	Name      string        `json:"company"`
	Operation JSONOperation `json:"operation"`
	JSONOperation
}

type JSONOperation struct {
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
	ID        interface{} `json:"id"`
	CreatedAt string      `json:"created_at"`
}

type ResultCompany struct {
	Name                  string   `json:"company"`
	ValidOperationsCount  int      `json:"valid_operations_count"`
	Balance               int      `json:"balance"`
	JSONInvalidOperations []string `json:"invalid_operations,omitempty"`
	operations            []InvalidOperation
}

type InvalidOperation struct {
	ID          string
	createdTime time.Time
}

func (c *ResultCompany) addInvalidOperation(operation InvalidOperation) {
	c.operations = append(c.operations, operation)
}

func (c *ResultCompany) applyOperation(jsonCompany JSONCompany) {
	parsedTime, _ := time.Parse(time.RFC3339, jsonCompany.CreatedAt)
	invalidOperation := InvalidOperation{
		ID:          fmt.Sprint(jsonCompany.ID),
		createdTime: parsedTime,
	}

	value, err := strconv.Atoi(fmt.Sprint(jsonCompany.Value))
	if err != nil {
		c.addInvalidOperation(invalidOperation)
		return
	}

	switch jsonCompany.Type {
	case "+":
		c.Balance += value
	case "-":
		c.Balance -= value
	case "income":
		c.Balance += value
	case "outcome":
		c.Balance -= value
	default:
		c.addInvalidOperation(invalidOperation)
		return
	}
	c.ValidOperationsCount++
}

func mapToSlice(m map[string]ResultCompany) []ResultCompany {
	s := make([]ResultCompany, 0, len(m))
	for _, value := range m {
		s = append(s, value)
	}
	return s
}

func writeResult(companies map[string]ResultCompany, w io.Writer) {
	companiesSlice := mapToSlice(companies)
	sort.SliceStable(companiesSlice, func(i, j int) bool {
		return companiesSlice[i].Name < companiesSlice[j].Name
	})

	for i, c := range companiesSlice {
		sort.SliceStable(c.operations, func(i, j int) bool {
			return c.operations[i].createdTime.Before(c.operations[j].createdTime)
		})
		c.JSONInvalidOperations = make([]string, 0, len(c.operations))
		for _, o := range c.operations {
			c.JSONInvalidOperations = append(c.JSONInvalidOperations, o.ID)
		}
		companiesSlice[i] = c
	}

	marshalIndent, err := json.MarshalIndent(companiesSlice, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := w.Write(marshalIndent); err != nil {
		log.Fatal(err)
	}
}

func collectDataFromOperation(jsonCompany *JSONCompany) {
	operation := jsonCompany.Operation
	if operation.Type != "" {
		jsonCompany.Type = operation.Type
	}
	if operation.Value != nil {
		jsonCompany.Value = operation.Value
	}
	if operation.ID != nil {
		jsonCompany.ID = operation.ID
	}
	if operation.CreatedAt != "" {
		jsonCompany.CreatedAt = operation.CreatedAt
	}
}

func isOperationValid(jsonCompany JSONCompany) bool {
	if jsonCompany.Name == "" {
		return false
	}
	if jsonCompany.ID == nil {
		return false
	}
	if jsonCompany.CreatedAt == "" {
		return false
	} else if _, err := time.Parse(time.RFC3339, jsonCompany.CreatedAt); err != nil {
		return false
	}
	return true
}

func collectData(jsonCompanies []JSONCompany) map[string]ResultCompany {
	allCompanies := make(map[string]ResultCompany)
	for _, jsonCompany := range jsonCompanies {
		collectDataFromOperation(&jsonCompany)
		if isOperationValid(jsonCompany) {
			if _, ok := allCompanies[jsonCompany.Name]; !ok {
				allCompanies[jsonCompany.Name] = ResultCompany{Name: jsonCompany.Name}
			}
			company := allCompanies[jsonCompany.Name]
			company.applyOperation(jsonCompany)
			allCompanies[jsonCompany.Name] = company
		}
	}
	return allCompanies
}

func readFromFile(filePath string) []byte {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	b, _ := io.ReadAll(file)
	if err := file.Close(); err != nil {
		return []byte{}
	}
	return b
}

func readFromConsole(stopSymbol string) []byte {
	var data []byte
	var line string
	for {
		if _, err := fmt.Scan(&line); err == io.EOF {
			break
		}
		data = append(data, []byte(line)...)
	}
	return data
}

func readDataFromAllSources(data *[]byte) {
	filePath := ""
	filePathPointer := flag.String("file", "", "Path to the json file")
	flag.Parse()

	filePath = *filePathPointer
	if filePath == "" {
		FILE := "FILE"
		if s, ok := os.LookupEnv(FILE); ok {
			filePath = s
		}
	}
	if filePath != "" {
		*data = readFromFile(filePath)
	} else {
		*data = readFromConsole("]")
	}
}

func main() {
	var data []byte
	readDataFromAllSources(&data)

	var jsonCompanies []JSONCompany
	_ = json.Unmarshal(data, &jsonCompanies)
	allCompanies := collectData(jsonCompanies)

	out, _ := os.Create("out.json")
	writeResult(allCompanies, out)
	if err := out.Close(); err != nil {
		log.Fatal(err)
	}
}
