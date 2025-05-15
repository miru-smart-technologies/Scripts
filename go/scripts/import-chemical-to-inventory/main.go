package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// --- database models ---

type PortalChemical struct { // <<<<<<<
	ID              string           `json:"id"`
	Name            string           `json:"name"`          // 1-butanol
	Formula         string           `json:"formula"`       // nice to have
	StateOfMatter   string           `json:"stateOfMatter"` // solid, liquid, gas (how to know that???) maybe
	IsProduct       bool             `json:"isProduct"`
	Type            string           `json:"type"` //IGNORE whatever is not on alchemy portal
	Description     string           `json:"description"`
	MolecularWeight int64            `json:"molecularWeight"`
	Density         float64          `json:"density"`
	SafetyInfo      PortalSafetyInfo `json:"safetyInfo"` // JSON object
}

type PortalSafetyInfo struct { // <<<<<<<
	CasNumber   string `json:"casNumber"`   // 71-36-3
	UNNumber    string `json:"unNumber"`    // UN1120
	HazardClass string `json:"hazardClass"` // 3
	SafetyNotes string `json:"safetyNotes"` // put other safety info here: "GHS Flammable liquid cateogry: 3;"
}

type PortalComponent struct { // <<<<<<<
	RecipeUUID  uuid.UUID `json:"recipeUUID"`
	RecipeTitle string    `json:"recipeTitle"`
	Fraction    float64   `json:"fraction"`
}

type PortalChemicalRecipe struct { // <<<<<<<
	ID           string    `json:"id"`
	Title        string    `json:"title"`        // 99.9%
	ChemicalUUID uuid.UUID `json:"chemicalUUID"` // UUID of butuanol
	Description  string    `json:"description"`  // ignore
	Tags         []string  `json:"tags"`         // ignore
	// ProcessSet   *PortalProcessSet `json:"processSet"`
	Components []PortalComponent `json:"components"` // emtpy if it's a supplied chemical - list of inputs
}

// --- payload models ---

type PayloadSafetyInfo struct { // <<<<<<<
	CasNumber   string `json:"casNumber"`   // 71-36-3
	UNNumber    string `json:"unNumber"`    // UN1120
	HazardClass string `json:"hazardClass"` // 3
	SafetyNotes string `json:"safetyNotes"` // put other safety info here: "GHS Flammable liquid cateogry: 3;"
}

type PayloadChemical struct { // <<<<<<<
	Name        string           `json:"name"` // 1-butanol
	Description string           `json:"description"`
	SafetyInfo  PortalSafetyInfo `json:"safetyInfo"` // JSON object
}

type PayloadComponent struct { // <<<<<<<
	RecipeUUID  uuid.UUID `json:"recipeUUID"`
	RecipeTitle string    `json:"recipeTitle"`
	Fraction    float64   `json:"fraction"`
}

type PayloadChemicalRecipe struct { // <<<<<<<
	ID           string            `json:"id"`
	Title        string            `json:"title"`        // 99.9%
	ChemicalUUID uuid.UUID         `json:"chemicalUUID"` // UUID of butuanol
	Components   []PortalComponent `json:"components"`   // emtpy if it's a supplied chemical - list of inputs
}

type ProcessingResult struct {
	FileRowNum  int
	Step        string // check chemical, create chemical, or etc.
	Status      string // success or error
	DatabaseID  string // ID, if successfully pushed to the database
	ErrorMsg    string
	ProcessedAt time.Time
}

