### About

This script imports the [Processed chemical inventory](https://docs.google.com/spreadsheets/d/1kicRikcw6DeY8L0TqNl8FAI3x7T6mW7mTNaMyLkt6zM/edit?gid=390417706#gid=390417706) spreadsheet into the Alchemy Portal database using APIs. We will use this script to load data for both testing and production stages.

### Script design

This script processes chemical data from a CSV file, creating database entries for chemicals, recipes, and instances while tracking processing results.

**Input File**

- Read from a downloaded CSV file (Google Sheet export)
- Rename the file to "chemicals-YYYY-MM-DD-HH-MM.csv" for versioning

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
  - Compare it with the existing database entry
  - Flag any differences but do not update the database
- If no recipe exists:
  - Create a new recipe using: - Title - Description (any notes that might be useful from the spreadsheet) - UUID of the chemical
    Note: All chemicals in this sheet are believed to be supplier chemicals, so the Components field will be empty

**Step 3: Instance Creation**

- Create a chemical instance for each chemical
- Create a recipe instance for each chemical instance
- This step will be implemented after the first two steps are completed (details TBD)

**Result Tracking**

- The script will generate a processed log CSV with the following information:
  - type of record (chemical, recipe, safetyNote)
  - UUID of each created chemical and recipe
  - Processing status (success/failure)
  - Detailed error messages for failed operations
  - Differences between existing and new recipes/chemicals (maybe)
  - Timestamp ("Processed At") for each record

**Processing Flow**

- The script will process the entire CSV file three times (once for each step)
- Line by line: Each step will be completed for all rows before moving to the next step

### Run the script

`go run main.go`
