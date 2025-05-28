package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"./names" // Import our names package
	"github.com/PuerkitoBio/goquery"
)

// Job represents a job posting
type Job struct {
	Title               string     `json:"title"`
	Company             string     `json:"company"`
	Location            string     `json:"location"`
	JobURL              string     `json:"job_url"`
	CompanyURL          string     `json:"company_url"`
	HasIndonesian       bool       `json:"has_indonesian"`
	IndonesianEmployees []Employee `json:"indonesian_employees"`
	EmployeeCount       int        `json:"employee_count"`
	CheckDuration       string     `json:"check_duration"`
}

// Employee represents an Indonesian employee found
type Employee struct {
	Name         string   `json:"name"`
	Position     string   `json:"position,omitempty"`
	MatchReasons []string `json:"match_reasons"`
	Confidence   float64  `json:"confidence"`
}

// LinkedInScraper handles the scraping logic with efficient name detection
type LinkedInScraper struct {
	client *http.Client
	delay  time.Duration
	nameDB *names.NameDB
}

// NewLinkedInScraper creates a new scraper instance
func NewLinkedInScraper() (*LinkedInScraper, error) {
	// Initialize the Indonesian names database
	nameDB, err := names.NewNameDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize names database: %v", err)
	}

	// Print database statistics
	stats := nameDB.GetStats()
	log.Printf("Loaded Indonesian names database: %+v", stats)

	return &LinkedInScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		delay:  2 * time.Second,
		nameDB: nameDB,
	}, nil
}

// makeRequest makes an HTTP request with proper headers and user agent rotation
func (s *LinkedInScraper) makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Rotate user agents for better stealth
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
	}

	ua := userAgents[time.Now().UnixNano()%int64(len(userAgents))]
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.7,id;q=0.3")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")

	return s.client.Do(req)
}

// SearchJobs searches for jobs in LinkedIn with improved parsing
func (s *LinkedInScraper) SearchJobs(country, jobTitle string, limit int) ([]Job, error) {
	baseURL := "https://www.linkedin.com/jobs/search"

	var allJobs []Job
	start := 0

	for len(allJobs) < limit {
		params := url.Values{}
		params.Add("keywords", jobTitle)
		params.Add("location", country)
		params.Add("f_TPR", "r604800") // Jobs posted in last week
		params.Add("start", fmt.Sprintf("%d", start))
		params.Add("sortBy", "R") // Most recent first

		searchURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
		log.Printf("Searching jobs (page %d): %s", start/25+1, searchURL)

		resp, err := s.makeRequest(searchURL)
		if err != nil {
			return nil, fmt.Errorf("failed to make search request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("search request failed with status: %d", resp.StatusCode)
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML: %v", err)
		}

		pageJobs := s.extractJobsFromDocument(doc)
		if len(pageJobs) == 0 {
			log.Printf("No more jobs found on page %d", start/25+1)
			break
		}

		allJobs = append(allJobs, pageJobs...)

		if len(allJobs) >= limit {
			allJobs = allJobs[:limit]
			break
		}

		start += 25
		time.Sleep(s.delay)
	}

	log.Printf("Found %d jobs total", len(allJobs))
	return allJobs, nil
}

// extractJobsFromDocument extracts job listings from HTML document
func (s *LinkedInScraper) extractJobsFromDocument(doc *goquery.Document) []Job {
	var jobs []Job

	// Try multiple selectors for job cards
	selectors := []string{
		".job-search-card",
		".jobs-search__results-list li",
		"[data-entity-urn*='jobPosting']",
		".job-result-card",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
			job := Job{}

			// Extract job title and URL with multiple fallbacks
			titleSelectors := []string{
				".base-search-card__title a",
				".job-result-card__title a",
				"h3 a",
				"[data-control-name='job_search_job_result_title']",
			}

			for _, titleSel := range titleSelectors {
				titleLink := sel.Find(titleSel).First()
				if titleLink.Length() > 0 {
					job.Title = strings.TrimSpace(titleLink.Text())
					if href, exists := titleLink.Attr("href"); exists {
						job.JobURL = s.normalizeURL(href)
					}
					break
				}
			}

			// Extract company name and URL
			companySelectors := []string{
				".hidden-nested-link",
				".job-result-card__subtitle a",
				"h4 a",
				"[data-control-name='job_search_company_name']",
			}

			for _, companySel := range companySelectors {
				companyLink := sel.Find(companySel).First()
				if companyLink.Length() > 0 {
					job.Company = strings.TrimSpace(companyLink.Text())
					if href, exists := companyLink.Attr("href"); exists {
						job.CompanyURL = s.normalizeURL(href)
					}
					break
				}
			}

			// Extract location
			locationSelectors := []string{
				".job-search-card__location",
				".job-result-card__location",
				"[data-test='job-location']",
			}

			for _, locSel := range locationSelectors {
				location := sel.Find(locSel).First()
				if location.Length() > 0 {
					job.Location = strings.TrimSpace(location.Text())
					break
				}
			}

			if job.Title != "" && job.Company != "" {
				jobs = append(jobs, job)
			}
		})

		if len(jobs) > 0 {
			break // Found jobs with this selector, no need to try others
		}
	}

	return jobs
}

