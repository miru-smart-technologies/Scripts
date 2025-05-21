package main

import (
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
	Title        string    `json:"name"`         // 99.9%
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
