package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"resty.dev/v3"
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
type PortalChemicalInstance struct {
	UUID             uuid.UUID                 `json:"uuid"`
	ID               int64                     `json:"id"`
	RecipeUUID       uuid.UUID                 `json:"recipeUUID"`
	Amount           float64                   `json:"amount"`
	Owner            uuid.UUID                 `json:"owner"`
	Components       []PortalComponentInstance `json:"inputComponentInstances"`
	HomeLocationUUID uuid.UUID                 `json:"locationUUID"`
	SupplierUUID     uuid.UUID                 `json:"supplierUUID"`
	ParentUUID       uuid.UUID                 `json:"parentUUID"`

	ManufactureDate string  `json:"manufactureDate"`
	ExpirationDate  string  `json:"expirationDate"`
	LotNumber       string  `json:"lotNumber"`
	Label           string  `json:"label"`
	GrossWeight     float64 `json:"grossWeight"`
	NetWeight       float64 `json:"netWeight"`
	Notes           string  `json:"notes"`
}

type PortalComponentInstance struct {
	ChemicalInstanceUUID uuid.UUID `json:"chemicalInstanceUUID"`
	Amount               float64   `json:"amount"`
	Unit                 string    `json:"unit"`
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
	Title        string            `json:"name"`         // 99.9%
	ChemicalUUID uuid.UUID         `json:"chemicalUUID"` // UUID of butuanol
	Components   []PortalComponent `json:"components"`   // emtpy if it's a supplied chemical - list of inputs
}

