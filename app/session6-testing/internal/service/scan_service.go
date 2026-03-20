package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mini-asm/internal/model"
	"mini-asm/internal/scanner"
	"mini-asm/internal/storage"

	"github.com/google/uuid"
)

// ScanService handles scan operations
type ScanService struct {
	storage          storage.Storage
	scanStorage      storage.ScanStorage
	dnsScanner       *scanner.DNSScanner
	whoisScanner     *scanner.WHOISScanner
	subdomainScanner *scanner.SubdomainScanner
}

// NewScanService creates a new scan service instance
func NewScanService(store storage.Storage, scanStore storage.ScanStorage) (*ScanService, error) {
	subdomainScanner, err := scanner.NewSubdomainScanner()
	if err != nil {
		return nil, fmt.Errorf("failed to create subdomain scanner: %w", err)
	}

	return &ScanService{
		storage:          store,
		scanStorage:      scanStore,
		dnsScanner:       scanner.NewDNSScanner(),
		whoisScanner:     scanner.NewWHOISScanner(),
		subdomainScanner: subdomainScanner,
	}, nil
}

// StartScan initiates a scan for an asset
// Returns the scan job ID immediately (async pattern)
func (s *ScanService) StartScan(assetID string, scanType model.ScanType) (*model.ScanJob, error) {
	// Validate asset exists
	asset, err := s.storage.GetByID(assetID)
	if err != nil {
		return nil, fmt.Errorf("asset not found: %w", err)
	}

	// Validate scan type
	if !model.IsValidScanType(scanType) {
		return nil, fmt.Errorf("invalid scan type: %s", scanType)
	}

	// Create scan job
	now := time.Now()
	job := &model.ScanJob{
		ID:        uuid.New().String(),
		AssetID:   assetID,
		ScanType:  scanType,
		Status:    model.ScanStatusPending,
		StartedAt: now,
		CreatedAt: now,
	}

	// Save job to database
	if err := s.scanStorage.CreateScanJob(job); err != nil {
		return nil, fmt.Errorf("failed to create scan job: %w", err)
	}

	// Start scan in background
	//go s.performScan(asset, job)

	s.performScan(asset, job)

	return job, nil
}

// performScan executes the actual scanning in the background
func (s *ScanService) performScan(asset *model.Asset, job *model.ScanJob) {
	// Update status to running
	job.Status = model.ScanStatusRunning
	s.scanStorage.UpdateScanJob(job)

	// Perform scan based on type
	var err error
	switch job.ScanType {
	case model.ScanTypeAll:
		err = s.performAllScans(asset, job)
	case model.ScanTypeDNS:
		err = s.performDNSScan(asset, job)
	case model.ScanTypeWHOIS:
		err = s.performWHOISScan(asset, job)
	case model.ScanTypeSubdomain:
		err = s.performSubdomainScan(asset, job)
	default:
		err = fmt.Errorf("unsupported scan type: %s", job.ScanType)
	}

	// Update job with results
	now := time.Now()
	job.EndedAt = &now

	if err != nil {
		job.Status = model.ScanStatusFailed
		job.Error = err.Error()
	} else {
		if job.Results == 0 {
			job.Status = model.ScanStatusPartial
			job.Error = "no results found"
		} else {
			job.Status = model.ScanStatusCompleted
		}
	}

	s.scanStorage.UpdateScanJob(job)
}

// performDNSScan executes DNS scanning
func (s *ScanService) performDNSScan(asset *model.Asset, job *model.ScanJob) error {
	// Scan DNS records
	records, err := s.dnsScanner.Scan(asset)
	if err != nil {
		return fmt.Errorf("DNS scan failed: %w", err)
	}

	// Save results
	for _, record := range records {
		record.ID = uuid.New().String()
		record.AssetID = asset.ID
		record.ScanJobID = job.ID
		record.CreatedAt = time.Now()

		if err := s.scanStorage.CreateDNSRecord(record); err != nil {
			return fmt.Errorf("failed to save DNS record: %w", err)
		}
	}

	job.Results = len(records)
	return nil
}

