package main

import (
	"fmt"
	"os"
	"strings"
)

// generateFirstNames creates a comprehensive list of Indonesian first names
func generateFirstNames() []string {
	// Male names
	maleNames := []string{
		// Arabic/Islamic influenced names (very common)
		"Abdul", "Abdulah", "Abdurrahman", "Abdu", "Abidin", "Abrar", "Achmat", "Ahmad", "Ahmed", "Ahmat",
		"Aidil", "Ainun", "Akbar", "Akhmad", "Ali", "Aman", "Amran", "Anwar", "Arief", "Arif", "Asep",
		"Baharuddin", "Chairul", "Darmawan", "Endang", "Fachri", "Fadil", "Fajar", "Fandi", "Farel",
		"Faris", "Fauzan", "Febri", "Ferdian", "Firman", "Hafiz", "Haikal", "Hakim", "Hamdan", "Haris",
		"Hasan", "Hilman", "Ikhsan", "Ilham", "Imam", "Irfan", "Kamil", "Kemal", "Kurnia", "Maulana",
		"Mohammad", "Muhammad", "Muhamad", "Nurdin", "Omar", "Rasyid", "Rayhan", "Reza", "Rezki", "Rheza",
		"Ridho", "Rizal", "Rizki", "Rizky", "Taufik", "Umar", "Wahyu", "Yusuf", "Zainal",

		// Javanese names
		"Adi", "Aditya", "Agung", "Agus", "Andi", "Andre", "Andrew", "Andy", "Angga", "Anggoro",
		"Anton", "Anto", "Ari", "Ario", "Arman", "Arya", "Bagus", "Bambang", "Banu", "Bayu",
		"Beni", "Billy", "Bimo", "Bintang", "Boby", "Budi", "Budianto", "Caesar", "Cahya", "Cahyo",
		"Christian", "Christopher", "Daffa", "Dani", "Daniel", "Danny", "Danang", "David", "Dedi",
		"Denny", "Deny", "Dharma", "Dian", "Dicky", "Dimas", "Dino", "Dirgantara", "Dwi", "Edi",
		"Edy", "Eka", "Eko", "Erik", "Erwin", "Fabian", "Fernando", "Ferry", "Frans", "Galih",
		"Giovanni", "Harry", "Hendri", "Hendra", "Hendy", "Herman", "Hero", "Heru", "Indra", "Ivan",
		"Jaka", "Joko", "Jonathan", "Joshua", "Julian", "Julianto", "Kevin", "Krisna", "Leonardo",
		"Lucky", "Lukman", "Michael", "Nanda", "Nathan", "Nico", "Oscar", "Patrick", "Paul", "Putra",
		"Rama", "Rangga", "Raihan", "Randy", "Rio", "Robert", "Ryan", "Sandi", "Santo", "Satria",
		"Sebastian", "Sigit", "Surya", "Teguh", "Tedy", "Tri", "William", "Wisnu", "Yoga",

		// Balinese names
		"Gede", "Kadek", "Ketut", "Komang", "Made", "Nyoman", "Putu", "Wayan", "Dewa",

		// Sundanese names
		"Asep", "Dede", "Ujang", "Yayat", "Cecep", "Usep",

		// Batak names
		"Hotma", "Juntak", "Parlindungan", "Tigor", "Togi", "Binsar", "Mangatas",

		// Minang names
		"Sutan", "Tengku", "Datuak", "Bagindo",

		// Modern/Western influenced
		"Albert", "Alex", "Alexander", "Alvin", "Brian", "Calvin", "Dennis", "Felix", "George",
		"Henry", "Jack", "James", "John", "Kevin", "Louis", "Marco", "Nicholas", "Oliver",
		"Peter", "Richard", "Samuel", "Steven", "Thomas", "Victor", "Wilson",
	}

	// Female names
	femaleNames := []string{
		// Arabic/Islamic influenced names
		"Aisyah", "Aida", "Alika", "Amelia", "Ana", "Andini", "Anisa", "Annisa", "Aprilia",
		"Fatimah", "Fadila", "Farah", "Laila", "Nadia", "Nadya", "Nayla", "Nurul", "Siti",
		"Zahra", "Zara", "Sari", "Kartika", "Ratna", "Dewi", "Putri", "Rani", "Suci",

		// Javanese names
		"Arum", "Astrid", "Ayu", "Bella", "Bunga", "Cantika", "Carla", "Citra", "Cynthia",
		"Dela", "Desi", "Diana", "Dina", "Dinda", "Dini", "Dita", "Diva", "Dwi", "Eka",
		"Ela", "Elsa", "Ema", "Evi", "Galuh", "Gita", "Hana", "Indah", "Inez", "Ira",
		"Irma", "Jessica", "Julia", "Karina", "Linda", "Lisa", "Lita", "Luna", "Maya",
		"Mega", "Mila", "Nanda", "Nia", "Nila", "Nina", "Nisa", "Novi", "Octavia",
		"Prita", "Ria", "Rika", "Rini", "Rita", "Rosa", "Shinta", "Sinta", "Siska",
		"Sri", "Tari", "Tika", "Tina", "Titik", "Umi", "Vera", "Vina", "Wati",
		"Wulan", "Yanti", "Yeni", "Yulia",

		// Balinese names
		"Kadek", "Ketut", "Komang", "Made", "Nyoman", "Putu", "Wayan",

		// Modern/Western influenced
		"Amanda", "Angela", "Anna", "Bella", "Christina", "Diana", "Emma", "Grace",
		"Hannah", "Isabella", "Jennifer", "Katherine", "Laura", "Michelle", "Natasha",
		"Olivia", "Patricia", "Rachel", "Stephanie", "Victoria",
	}

	var allNames []string
	allNames = append(allNames, maleNames...)
	allNames = append(allNames, femaleNames...)

	return allNames
}

