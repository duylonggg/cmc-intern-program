package handler

import (
	"encoding/json"
	"net/http"

	"mini-asm/internal/model"
	"mini-asm/internal/service"
)

// ScanHandler handles scan-related HTTP requests
type ScanHandler struct {
	scanService *service.ScanService
}

// PaginatedScanResults represents paginated scan results
type PaginatedScanResults struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// NewScanHandler creates a new scan handler
func NewScanHandler(scanService *service.ScanService) *ScanHandler {
	return &ScanHandler{
		scanService: scanService,
	}
}

// StartScan initiates a scan for an asset
// POST /assets/{id}/scan
func (h *ScanHandler) StartScan(w http.ResponseWriter, r *http.Request) {
	// Extract asset ID from path
	assetID := r.PathValue("id")
	if assetID == "" {
		http.Error(w, "asset ID required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		ScanType model.ScanType `json:"scan_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate scan type
	if !model.IsValidScanType(req.ScanType) {
		http.Error(w, "invalid scan type", http.StatusBadRequest)
		return
	}

	// ⚠️ WARNING CHECK: Active scans require authorization
	//if req.ScanType.RequiresPermission() {
	// Log warning for active scans
	//log.Printf("⚠️ ACTIVE SCAN REQUESTED: %s on asset %s", req.ScanType, assetID)
	//log.Printf("⚠️ Active scanning requires authorization from target owner")
	//log.Printf("⚠️ Ensure proper authorization is documented before proceeding")

	// In production, you might want to:
	// 1. Check if user has permission in database
	// 2. Require additional authorization token
	// 3. Send notification to security team
	// 4. Require explicit confirmation parameter

	// For training purposes, we allow it but warn heavily
	// Students should only scan localhost
	//}

	// Start scan
	job, err := h.scanService.StartScan(assetID, req.ScanType)
	if err != nil {
		status := mapErrorToStatus(err)
		http.Error(w, err.Error(), status)
		return
	}

	// Return job info
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202 Accepted (async operation)
	json.NewEncoder(w).Encode(job)
}

// GetScanJob retrieves scan job status
// GET /scan-jobs/{id}
func (h *ScanHandler) GetScanJob(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from path
	jobID := r.PathValue("id")
	if jobID == "" {
		http.Error(w, "job ID required", http.StatusBadRequest)
		return
	}

	// Get job
	job, err := h.scanService.GetScanJob(jobID)
	if err != nil {
		status := mapErrorToStatus(err)
		http.Error(w, err.Error(), status)
		return
	}

	// Return job
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// GetScanResults retrieves results for a scan job
// GET /scan-jobs/{id}/results
func (h *ScanHandler) GetScanResults(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from path
	jobID := r.PathValue("id")
	if jobID == "" {
		http.Error(w, "job ID required", http.StatusBadRequest)
		return
	}

	// Get results
	results, err := h.scanService.GetScanResults(jobID)
	if err != nil {
		status := mapErrorToStatus(err)
		http.Error(w, err.Error(), status)
		return
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// ListScanJobs retrieves all scan jobs for an asset
// GET /assets/{id}/scans?page=1&page_size=20&scan_type=dns&status=completed
func (h *ScanHandler) ListScanJobs(w http.ResponseWriter, r *http.Request) {
	// Extract asset ID from path
	assetID := r.PathValue("id")
	if assetID == "" {
		http.Error(w, "asset ID required", http.StatusBadRequest)
		return
	}

	// Parse pagination params
	page := parseIntParam(r, "page", 1)
	pageSize := parseIntParam(r, "page_size", 20)
	scanType := r.URL.Query().Get("scan_type") // Optional filter
	status := r.URL.Query().Get("status")      // Optional filter

	// Get all jobs
	jobs, err := h.scanService.ListScanJobs(assetID)
	if err != nil {
		status := mapErrorToStatus(err)
		http.Error(w, err.Error(), status)
		return
	}

	// Apply filters
	if scanType != "" || status != "" {
		filteredJobs := make([]*model.ScanJob, 0)
		for _, job := range jobs {
			// Filter by scan type
			if scanType != "" && string(job.ScanType) != scanType {
				continue
			}
			// Filter by status
			if status != "" && string(job.Status) != status {
				continue
			}
			filteredJobs = append(filteredJobs, job)
		}
		jobs = filteredJobs
	}

	// Apply pagination
	paginatedResult := paginateResults(jobs, page, pageSize)

	// Return paginated results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginatedResult)
}

// GetAssetSubdomains retrieves all subdomains for an asset
// GET /assets/{id}/subdomains?page=1&page_size=20
func (h *ScanHandler) GetAssetSubdomains(w http.ResponseWriter, r *http.Request) {
	// Extract asset ID from path
	assetID := r.PathValue("id")
	if assetID == "" {
		http.Error(w, "asset ID required", http.StatusBadRequest)
		return
	}

	// Parse pagination params
	page := parseIntParam(r, "page", 1)
	pageSize := parseIntParam(r, "page_size", 20)

	// Get all subdomains
	subdomains, err := h.scanService.GetAssetSubdomains(assetID)
	if err != nil {
		status := mapErrorToStatus(err)
		http.Error(w, err.Error(), status)
		return
	}

	// Apply pagination
	paginatedResult := paginateResults(subdomains, page, pageSize)

	// Return paginated results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginatedResult)
}

// GetAssetDNS retrieves all DNS records for an asset
// GET /assets/{id}/dns?page=1&page_size=20&record_type=A
func (h *ScanHandler) GetAssetDNS(w http.ResponseWriter, r *http.Request) {
	// Extract asset ID from path
	assetID := r.PathValue("id")
	if assetID == "" {
		http.Error(w, "asset ID required", http.StatusBadRequest)
		return
	}

	// Parse pagination params
	page := parseIntParam(r, "page", 1)
	pageSize := parseIntParam(r, "page_size", 20)
	recordType := r.URL.Query().Get("record_type") // Optional filter

	// Get all DNS records
	records, err := h.scanService.GetAssetDNSRecords(assetID)
	if err != nil {
		status := mapErrorToStatus(err)
		http.Error(w, err.Error(), status)
		return
	}

	// Filter by record type if specified
	if recordType != "" {
		filteredRecords := make([]*model.DNSRecord, 0)
		for _, record := range records {
			if record.RecordType == recordType {
				filteredRecords = append(filteredRecords, record)
			}
		}
		records = filteredRecords
	}

	// Apply pagination
	paginatedResult := paginateResults(records, page, pageSize)

	// Return paginated results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginatedResult)
}

// GetAssetWHOIS retrieves WHOIS information for an asset
// GET /assets/{id}/whois
func (h *ScanHandler) GetAssetWHOIS(w http.ResponseWriter, r *http.Request) {
	// Extract asset ID from path
	assetID := r.PathValue("id")
	if assetID == "" {
		http.Error(w, "asset ID required", http.StatusBadRequest)
		return
	}

	// Get WHOIS record
	record, err := h.scanService.GetAssetWHOIS(assetID)
	if err != nil {
		status := mapErrorToStatus(err)
		http.Error(w, err.Error(), status)
		return
	}

	// Return record
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

// Helper functions for pagination

// paginateResults applies in-memory pagination to any slice
func paginateResults(data interface{}, page int, pageSize int) PaginatedScanResults {
	// Use reflection to handle any slice type
	// In production, you'd implement database-level pagination for efficiency

	// For simplicity, convert to []interface{} to count and slice
	// This is a teaching implementation - production should paginate at DB level

	var items []interface{}
	var total int

	// Type assertion for common types
	switch v := data.(type) {
	case []*model.ScanJob:
		total = len(v)
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		return PaginatedScanResults{
			Data:       v[start:end],
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: (total + pageSize - 1) / pageSize,
		}
	case []*model.Subdomain:
		total = len(v)
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		return PaginatedScanResults{
			Data:       v[start:end],
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: (total + pageSize - 1) / pageSize,
		}
	case []*model.DNSRecord:
		total = len(v)
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		return PaginatedScanResults{
			Data:       v[start:end],
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: (total + pageSize - 1) / pageSize,
		}
	default:
		// Fallback: return empty result
		return PaginatedScanResults{
			Data:       items,
			Total:      0,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: 0,
		}
	}
}

// GetAssetAllResults retrieves ALL scan results for an asset
// GET /assets/{id}/results
//
// Returns combined view of DNS records, WHOIS info, and subdomains.
// This is useful for getting complete reconnaissance data in one request.
func (h *ScanHandler) GetAssetAllResults(w http.ResponseWriter, r *http.Request) {
	// Extract asset ID from path
	assetID := r.PathValue("id")
	if assetID == "" {
		http.Error(w, "asset ID required", http.StatusBadRequest)
		return
	}

	// Get all scan results
	results, err := h.scanService.GetAssetAllScanResults(assetID)
	if err != nil {
		status := mapErrorToStatus(err)
		http.Error(w, err.Error(), status)
		return
	}

	// Return combined results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// DemoSyncVsAsync demonstrates sync vs async scanning performance
// POST /assets/{id}/scan/demo
//
// This is a teaching endpoint that runs the same scans twice:
// 1. Synchronously (sequential execution)
// 2. Asynchronously (concurrent execution with goroutines)
//
// Output is printed to server console for comparison.
func (h *ScanHandler) DemoSyncVsAsync(w http.ResponseWriter, r *http.Request) {
	// Extract asset ID from path
	assetID := r.PathValue("id")
	if assetID == "" {
		http.Error(w, "asset ID required", http.StatusBadRequest)
		return
	}

	// Run the demo (output goes to console)
	err := h.scanService.DemoSyncVsAsync(assetID)
	if err != nil {
		status := mapErrorToStatus(err)
		http.Error(w, err.Error(), status)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Demo completed! Check server console for detailed output.",
		"note":    "This endpoint demonstrates sync vs async performance comparison.",
	})
}
