package names

import (
	"bufio"
	"os"
	"strings"
	"unicode"
)

// NameDB represents the Indonesian names database with efficient lookup
type NameDB struct {
	firstNames     map[string]bool
	lastNames      map[string]bool
	commonPatterns map[string]bool
	prefixes       map[string]bool
	suffixes       map[string]bool
}

// NewNameDB creates and initializes the Indonesian names database
func NewNameDB() (*NameDB, error) {
	db := &NameDB{
		firstNames:     make(map[string]bool),
		lastNames:      make(map[string]bool),
		commonPatterns: make(map[string]bool),
		prefixes:       make(map[string]bool),
		suffixes:       make(map[string]bool),
	}

	// Load all name files
	files := []struct {
		filename string
		target   map[string]bool
	}{
		{"data/first_names.txt", db.firstNames},
		{"data/last_names.txt", db.lastNames},
		{"data/common_patterns.txt", db.commonPatterns},
		{"data/prefixes.txt", db.prefixes},
		{"data/suffixes.txt", db.suffixes},
	}

	for _, file := range files {
		if err := db.loadNamesFromFile(file.filename, file.target); err != nil {
			return nil, err
		}
	}

	return db, nil
}

// loadNamesFromFile loads names from file into the target map
func (db *NameDB) loadNamesFromFile(filename string, target map[string]bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			// Store in lowercase for case-insensitive lookup
			target[strings.ToLower(line)] = true
		}
	}

	return scanner.Err()
}

// IsIndonesianName checks if a full name appears to be Indonesian using efficient lookup
func (db *NameDB) IsIndonesianName(fullName string) (bool, []string) {
	if fullName == "" {
		return false, nil
	}

	// Clean and normalize the name
	cleanName := db.cleanName(fullName)
	nameParts := strings.Fields(cleanName)

	if len(nameParts) == 0 {
		return false, nil
	}

	var matches []string
	score := 0
	totalParts := len(nameParts)

	// Check each part of the name
	for i, part := range nameParts {
		partLower := strings.ToLower(part)

		// Check first names (higher weight for first position)
		if db.firstNames[partLower] {
			if i == 0 {
				score += 3 // Higher weight for first position
			} else {
				score += 2
			}
			matches = append(matches, "first_name:"+part)
		}

		// Check last names (higher weight for last position)
		if db.lastNames[partLower] {
			if i == totalParts-1 {
				score += 3 // Higher weight for last position
			} else {
				score += 2
			}
			matches = append(matches, "last_name:"+part)
		}

		// Check common patterns
		if db.commonPatterns[partLower] {
			score += 2
			matches = append(matches, "pattern:"+part)
		}

		// Check prefixes and suffixes
		if db.checkPrefixSuffix(partLower) {
			score += 1
			matches = append(matches, "affix:"+part)
		}
	}

	// Check for Indonesian-specific patterns in the full name
	if db.hasIndonesianPatterns(cleanName) {
		score += 1
		matches = append(matches, "indonesian_pattern")
	}

	// Determine if name is Indonesian based on score
	// Threshold based on name length and matches
	threshold := 1
	if totalParts > 2 {
		threshold = 2
	}

	isIndonesian := score >= threshold

	return isIndonesian, matches
}

// checkPrefixSuffix checks if a name part contains Indonesian prefixes or suffixes
func (db *NameDB) checkPrefixSuffix(part string) bool {
	// Check prefixes
	for prefix := range db.prefixes {
		if strings.HasPrefix(part, prefix) {
			return true
		}
	}

	// Check suffixes
	for suffix := range db.suffixes {
		if strings.HasSuffix(part, suffix) {
			return true
		}
	}

	return false
}

// hasIndonesianPatterns checks for Indonesian-specific naming patterns
func (db *NameDB) hasIndonesianPatterns(name string) bool {
	nameLower := strings.ToLower(name)

	// Indonesian naming patterns
	patterns := []string{
		"bin ", "binti ", // Arabic influence
		"van ", "de ", // Dutch colonial influence
		"abdul", "muhammad", "ahmad", // Arabic names common in Indonesia
	}

	for _, pattern := range patterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}

	// Check for sequential naming (I Made, I Gede, etc.)
	if strings.HasPrefix(nameLower, "i ") {
		return true
	}

	return false
}

// cleanName cleans and normalizes a name string
func (db *NameDB) cleanName(name string) string {
	// Remove extra whitespace and normalize
	name = strings.TrimSpace(name)
	name = strings.Join(strings.Fields(name), " ")

	// Remove common titles and suffixes
	titles := []string{
		"dr.", "dr", "prof.", "prof", "ir.", "ir", "st.", "mt.", "s.kom", "s.t",
		"m.kom", "m.t", "ph.d", "phd", "mba", "cpa", "ca", "jr.", "sr.", "ii", "iii",
	}

	for _, title := range titles {
		// Remove from beginning
		if strings.HasPrefix(strings.ToLower(name), title+" ") {
			name = strings.TrimSpace(name[len(title):])
		}
		// Remove from end
		if strings.HasSuffix(strings.ToLower(name), " "+title) {
			name = strings.TrimSpace(name[:len(name)-len(title)])
		}
	}

	// Remove special characters but keep spaces and hyphens
	var cleaned strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsSpace(r) || r == '-' || r == '\'' {
			cleaned.WriteRune(r)
		}
	}

	return strings.TrimSpace(cleaned.String())
}

// GetStats returns statistics about the name database
func (db *NameDB) GetStats() map[string]int {
	return map[string]int{
		"first_names":     len(db.firstNames),
		"last_names":      len(db.lastNames),
		"common_patterns": len(db.commonPatterns),
		"prefixes":        len(db.prefixes),
		"suffixes":        len(db.suffixes),
		"total":           len(db.firstNames) + len(db.lastNames) + len(db.commonPatterns) + len(db.prefixes) + len(db.suffixes),
	}
}