// generateLastNames creates a comprehensive list of Indonesian last names
func generateLastNames() []string {
	return []string{
		// Common Indonesian surnames
		"Adiputra", "Adisaputra", "Adriansyah", "Agustina", "Ahmadi", "Aisyah", "Akbar", "Alamsyah",
		"Amelia", "Andriani", "Anggraeni", "Anwar", "Aprilia", "Ardiansyah", "Ariyanto", "Ashari",
		"Aulia", "Azzahra", "Baharuddin", "Budiman", "Budiarto", "Chandra", "Darmawan", "Dharma",
		"Dwiyanto", "Fahreza", "Firdaus", "Gunawan", "Hakim", "Halim", "Handoko", "Harahap",
		"Haryanto", "Hasanah", "Hidayat", "Hidayanto", "Hutapea", "Ibrahim", "Indrawati", "Iskandar",
		"Jaya", "Kartika", "Kencana", "Kurniawan", "Kusuma", "Laksana", "Maharani", "Mahendra",
		"Maulana", "Mulyadi", "Nasution", "Nugroho", "Nurhasanah", "Oktaviani", "Prabowo", "Pratama",
		"Purnama", "Putra", "Putri", "Rahayu", "Rahman", "Rahmawati", "Ramadhan", "Santoso",
		"Saputra", "Setiawan", "Siregar", "Situmorang", "Subekti", "Suharto", "Suryadi", "Susanto",
		"Utama", "Utomo", "Wardani", "Wibowo", "Widodo", "Wijaya", "Wulandari", "Yanto", "Yulianti",

		// Batak surnames
		"Nababan", "Panggabean", "Pardede", "Purba", "Siagian", "Siahaan", "Simanjuntak", "Sinaga",
		"Situmeang", "Tambunan", "Hutagalung", "Manullang", "Pasaribu", "Sagala", "Samosir",
		"Tampubolon", "Turnip", "Bangun", "Damanik", "Ginting", "Karo", "Perangin", "Sembiring",
		"Tarigan", "Brahmana", "Keliat", "Milala", "Munte", "Pinem", "Sebayang", "Sinulingga",

		// Javanese surnames
		"Suharto", "Sukarno", "Widodo", "Susilo", "Bambang", "Yudhoyono", "Megawati", "Soekarnoputri",
		"Wahid", "Habibie", "Abdurrahman", "Gus", "Dur", "Jokowi", "Prabowo", "Subianto",

		// Sundanese surnames
		"Supratman", "Surapati", "Sudirman", "Siliwangi", "Hasanudin", "Kartawisastra", "Suryakencana",

		// Minang surnames
		"Chatib", "Datuk", "Gelar", "Indra", "Maharajo", "Pangulu", "Sutan", "Tuanku",

		// Chinese Indonesian surnames
		"Tanoto", "Riady", "Lippo", "Salim", "Eka", "Tjipta", "Widjaja", "Djojohadikusumo",
		"Hartono", "Wanandi", "Murdaya", "Tahir", "Panigoro", "Gondokusumo", "Soeryadjaya",

		// Modern compound surnames
		"Wibisono", "Atmadja", "Kusumawardani", "Prasetyo", "Setiabudhi", "Mangunkusumo",
		"Kartawiria", "Kusumawardhana", "Purbonegoro", "Kusumo", "Wardana", "Prawira",
		"Kusumadewa", "Permana", "Hermawan", "Lestari", "Fitriani", "Kusnadi", "Wicaksono",
		"Retnowati", "Widyastuti", "Handayani", "Sulistyowati", "Purnamasari", "Kurniasari",
	}
}