// performWHOISScan executes WHOIS scanning
func (s *ScanService) performWHOISScan(asset *model.Asset, job *model.ScanJob) error {
	// Scan WHOIS information
	record, err := s.whoisScanner.Scan(asset)
	if err != nil {
		return fmt.Errorf("WHOIS scan failed: %w", err)
	}

	// Save result
	record.ID = uuid.New().String()
	record.AssetID = asset.ID
	record.ScanJobID = job.ID
	record.CreatedAt = time.Now()

	if err := s.scanStorage.CreateWHOISRecord(record); err != nil {
		return fmt.Errorf("failed to save WHOIS record: %w", err)
	}

	job.Results = 1
	return nil
}

// performSubdomainScan executes subdomain enumeration
func (s *ScanService) performSubdomainScan(asset *model.Asset, job *model.ScanJob) error {
	// Create context with timeout (5 minutes max)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Scan subdomains
	subdomains, err := s.subdomainScanner.Scan(asset, ctx)
	if err != nil {
		return fmt.Errorf("subdomain scan failed: %w", err)
	}

	// Save results
	for _, subdomain := range subdomains {
		subdomain.ID = uuid.New().String()
		subdomain.AssetID = asset.ID
		subdomain.ScanJobID = job.ID
		subdomain.CreatedAt = time.Now()

		if err := s.scanStorage.CreateSubdomain(subdomain); err != nil {
			return fmt.Errorf("failed to save subdomain: %w", err)
		}
	}

	job.Results = len(subdomains)
	return nil
}

// performAllScans executes all available passive scans (DNS, WHOIS, Subdomain)
// This is a convenience method to run comprehensive reconnaissance
func (s *ScanService) performAllScans(asset *model.Asset, job *model.ScanJob) error {
	totalResults := 0
	errors := []string{}

	// Run DNS scan
	if err := s.performDNSScan(asset, job); err != nil {
		errors = append(errors, fmt.Sprintf("DNS scan: %v", err))
	} else {
		totalResults += job.Results
	}

	// Reset job.Results for next scan
	job.Results = 0

	// Run WHOIS scan
	if err := s.performWHOISScan(asset, job); err != nil {
		errors = append(errors, fmt.Sprintf("WHOIS scan: %v", err))
	} else {
		totalResults += job.Results
	}

	// Reset job.Results for next scan
	job.Results = 0

	// Run Subdomain scan
	if err := s.performSubdomainScan(asset, job); err != nil {
		errors = append(errors, fmt.Sprintf("Subdomain scan: %v", err))
	} else {
		totalResults += job.Results
	}

	// Set final total results
	job.Results = totalResults

	// If all scans failed, return error
	if len(errors) == 3 {
		return fmt.Errorf("all scans failed: %v", errors)
	}

	// If some scans failed, return partial error
	if len(errors) > 0 {
		return fmt.Errorf("some scans failed: %v", errors)
	}

	return nil
}

// ============================================================================
// SYNC VS ASYNC COMPARISON - Teaching Examples
// ============================================================================