// normalizeURL normalizes LinkedIn URLs
func (s *LinkedInScraper) normalizeURL(href string) string {
	if strings.HasPrefix(href, "http") {
		return href
	}
	if strings.HasPrefix(href, "/") {
		return "https://www.linkedin.com" + href
	}
	return href
}

// CheckIndonesianEmployees efficiently checks if a company has Indonesian employees
func (s *LinkedInScraper) CheckIndonesianEmployees(companyURL string) (bool, []Employee, error) {
	startTime := time.Now()

	if companyURL == "" {
		return false, nil, fmt.Errorf("empty company URL")
	}

	time.Sleep(s.delay)

	// Try multiple approaches for finding employees
	approaches := []func(string) ([]Employee, error){
		s.checkCompanyPeoplePage,
		s.checkCompanyAboutPage,
		s.searchEmployeesDirectly,
	}

	var allEmployees []Employee

	for i, approach := range approaches {
		log.Printf("Trying approach %d for employee detection", i+1)
		employees, err := approach(companyURL)
		if err != nil {
			log.Printf("Approach %d failed: %v", i+1, err)
			continue
		}

		allEmployees = append(allEmployees, employees...)

		if len(allEmployees) > 0 {
			break // Found employees, no need to try other approaches
		}
	}

	// Remove duplicates and sort by confidence
	allEmployees = s.deduplicateEmployees(allEmployees)

	duration := time.Since(startTime)
	hasIndonesian := len(allEmployees) > 0

	log.Printf("Employee check completed in %v, found %d Indonesian employees", duration, len(allEmployees))

	return hasIndonesian, allEmployees, nil
}

// checkCompanyPeoplePage checks the company's people page
func (s *LinkedInScraper) checkCompanyPeoplePage(companyURL string) ([]Employee, error) {
	peopleURL := strings.Replace(companyURL, "/company/", "/company/", 1) + "/people/"

	resp, err := s.makeRequest(peopleURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("people page returned status %d", resp.StatusCode)
	}

	return s.extractEmployeesFromHTML(resp.Body)
}

// checkCompanyAboutPage checks the company's about page
func (s *LinkedInScraper) checkCompanyAboutPage(companyURL string) ([]Employee, error) {
	aboutURL := strings.Replace(companyURL, "/company/", "/company/", 1) + "/about/"

	resp, err := s.makeRequest(aboutURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("about page returned status %d", resp.StatusCode)
	}

	return s.extractEmployeesFromHTML(resp.Body)
}