// generateCommonPatterns creates patterns commonly found in Indonesian names
func generateCommonPatterns() []string {
	return []string{
		// Religious/Arabic patterns
		"Abdul", "Abdur", "Ainul", "Amal", "Amin", "Bahar", "Darul", "Fajar", "Hadir", "Kasih",
		"Mulia", "Nur", "Rahmat", "Sakti", "Suci", "Tulus",

		// Balinese patterns
		"Kadek", "Ketut", "Komang", "Made", "Nyoman", "Putu", "Wayan", "Gede", "Luh",

		// Javanese patterns
		"Bambang", "Budi", "Dwi", "Eko", "Endang", "Heru", "Joko", "Rini", "Sari", "Sri",
		"Sukamto", "Tri", "Yani", "Yono", "Yudi", "Yuli", "Yuni", "Yono",

		// Sundanese patterns
		"Asep", "Dedi", "Iin", "Neng", "Teteh", "Aa", "Akang", "Cing", "Euis",

		// Batak patterns
		"Hotma", "Juntak", "Nababan", "Panggabean", "Pardede", "Purba", "Siagian", "Siahaan",
		"Simanjuntak", "Sinaga", "Siregar", "Situmeang", "Situmorang", "Tambunan",

		// Minang patterns
		"Sutan", "Tengku", "Datuak", "Bagindo", "Chatib", "Gelar", "Indra", "Sari",

		// Betawi patterns
		"Mpok", "Nyai", "Bang", "Uda", "Encik",

		// Generic patterns
		"Bening", "Cahya", "Fitri", "Indah", "Jati", "Lestari", "Putih", "Utami", "Wening",
		"Agung", "Bayu", "Dharma", "Eka", "Fajar", "Galih", "Hadi", "Indra", "Jaya", "Krisna",
		"Lingga", "Mahendra", "Nanda", "Okta", "Panca", "Rama", "Satya", "Tirta", "Vira", "Wisnu",
	}
}

// generatePrefixes creates Indonesian name prefixes
func generatePrefixes() []string {
	return []string{
		"abdul", "abdur", "abdu", "bin", "binti", "darul", "fathul", "khairu", "khairul",
		"nuru", "nurul", "rahmatu", "saiful", "siti", "nyai", "haji", "hajjah",
	}
}

// generateSuffixes creates Indonesian name suffixes
func generateSuffixes() []string {
	return []string{
		"din", "man", "nto", "wan", "wati", "yah", "yanto", "anto", "adi", "ati", "ani",
		"ina", "sari", "wulan", "dewi", "putri", "ningrum", "ningrat", "ayu", "utami",
	}
}

// writeNamesToFile writes names to a file
func writeNamesToFile(filename string, names []string, header string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("# %s\n", header))
	for _, name := range names {
		file.WriteString(name + "\n")
	}

	return nil
}

