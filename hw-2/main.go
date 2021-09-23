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

type Operation struct {
	Type        string
	Value       int
	ID          string
	CreatedTime time.Time
}

type Company struct {
	Name                 string
	validOperationsCount int
	balance              int
	invalidOperations    []Operation
}

func (c Company) writeTo(w io.Writer) {
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

func (c *Company) applyOperation(operation Operation) {
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

func (c *Company) addInvalidOperation(operation Operation) {
	c.invalidOperations = append(c.invalidOperations, operation)
}

func writeString(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		return
	}
}

func mapToSlice(m map[string]Company) []Company {
	s := make([]Company, 0, len(m))
	for _, value := range m {
		s = append(s, value)
	}
	return s
}

func writeResult(companies map[string]Company, w io.Writer) {
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

func collectDataFromOperation(company map[string]interface{}, operation map[string]interface{}) {
	for k, v := range operation {
		company[k] = v
	}
}

func isOperationValid(company map[string]interface{}) bool {
	if _, ok := company["company"]; !ok {
		return false
	}
	if _, ok := company["id"]; !ok {
		return false
	}
	if _, ok := company["created_at"]; !ok {
		return false
	} else if _, err := time.Parse(time.RFC3339, fmt.Sprint(company["created_at"])); err != nil {
		return false
	}
	return true
}

func collectData(jsonCompanyMaps []jsonMap, jsonCompanyOperations []map[string]jsonMap) map[string]Company {
	allCompanies := make(map[string]Company)
	for i, companyMap := range jsonCompanyMaps {
		collectDataFromOperation(companyMap, jsonCompanyOperations[i]["operation"])
		if isOperationValid(companyMap) {
			name := fmt.Sprint(companyMap["company"])
			if _, ok := allCompanies[name]; !ok {
				allCompanies[name] = Company{Name: name}
			}

			company := allCompanies[name]
			parsedTime, _ := time.Parse(time.RFC3339, fmt.Sprint(companyMap["created_at"]))
			value, err := strconv.Atoi(fmt.Sprint(companyMap["value"]))
			if err != nil {
				company.addInvalidOperation(Operation{
					ID:          fmt.Sprint(companyMap["id"]),
					CreatedTime: parsedTime,
				})
			} else {
				company.applyOperation(Operation{
					Type:        fmt.Sprint(companyMap["type"]),
					Value:       value,
					ID:          fmt.Sprint(companyMap["id"]),
					CreatedTime: parsedTime,
				})
			}
			allCompanies[name] = company
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

	var jsonCompanyMaps []jsonMap
	var jsonCompanyOperations []map[string]jsonMap

	_ = json.Unmarshal(data, &jsonCompanyMaps)
	_ = json.Unmarshal(data, &jsonCompanyOperations)

	allCompanies := collectData(jsonCompanyMaps, jsonCompanyOperations)

	out, _ := os.Create("out.json")
	writeResult(allCompanies, out)
	if err := out.Close(); err != nil {
		return
	}
}