// searchEmployeesDirectly searches for employees using LinkedIn search
func (s *LinkedInScraper) searchEmployeesDirectly(companyURL string) ([]Employee, error) {
	companyName := s.extractCompanyName(companyURL)
	if companyName == "" {
		return nil, fmt.Errorf("could not extract company name")
	}

	searchURL := fmt.Sprintf("https://www.linkedin.com/search/results/people/?currentCompany=%%5B%%22%s%%22%%5D", url.QueryEscape(companyName))

	resp, err := s.makeRequest(searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return s.extractEmployeesFromHTML(resp.Body)
}

// extractEmployeesFromHTML extracts Indonesian employees from HTML content using efficient name lookup
func (s *LinkedInScraper) extractEmployeesFromHTML(body io.Reader) ([]Employee, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	var employees []Employee
	processedNames := make(map[string]bool)

	// Multiple selectors for finding employee names
	nameSelectors := []string{
		".org-people-profile-card__profile-title",
		".profile-card__title",
		".member-name",
		".name",
		"[data-anonymize='person-name']",
		".entity-result__title-text a span[aria-hidden='true']",
		".app-aware-link span[aria-hidden='true']",
		".actor-name",
		".search-result__result-link h3 span[aria-hidden='true']",
	}

	for _, selector := range nameSelectors {
		doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
			name := strings.TrimSpace(sel.Text())
			if name == "" || processedNames[name] {
				return
			}

			// Use efficient name lookup
			isIndonesian, matchReasons := s.nameDB.IsIndonesianName(name)
			if isIndonesian {
				// Calculate confidence based on match reasons
				confidence := s.calculateConfidence(matchReasons)

				// Try to extract position/title
				position := s.extractPosition(sel)

				employee := Employee{
					Name:         name,
					Position:     position,
					MatchReasons: matchReasons,
					Confidence:   confidence,
				}

				employees = append(employees, employee)
				processedNames[name] = true
			}
		})
	}

	// Also check page content with regex for names in text
	htmlContent := doc.Text()
	regexEmployees := s.findNamesWithRegex(htmlContent, processedNames)
	employees = append(employees, regexEmployees...)

	return employees, nil
}

// findNamesWithRegex uses regex to find Indonesian names in text content
func (s *LinkedInScraper) findNamesWithRegex(content string, processedNames map[string]bool) []Employee {
	var employees []Employee

	// Pattern for finding names (2-4 words, capitalized)
	namePattern := regexp.MustCompile(`\b[A-Z][a-z]+(?:\s+[A-Z][a-z]+){1,3}\b`)
	matches := namePattern.FindAllString(content, -1)

	for _, match := range matches {
		if processedNames[match] {
			continue
		}

		isIndonesian, matchReasons := s.nameDB.IsIndonesianName(match)
		if isIndonesian {
			confidence := s.calculateConfidence(matchReasons)

			employee := Employee{
				Name:         match,
				MatchReasons: matchReasons,
				Confidence:   confidence,
			}

			employees = append(employees, employee)
			processedNames[match] = true
		}
	}

	return employees
}

// extractPosition tries to extract job position/title for an employee
func (s *LinkedInScraper) extractPosition(sel *goquery.Selection) string {
	// Look for position in nearby elements
	positionSelectors := []string{
		".org-people-profile-card__profile-info .profile-card__subtitle",
		".profile-card__subtitle",
		".member-title",
		".position",
	}

	parent := sel.Parent()
	for i := 0; i < 3; i++ { // Check up to 3 levels up
		for _, posSel := range positionSelectors {
			position := parent.Find(posSel).First().Text()
			if position != "" {
				return strings.TrimSpace(position)
			}
		}
		parent = parent.Parent()
		if parent.Length() == 0 {
			break
		}
	}

	return ""
}