// --- main function ---
func main() {
	// 1. prepare the processed log file
	processedLog, err := os.Create("log-" + time.Now().Format("2006-01-02-15-04") + ".csv")

	if err != nil {
		log.Fatalf("failed to create log file: %s", err)

	}
	defer processedLog.Close()

	writer := csv.NewWriter(processedLog)
	defer writer.Flush()

	writer.Write([]string{"FileRowNum",
		"Type",
		"Status",
		"DatabaseID",
		"ErrorMsg",
		"ProcessedAt"})

	// 2. open the CSV file
	filename := "chemicals-05-12-11-17.csv"

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read() // Skip header line
	if err != nil {
		log.Fatalf("failed to read header: %v", err)
	}

	// 3. read the CSV file line by line
	rowNum := 1
	for {
		fmt.Printf("Processing row %d\n", rowNum)
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break // End of file
			}
			fmt.Printf("Error reading row %d: %v - skipping\n", rowNum, err)
			writeProcessedLog(writer, rowNum, "Read row", "cannot read", "", err.Error())
			rowNum++
			continue
		}

		fmt.Println("Step 0: Validating required fields")
		err = checkIfRequiredFieldsPresent(row)
		if err != nil {
			fmt.Printf("Validation error in row %d: %v - skipping\n", rowNum, err)
			writeProcessedLog(writer, rowNum, "Validate row", "missing required fields", "", err.Error())
			rowNum++
			continue
		}

		fmt.Println("Step 1: Processing chemical data and safety info")

		notes := ""
		if row[18] != "" {
			notes = "GHS Flammable liquid category: " + row[18]
		}

		pChemical := PayloadChemical{
			Name: row[4],
			SafetyInfo: PortalSafetyInfo{
				CasNumber:   row[6],
				UNNumber:    row[15],
				HazardClass: row[16],
				SafetyNotes: notes,
			},
		}

		res, chemicalID, err := checkIfChemicalExistsInDB(pChemical.Name)
		if err != nil {
			fmt.Printf("Error checking if chemical exists in DB: %v - skipping\n", err)
			writeProcessedLog(writer, rowNum, "Check if chemical already exists", "cannot check if chemical exists", "", err.Error())
			rowNum++
			continue
		}

		if res {
			fmt.Printf("Chemical %s already exists in DB - skipping\n", pChemical.Name)
			writeProcessedLog(writer, rowNum, "Check if chemical already exists", "success", chemicalID, "")
		} else {
			chemicalID, err = createNewChemical(pChemical)
			if err != nil {
				fmt.Printf("Error creating new chemical: %v - skipping\n", err)
				writeProcessedLog(writer, rowNum, "Create new chemical", "cannot create new chemical", "", err.Error())
				rowNum++
				continue
			}
			fmt.Printf("Created new chemical %s with ID %s\n", pChemical.Name, chemicalID)
			writeProcessedLog(writer, rowNum, "Create new chemical", "success", chemicalID, "")
		}

		//TODO: step 2: create chemical recipe
		rowNum++

	}
}

// --- helper functions ---

func checkIfChemicalExistsInDB(name string) (bool, string, error) {
	if strings.Contains(name, "(") || strings.Contains(name, ")") {
		name = url.PathEscape(name)
	}
	apiURL := "http://192.168.2.2:8092/chemicals/name/" + name
	resp, err := http.Get(apiURL)

	if err != nil {
		return false, "", fmt.Errorf("failed to check if chemical exists: %w", err)
	}

	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	// TODO: we will update the API to return 404 if chemical not found
	if resp.StatusCode == http.StatusInternalServerError {
		return false, "", nil
	}

	if resp.StatusCode == http.StatusOK {
		var result PortalChemical
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return false, "", fmt.Errorf("failed to decode response: %w", err)
		}
		return true, result.ID, nil
	}

	return false, "", fmt.Errorf("unexpected response code: %d, body: %s", resp.StatusCode, bodyStr)
}

func createNewChemical(pChemical PayloadChemical) (string, error) {
	payload, err := json.Marshal(pChemical)
	if err != nil {
		return "", fmt.Errorf("failed to marshal chemical data: %w", err)
	}

	apiURL := "http://192.168.2.2:8092/chemicals"

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(payload))

	if err != nil {
		return "", fmt.Errorf("failed to create new chemical: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		var result PortalChemical
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", fmt.Errorf("failed to decode response: %w", err)
		}
		return result.ID, nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return "", fmt.Errorf("failed to create new chemical, status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
}

func checkIfRequiredFieldsPresent(row []string) error {
	if row[4] == "" {
		return fmt.Errorf("missing chemical name")
	}

	if row[5] == "" {
		return fmt.Errorf("missing recipe title")
	}

	return nil
}

func writeProcessedLog(writer *csv.Writer,
	fileRowNum int,
	recordType string,
	status string,
	databaseID string,
	errorMsg string) {

	entry := ProcessingResult{
		FileRowNum:  fileRowNum,
		Step:        recordType,
		Status:      status,
		DatabaseID:  databaseID,
		ErrorMsg:    errorMsg,
		ProcessedAt: time.Now(),
	}

	writer.Write([]string{
		strconv.Itoa(entry.FileRowNum),
		entry.Step,
		entry.Status,
		entry.DatabaseID,
		entry.ErrorMsg,
		entry.ProcessedAt.Format(time.RFC3339),
	})
}
