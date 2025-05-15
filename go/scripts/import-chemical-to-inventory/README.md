### About

This script imports the [Processed chemical inventory](https://docs.google.com/spreadsheets/d/1kicRikcw6DeY8L0TqNl8FAI3x7T6mW7mTNaMyLkt6zM/edit?gid=390417706#gid=390417706) spreadsheet into the Alchemy Portal database using APIs. We will use this script to load data for both testing and production stages.

### Script design

This script processes chemical data from a CSV file, creating database entries for chemicals, recipes, and instances while tracking processing results.

**Input File**

- Read from a downloaded CSV file (Google Sheet export)
- Rename the file to "chemicals-YYYY-MM-DD-HH-MM.csv" for versioning

**For each row:**

**Step 1: Chemical Processing**

- Process each chemical record in the CSV
- Check if the chemical already exists in the database using the CAS number
- If not found, create a new chemical entry using:
  - Chemical name
  - CAS number
  - Hazard class
  - UN number
  - Safety notes
    Note: Chemical formulas will not be included in this initial version

**Step 2: Recipe Processing**

- Create recipes for each chemical
- Check if a recipe already exists for the given chemical
- If a recipe exists:
  - do nothing.
- If no recipe exists:
  - create a new recipe using: - Title - Description (any notes that might be useful from the spreadsheet) - UUID of the chemical
- Note: All chemicals in this sheet are believed to be supplier chemicals, so the Components field will be empty

** Step 3: Instance Creation(TBD)**

- Create a chemical instance for each chemical based on chemical and recipe

**Result Tracking**

- The script will generate a processed log CSV with the following information:

```
type ProcessingResult struct {
	FileRowNum  int
	Step        string // check chemical, create chemical, or etc.
	Status      string // success or error
	DatabaseID  string // ID, if successfully pushed to the database
	ErrorMsg    string
	ProcessedAt time.Time
}
```

**Processing Flow**

- The script will process the entire CSV file only one time
- Line by line: Each step will be completed for all rows before moving to the next step
- A summary of the processing will be printed at the end of the script. Example:

```
=== Processing Summary ===
Log file created:              log-2025-05-15-15-25.csv
Total rows processed:          1113
Chemicals created:             502
Chemical recipes created:      0
Empty recipe rows:             331

=== Error Summary ===
Total errors:                        781
Breakdown:
        - Missing chemical name errors:    2
        - Check chemical errors:           0
        - Create chemical errors:          0
        - Missing chemical ID errors:      0
        - Check recipe errors:             779
        - Create recipe errors:            0

=== Consistency Check ===
Is total error count correct?  true
```

### Run the script

`go run main.go`