// performAllScansSync runs all passive scans SEQUENTIALLY (one after another)
// This is a SYNCHRONOUS approach - each scan waits for previous to complete
func (s *ScanService) performAllScansSync(asset *model.Asset, job *model.ScanJob) error {
	startTime := time.Now()
	totalResults := 0
	errors := []string{}

	fmt.Printf("🔵 SYNC: Starting sequential scans for %s\n", asset.Name)

	// Run DNS scan - WAIT for completion
	fmt.Printf("  → DNS scan starting...\n")
	dnsStart := time.Now()
	if err := s.performDNSScan(asset, job); err != nil {
		errors = append(errors, fmt.Sprintf("DNS scan: %v", err))
	} else {
		totalResults += job.Results
		fmt.Printf("  ✓ DNS scan completed in %v (found %d records)\n", time.Since(dnsStart), job.Results)
	}
	job.Results = 0

	// Run WHOIS scan - WAIT for completion
	fmt.Printf("  → WHOIS scan starting...\n")
	whoisStart := time.Now()
	if err := s.performWHOISScan(asset, job); err != nil {
		errors = append(errors, fmt.Sprintf("WHOIS scan: %v", err))
	} else {
		totalResults += job.Results
		fmt.Printf("  ✓ WHOIS scan completed in %v (found %d records)\n", time.Since(whoisStart), job.Results)
	}
	job.Results = 0

	// Run Subdomain scan - WAIT for completion
	fmt.Printf("  → Subdomain scan starting...\n")
	subdomainStart := time.Now()
	if err := s.performSubdomainScan(asset, job); err != nil {
		errors = append(errors, fmt.Sprintf("Subdomain scan: %v", err))
	} else {
		totalResults += job.Results
		fmt.Printf("  ✓ Subdomain scan completed in %v (found %d records)\n", time.Since(subdomainStart), job.Results)
	}

	job.Results = totalResults

	fmt.Printf("🔵 SYNC: Total time: %v (found %d total results)\n\n", time.Since(startTime), totalResults)

	// Handle errors
	if len(errors) == 3 {
		return fmt.Errorf("all scans failed: %v", errors)
	}
	if len(errors) > 0 {
		return fmt.Errorf("some scans failed: %v", errors)
	}

	return nil
}

// performAllScansAsync runs all passive scans CONCURRENTLY using goroutines
// This is an ASYNCHRONOUS approach - all scans run at the same time
func (s *ScanService) performAllScansAsync(asset *model.Asset, job *model.ScanJob) error {
	startTime := time.Now()

	fmt.Printf("🟢 ASYNC: Starting concurrent scans for %s\n", asset.Name)

	// Shared state protection
	var (
		mu           sync.Mutex     // Protects shared variables
		wg           sync.WaitGroup // Waits for all goroutines
		totalResults int
		errors       []string
	)

	// Channel for scan results (optional pattern)
	type scanResult struct {
		scanType string
		results  int
		duration time.Duration
		err      error
	}
	resultChan := make(chan scanResult, 3)

	// Launch DNS scan in goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanStart := time.Now()

		// Create temporary job for this scan
		tempJob := &model.ScanJob{
			ID:        job.ID,
			AssetID:   asset.ID,
			ScanType:  model.ScanTypeDNS,
			Status:    model.ScanStatusRunning,
			CreatedAt: job.CreatedAt,
		}

		err := s.performDNSScan(asset, tempJob)
		duration := time.Since(scanStart)

		// Send result to channel
		resultChan <- scanResult{
			scanType: "DNS",
			results:  tempJob.Results,
			duration: duration,
			err:      err,
		}
	}()

	// Launch WHOIS scan in goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanStart := time.Now()

		// Create temporary job for this scan
		tempJob := &model.ScanJob{
			ID:        job.ID,
			AssetID:   asset.ID,
			ScanType:  model.ScanTypeWHOIS,
			Status:    model.ScanStatusRunning,
			CreatedAt: job.CreatedAt,
		}

		err := s.performWHOISScan(asset, tempJob)
		duration := time.Since(scanStart)

		// Send result to channel
		resultChan <- scanResult{
			scanType: "WHOIS",
			results:  tempJob.Results,
			duration: duration,
			err:      err,
		}
	}()

	// Launch Subdomain scan in goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanStart := time.Now()

		// Create temporary job for this scan
		tempJob := &model.ScanJob{
			ID:        job.ID,
			AssetID:   asset.ID,
			ScanType:  model.ScanTypeSubdomain,
			Status:    model.ScanStatusRunning,
			CreatedAt: job.CreatedAt,
		}

		err := s.performSubdomainScan(asset, tempJob)
		duration := time.Since(scanStart)

		// Send result to channel
		resultChan <- scanResult{
			scanType: "Subdomain",
			results:  tempJob.Results,
			duration: duration,
			err:      err,
		}
	}()

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results from channel
	for result := range resultChan {
		if result.err != nil {
			mu.Lock()
			errors = append(errors, fmt.Sprintf("%s scan: %v", result.scanType, result.err))
			mu.Unlock()
			fmt.Printf("  ✗ %s scan failed in %v: %v\n", result.scanType, result.duration, result.err)
		} else {
			mu.Lock()
			totalResults += result.results
			mu.Unlock()
			fmt.Printf("  ✓ %s scan completed in %v (found %d records)\n", result.scanType, result.duration, result.results)
		}
	}

	job.Results = totalResults

	fmt.Printf("🟢 ASYNC: Total time: %v (found %d total results)\n\n", time.Since(startTime), totalResults)

	// Handle errors
	if len(errors) == 3 {
		return fmt.Errorf("all scans failed: %v", errors)
	}
	if len(errors) > 0 {
		return fmt.Errorf("some scans failed: %v", errors)
	}

	return nil
}

