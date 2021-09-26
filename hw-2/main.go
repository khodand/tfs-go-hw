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

type jsonMap map[string]interface{}

type ResultOperation struct {
	Type        string
	Value       int
	ID          string
	CreatedTime time.Time
}

type JsonCompany struct {
	Name      string        `json:"company"`
	Operation JsonOperation `json:"operation"`
	Type      string        `json:"type"`
	Value     interface{}   `json:"value"`
	Id        interface{}   `json:"id"`
	CreatedAt string        `json:"created_at"`
}

type JsonOperation struct {
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
	Id        interface{} `json:"id"`
	CreatedAt string      `json:"created_at"`
}

type ResultCompany struct {
	Name                 string
	validOperationsCount int
	balance              int
	invalidOperations    []ResultOperation
}

func (c ResultCompany) writeTo(w io.Writer) {
	writeString(w, fmt.Sprintf("		\"company\":\"%s\",\n", c.Name))
	writeString(w, fmt.Sprintf("		\"valid_operations_count\":%v,\n", c.validOperationsCount))
	writeString(w, fmt.Sprintf("		\"balance\":%v", c.balance))
	if len(c.invalidOperations) != 0 {
		writeString(w, ",\n		\"invalid_operations\":[")

		sort.SliceStable(c.invalidOperations, func(i, j int) bool {
			return c.invalidOperations[i].CreatedTime.Before(c.invalidOperations[j].CreatedTime)
		})

		for i := 0; i < len(c.invalidOperations)-1; i++ {
			writeString(w, fmt.Sprintf("\"%s\", ", c.invalidOperations[i].ID))
		}
		writeString(w, fmt.Sprintf("\"%s\"]\n", c.invalidOperations[len(c.invalidOperations)-1].ID))
	} else {
		writeString(w, "\n")
	}
}

func (c *ResultCompany) applyOperation(operation ResultOperation) {
	switch operation.Type {
	case "+":
		c.balance += operation.Value
	case "-":
		c.balance -= operation.Value
	case "income":
		c.balance += operation.Value
	case "outcome":
		c.balance -= operation.Value
	default:
		c.addInvalidOperation(operation)
		return
	}
	c.validOperationsCount++
}

func (c *ResultCompany) addInvalidOperation(operation ResultOperation) {
	c.invalidOperations = append(c.invalidOperations, operation)
}

func writeString(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		return
	}
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

	writeString(w, "[\n")
	for i, company := range companiesSlice {
		writeString(w, "\t{\n")
		company.writeTo(w)
		writeString(w, "\t}")
		if i != len(companies)-1 {
			writeString(w, ",")
		}
		writeString(w, "\n")
	}
	writeString(w, "]\n")
}

func collectDataFromOperation(company *JsonCompany) {
	operation := company.Operation
	if operation.Type != "" {
		company.Type = operation.Type
	}
	if operation.Value != nil {
		company.Value = operation.Value
	}
	if operation.Id != nil {
		company.Id = operation.Id
	}
	if operation.CreatedAt != "" {
		company.CreatedAt = operation.CreatedAt
	}
}

func isOperationValid(company JsonCompany) bool {
	if company.Name == "" {
		return false
	}
	if company.Id == nil {
		return false
	}
	if company.CreatedAt == "" {
		return false
	} else if _, err := time.Parse(time.RFC3339, company.CreatedAt); err != nil {
		return false
	}
	return true
}

func collectData(jsonCompanies []JsonCompany) map[string]ResultCompany {
	allCompanies := make(map[string]ResultCompany)
	for _, jsonCompany := range jsonCompanies {
		collectDataFromOperation(&jsonCompany)
		if isOperationValid(jsonCompany) {
			if _, ok := allCompanies[jsonCompany.Name]; !ok {
				allCompanies[jsonCompany.Name] = ResultCompany{Name: jsonCompany.Name}
			}

			company := allCompanies[jsonCompany.Name]
			parsedTime, _ := time.Parse(time.RFC3339, jsonCompany.CreatedAt)
			value, err := strconv.Atoi(fmt.Sprint(jsonCompany.Value))
			if err != nil {
				company.addInvalidOperation(ResultOperation{
					ID:          fmt.Sprint(jsonCompany.Id),
					CreatedTime: parsedTime,
				})
			} else {
				company.applyOperation(ResultOperation{
					Type:        jsonCompany.Type,
					Value:       value,
					ID:          fmt.Sprint(jsonCompany.Id),
					CreatedTime: parsedTime,
				})
			}
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
	for line != stopSymbol {
		if _, err := fmt.Scan(&line); err != nil {
			return nil
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
	fmt.Println("Options =" + filePath)
	if filePath == "" {
		FILE := "FILE"
		if s, ok := os.LookupEnv(FILE); ok {
			filePath = s
		}
		fmt.Println("ENV", filePath)
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

	var jsonCompanies []JsonCompany
	fmt.Println(string(data))
	_ = json.Unmarshal(data, &jsonCompanies)

	allCompanies := collectData(jsonCompanies)
	fmt.Println(jsonCompanies)

	out, _ := os.Create("out.json")
	writeResult(allCompanies, out)
	if err := out.Close(); err != nil {
		return
	}
}