// generateExpandedNames creates variations and combinations
func generateExpandedNames(baseNames []string) []string {
	var expanded []string
	expanded = append(expanded, baseNames...)

	// Add common variations
	variations := map[string][]string{
		"Muhammad": {"Muhamad", "Mohamed", "Mohammed"},
		"Ahmad":    {"Ahmed", "Achmad", "Achmat"},
		"Yusuf":    {"Yosef", "Yoseph", "Jusuf"},
		"Ibrahim":  {"Ibraheem", "Abrahim"},
		"Iskandar": {"Iskander", "Alexander"},
		"Umar":     {"Omar"},
		"Ali":      {"Aly"},
		"Hasan":    {"Hassan"},
		"Husein":   {"Hussein", "Husain"},
		"Abdulah":  {"Abdullah", "Abdulloh"},
		"Rahman":   {"Rachman", "Rohman"},
		"Rahim":    {"Rachim", "Rohim"},
		"Ridho":    {"Ridho", "Rido", "Rizho"},
		"Rizki":    {"Rizky", "Risky", "Rizqi"},
		"Fajar":    {"Fajr", "Fadjr"},
		"Bayu":     {"Baio"},
		"Galih":    {"Galeh"},
		"Satria":   {"Satrya"},
		"Aditya":   {"Adytya"},
		"Angga":    {"Anga"},
		"Yoga":     {"Yuga"},
		"Wisnu":    {"Vishnu"},
		"Krisna":   {"Krishna"},
		"Rama":     {"Ramadhani", "Ramadhan"},
		"Dewi":     {"Devi"},
		"Sari":     {"Saree"},
		"Ratna":    {"Ratni"},
		"Indah":    {"Indha"},
		"Wulan":    {"Bulan"},
	}

	for base, vars := range variations {
		expanded = append(expanded, vars...)
	}

	return expanded
}

// generateCompoundNames creates compound Indonesian names
func generateCompoundNames() []string {
	prefixes := []string{"Adi", "Dwi", "Tri", "Catur", "Panca", "Eka", "Dua", "Tiga"}
	suffixes := []string{"putra", "putri", "wati", "yanto", "sari", "dewi", "ningrum", "utami"}

	var compounds []string

	for _, prefix := range prefixes {
		for _, suffix := range suffixes {
			compounds = append(compounds, prefix+suffix)
		}
	}

	// Add specific compound patterns
	compounds = append(compounds, []string{
		"Adiputra", "Adiputri", "Adiwati", "Adiyanto",
		"Dwiputra", "Dwiputri", "Dwiwati", "Dwiyanto",
		"Triputra", "Triputri", "Triwati", "Triyanto",
		"Ekaputra", "Ekaputri", "Ekawati", "Ekayanto",
		"Pancaputra", "Pancaputri", "Pancawati", "Pancayanto",
	}...)

	return compounds
}

