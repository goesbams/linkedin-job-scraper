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

	"github.com/PuerkitoBio/goquery"
	"github.com/goesbams/linkedin-job-scraper/names"
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

// LinkedInScraper handles the scraping logic with enhanced debugging
type LinkedInScraper struct {
	client *http.Client
	delay  time.Duration
	nameDB *names.NameDB
	debug  bool
}

// NewLinkedInScraper creates a new scraper instance with debug mode
func NewLinkedInScraper() (*LinkedInScraper, error) {
	// Initialize the Indonesian names database
	nameDB, err := names.NewNameDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize names database: %v", err)
	}

	// Print database statistics
	stats := nameDB.GetStats()
	log.Printf("Loaded Indonesian names database: %+v", stats)

	// Check for debug mode
	debug := os.Getenv("DEBUG") == "true" || os.Getenv("SCRAPER_DEBUG") == "true"

	return &LinkedInScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		delay:  3 * time.Second, // Increased delay to be more respectful
		nameDB: nameDB,
		debug:  debug,
	}, nil
}

// debugLog prints debug information if debug mode is enabled
func (s *LinkedInScraper) debugLog(format string, args ...interface{}) {
	if s.debug {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// makeRequest makes an HTTP request with enhanced headers and debugging
func (s *LinkedInScraper) makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Enhanced headers to appear more like a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("DNT", "1")

	s.debugLog("Making request to: %s", url)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	s.debugLog("Response status: %d %s", resp.StatusCode, resp.Status)

	return resp, nil
}

// testLinkedInAccess tests if LinkedIn is accessible and not blocking us
func (s *LinkedInScraper) testLinkedInAccess() error {
	testURL := "https://www.linkedin.com"
	s.debugLog("Testing LinkedIn access...")

	resp, err := s.makeRequest(testURL)
	if err != nil {
		return fmt.Errorf("failed to access LinkedIn: %v", err)
	}
	defer resp.Body.Close()

	s.debugLog("LinkedIn access test: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode == 999 {
		return fmt.Errorf("LinkedIn is blocking requests (status 999) - try using VPN or different IP")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("LinkedIn returned status %d - may be blocking requests", resp.StatusCode)
	}

	return nil
}

// SearchJobs searches for jobs with enhanced debugging and multiple fallback strategies
func (s *LinkedInScraper) SearchJobs(country, jobTitle string, limit int) ([]Job, error) {
	// Test LinkedIn access first
	if err := s.testLinkedInAccess(); err != nil {
		log.Printf("⚠️  Warning: %v", err)
	}

	baseURL := "https://www.linkedin.com/jobs/search"

	var allJobs []Job
	start := 0

	for len(allJobs) < limit {
		// Try multiple parameter combinations for better success rate
		paramSets := []url.Values{
			// Without time restriction (more results) - PRIMARY APPROACH
			{
				"keywords": {jobTitle},
				"location": {country},
				"start":    {fmt.Sprintf("%d", start)},
				"sortBy":   {"R"},
			},
			// Original parameters with time filter
			{
				"keywords": {jobTitle},
				"location": {country},
				"f_TPR":    {"r604800"}, // Last week
				"start":    {fmt.Sprintf("%d", start)},
				"sortBy":   {"R"},
			},
			// Simple parameters
			{
				"keywords": {jobTitle},
				"location": {country},
				"start":    {fmt.Sprintf("%d", start)},
			},
			// Most basic approach
			{
				"keywords": {jobTitle},
				"location": {country},
			},
		}

		var pageJobs []Job

		// Try each parameter set until one works
		for i, params := range paramSets {
			searchURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

			if i == 0 {
				log.Printf("🔍 Searching jobs (page %d): %s", start/25+1, searchURL)
			} else {
				s.debugLog("Trying fallback approach %d: %s", i+1, searchURL)
			}

			resp, err := s.makeRequest(searchURL)
			if err != nil {
				s.debugLog("Request failed for approach %d: %v", i+1, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == 999 {
				return nil, fmt.Errorf("LinkedIn is blocking requests (status 999). Try:\n1. Using a VPN\n2. Waiting and trying later\n3. Using different search terms")
			}

			if resp.StatusCode != 200 {
				s.debugLog("Non-200 response for approach %d: %d", i+1, resp.StatusCode)
				continue
			}

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				s.debugLog("Parse failed for approach %d: %v", i+1, err)
				continue
			}

			// Debug: save page content if in debug mode
			if s.debug {
				html, _ := doc.Html()
				filename := fmt.Sprintf("debug_approach_%d_page_%d.html", i+1, start/25+1)
				os.WriteFile(filename, []byte(html), 0644)
				s.debugLog("Saved debug page to: %s", filename)
			}

			jobs := s.extractJobsFromDocument(doc)
			s.debugLog("Approach %d extracted %d jobs from page %d", i+1, len(jobs), start/25+1)

			if len(jobs) > 0 {
				pageJobs = jobs
				if i > 0 {
					log.Printf("✅ Success with fallback approach %d - found %d jobs", i+1, len(jobs))
				}
				break // Found jobs, use this approach
			}
		}

		if len(pageJobs) == 0 {
			s.debugLog("No jobs found with any approach on page %d", start/25+1)

			// Enhanced blocking detection
			if start == 0 { // Only check on first page
				log.Printf("🔍 Analyzing why no jobs were found...")
				// The debug files will contain the actual page content for analysis
			}

			log.Printf("ℹ️  No more jobs found on page %d", start/25+1)
			break
		}

		allJobs = append(allJobs, pageJobs...)

		if len(allJobs) >= limit {
			allJobs = allJobs[:limit]
			break
		}

		start += 25

		s.debugLog("Waiting %v before next request...", s.delay)
		time.Sleep(s.delay)
	}

	log.Printf("✅ Found %d jobs total", len(allJobs))
	return allJobs, nil
}

// extractJobsFromDocument extracts job listings with current LinkedIn selectors (2025)
func (s *LinkedInScraper) extractJobsFromDocument(doc *goquery.Document) []Job {
	var jobs []Job

	// UPDATED: Current LinkedIn selectors as of 2025
	// Based on LinkedIn's current page structure
	selectors := []string{
		// Primary current selectors (2025)
		".jobs-search-results-list .jobs-search-results__list-item",
		".jobs-search-results__list-item",
		".scaffold-layout__list-item",
		".artdeco-list__item",
		"[data-entity-urn*='jobPosting']",
		"[data-occludable-job-id]",

		// Secondary current selectors
		".job-card-container",
		".job-card-list__entity-lockup",
		".entity-result",
		".search-result",
		"[data-job-id]",

		// Backup selectors for different page layouts
		".base-search-card",
		".job-search-card",
		".jobs-search__results-list li",
		".job-result-card",
		".base-card",
		"li[data-occludable-job-id]",
		".scaffold-layout__list-container li",

		// Generic fallbacks
		"article",
		".card",
		"[data-urn]",
		"li[data-urn]",
		"div[class*='job']",
		"li[class*='job']",
		"[class*='search-result']",
		"li[class*='entity']",
		".entity-lockup",
	}

	s.debugLog("Trying %d different selectors for job extraction", len(selectors))

	for i, selector := range selectors {
		s.debugLog("Trying selector %d: %s", i+1, selector)

		jobElements := doc.Find(selector)
		s.debugLog("Found %d elements with selector: %s", jobElements.Length(), selector)

		if jobElements.Length() > 0 {
			var selectorJobs []Job
			jobElements.Each(func(j int, sel *goquery.Selection) {
				job := s.extractJobFromElement(sel)
				if job.Title != "" && job.Company != "" {
					selectorJobs = append(selectorJobs, job)
					s.debugLog("Extracted job %d: %s at %s", j+1, job.Title, job.Company)
				}
			})

			if len(selectorJobs) > 0 {
				s.debugLog("Successfully extracted %d jobs with selector: %s", len(selectorJobs), selector)
				jobs = append(jobs, selectorJobs...)
				break // Found jobs with this selector
			}
		}
	}

	if len(jobs) == 0 {
		s.debugLog("⚠️  No jobs extracted with any selector")
		if s.debug {
			// Enhanced page structure analysis
			pageStructure := s.analyzePageStructure(doc)
			log.Printf("Page structure analysis:\n%s", pageStructure)
		}
	}

	return jobs
}

// extractJobFromElement extracts job details with current LinkedIn selectors (2025)
func (s *LinkedInScraper) extractJobFromElement(sel *goquery.Selection) Job {
	job := Job{}

	// UPDATED: Current LinkedIn title selectors (2025)
	titleSelectors := []string{
		// Primary current selectors
		".job-card-list__title a",
		".job-card-container__link",
		".artdeco-entity-lockup__title a",
		".entity-result__title-text a",
		".base-search-card__title a",
		"[data-control-name='job_search_job_result_title']",

		// Secondary selectors
		".job-result-card__title a",
		"h3 a",
		"h2 a",
		".job-search-card__title a",
		"a[data-control-name='job_search_job_result_title']",
		".job-card__title a",
		".job-title a",
		".jobs-unified-top-card__job-title a",
		"[aria-label*='job']",

		// Generic fallbacks
		"a[href*='/jobs/view/']",
		"a[href*='/jobs/collections/']",
	}

	for _, titleSel := range titleSelectors {
		titleLink := sel.Find(titleSel).First()
		if titleLink.Length() > 0 {
			title := strings.TrimSpace(titleLink.Text())
			if title != "" && len(title) > 3 { // Basic validation
				job.Title = title
				if href, exists := titleLink.Attr("href"); exists {
					job.JobURL = s.normalizeURL(href)
				}
				s.debugLog("Found title with selector %s: %s", titleSel, job.Title)
				break
			}
		}
	}

	// UPDATED: Current LinkedIn company selectors (2025)
	companySelectors := []string{
		// Primary current selectors
		".job-card-container__primary-description",
		".artdeco-entity-lockup__subtitle a",
		".entity-result__primary-subtitle a",
		".base-search-card__subtitle a",
		"[data-control-name='job_search_company_name']",

		// Secondary selectors
		".hidden-nested-link",
		".job-result-card__subtitle a",
		"h4 a",
		".job-search-card__subtitle a",
		"a[data-control-name='job_search_company_name']",
		".job-card__subtitle a",
		".company-name a",
		".jobs-unified-top-card__company-name a",
		".job-result-card__subtitle-link",

		// Generic fallbacks
		"a[href*='/company/']",
		"a[href*='/school/']",
	}

	for _, companySel := range companySelectors {
		companyLink := sel.Find(companySel).First()
		if companyLink.Length() > 0 {
			company := strings.TrimSpace(companyLink.Text())
			if company != "" && len(company) > 1 { // Basic validation
				job.Company = company
				if href, exists := companyLink.Attr("href"); exists {
					job.CompanyURL = s.normalizeURL(href)
				}
				s.debugLog("Found company with selector %s: %s", companySel, job.Company)
				break
			}
		}
	}

	// UPDATED: Current LinkedIn location selectors (2025)
	locationSelectors := []string{
		// Primary current selectors
		".job-card-container__metadata-wrapper",
		".artdeco-entity-lockup__caption",
		".entity-result__secondary-subtitle",
		".base-search-card__metadata",
		".job-result-card__location",

		// Secondary selectors
		".job-search-card__location",
		"[data-test='job-location']",
		".job-search-card__location span",
		".job-card__location",
		".job-location",
		".jobs-unified-top-card__bullet",
		".location",

		// Generic fallbacks that might contain location
		".job-card-container__metadata",
		".metadata",
	}

	for _, locSel := range locationSelectors {
		location := sel.Find(locSel).First()
		if location.Length() > 0 {
			loc := strings.TrimSpace(location.Text())
			if loc != "" && len(loc) > 2 { // Basic validation
				job.Location = loc
				s.debugLog("Found location with selector %s: %s", locSel, job.Location)
				break
			}
		}
	}

	return job
}

// analyzePageStructure analyzes the page structure for debugging with enhanced detection
func (s *LinkedInScraper) analyzePageStructure(doc *goquery.Document) string {
	var analysis strings.Builder

	analysis.WriteString("=== ENHANCED PAGE STRUCTURE ANALYSIS ===\n")

	// Check for various job-related classes and IDs
	checkElements := []string{
		".job-search-card", ".jobs-search__results-list", ".job-result-card",
		".base-card", ".scaffold-layout__list-container", "[data-entity-urn]",
		".jobs-search-results__list-item", ".job-card-container", ".job-card-search",
		"[data-job-id]", ".artdeco-entity-lockup", ".job-card-square",
		".jobs-search__job-card", ".jobs-unified-top-card", "article", ".card",
		"[data-urn]", "li[data-urn]", "li", "div[class*='job']", "[class*='search']",
		"main", "#main", ".main-content", ".content", ".jobs-search",
	}

	foundElements := 0
	for _, selector := range checkElements {
		count := doc.Find(selector).Length()
		if count > 0 {
			analysis.WriteString(fmt.Sprintf("✅ Found %d elements with '%s'\n", count, selector))
			foundElements += count
		}
	}

	if foundElements == 0 {
		analysis.WriteString("❌ No job-related elements found with standard selectors\n")
	}

	// Check page title and meta info
	title := doc.Find("title").Text()
	analysis.WriteString(fmt.Sprintf("📄 Page title: %s\n", strings.TrimSpace(title)))

	// Check for specific LinkedIn indicators
	indicators := []string{"linkedin.com/jobs", "job search", "job results", "jobs found", "no jobs", "search results"}
	pageText := strings.ToLower(doc.Text())
	for _, indicator := range indicators {
		if strings.Contains(pageText, indicator) {
			analysis.WriteString(fmt.Sprintf("🔍 Found indicator: '%s'\n", indicator))
		}
	}

	// Check for error/blocking messages
	errorKeywords := []string{"blocked", "captcha", "verification", "error", "access denied", "sign in", "login", "please try again"}
	for _, keyword := range errorKeywords {
		if strings.Contains(pageText, keyword) {
			analysis.WriteString(fmt.Sprintf("⚠️  Found warning keyword: '%s'\n", keyword))
		}
	}

	// Count total elements for context
	totalElements := doc.Find("*").Length()
	analysis.WriteString(fmt.Sprintf("📊 Total page elements: %d\n", totalElements))

	return analysis.String()
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
		s.debugLog("Trying approach %d for employee detection", i+1)
		employees, err := approach(companyURL)
		if err != nil {
			s.debugLog("Approach %d failed: %v", i+1, err)
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

	s.debugLog("Employee check completed in %v, found %d Indonesian employees", duration, len(allEmployees))

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

// ================================
// ORIGINAL FUNCTION: ProcessJobs
// ================================
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

// ========================================
// NEW ENHANCED FUNCTION: ProcessJobsWithFallback
// ========================================
// ProcessJobsWithFallback processes jobs with Indonesian detection but ALWAYS returns results
func (s *LinkedInScraper) ProcessJobsWithFallback(jobs []Job) []Job {
	var processedJobs []Job
	totalJobs := len(jobs)
	indonesianJobs := 0

	log.Printf("🔄 Processing %d jobs for Indonesian employee detection...", totalJobs)
	log.Printf("💡 Strategy: ALL jobs will be returned (prioritized by Indonesian employees)")

	for i, job := range jobs {
		log.Printf("[%d/%d] Processing: %s at %s", i+1, totalJobs, job.Title, job.Company)

		startTime := time.Now()
		hasIndonesian, employees, err := s.CheckIndonesianEmployees(job.CompanyURL)
		duration := time.Since(startTime)

		job.CheckDuration = duration.String()
		job.EmployeeCount = len(employees)

		if err != nil {
			s.debugLog("Error checking employees for %s: %v", job.Company, err)
			job.HasIndonesian = false
			job.IndonesianEmployees = []Employee{}
		} else {
			job.HasIndonesian = hasIndonesian
			job.IndonesianEmployees = employees

			if hasIndonesian {
				indonesianJobs++
				log.Printf("✅ Found %d Indonesian employees at %s", len(employees), job.Company)
			} else {
				log.Printf("➖ No Indonesian employees found at %s (still adding to results)", job.Company)
			}
		}

		// ALWAYS add the job to results, regardless of Indonesian employees
		processedJobs = append(processedJobs, job)

		// Progress indicator
		progress := float64(i+1) / float64(totalJobs) * 100
		log.Printf("Progress: %.1f%% (%d/%d)", progress, i+1, totalJobs)

		time.Sleep(s.delay)
	}

	// Enhanced summary log
	log.Printf("\n📊 PROCESSING SUMMARY:")
	log.Printf("   Total Jobs Processed: %d", totalJobs)
	log.Printf("   🎯 Jobs with Indonesian Employees: %d", indonesianJobs)
	log.Printf("   💼 Jobs without Indonesian Employees: %d", totalJobs-indonesianJobs)
	log.Printf("   ✅ All jobs included in results for your review!")

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

// ================================
// ORIGINAL FUNCTION: printResults
// ================================
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

// ==========================================
// NEW ENHANCED FUNCTION: printResultsEnhanced
// ==========================================
func printResultsEnhanced(jobs []Job) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("🇮🇩 ENHANCED LINKEDIN JOB SEARCH RESULTS WITH SMART PRIORITIZATION")
	fmt.Println(strings.Repeat("=", 100))

	totalJobs := len(jobs)
	jobsWithIndonesians := 0
	jobsWithoutIndonesians := 0
	totalIndonesianEmployees := 0

	for _, job := range jobs {
		if job.HasIndonesian {
			jobsWithIndonesians++
			totalIndonesianEmployees += len(job.IndonesianEmployees)
		} else {
			jobsWithoutIndonesians++
		}
	}

	fmt.Printf("📊 ENHANCED SUMMARY:\n")
	fmt.Printf("   Total Jobs Found: %d\n", totalJobs)
	fmt.Printf("   🎯 PRIORITY Jobs (Indonesian Employees): %d (%.1f%%) - APPLY FIRST!\n",
		jobsWithIndonesians, float64(jobsWithIndonesians)/float64(totalJobs)*100)
	fmt.Printf("   💼 ALTERNATIVE Jobs (No Indonesian Detected): %d (%.1f%%) - BACKUP OPTIONS\n",
		jobsWithoutIndonesians, float64(jobsWithoutIndonesians)/float64(totalJobs)*100)
	fmt.Printf("   👥 Total Indonesian Employees Found: %d\n", totalIndonesianEmployees)

	// Sort jobs: Indonesian employees first, then others
	sortedJobs := make([]Job, len(jobs))
	copy(sortedJobs, jobs)

	for i := 0; i < len(sortedJobs)-1; i++ {
		for j := i + 1; j < len(sortedJobs); j++ {
			if !sortedJobs[i].HasIndonesian && sortedJobs[j].HasIndonesian {
				sortedJobs[i], sortedJobs[j] = sortedJobs[j], sortedJobs[i]
			}
		}
	}

	// Print PRIORITY jobs with Indonesian employees first
	if jobsWithIndonesians > 0 {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("🎯 PRIORITY JOBS - COMPANIES WITH INDONESIAN EMPLOYEES")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("💡 Apply to these first! You can network with Indonesian colleagues.")

		priorityCount := 0
		for _, job := range sortedJobs {
			if job.HasIndonesian {
				priorityCount++
				fmt.Printf("\n🌟 PRIORITY #%d: %s\n", priorityCount, job.Title)
				fmt.Printf("   🏢 Company: %s\n", job.Company)
				fmt.Printf("   📍 Location: %s\n", job.Location)
				fmt.Printf("   🔗 Job URL: %s\n", job.JobURL)
				fmt.Printf("   🇮🇩 Indonesian Employees Found: %d\n", len(job.IndonesianEmployees))

				if len(job.IndonesianEmployees) > 0 {
					fmt.Printf("   👥 Indonesian Staff (for networking):\n")
					for _, emp := range job.IndonesianEmployees {
						confidenceStr := fmt.Sprintf("%.0f%%", emp.Confidence*100)
						fmt.Printf("      • %s", emp.Name)
						if emp.Position != "" {
							fmt.Printf(" (%s)", emp.Position)
						}
						fmt.Printf(" [Match Confidence: %s]\n", confidenceStr)
					}
				}

				fmt.Printf("   ⭐ STRATEGY: Mention Indonesian connection in your application!\n")
				fmt.Printf("   ⏱️  Detection Time: %s\n", job.CheckDuration)
				fmt.Println(strings.Repeat("-", 60))
			}
		}
	}

	// Print ALTERNATIVE jobs (without detected Indonesian employees)
	if jobsWithoutIndonesians > 0 {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("💼 ALTERNATIVE JOBS - STILL EXCELLENT OPPORTUNITIES")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("ℹ️  No Indonesian employees detected, but these are still great opportunities!")
		fmt.Println("💡 Indonesian employees may exist but weren't found in our search.")

		altCount := 0
		for _, job := range sortedJobs {
			if !job.HasIndonesian {
				altCount++
				fmt.Printf("\n💼 ALTERNATIVE #%d: %s\n", altCount, job.Title)
				fmt.Printf("   🏢 Company: %s\n", job.Company)
				fmt.Printf("   📍 Location: %s\n", job.Location)
				fmt.Printf("   🔗 Job URL: %s\n", job.JobURL)
				fmt.Printf("   🔍 Indonesian Check: No employees detected in our search\n")
				fmt.Printf("   💡 TIP: Research company manually or apply with standard approach\n")
				fmt.Printf("   🚀 OPPORTUNITY: Could be the first Indonesian employee!\n")
				fmt.Printf("   ⏱️  Detection Time: %s\n", job.CheckDuration)
				fmt.Println(strings.Repeat("-", 60))
			}
		}

		fmt.Println("\n💡 WHY ALTERNATIVE JOBS ARE STILL VALUABLE:")
		fmt.Println("   • Indonesian employees might exist but use different names")
		fmt.Println("   • Company may be open to hiring Indonesian employees")
		fmt.Println("   • Great opportunity to be a pioneer and build Indonesian community")
		fmt.Println("   • Still excellent career opportunities regardless of employee demographics")
	}

	// Final strategic recommendations
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📋 STRATEGIC ACTION PLAN")
	fmt.Println(strings.Repeat("=", 80))

	if jobsWithIndonesians > 0 {
		fmt.Printf("1. 🎯 IMMEDIATE ACTION: Apply to %d PRIORITY jobs with Indonesian employees\n", jobsWithIndonesians)
		fmt.Println("   → Mention Indonesian connection in your cover letter")
		fmt.Println("   → Reach out to Indonesian employees for referrals")
		fmt.Println("   → Use Indonesian community networks")
		fmt.Println("")
	}

	if jobsWithoutIndonesians > 0 {
		fmt.Printf("2. 💼 BACKUP STRATEGY: Consider %d ALTERNATIVE jobs as excellent options\n", jobsWithoutIndonesians)
		fmt.Println("   → Research companies thoroughly")
		fmt.Println("   → Apply with standard professional approach")
		fmt.Println("   → Could be opportunity to build Indonesian presence")
		fmt.Println("")
	}

	fmt.Println("3. 🔄 EXPAND SEARCH: Try different keywords or locations for more opportunities")
	fmt.Println("4. 📈 IMPROVE DATABASE: Report any Indonesian names we missed to enhance detection")
	fmt.Println("5. 🌐 NETWORK: Use Indonesian professional communities for additional opportunities")

	fmt.Printf("\n📊 SUCCESS METRICS: You now have %d total opportunities with clear prioritization!\n", totalJobs)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("🇮🇩 Enhanced LinkedIn Indonesian Employee Job Scraper v2.0")
		fmt.Println("===========================================================")
		fmt.Println("Usage: go run main.go <country> <job_title> [limit]")
		fmt.Println("Example: go run main.go \"Germany\" \"software engineer\" 25")
		fmt.Println("")
		fmt.Println("🆕 ENHANCED FEATURES:")
		fmt.Println("✅ ALWAYS returns ALL jobs found (no more empty results!)")
		fmt.Println("🎯 PRIORITIZES jobs with Indonesian employees")
		fmt.Println("💼 SHOWS alternative jobs as backup options")
		fmt.Println("📊 CLEAR action plan and strategic recommendations")
		fmt.Println("🔍 ENHANCED job search with multiple fallback strategies")
		fmt.Println("🛡️  IMPROVED LinkedIn blocking detection and avoidance")
		fmt.Println("")
		fmt.Println("🐛 DEBUG MODE:")
		fmt.Println("DEBUG=true go run main.go \"Germany\" \"software engineer\" 5")
		fmt.Println("")
		fmt.Println("💡 STRATEGY MODES:")
		fmt.Println("• Enhanced strategy (default): Always returns results with prioritization")
		fmt.Println("• Original strategy: Set useEnhancedStrategy = false in code")
		os.Exit(1)
	}

	country := os.Args[1]
	jobTitle := os.Args[2]
	limit := 25

	if len(os.Args) > 3 {
		fmt.Sscanf(os.Args[3], "%d", &limit)
	}

	// Strategy selection - you can change this
	useEnhancedStrategy := true // Set to false for original behavior

	fmt.Println("🇮🇩 Enhanced LinkedIn Indonesian Employee Job Scraper")
	fmt.Println("====================================================")
	fmt.Printf("🔍 Searching for '%s' jobs in %s (limit: %d)...\n", jobTitle, country, limit)

	if useEnhancedStrategy {
		fmt.Println("🎯 ENHANCED STRATEGY: Find jobs with Indonesian employees + show alternatives")
	} else {
		fmt.Println("🔍 ORIGINAL STRATEGY: Indonesian employee detection only")
	}

	if os.Getenv("DEBUG") == "true" || os.Getenv("SCRAPER_DEBUG") == "true" {
		fmt.Println("🐛 DEBUG MODE ENABLED - Detailed logging activated")
	}

	scraper, err := NewLinkedInScraper()
	if err != nil {
		log.Fatalf("Failed to initialize scraper: %v", err)
	}

	// Search for jobs with enhanced fallback strategies
	jobs, err := scraper.SearchJobs(country, jobTitle, limit)
	if err != nil {
		log.Fatalf("Failed to search jobs: %v", err)
	}

	if len(jobs) == 0 {
		fmt.Println("❌ No jobs found with current search terms.")
		fmt.Println("\n🔧 ENHANCED TROUBLESHOOTING SUGGESTIONS:")
		fmt.Println("1. 🎯 Try broader search terms:")
		fmt.Println("   • 'software engineer' instead of 'golang developer'")
		fmt.Println("   • 'developer' instead of specific technologies")
		fmt.Println("   • 'programmer' or 'engineer' for wider results")
		fmt.Println("")
		fmt.Println("2. 🗺️  Try specific locations:")
		fmt.Println("   • 'Berlin, Germany' instead of 'Germany'")
		fmt.Println("   • 'Amsterdam, Netherlands'")
		fmt.Println("   • 'London, United Kingdom'")
		fmt.Println("")
		fmt.Println("3. 🛡️  LinkedIn may be rate limiting:")
		fmt.Println("   • Try using a VPN (different location)")
		fmt.Println("   • Wait 30-60 minutes and retry")
		fmt.Println("   • Use DEBUG=true for detailed diagnostics")
		fmt.Println("")
		fmt.Println("4. 🐛 Enable debug mode for detailed analysis:")
		fmt.Println("   DEBUG=true go run main.go \"Germany\" \"software engineer\" 5")
		fmt.Println("")
		fmt.Println("5. 📊 Check debug output:")
		fmt.Println("   • Debug HTML files will be saved for analysis")
		fmt.Println("   • Page structure analysis will show what LinkedIn returned")
		fmt.Println("   • Response status codes will indicate blocking")
		return
	}

	fmt.Printf("✅ Found %d jobs! Now checking for Indonesian employees...\n", len(jobs))

	if useEnhancedStrategy {
		fmt.Println("💡 Enhanced Strategy: ALL jobs will be included in results")
		// Use enhanced strategy - always returns results
		processedJobs := scraper.ProcessJobsWithFallback(jobs)
		printResultsEnhanced(processedJobs)

		// Save results
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("linkedin_jobs_enhanced_%s_%s_%s.json",
			strings.ReplaceAll(strings.ToLower(country), " ", "_"),
			strings.ReplaceAll(strings.ToLower(jobTitle), " ", "_"),
			timestamp)

		if err := SaveResults(processedJobs, filename); err != nil {
			log.Printf("❌ Failed to save results: %v", err)
		} else {
			fmt.Printf("\n💾 Enhanced results saved to: %s\n", filename)
		}

		fmt.Println("\n✅ Enhanced job search completed!")
		fmt.Println("📊 Review both PRIORITY jobs (with Indonesian employees) and ALTERNATIVES")

	} else {
		fmt.Println("💡 Original Strategy: Indonesian employee focused results")
		// Use original strategy
		processedJobs := scraper.ProcessJobs(jobs)
		printResults(processedJobs)

		// Save results
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
	}

	fmt.Println("📈 Use the JSON file for further analysis or integration with other tools.")
}