type PayloadChemicalInstance struct {
	ID               int64                     `json:"id"`
	RecipeUUID       uuid.UUID                 `json:"recipeUUID"`
	Amount           float64                   `json:"amount"`
	Owner            uuid.UUID                 `json:"owner"`
	Components       []PortalComponentInstance `json:"inputComponentInstances"`
	HomeLocationUUID uuid.UUID                 `json:"locationUUID"`
	SupplierUUID     uuid.UUID                 `json:"supplierUUID"`
	ParentUUID       uuid.UUID                 `json:"parentUUID"`

	ManufactureDate string  `json:"manufactureDate"`
	ExpirationDate  string  `json:"expirationDate"`
	LotNumber       string  `json:"lotNumber"`
	Label           string  `json:"label"`
	GrossWeight     float64 `json:"grossWeight"`
	NetWeight       float64 `json:"netWeight"`
	Notes           string  `json:"notes"`
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
	createdChemicalCount := 0
	createdRecipeCount := 0
	chemicalValidationErrorCount := 0
	emptyRecipeCount := 0
	checkChemicalErrorCount := 0
	createChemicalErrorCount := 0
	missingChemicalIDErrorCount := 0
	createRecipeErrorCount := 0
	checkRecipeErrorCount := 0
	errorCount := 0
	for {
		fmt.Printf("\rProcessing row %d \n", rowNum)
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break // End of file
			}
			fmt.Printf("Error reading row %d: %v - skipping\n", rowNum, err)
			writeProcessedLog(writer, rowNum, "Read row", "cannot read", "", err.Error())
			rowNum++
			errorCount++
			continue
		}

		fmt.Println("Step 0: Validating required fields")
		err = checkIfRequiredFieldsPresent("chemical", row)
		if err != nil {
			fmt.Printf("Validation error in row %d: %v - skipping\n", rowNum, err)
			writeProcessedLog(writer, rowNum, "Validate row ", "missing chemical name", "", err.Error())
			rowNum++
			errorCount++
			chemicalValidationErrorCount++
			continue
		}

		fmt.Println("Step 1: Processing chemical data and safety info")

		notes := ""
		if row[18] != "" {
			notes = "GHS Flammable liquid category: " + row[18]
		}

		pChemical := PayloadChemical{
			Name: removeExtraSpace(row[4]),
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
			errorCount++
			checkChemicalErrorCount++
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
				createChemicalErrorCount++
				errorCount++
				continue
			}
			fmt.Printf("Created new chemical %s with ID %s\n", pChemical.Name, chemicalID)
			writeProcessedLog(writer, rowNum, "Create new chemical", "success", chemicalID, "")
			createdChemicalCount++
		}

		fmt.Println("Step 2: Processing chemical recipe data - if there's chem recipe information to process")

		err = checkIfRequiredFieldsPresent("recipe", row)
		if err != nil {
			fmt.Printf("Recipe title is empty - skipping\n")
			writeProcessedLog(writer, rowNum, "Validate row", "missing recipe title", "", "recipe title is empty")
			emptyRecipeCount++
		} else {
			// this checmicalID check is cuz sometimes, the check if chemical exist step fails unexpectedly
			// main hypothesis is due to special character
			// will look into this later; for now, we will have this chemicalID check
			if chemicalID == "" {
				fmt.Printf("Error - chemicalID is empty - skipping\n")
				writeProcessedLog(writer, rowNum, "Validate chemical ID", "missing chemical ID", "", "no chemical ID available")
				missingChemicalIDErrorCount++
				errorCount++
				continue
			}

			recipeTitle := removeExtraSpace(row[5])
			chemicalUUID := uuid.MustParse(chemicalID)

			pRecipe := PayloadChemicalRecipe{
				Title:        recipeTitle,
				ChemicalUUID: chemicalUUID,
			}

			fmt.Printf("double check payload info as right now the check won't work any ways - title: %s\n", pRecipe.Title)
			fmt.Printf("double check  chemicalID: %s\n", pRecipe.ChemicalUUID)
			fmt.Printf("ChemicalUUID type: %T\n", pRecipe.ChemicalUUID)

			res, recipeID, err := checkIfChemicalRecipeExistsInDB(pRecipe.Title, chemicalID)
			if err != nil {
				fmt.Printf("Error checking if chemical recipe exists in DB: %v - skipping\n", err)
				writeProcessedLog(writer, rowNum, "Check if chemical recipe already exists", "cannot check if chemical recipe exists", "", err.Error())
				rowNum++
				errorCount++
				checkRecipeErrorCount++
				continue
			}

			if res {
				fmt.Printf("Chemical recipe %s already exists in DB - skipping\n", pRecipe.Title)
				writeProcessedLog(writer, rowNum, "Check if chemical recipe already exists", "success", recipeID, "")
			} else {
				recipeID, err := createNewChemicalRecipe(pRecipe)
				if err != nil {
					fmt.Printf("Error creating new chemical recipe: %v - skipping\n", err)
					writeProcessedLog(writer, rowNum, "Create new chemical recipe", "cannot create new chemical recipe", "", err.Error())
					createRecipeErrorCount++
					errorCount++
					rowNum++
					continue
				}
				fmt.Printf("Created new chemical recipe %s with ID %s\n", pRecipe.Title, recipeID)
				writeProcessedLog(writer, rowNum, "Create new chemical recipe", "success", recipeID, "")
				createdRecipeCount++
			}
		}
		// TODO -3. create instance

		rowNum++

	}
	fmt.Println()

	fmt.Println("\n=== Processing Summary ===")
	fmt.Printf("Log file created:              %s\n", processedLog.Name())
	fmt.Printf("Total rows processed:          %d\n", rowNum)
	fmt.Printf("Chemicals created:             %d\n", createdChemicalCount)
	fmt.Printf("Chemical recipes created:      %d\n", createdRecipeCount)
	fmt.Printf("Empty recipe rows:             %d\n", emptyRecipeCount)

	fmt.Println("\n=== Error Summary ===")
	fmt.Printf("Total errors:                        %d\n", errorCount)
	fmt.Printf("Breakdown:\n")
	fmt.Printf("\t- Missing chemical name errors:    %d\n", chemicalValidationErrorCount)
	fmt.Printf("\t- Check chemical errors:           %d\n", checkChemicalErrorCount)
	fmt.Printf("\t- Create chemical errors:          %d\n", createChemicalErrorCount)
	fmt.Printf("\t- Missing chemical ID errors:      %d\n", missingChemicalIDErrorCount)
	fmt.Printf("\t- Check recipe errors:             %d\n", checkRecipeErrorCount)
	fmt.Printf("\t- Create recipe errors:            %d\n", createRecipeErrorCount)

	fmt.Println("\n=== Consistency Check ===")
	fmt.Printf("Is total error count correct?  %t\n",
		errorCount == (chemicalValidationErrorCount+
			checkChemicalErrorCount+
			createChemicalErrorCount+
			missingChemicalIDErrorCount+
			checkRecipeErrorCount+
			createRecipeErrorCount),
	)

}