// createDirectoryStructure creates the data directory and files
func createDirectoryStructure() error {
	// Create data directory
	err := os.MkdirAll("data", 0755)
	if err != nil {
		return err
	}

	fmt.Println("📁 Created data directory structure")

	// Generate and write first names (expanded)
	firstNames := generateFirstNames()
	expandedFirstNames := generateExpandedNames(firstNames)
	compounds := generateCompoundNames()
	allFirstNames := append(expandedFirstNames, compounds...)

	err = writeNamesToFile("data/first_names.txt", allFirstNames, "Indonesian First Names Database - "+fmt.Sprintf("%d names", len(allFirstNames)))
	if err != nil {
		return err
	}
	fmt.Printf("✅ Generated first_names.txt with %d names\n", len(allFirstNames))

	// Generate and write last names
	lastNames := generateLastNames()
	expandedLastNames := generateExpandedNames(lastNames)

	err = writeNamesToFile("data/last_names.txt", expandedLastNames, "Indonesian Last Names Database - "+fmt.Sprintf("%d names", len(expandedLastNames)))
	if err != nil {
		return err
	}
	fmt.Printf("✅ Generated last_names.txt with %d names\n", len(expandedLastNames))

	// Generate and write common patterns
	patterns := generateCommonPatterns()
	err = writeNamesToFile("data/common_patterns.txt", patterns, "Indonesian Common Name Patterns - "+fmt.Sprintf("%d patterns", len(patterns)))
	if err != nil {
		return err
	}
	fmt.Printf("✅ Generated common_patterns.txt with %d patterns\n", len(patterns))

	// Generate and write prefixes
	prefixes := generatePrefixes()
	err = writeNamesToFile("data/prefixes.txt", prefixes, "Indonesian Name Prefixes - "+fmt.Sprintf("%d prefixes", len(prefixes)))
	if err != nil {
		return err
	}
	fmt.Printf("✅ Generated prefixes.txt with %d prefixes\n", len(prefixes))

	// Generate and write suffixes
	suffixes := generateSuffixes()
	err = writeNamesToFile("data/suffixes.txt", suffixes, "Indonesian Name Suffixes - "+fmt.Sprintf("%d suffixes", len(suffixes)))
	if err != nil {
		return err
	}
	fmt.Printf("✅ Generated suffixes.txt with %d suffixes\n", len(suffixes))

	totalNames := len(allFirstNames) + len(expandedLastNames) + len(patterns) + len(prefixes) + len(suffixes)
	fmt.Printf("\n🎉 Total database size: %d entries\n", totalNames)
	fmt.Println("📊 Database breakdown:")
	fmt.Printf("   - First Names: %d\n", len(allFirstNames))
	fmt.Printf("   - Last Names: %d\n", len(expandedLastNames))
	fmt.Printf("   - Patterns: %d\n", len(patterns))
	fmt.Printf("   - Prefixes: %d\n", len(prefixes))
	fmt.Printf("   - Suffixes: %d\n", len(suffixes))

	return nil
}

// generateTestCases creates test cases for validation
func generateTestCases() []string {
	return []string{
		"Budi Santoso",
		"Siti Nurhaliza",
		"Ahmad Dhani",
		"Dewi Sartika",
		"Joko Widodo",
		"Megawati Soekarnoputri",
		"Susilo Bambang Yudhoyono",
		"Prabowo Subianto",
		"Ridwan Kamil",
		"Anies Baswedan",
		"Tri Rismaharini",
		"Ganjar Pranowo",
		"Khofifah Indar Parawansa",
		"Mahfud MD",
		"Sri Mulyani Indrawati",
		"Luhut Binsar Pandjaitan",
		"Retno Marsudi",
		"Coordinating Minister",
		"Made Pastika",
		"Nyoman Nuarta",
		"Ketut Liyer",
		"Wayan Mirna",
		"Gede Prama",
		"I Putu Gede",
		"Kadek Devi",
		"John Smith",          // Should not match
		"Michael Johnson",     // Should not match
		"Zhang Wei",           // Should not match
		"Hiroshi Tanaka",      // Should not match
		"Abdullah Rahman",     // Should match
		"Sari Dewi Fortuna",   // Should match
		"Rizki Pratama",       // Should match
		"Indira Kencana Sari", // Should match
	}
}

// testNameDetection tests the name detection with sample names
func testNameDetection() {
	fmt.Println("\n🧪 Testing name detection...")
	testCases := generateTestCases()

	// This would require the names package to be initialized
	// For now, just print the test cases
	fmt.Println("Test cases generated:")
	for i, testCase := range testCases {
		fmt.Printf("%d. %s\n", i+1, testCase)
	}
}

func main() {
	fmt.Println("🇮🇩 Indonesian Names Database Generator")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Create the directory structure and generate files
	err := createDirectoryStructure()
	if err != nil {
		fmt.Printf("❌ Error creating directory structure: %v\n", err)
		os.Exit(1)
	}

	// Generate test cases
	testNameDetection()

	fmt.Println("\n✅ Database generation completed!")
	fmt.Println("📁 Files created in ./data/ directory:")
	fmt.Println("   - first_names.txt")
	fmt.Println("   - last_names.txt")
	fmt.Println("   - common_patterns.txt")
	fmt.Println("   - prefixes.txt")
	fmt.Println("   - suffixes.txt")
	fmt.Println("\n🚀 You can now use this database with the LinkedIn scraper!")
	fmt.Println("💡 Run: go run scraper.go \"Germany\" \"golang developer\"")
}
