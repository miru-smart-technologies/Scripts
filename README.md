# Scripts Repository

This repo contains scripts for varoius tasks. The structure of the repo is as follows so that it is easy to add new scripts in the future. The scripts are organized by language, and each script has its own directory with a README file for documentation. There are also shared utilities for each language.

```
Scripts/
├── README.md                    # Repository documentation
├── .gitignore                   # Common files to ignore for multiple languages
├── go/                          # Go scripts directory
│   ├── go.mod                   # Go module definition
│   ├── go.sum                   # Go dependency lockfile
│   ├── scripts/                 # Go scripts
│   │   └── new-go-script/       # Directory for a Go script
│   │       ├── README.md        # Documentation
│   │       └── main.go          # Script implementation
│   └── pkg/                     # Shared Go packages
│       └── common/              # Common utilities for Go scripts
│           └── e.g.common.go
├── python/                      # Python scripts directory
│   ├── requirements.txt         # Shared Python dependencies
│   ├── scripts/                 # Python scripts
│   │   └── new-python-script/   # Directory for a Python script
│   │       ├── README.md        # Documentation
│   │       ├── script.py        # Script implementation
│   │       └── requirements.txt # Script-specific dependencies
│   └── utils/                   # Shared Python utilities
│       └── e.g.common.py        # Common utilities for Python scripts
└── docs/                        # General documentation
    └── contributing.md          # Guidelines for adding new scripts
```
