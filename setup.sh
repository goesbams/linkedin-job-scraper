#!/bin/bash

# LinkedIn Indonesian Employee Scraper Setup Guide
# ================================================

echo "🇮🇩 LinkedIn Indonesian Employee Scraper Setup"
echo "================================================"

# Create project directory structure
echo "📁 Creating project directory structure..."
mkdir -p linkedin-scraper/{names,data,results}
cd linkedin-scraper

# Initialize Go module
echo "🚀 Initializing Go module..."
go mod init linkedin-scraper
go get github.com/PuerkitoBio/goquery

# Create the project structure
echo "📝 Creating project files..."

# Create go.mod if not exists
cat > go.mod << 'EOF'
module linkedin-scraper

go 1.21

require github.com/PuerkitoBio/goquery v1.8.1

require (
	github.com/andybalholm/cascadia v1.3.1 // indirect
	golang.org/x/net v0.7.0 // indirect
)
EOF

# Create main scraper file
echo "Creating main.go..."
# (The main scraper code would go here - reference to the artifact)

# Create names package
echo "Creating names package..."
mkdir -p names
# (The names package code would go here - reference to the artifact)

# Generate names database
echo "🗃️ Generating Indonesian names database..."
# (The database generator would run here)

# Create run script
cat > run.sh << 'EOF'
#!/bin/bash

# LinkedIn Scraper Runner Script
# Usage: ./run.sh <country> <job_title> [limit]

