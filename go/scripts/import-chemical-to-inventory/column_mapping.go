package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Columns holds all possible column indices
type Columns struct {
	// Chemical columns
	ChemicalName               int
	CasNumber                  int
	UnNumber                   int
	HazardClass                int
	GhsFlammableLiquidCategory int

	// Recipe columns
	RecipeTitle int

	// Supplier columns
	SupplierName int

	// Location columns
	LocationName int

	// Instance columns
	Ciid           int
	LotNumber      int
	Amount         int
	ExpirationDate int
	ParentID       int
	Label          int

	// API configuration
	ApiBaseUrl string
	RawCsv     string
}

// NewColumns creates a new Columns structure with all indices initialized to -1
func NewColumns() *Columns {
	return &Columns{
		ChemicalName:               -1,
		CasNumber:                  -1,
		UnNumber:                   -1,
		HazardClass:                -1,
		GhsFlammableLiquidCategory: -1,
		RecipeTitle:                -1,
		SupplierName:               -1,
		LocationName:               -1,
		Ciid:                       -1,
		LotNumber:                  -1,
		Amount:                     -1,
		ExpirationDate:             -1,
		ParentID:                   -1,
		Label:                      -1,
	}
}

// LetterToIndex converts Excel-style column letter to a zero-based index
// e.g., A->0, B->1, ..., Z->25, AA->26, etc.
func LetterToIndex(colLetter string) int {
	colLetter = strings.ToUpper(colLetter)
	result := 0
	for _, char := range colLetter {
		result = result*26 + int(char-'A'+1)
	}
	return result - 1 // Convert to 0-based index
}

// LoadFromEnv loads column mappings from environment variables
func (c *Columns) LoadFromEnv(filename string) error {
	if err := godotenv.Load(filename); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	envMappings := map[string]*int{
		"COLUMN_CHEMICAL_NAME":                 &c.ChemicalName,
		"COLUMN_CAS_NUMBER":                    &c.CasNumber,
		"COLUMN_UN_NUMBER":                     &c.UnNumber,
		"COLUMN_HAZARD_CLASS":                  &c.HazardClass,
		"COLUMN_GHS_FLAMMABLE_LIQUID_CATEGORY": &c.GhsFlammableLiquidCategory,
		"COLUMN_RECIPE_TITLE":                  &c.RecipeTitle,
		"COLUMN_SUPPLIER_NAME":                 &c.SupplierName,
		"COLUMN_LOCATION_NAME":                 &c.LocationName,
		"COLUMN_CIID":                          &c.Ciid,
		"COLUMN_LOT_NUMBER":                    &c.LotNumber,
		"COLUMN_AMOUNT":                        &c.Amount,
		"COLUMN_EXPIRATION_DATE":               &c.ExpirationDate,
		"COLUMN_PARENT_ID":                     &c.ParentID,
		"COLUMN_LABEL":                         &c.Label,
	}

	for envName, columnPtr := range envMappings {
		colLetter := os.Getenv(envName)
		if colLetter != "" {
			*columnPtr = LetterToIndex(colLetter)
		}
	}

	return nil
}

func (c *Columns) HasColumn(columnIndex int) bool {
	return columnIndex >= 0
}

// GetValueFromRow safely gets a value from a row using a column index
func (c *Columns) GetValueFromRow(row []string, columnIndex int) (string, error) {
	if !c.HasColumn(columnIndex) {
		return "", fmt.Errorf("column index %d is not available", columnIndex)
	}

	if columnIndex >= len(row) {
		return "", fmt.Errorf("index %d is out of range for row with length %d", columnIndex, len(row))
	}

	return row[columnIndex], nil
}

// GetOptionalValueFromRow gets a value from a row with a default fallback
func (c *Columns) GetOptionalValueFromRow(row []string, columnIndex int, defaultValue string) string {
	if !c.HasColumn(columnIndex) {
		return defaultValue
	}

	if columnIndex >= len(row) {
		return defaultValue
	}

	value := row[columnIndex]
	if value == "" {
		return defaultValue
	}

	return value
}