// --- helper functions ---

func removeExtraSpace(s string) string {
	s = strings.TrimRight(s, " ")
	s = strings.TrimLeft(s, " ")
	return s
}

var client = resty.New().SetBaseURL("http://192.168.2.2:8092")

func checkIfChemicalExistsInDB(name string) (bool, string, error) {
	var result PortalChemical

	resp, err := client.R().
		SetQueryParam("name", name).
		SetResult(&result).
		Get("/chemicals/name")

	if err != nil {
		return false, "", fmt.Errorf("failed to check if chemical exists: %w", err)
	}

	// TODO: we will update the API to return 404 if chemical not found
	if resp.StatusCode() == 500 {
		return false, "", nil
	}

	if resp.StatusCode() == 200 {
		return true, result.ID, nil
	}

	return false, "", fmt.Errorf("unexpected response code: %d, body: %s", resp.StatusCode(), resp.String())
}

func checkIfChemicalRecipeExistsInDB(name string, chemicalID string) (bool, string, error) {
	/**
	 * the reason why we are checking by looking up all the recipes given a chemical ID and see if title matches
	 * is cuz we don't have a GET API to check if a recipe exists by recipe name and chemical ID
	 */
	var result []PortalChemicalRecipe

	resp, err := client.R().
		SetResult(&result).
		Get("/chemicals/" + chemicalID + "/recipes")

	if err != nil {
		return false, "", fmt.Errorf("failed to check if chemical recipe exists: %w", err)
	}

	if resp.StatusCode() == 500 {
		return false, "", nil
	}

	if resp.StatusCode() == 200 {
		for _, recipe := range result {
			if recipe.Title == name {
				return true, recipe.ID, nil
			}
		}
		return false, "", nil
	}

	return false, "", fmt.Errorf("unexpected response code: %d, body: %s", resp.StatusCode(), resp.String())
}

func createNewChemical(pChemical PayloadChemical) (string, error) {
	var result PortalChemical

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(pChemical).
		SetResult(&result).
		Post("/chemicals")

	if err != nil {
		return "", fmt.Errorf("failed to create new chemical: %w", err)
	}

	if resp.StatusCode() == 201 {
		return result.ID, nil
	}

	return "", fmt.Errorf("failed to create new chemical, status code: %d, response: %s",
		resp.StatusCode(), resp.String())
}

func createNewChemicalRecipe(pRecipe PayloadChemicalRecipe) (string, error) {
	var result PortalChemicalRecipe

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(pRecipe).
		SetResult(&result).
		Post("/recipes/")

	if err != nil {
		return "", fmt.Errorf("failed to create new chemical recipe: %w", err)
	}

	if resp.StatusCode() == 201 {
		return result.ID, nil
	}

	return "", fmt.Errorf("failed to create new chemical recipe, status code: %d, response: %s",
		resp.StatusCode(), resp.String())
}