if [ $# -lt 2 ]; then
    echo "Usage: $0 <country> <job_title> [limit]"
    echo "Examples:"
    echo "  $0 \"Germany\" \"golang developer\" 50"
    echo "  $0 \"Austria\" \"software engineer\" 25"
    echo "  $0 \"Estonia\" \"backend developer\" 30"
    exit 1
fi

COUNTRY="$1"
JOB_TITLE="$2"
LIMIT="${3:-25}"

echo "🔍 Searching for '$JOB_TITLE' jobs in $COUNTRY (limit: $LIMIT)"
echo "🇮🇩 Checking for Indonesian employees..."

# Run the scraper
go run main.go "$COUNTRY" "$JOB_TITLE" "$LIMIT"

# Check if results were generated
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULT_FILE="linkedin_jobs_$(echo "$COUNTRY" | tr '[:upper:]' '[:lower:]' | tr ' ' '_')_$(echo "$JOB_TITLE" | tr '[:upper:]' '[:lower:]' | tr ' ' '_')_*.json"

if ls $RESULT_FILE 1> /dev/null 2>&1; then
    echo "📁 Moving results to results directory..."
    mv $RESULT_FILE results/
    echo "✅ Results saved in results directory"
else
    echo "⚠️  No results file found"
fi
EOF

chmod +x run.sh

# Create analysis script
cat > analyze.py << 'EOF'
#!/usr/bin/env python3
"""
LinkedIn Scraper Results Analyzer
Analyzes the JSON results from the scraper
"""

import json
import sys
import os
from datetime import datetime
from collections import Counter

def analyze_results(filename):
    """Analyze scraper results"""
    try:
        with open(filename, 'r', encoding='utf-8') as f:
            data = json.load(f)
    except FileNotFoundError:
        print(f"❌ File not found: {filename}")
        return
    except json.JSONDecodeError:
        print(f"❌ Invalid JSON file: {filename}")
        return
    
    summary = data.get('summary', {})
    jobs = data.get('jobs', [])
    
    print("📊 ANALYSIS RESULTS")
    print("=" * 50)
    print(f"📁 File: {filename}")
    print(f"📅 Generated: {summary.get('generated_at', 'Unknown')}")
    print(f"🔍 Total Jobs: {summary.get('total_jobs', 0)}")
    print(f"🇮🇩 Jobs with Indonesians: {summary.get('jobs_with_indonesians', 0)}")
    print(f"👥 Total Indonesian Employees: {summary.get('total_indonesian_employees', 0)}")
    
    if not jobs:
        print("⚠️  No job data found")
        return
    
    # Analyze companies with Indonesian employees
    indonesian_companies = [job for job in jobs if job.get('has_indonesian', False)]
    
    if indonesian_companies:
        print(f"\n🎯 TOP COMPANIES WITH INDONESIAN EMPLOYEES:")
        print("-" * 50)
        for i, job in enumerate(indonesian_companies[:10], 1):
            emp_count = len(job.get('indonesian_employees', []))
            print(f"{i:2d}. {job.get('company', 'Unknown')} ({emp_count} Indonesian employees)")
            print(f"    💼 Position: {job.get('title', 'Unknown')}")
            print(f"    📍 Location: {job.get('location', 'Unknown')}")
            
            # Show employee names
            employees = job.get('indonesian_employees', [])
            if employees:
                names = [emp.get('name', 'Unknown') for emp in employees[:3]]
                if len(employees) > 3:
                    names.append(f"and {len(employees) - 3} more...")
                print(f"    👥 Employees: {', '.join(names)}")
            print()
    
    # Analyze most common Indonesian names found
    all_names = []
    for job in jobs:
        for emp in job.get('indonesian_employees', []):
            name = emp.get('name', '')
            if name:
                all_names.extend(name.split())
    
    if all_names:
        print(f"\n📈 MOST COMMON INDONESIAN NAMES FOUND:")
        print("-" * 40)
        name_counts = Counter(all_names)
        for name, count in name_counts.most_common(10):
            print(f"   {name}: {count} occurrences")
    
    # Location analysis
    locations = [job.get('location', 'Unknown') for job in indonesian_companies]
    if locations:
        print(f"\n🌍 LOCATIONS WITH INDONESIAN EMPLOYEES:")
        print("-" * 40)
        location_counts = Counter(locations)
        for location, count in location_counts.most_common(5):
            print(f"   {location}: {count} companies")

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 analyze.py <results_file.json>")
        print("\nAvailable result files:")
        results_dir = "results"
        if os.path.exists(results_dir):
            for file in os.listdir(results_dir):
                if file.endswith('.json'):
                    print(f"  - {os.path.join(results_dir, file)}")
        sys.exit(1)
    
    filename = sys.argv[1]
    analyze_results(filename)

if __name__ == "__main__":
    main()
EOF

chmod +x analyze.py

# Create README
cat > README.md << 'EOF'
# LinkedIn Indonesian Employee Scraper

An efficient Go-based web scraper that searches LinkedIn for job postings and identifies companies with Indonesian employees.

## Features

- 🔍 **Job Search**: Search jobs by country and job title
- 🇮🇩 **Indonesian Detection**: Efficiently identifies Indonesian employees using a comprehensive name database
- 📊 **Detailed Analysis**: Provides confidence scores and match reasons
- 💾 **Export Results**: Saves results in JSON format for further analysis
- ⚡ **High Performance**: Uses hash maps for O(1) name lookups instead of loops
- 🗃️ **Modular Database**: 10,000+ Indonesian names organized in separate files

## Quick Start

1. **Setup the project:**
   ```bash
   ./setup.sh
   ```

2. **Run a job search:**
   ```bash
   ./run.sh "Germany" "golang developer" 50
   ```

3. **Analyze results:**
   ```bash
   python3 analyze.py results/linkedin_jobs_germany_golang_developer_*.json
   ```

## Usage Examples

```bash
# Search for Go developers in Germany
./run.sh "Germany" "golang developer" 50

# Search for software engineers in Austria  
./run.sh "Austria" "software engineer" 25

# Search for backend developers in Estonia
./run.sh "Estonia" "backend developer" 30

# Search for any software jobs in Netherlands
./run.sh "Netherlands" "software" 100
```

## Database Structure

The Indonesian names database is organized into modular files:

- `data/first_names.txt` - 3,000+ Indonesian first names
- `data/last_names.txt` - 2,000+ Indonesian last names  
- `data/common_patterns.txt` - 500+ common patterns
- `data/prefixes.txt` - Name prefixes (Abdul, Nur, etc.)
- `data/suffixes.txt` - Name suffixes (wan, wati, etc.)

## Efficiency Features

- **Hash Map Lookups**: O(1) name detection instead of O(n) loops
- **Confidence Scoring**: Each match includes confidence percentage
- **Multiple Fallbacks**: Uses different methods to find employee data
- **Rate Limiting**: Built-in delays to avoid being blocked
- **User Agent Rotation**: Reduces detection probability

## Output Format

Results include:
- Job details (title, company, location, URL)
- Indonesian employee detection status
- List of found Indonesian employees with confidence scores
- Match reasons (first_name, last_name, pattern, etc.)
- Processing time and statistics

## Legal Notes

- Always respect LinkedIn's Terms of Service
- Use reasonable delays between requests
- Consider using LinkedIn's official API for production use
- This tool is for educational and research purposes

## Requirements

- Go 1.21+
- Python 3.6+ (for analysis script)
- Internet connection
- github.com/PuerkitoBio/goquery

## Contributing

Feel free to add more Indonesian names to the database files or improve the detection algorithms!
EOF

echo "✅ Project setup completed!"
echo ""
echo "📁 Directory structure created:"
echo "   linkedin-scraper/"
echo "   ├── main.go (scraper)"
echo "   ├── names/ (names package)"
echo "   ├── data/ (names database)"
echo "   ├── results/ (output directory)"
echo "   ├── run.sh (runner script)"
echo "   ├── analyze.py (analysis tool)"
echo "   └── README.md (documentation)"
echo ""
echo "🚀 Next steps:"
echo "1. Copy the scraper code to main.go"
echo "2. Copy the names package code to names/"
echo "3. Run the database generator"
echo "4. Test with: ./run.sh \"Germany\" \"golang developer\" 25"
echo ""
echo "💡 Remember to respect LinkedIn's Terms of Service!"