// DemoSyncVsAsync demonstrates the difference between sync and async scanning
// This is a teaching function to show performance comparison
func (s *ScanService) DemoSyncVsAsync(assetID string) error {
	// Validate asset exists
	asset, err := s.storage.GetByID(assetID)
	if err != nil {
		return fmt.Errorf("asset not found: %w", err)
	}

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║         SYNC vs ASYNC PERFORMANCE COMPARISON                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create job for sync test
	syncJob := &model.ScanJob{
		ID:        uuid.New().String(),
		AssetID:   assetID,
		ScanType:  model.ScanTypeAll,
		Status:    model.ScanStatusRunning,
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	// Save sync job to database (required for foreign key constraints)
	if err := s.scanStorage.CreateScanJob(syncJob); err != nil {
		return fmt.Errorf("failed to create sync job: %w", err)
	}

	// Test SYNCHRONOUS approach
	fmt.Println("TEST 1: SYNCHRONOUS (Sequential) Execution")
	fmt.Println("-------------------------------------------")
	syncErr := s.performAllScansSync(asset, syncJob)

	// Update sync job status in database
	now := time.Now()
	syncJob.EndedAt = &now
	if syncErr != nil {
		syncJob.Status = model.ScanStatusFailed
		syncJob.Error = syncErr.Error()
	} else {
		syncJob.Status = model.ScanStatusCompleted
	}
	s.scanStorage.UpdateScanJob(syncJob)

	// Create job for async test
	asyncJob := &model.ScanJob{
		ID:        uuid.New().String(),
		AssetID:   assetID,
		ScanType:  model.ScanTypeAll,
		Status:    model.ScanStatusRunning,
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	// Save async job to database (required for foreign key constraints)
	if err := s.scanStorage.CreateScanJob(asyncJob); err != nil {
		return fmt.Errorf("failed to create async job: %w", err)
	}

	// Test ASYNCHRONOUS approach
	fmt.Println("TEST 2: ASYNCHRONOUS (Concurrent) Execution")
	fmt.Println("--------------------------------------------")
	asyncErr := s.performAllScansAsync(asset, asyncJob)

	// Update async job status in database
	now2 := time.Now()
	asyncJob.EndedAt = &now2
	if asyncErr != nil {
		asyncJob.Status = model.ScanStatusFailed
		asyncJob.Error = asyncErr.Error()
	} else {
		asyncJob.Status = model.ScanStatusCompleted
	}
	s.scanStorage.UpdateScanJob(asyncJob)

	// Summary
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                         SUMMARY                                ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("SYNC (Sequential):")
	fmt.Printf("  - Scans run one after another\n")
	fmt.Printf("  - Total time = DNS time + WHOIS time + Subdomain time\n")
	fmt.Printf("  - Results: %d\n", syncJob.Results)
	if syncErr != nil {
		fmt.Printf("  - Error: %v\n", syncErr)
	}
	fmt.Println()
	fmt.Println("ASYNC (Concurrent):")
	fmt.Printf("  - All scans run simultaneously\n")
	fmt.Printf("  - Total time = MAX(DNS time, WHOIS time, Subdomain time)\n")
	fmt.Printf("  - Results: %d\n", asyncJob.Results)
	if asyncErr != nil {
		fmt.Printf("  - Error: %v\n", asyncErr)
	}
	fmt.Println()

	return nil
}

// GetScanJob retrieves a scan job by ID
func (s *ScanService) GetScanJob(jobID string) (*model.ScanJob, error) {
	return s.scanStorage.GetScanJob(jobID)
}

// ListScanJobs retrieves all scan jobs for an asset
func (s *ScanService) ListScanJobs(assetID string) ([]*model.ScanJob, error) {
	// Validate asset exists
	if _, err := s.storage.GetByID(assetID); err != nil {
		return nil, fmt.Errorf("asset not found: %w", err)
	}

	return s.scanStorage.ListScanJobsByAsset(assetID)
}

// GetScanResults retrieves results for a scan job
func (s *ScanService) GetScanResults(jobID string) (interface{}, error) {
	// Get job to determine type
	job, err := s.scanStorage.GetScanJob(jobID)
	if err != nil {
		return nil, fmt.Errorf("scan job not found: %w", err)
	}

	// Return results based on scan type
	switch job.ScanType {
	case model.ScanTypeAll:
		// For "all" scan type, return combined results
		dnsRecords, _ := s.scanStorage.GetDNSRecordsByScan(jobID)
		whoisRecords, _ := s.scanStorage.GetWHOISRecordsByScan(jobID)
		subdomains, _ := s.scanStorage.GetSubdomainsByScan(jobID)

		return map[string]interface{}{
			"dns_records":   dnsRecords,
			"whois_records": whoisRecords,
			"subdomains":    subdomains,
		}, nil
	case model.ScanTypeDNS:
		return s.scanStorage.GetDNSRecordsByScan(jobID)
	case model.ScanTypeWHOIS:
		return s.scanStorage.GetWHOISRecordsByScan(jobID)
	case model.ScanTypeSubdomain:
		return s.scanStorage.GetSubdomainsByScan(jobID)
	default:
		return nil, fmt.Errorf("unsupported scan type: %s", job.ScanType)
	}
}

// GetAssetSubdomains retrieves all subdomains for an asset
func (s *ScanService) GetAssetSubdomains(assetID string) ([]*model.Subdomain, error) {
	return s.scanStorage.GetSubdomainsByAsset(assetID)
}

// GetAssetDNSRecords retrieves all DNS records for an asset
func (s *ScanService) GetAssetDNSRecords(assetID string) ([]*model.DNSRecord, error) {
	return s.scanStorage.GetDNSRecordsByAsset(assetID)
}

// GetAssetWHOIS retrieves the latest WHOIS record for an asset
func (s *ScanService) GetAssetWHOIS(assetID string) (*model.WHOISRecord, error) {
	return s.scanStorage.GetWHOISRecordByAsset(assetID)
}

// GetAssetAllScanResults retrieves ALL scan results for an asset (DNS, WHOIS, Subdomains)
// This is a convenience method to get comprehensive view of all discovered information
func (s *ScanService) GetAssetAllScanResults(assetID string) (map[string]interface{}, error) {
	// Validate asset exists
	if _, err := s.storage.GetByID(assetID); err != nil {
		return nil, fmt.Errorf("asset not found: %w", err)
	}

	// Get all scan results (allow partial failures)
	dnsRecords, _ := s.scanStorage.GetDNSRecordsByAsset(assetID)
	whoisRecord, _ := s.scanStorage.GetWHOISRecordByAsset(assetID)
	subdomains, _ := s.scanStorage.GetSubdomainsByAsset(assetID)

	// Return combined results
	return map[string]interface{}{
		"dns_records":   dnsRecords,
		"whois_records": whoisRecord,
		"subdomains":    subdomains,
	}, nil
}

/*
 NOTES - Scan Service

=== ASYNC PATTERN ===

Problem: Scans take time (seconds to minutes)
Solution: Job/Task pattern with background processing

Flow:
```
Client:
  POST /assets/{id}/scan → scan service

Service:
  1. Create scan job (status: pending)
  2. Save to database
  3. Return job ID immediately
  4. Start scan in background goroutine

Client:
  GET /scan-jobs/{jobID} → check status
  {
    "status": "running",
    "results": 5,
    ...
  }
```

Benefits:
- HTTP request returns immediately (no timeout)
- Client can poll for progress
- Can queue multiple scans
- Better UX (progress updates)

=== GOROUTINE FOR BACKGROUND WORK ===

```go
// Start scan in background
go s.performScan(asset, job)
```

Key points:
1. **Non-blocking**: Main function returns immediately
2. **Concurrent**: Multiple scans can run simultaneously
3. **Error handling**: Errors saved to job.Error, not returned
4. **Database updates**: Job status tracked in database

Lifecycle:
```
pending → running → completed/failed/partial
```

=== ERROR HANDLING STRATEGIES ===

1. **Scan Errors**:
   ```go
   if err != nil {
       job.Status = ScanStatusFailed
       job.Error = err.Error()
   }
   ```
   - Save error message for debugging
   - User can see what went wrong

2. **No Results**:
   ```go
   if job.Results == 0 {
       job.Status = ScanStatusPartial
       job.Error = "no results found"
   }
   ```
   - Not a failure, but note it
   - Helps distinguish "scan worked but found nothing" vs "scan failed"

3. **Database Errors**:
   - Return error to caller
   - Job might not be saved
   - Client will get HTTP error

=== CONTEXT FOR CANCELLATION ===

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

subdomains, err := s.subdomainScanner.Scan(asset, ctx)
```

Why timeout?
- Subdomain scan could take very long
- Prevent infinite scans
- Resource management

5 minutes chosen because:
- 100 subdomains * 5s timeout = 500s = 8 minutes (worst case)
- With concurrency: much faster
- Balance: thorough vs timely

=== SCANNER INITIALIZATION ===

```go
func NewScanService(...) (*ScanService, error) {
    subdomainScanner, err := scanner.NewSubdomainScanner()
    if err != nil {
        return nil, err  // Wordlist loading failed
    }

    return &ScanService{
        dnsScanner: scanner.NewDNSScanner(),
        whoisScanner: scanner.NewWHOISScanner(),
        subdomainScanner: subdomainScanner,
    }
}
```

One-time initialization:
- Load wordlists once
- Reuse scanners for all scans
- Efficient (no reload per scan)

=== SCAN TYPE DISPATCH ===

```go
switch job.ScanType {
case model.ScanTypeDNS:
    err = s.performDNSScan(asset, job)
case model.ScanTypeWHOIS:
    err = s.performWHOISScan(asset, job)
case model.ScanTypeSubdomain:
    err = s.performSubdomainScan(asset, job)
default:
    err = fmt.Errorf("unsupported scan type: %s", job.ScanType)
}
```

Dispatch pattern:
- Single entry point (performScan)
- Route to specific implementation
- Easy to add new scan types

Alternative (more extensible):
```go
type Scanner interface {
    Scan(asset, job) error
}

scanners := map[ScanType]Scanner{
    ScanTypeDNS: dnsScanner,
    ScanTypeWHOIS: whoisScanner,
}

scanner := scanners[job.ScanType]
err := scanner.Scan(asset, job)
```

=== RESULT STORAGE PATTERN ===

```go
for _, record := range records {
    record.ID = uuid.New().String()
    record.AssetID = asset.ID
    record.ScanJobID = job.ID
    record.CreatedAt = time.Now()

    s.scanStorage.CreateDNSRecord(record)
}

job.Results = len(records)
```

Pattern:
1. Scanner returns results (without IDs)
2. Service assigns IDs, foreign keys, timestamps
3. Service saves to database
4. Service updates job with count

Separation of concerns:
- Scanner: Discovery logic
- Service: Orchestration, persistence
- Storage: Database operations

=== RETRIEVAL METHODS ===

Multiple ways to get results:

1. **By Job ID** (what was found in this scan):
   ```go
   GetScanResults(jobID)
   ```

2. **By Asset ID** (all historical data for asset):
   ```go
   GetAssetSubdomains(assetID)
   GetAssetDNSRecords(assetID)
   GetAssetWHOIS(assetID)
   ```

Use cases:
- Job ID: "Show me what this scan found"
- Asset ID: "Show me everything we know about this asset"

=== VALIDATION ===

```go
// Validate asset exists
asset, err := s.storage.GetByID(assetID)
if err != nil {
    return nil, fmt.Errorf("asset not found: %w", err)
}

// Validate scan type
if !model.IsValidScanType(scanType) {
    return nil, fmt.Errorf("invalid scan type: %s", scanType)
}
```

Fail fast:
- Check preconditions before starting work
- Clear error messages
- Prevent wasted computation

=== COMPARISON WITH PREVIOUS SESSIONS ===

Session 4: Synchronous CRUD operations
Session 5:
  - Async operations (background goroutines)
  - Job tracking (pending → running → completed)
  - Multiple scanners coordinated
  - Complex error handling

New concepts:
  - Goroutines for background work
  - Context for cancellation/timeout
  - Job/task pattern
  - Scanner abstraction

=== PRODUCTION CONSIDERATIONS ===

For production systems, consider:

1. **Job Queue**:
   - Current: goroutine per scan
   - Better: Queue (Redis, RabbitMQ)
   - Benefits: Distributed, persistent, rate limiting

2. **Worker Pool**:
   - Current: Unlimited goroutines
   - Better: Fixed number of workers
   - Benefits: Resource control, fairness

3. **Retry Logic**:
   - Current: Single attempt
   - Better: Retry with exponential backoff
   - Benefits: Handle transient failures

4. **Distributed Scanning**:
   - Current: Single server
   - Better: Multiple scan workers
   - Benefits: Scalability, fault tolerance

5. **Scan Scheduling**:
   - Current: Manual trigger
   - Better: Cron-like scheduling
   - Benefits: Automated monitoring

6. **Rate Limiting**:
   - Current: Per-scanner limits
   - Better: Global rate limiter
   - Benefits: Protect external services

Our implementation: Educational, great foundation!
Students learn patterns applicable to production.

=== STUDENT EXERCISES ===

1. **Add Scan Cancellation**:
   ```go
   func (s *ScanService) CancelScan(jobID string) error
   // Set context.Cancel
   // Update job status to cancelled
   ```

2. **Add Scan Retry**:
   ```go
   func (s *ScanService) RetryScan(jobID string) error
   // Create new job based on failed job
   // Restart scan
   ```

3. **Add Bulk Scan**:
   ```go
   func (s *ScanService) ScanAllAssets(scanType ScanType) ([]*ScanJob, error)
   // Scan all assets of a type
   // Return slice of job IDs
   ```

4. **Add Scan Comparison**:
   ```go
   func (s *ScanService) CompareScanResults(job1ID, job2ID string) (*ScanComparison, error)
   // Find differences between two scans
   // Useful for change detection
   ```

5. **Add Webhook Notifications**:
   ```go
   func (s *ScanService) RegisterWebhook(url string) error
   // Call webhook when scan completes
   // Enable integrations (Slack, email, etc.)
   ```

=== TESTING STRATEGIES ===

1. **Unit Tests**:
   - Mock storage interfaces
   - Test each scan type
   - Test error conditions

2. **Integration Tests**:
   - Real scanners
   - Test database (Docker)
   - Verify complete flow

3. **Test Scenarios**:
   - Successful scan (results found)
   - Failed scan (network error)
   - Partial scan (timeout)
   - No results (valid but empty)
   - Invalid asset ID
   - Invalid scan type

=== KEY TAKEAWAYS ===

1. Async pattern via goroutines
2. Job tracking for long operations
3. Context for cancellation/timeout
4. Separation: scanner vs service vs storage
5. Error handling at multiple levels
6. Multiple retrieval patterns (by job, by asset)
7. Validation before work
8. Foundation for production systems

Service layer orchestrates everything - the conductor of the orchestra!
*/