func checkIfRequiredFieldsPresent(recordType string, row []string) error {
	if recordType == "chemical" {
		if row[4] == "" {
			return fmt.Errorf("missing chemical name")
		}

		return nil
	}

	if recordType == "recipe" {
		if row[5] == "" {
			return fmt.Errorf("missing recipe title")
		}

		return nil
	}

	return fmt.Errorf("unknown record type")
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

// ---- old code ----

// func checkIfChemicalExistsInDB(name string) (bool, string, error) {
// 	apiURL := "http://192.168.2.2:8092/chemicals/name?name=" + url.PathEscape(name)
// 	resp, err := http.Get(apiURL)

// 	if err != nil {
// 		return false, "", fmt.Errorf("failed to check if chemical exists: %w", err)
// 	}

// 	defer resp.Body.Close()
// 	bodyBytes, _ := io.ReadAll(resp.Body)
// 	bodyStr := string(bodyBytes)

// 	// TODO: we will update the API to return 404 if chemical not found
// 	if resp.StatusCode == http.StatusInternalServerError {
// 		return false, "", nil
// 	}

// 	if resp.StatusCode == http.StatusOK {
// 		var result PortalChemical
// 		if err := json.Unmarshal(bodyBytes, &result); err != nil {
// 			return false, "", fmt.Errorf("failed to decode response: %w", err)
// 		}
// 		return true, result.ID, nil
// 	}

// 	return false, "", fmt.Errorf("unexpected response code: %d, body: %s", resp.StatusCode, bodyStr)
// }

// func checkIfChemicalRecipeExistsInDB(name string, chemicalID string) (bool, string, error) {

// 	/**
// 	the reason why we are checking by looking up all the recipes given a chemical ID and see if title matches
// 	is cuz we don't have a GET API to check if a recipe exists by recipe name and chemical ID
// 	*/

// 	apiURL := "http://192.168.2.2:8092/chemicals/" + chemicalID + "/recipes/"
// 	resp, err := http.Get(apiURL)
// 	if err != nil {
// 		return false, "", fmt.Errorf("failed to check if chemical recipe exists: %w", err)
// 	}
// 	defer resp.Body.Close()
// 	bodyBytes, _ := io.ReadAll(resp.Body)
// 	bodyStr := string(bodyBytes)
// 	if resp.StatusCode == http.StatusInternalServerError {
// 		return false, "", nil
// 	}
// 	if resp.StatusCode == http.StatusOK {
// 		var result []PortalChemicalRecipe
// 		if err := json.Unmarshal(bodyBytes, &result); err != nil {
// 			return false, "", fmt.Errorf("failed to decode response: %w", err)
// 		}
// 		for _, recipe := range result {
// 			if recipe.Title == name {
// 				return true, recipe.ID, nil
// 			}
// 		}
// 		return false, "", nil
// 	}
// 	return false, "", fmt.Errorf("unexpected response code: %d, body: %s", resp.StatusCode, bodyStr)
// }

// func createNewChemical(pChemical PayloadChemical) (string, error) {
// 	payload, err := json.Marshal(pChemical)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal chemical data: %w", err)
// 	}

// 	apiURL := "http://192.168.2.2:8092/chemicals"

// 	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(payload))

// 	if err != nil {
// 		return "", fmt.Errorf("failed to create new chemical: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusCreated {
// 		var result PortalChemical
// 		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 			return "", fmt.Errorf("failed to decode response: %w", err)
// 		}
// 		return result.ID, nil
// 	}

// 	bodyBytes, _ := io.ReadAll(resp.Body)
// 	return "", fmt.Errorf("failed to create new chemical, status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
// }

// func createNewChemicalRecipe(pRecipe PayloadChemicalRecipe) (string, error) {
// 	payload, err := json.Marshal(pRecipe)

// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal chemical recipe data: %w", err)
// 	}

// 	apiURL := "http://192.168.2.2:8092/recipes/"

// 	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(payload))

// 	if err != nil {
// 		return "", fmt.Errorf("failed to create new chemical recipe: %w", err)
// 	}

// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusCreated {
// 		var result PortalChemicalRecipe
// 		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 			return "", fmt.Errorf("failed to decode response: %w", err)
// 		}
// 		return result.ID, nil
// 	}

// 	bodyBytes, _ := io.ReadAll(resp.Body)
// 	return "", fmt.Errorf("failed to create new chemical recipe, status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
// }