// calculateConfidence calculates confidence score based on match reasons
func (s *LinkedInScraper) calculateConfidence(matchReasons []string) float64 {
	if len(matchReasons) == 0 {
		return 0.0
	}

	score := 0.0
	for _, reason := range matchReasons {
		switch {
		case strings.Contains(reason, "first_name"):
			score += 0.4
		case strings.Contains(reason, "last_name"):
			score += 0.4
		case strings.Contains(reason, "pattern"):
			score += 0.3
		case strings.Contains(reason, "affix"):
			score += 0.2
		case strings.Contains(reason, "indonesian_pattern"):
			score += 0.1
		}
	}

	// Normalize to 0-1 range
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// deduplicateEmployees removes duplicate employees and sorts by confidence
func (s *LinkedInScraper) deduplicateEmployees(employees []Employee) []Employee {
	seen := make(map[string]*Employee)

	for _, emp := range employees {
		key := strings.ToLower(emp.Name)
		if existing, found := seen[key]; found {
			// Keep the one with higher confidence
			if emp.Confidence > existing.Confidence {
				seen[key] = &emp
			}
		} else {
			seen[key] = &emp
		}
	}

	var result []Employee
	for _, emp := range seen {
		result = append(result, *emp)
	}

	// Sort by confidence (highest first)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Confidence < result[j].Confidence {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// extractCompanyName extracts company name from LinkedIn URL
func (s *LinkedInScraper) extractCompanyName(companyURL string) string {
	parts := strings.Split(companyURL, "/")
	for i, part := range parts {
		if part == "company" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// ProcessJobs processes all jobs and checks for Indonesian employees with progress tracking
func (s *LinkedInScraper) ProcessJobs(jobs []Job) []Job {
	var processedJobs []Job
	totalJobs := len(jobs)

	log.Printf("Processing %d jobs for Indonesian employee detection...", totalJobs)

	for i, job := range jobs {
		log.Printf("[%d/%d] Processing: %s at %s", i+1, totalJobs, job.Title, job.Company)

		startTime := time.Now()
		hasIndonesian, employees, err := s.CheckIndonesianEmployees(job.CompanyURL)
		duration := time.Since(startTime)

		job.CheckDuration = duration.String()
		job.EmployeeCount = len(employees)

		if err != nil {
			log.Printf("❌ Error checking employees for %s: %v", job.Company, err)
			job.HasIndonesian = false
			job.IndonesianEmployees = []Employee{}
		} else {
			job.HasIndonesian = hasIndonesian
			job.IndonesianEmployees = employees

			if hasIndonesian {
				log.Printf("✅ Found %d Indonesian employees at %s", len(employees), job.Company)
			} else {
				log.Printf("➖ No Indonesian employees found at %s", job.Company)
			}
		}

		processedJobs = append(processedJobs, job)

		// Progress indicator
		progress := float64(i+1) / float64(totalJobs) * 100
		log.Printf("Progress: %.1f%% (%d/%d)", progress, i+1, totalJobs)

		time.Sleep(s.delay)
	}

	return processedJobs
}

// SaveResults saves the results to a JSON file with better formatting
func SaveResults(jobs []Job, filename string) error {
	// Create a summary
	summary := map[string]interface{}{
		"total_jobs":                 len(jobs),
		"jobs_with_indonesians":      0,
		"total_indonesian_employees": 0,
		"generated_at":               time.Now().Format("2006-01-02 15:04:05"),
	}

	for _, job := range jobs {
		if job.HasIndonesian {
			summary["jobs_with_indonesians"] = summary["jobs_with_indonesians"].(int) + 1
			summary["total_indonesian_employees"] = summary["total_indonesian_employees"].(int) + len(job.IndonesianEmployees)
		}
	}

	result := map[string]interface{}{
		"summary": summary,
		"jobs":    jobs,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// printResults prints the results in a formatted way with enhanced output
func printResults(jobs []Job) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("🇮🇩 LINKEDIN JOB SEARCH RESULTS WITH INDONESIAN EMPLOYEE DETECTION")
	fmt.Println(strings.Repeat("=", 100))

	totalJobs := len(jobs)
	jobsWithIndonesians := 0
	totalIndonesianEmployees := 0

	for _, job := range jobs {
		if job.HasIndonesian {
			jobsWithIndonesians++
			totalIndonesianEmployees += len(job.IndonesianEmployees)
		}
	}

	fmt.Printf("📊 SUMMARY:\n")
	fmt.Printf("   Total Jobs Found: %d\n", totalJobs)
	fmt.Printf("   Jobs with Indonesian Employees: %d (%.1f%%)\n", jobsWithIndonesians, float64(jobsWithIndonesians)/float64(totalJobs)*100)
	fmt.Printf("   Total Indonesian Employees Found: %d\n", totalIndonesianEmployees)
	fmt.Println(strings.Repeat("-", 100))

	// Sort jobs: those with Indonesian employees first
	sortedJobs := make([]Job, len(jobs))
	copy(sortedJobs, jobs)

	for i := 0; i < len(sortedJobs)-1; i++ {
		for j := i + 1; j < len(sortedJobs); j++ {
			if !sortedJobs[i].HasIndonesian && sortedJobs[j].HasIndonesian {
				sortedJobs[i], sortedJobs[j] = sortedJobs[j], sortedJobs[i]
			}
		}
	}

	for i, job := range sortedJobs {
		fmt.Printf("\n%d. %s\n", i+1, job.Title)
		fmt.Printf("   🏢 Company: %s\n", job.Company)
		fmt.Printf("   📍 Location: %s\n", job.Location)
		fmt.Printf("   🔗 Job URL: %s\n", job.JobURL)
		fmt.Printf("   🇮🇩 Indonesian Employees: %v (%d found)\n", job.HasIndonesian, len(job.IndonesianEmployees))

		if len(job.IndonesianEmployees) > 0 {
			fmt.Printf("   👥 Indonesian Staff:\n")
			for _, emp := range job.IndonesianEmployees {
				confidenceStr := fmt.Sprintf("%.0f%%", emp.Confidence*100)
				fmt.Printf("      • %s", emp.Name)
				if emp.Position != "" {
					fmt.Printf(" (%s)", emp.Position)
				}
				fmt.Printf(" [Confidence: %s]\n", confidenceStr)
				if len(emp.MatchReasons) > 0 {
					fmt.Printf("        Reasons: %s\n", strings.Join(emp.MatchReasons, ", "))
				}
			}
		}

		if job.HasIndonesian {
			fmt.Printf("   ⭐ HIGHLY RECOMMENDED: This company has Indonesian employees!\n")
		}

		fmt.Printf("   ⏱️  Check Duration: %s\n", job.CheckDuration)
		fmt.Println(strings.Repeat("-", 80))
	}

	// Print top recommendations
	if jobsWithIndonesians > 0 {
		fmt.Printf("\n🎯 TOP RECOMMENDATIONS (Companies with Indonesian Employees):\n")
		fmt.Println(strings.Repeat("-", 60))
		rank := 1
		for _, job := range sortedJobs {
			if job.HasIndonesian {
				fmt.Printf("%d. %s at %s (%d Indonesian employees)\n", rank, job.Title, job.Company, len(job.IndonesianEmployees))
				rank++
			}
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <country> <job_title> [limit]")
		fmt.Println("Example: go run main.go \"Germany\" \"golang developer\" 50")
		fmt.Println("Default limit: 25 jobs")
		os.Exit(1)
	}

	country := os.Args[1]
	jobTitle := os.Args[2]
	limit := 25

	if len(os.Args) > 3 {
		fmt.Sscanf(os.Args[3], "%d", &limit)
	}

	fmt.Printf("🔍 Searching for '%s' jobs in %s (limit: %d)...\n", jobTitle, country, limit)
	fmt.Println("🇮🇩 Will efficiently check for Indonesian employees at each company")

	scraper, err := NewLinkedInScraper()
	if err != nil {
		log.Fatalf("Failed to initialize scraper: %v", err)
	}

	// Search for jobs
	jobs, err := scraper.SearchJobs(country, jobTitle, limit)
	if err != nil {
		log.Fatalf("Failed to search jobs: %v", err)
	}

	if len(jobs) == 0 {
		fmt.Println("❌ No jobs found. Try different search terms or check if LinkedIn is accessible.")
		return
	}

	fmt.Printf("✅ Found %d jobs, now checking for Indonesian employees using efficient lookup...\n", len(jobs))

	// Process jobs to check for Indonesian employees
	processedJobs := scraper.ProcessJobs(jobs)

	// Print results
	printResults(processedJobs)

	// Save to JSON file
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("linkedin_jobs_%s_%s_%s.json",
		strings.ReplaceAll(strings.ToLower(country), " ", "_"),
		strings.ReplaceAll(strings.ToLower(jobTitle), " ", "_"),
		timestamp)

	if err := SaveResults(processedJobs, filename); err != nil {
		log.Printf("❌ Failed to save results: %v", err)
	} else {
		fmt.Printf("\n💾 Results saved to: %s\n", filename)
	}

	fmt.Println("\n✅ Job search completed!")
	fmt.Println("📈 Use the JSON file for further analysis or integration with other tools.")
}
