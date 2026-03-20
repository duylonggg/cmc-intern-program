# 🔍 Session 5: EASM (External Attack Surface Management) - Scanning

## Overview

**Prerequisites:** Sessions 1-4 (Foundation, Basic API, Database, Advanced Features)

This session introduces **External Attack Surface Management** concepts by implementing automated scanning capabilities. Students will learn how to discover and monitor external-facing assets (domains, subdomains, DNS records, WHOIS information) through programmatic scanning.

**Key Learning Objectives:**

- Understanding EASM and attack surface concepts
- Implementing async job/task patterns
- External API/service integration (DNS, WHOIS)
- Background processing with goroutines
- Scanner architecture and abstraction
- Multi-table database relationships

---

## 📋 Table of Contents

1. [What is EASM?](#what-is-easm)
2. [Session 5 Features](#session-5-features)
3. [Quick Start](#quick-start)
4. [Architecture Overview](#architecture-overview)
5. [Implemented Scanners](#implemented-scanners)
6. [API Endpoints Reference](#api-endpoints-reference)
7. [Student Exercises](#student-exercises)
8. [Troubleshooting](#troubleshooting)

---

## What is EASM?

**External Attack Surface Management (EASM)** helps organizations discover and monitor their external-facing assets:

### Why EASM Matters

```
Traditional Security: "Protect what we know we have"
EASM: "Discover what attackers can see"
```

### Key Problems EASM Solves

1. **Shadow IT Discovery**
   - Forgotten domains/subdomains
   - Unauthorized cloud instances
   - Abandoned services

2. **Continuous Monitoring**
   - New assets appearing
   - Configuration changes
   - Expiring domains/certificates

3. **Attack Surface Visibility**
   - What can attackers see?
   - What ports are open?
   - What services are exposed?

### EASM Discovery Chain

```
Domain (example.com)
  ├─► WHOIS Scan ────► Registration info, expiration dates
  ├─► DNS Scan ──────► IP addresses, mail servers
  │                    │
  │                    └─► Create IP assets
  │                         └─► Port scanning (student exercise)
  │
  └─► Subdomain Scan ► api.example.com, mail.example.com, ...
                       │
                       └─► Recursive scanning of each subdomain
```

---

## 🔐 Passive vs Active Scanning

**CRITICAL CONCEPT:** In EASM and security testing, scans are categorized as either **passive** or **active**. Understanding this difference is essential for legal, ethical, and practical reasons.

### 🔵 Passive Scanning (Safe, OSINT)

**Definition:** Using publicly available information without directly interacting with target systems.

**Characteristics:**

- ✅ **Legal** - Uses public data sources, no authorization needed
- ✅ **Stealth** - Target doesn't know you're scanning
- ✅ **Safe** - Cannot harm target systems
- ✅ **24/7** - Can run continuously without concerns
- ✅ **No Detection** - Won't trigger IDS/IPS alerts

**Examples in Session 5:**

| Scan Type                    | Description              | Data Source             |
| ---------------------------- | ------------------------ | ----------------------- |
| **WHOIS**                    | Domain registration info | Public WHOIS database   |
| **DNS**                      | Domain name resolution   | Public DNS servers      |
| **Subdomain Enum**           | DNS bruteforce/CT logs   | DNS queries, CT logs    |
| **Certificate Transparency** | Certificate discovery    | Public CT logs (crt.sh) |
| **ASN Lookup**               | Network ownership        | Public WHOIS/BGP data   |

**When to use:**

- Initial reconnaissance
- Continuous monitoring
- Bug bounty discovery phase
- Threat intelligence gathering
- Any time you DON'T have explicit permission

**Real-world tools:**

- Censys, Shodan, ZoomEye (search engines)
- Certificate Transparency (crt.sh)
- WHOIS databases
- DNS enumeration tools

---

### 🔴 Active Scanning (⚠️ Requires Permission)

**Definition:** Directly probing target systems with network requests or packets.

**Characteristics:**

- ⚠️ **Legal Risk** - May be illegal without authorization
- ⚠️ **Detectable** - Target sees your activity in logs/IDS/IPS
- ⚠️ **Potentially Harmful** - Can crash services or trigger alerts
- ⚠️ **Requires Approval** - Only run with explicit permission
- ⚠️ **Logged Evidence** - Your IP will be recorded

**Examples in Session 5 (Student Exercises):**

| Scan Type                   | Risk Level | Why It's Active                             |
| --------------------------- | ---------- | ------------------------------------------- |
| **Port Scanning**           | HIGH       | TCP/UDP connection attempts to target ports |
| **Service Fingerprinting**  | HIGH       | Probing services for versions/banners       |
| **SSL Certificate Probing** | MEDIUM     | TLS handshakes with target servers          |
| **Vulnerability Scanning**  | HIGH       | Testing for exploitable weaknesses          |

**Legal Considerations:**

⚖️ **Laws you should know:**

- 🇺🇸 **Computer Fraud and Abuse Act (CFAA)** - Unauthorized access is illegal
- 🇬🇧 **Computer Misuse Act** - Unauthorized access is illegal
- 🇪🇺 **NIS Directive** - Security testing regulations
- 🌍 **Local laws** - Every country has different rules

⚖️ **When active scanning is LEGAL:**

- ✅ You own the assets
- ✅ Written permission from asset owner
- ✅ Bug bounty programs (within scope!)
- ✅ Penetration testing contracts
- ✅ Authorized security assessments

⚖️ **When active scanning is ILLEGAL:**

- ❌ No permission from owner
- ❌ Beyond bug bounty scope
- ❌ "Testing" third-party websites
- ❌ Scanning your employer without approval
- ❌ Scanning your ISP/university network

**Real-world example - What could go wrong:**

```
❌ BAD: "Let me port scan example.com to see what services they run"
   → Can result in: Legal action, FBI visit, criminal charges

✅ GOOD: "Let me check example.com's DNS records and WHOIS info"
   → Completely legal, public information

✅ GOOD: "Let me port scan localhost (127.0.0.1) to learn"
   → Your own machine, totally fine

✅ GOOD: "Client gave written approval to port scan their /24 network"
   → Authorized, documented, legal
```

---

### 📋 Session 5 Scan Classification

| Scan Type         | Category      | Session          | Permission Required? |
| ----------------- | ------------- | ---------------- | -------------------- |
| WHOIS             | 🔵 Passive    | 5A (Implemented) | ❌ No                |
| DNS               | 🔵 Passive    | 5A (Implemented) | ❌ No                |
| Subdomain Enum    | 🔵 Passive    | 5A (Implemented) | ❌ No                |
| Cert Transparency | 🔵 Passive    | 5B (Exercise)    | ❌ No                |
| ASN Lookup        | 🔵 Passive    | 5B (Exercise)    | ❌ No                |
| Port Scanning     | 🔴 **Active** | 5B (Exercise)    | ✅ **YES**           |
| Service Detection | 🔴 **Active** | 5B (Exercise)    | ✅ **YES**           |
| SSL Probing       | 🔴 **Active** | 5B (Exercise)    | ✅ **YES**           |

---

## Session 5 Features

### ✅ Implemented (For Teaching)

1. **WHOIS Scanning**
   - Domain registration information
   - Expiration date monitoring
   - Registrar and nameserver discovery
   - Contact information extraction

2. **DNS Scanning**
   - A/AAAA records (IPv4/IPv6 addresses)
   - MX records (mail servers)
   - NS records (nameservers)
   - TXT records (SPF, DKIM, verification)
   - CNAME records (aliases)

3. **Subdomain Enumeration**
   - DNS bruteforce with wordlist
   - Concurrent scanning (50 workers)
   - Rate limiting (100 req/sec)
   - Context-based cancellation

4. **Async Job Pattern**
   - Background processing
   - Status tracking (pending → running → completed/failed)
   - Result counting
   - Error handling

### 🔄 Student Exercises

Students will implement:

1. **Port Scanning** - Discover open ports and services
2. **SSL Certificate Analysis** - Certificate validation and expiration
3. **ASN Lookup** - Network ownership information
4. **Certificate Transparency** - Advanced subdomain discovery
5. **Recursive Scanning** - Automatic scanning of discovered assets

---

## Quick Start

### 1. Start Database

```bash
# Start PostgreSQL with Docker
docker-compose up -d

# Check database is running
docker-compose ps

# Migrations run automatically on first start
#   - 001_create_assets.up.sql (from Session 3)
#   - 002_create_scan_tables.up.sql (Session 5)
```

### 2. Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Verify configuration
cat .env

# Output:
# DB_HOST=localhost
# DB_PORT=5432
# DB_USER=postgres
# DB_PASSWORD=postgres
# DB_NAME=mini_asm
```

### 3. Start Server

```bash
# Run server
go run cmd/server/main.go

# Output:
# 🚀 Starting Mini ASM Server (Session 5 - EASM Scanning)...
# ✅ Storage initialized: PostgreSQL
# ✅ Service initialized: AssetService with Validator
# ✅ Service initialized: ScanService with DNS, WHOIS, Subdomain scanners
# ✅ Handlers initialized
# ✅ Routes registered:
#    ...
# 🌐 Server listening on http://localhost:8080
```

### 4. Run Your First Scan

```bash
# Step 1: Create a domain asset
curl -X POST http://localhost:8080/assets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "example.com",
    "type": "domain"
  }'

# Response:
# {
#   "id": "550e8400-e29b-41d4-a716-446655440000",
#   "name": "example.com",
#   "type": "domain",
#   "status": "active",
#   ...
# }

# Save the asset ID!
ASSET_ID="550e8400-e29b-41d4-a716-446655440000"

# Step 2: Start a DNS scan
curl -X POST http://localhost:8080/assets/$ASSET_ID/scan \
  -H "Content-Type: application/json" \
  -d '{
    "scan_type": "dns"
  }'

# Response (202 Accepted):
# {
#   "id": "123e4567-e89b-12d3-a456-426614174000",
#   "asset_id": "550e8400-e29b-41d4-a716-446655440000",
#   "scan_type": "dns",
#   "status": "pending",
#   "started_at": "2024-03-05T10:00:00Z",
#   ...
# }

# Save the job ID!
JOB_ID="123e4567-e89b-12d3-a456-426614174000"

# Step 3: Check scan status (poll until complete)
curl http://localhost:8080/scan-jobs/$JOB_ID

# Response:
# {
#   "id": "123e4567-e89b-12d3-a456-426614174000",
#   "status": "running",  # or "completed"
#   "results": 5,
#   ...
# }

# Step 4: Get scan results
curl http://localhost:8080/scan-jobs/$JOB_ID/results

# Response:
# [
#   {
#     "id": "...",
#     "record_type": "A",
#     "name": "example.com",
#     "value": "93.184.216.34",
#     ...
#   },
#   {
#     "record_type": "AAAA",
#     "name": "example.com",
#     "value": "2606:2800:220:1:248:1893:25c8:1946",
#     ...
#   },
#   ...
# ]

# Step 5: Try other scan types
# WHOIS scan
curl -X POST http://localhost:8080/assets/$ASSET_ID/scan \
  -H "Content-Type: application/json" \
  -d '{"scan_type": "whois"}'

# Subdomain scan (takes longer)
curl -X POST http://localhost:8080/assets/$ASSET_ID/scan \
  -H "Content-Type: application/json" \
  -d '{"scan_type": "subdomain"}'

# Run ALL passive scans at once (recommended!)
curl -X POST http://localhost:8080/assets/$ASSET_ID/scan \
  -H "Content-Type: application/json" \
  -d '{"scan_type": "all"}'

# Get combined results for "all" scan
curl http://localhost:8080/scan-jobs/$JOB_ID/results
# Response:
# {
#   "dns_records": [...],
#   "whois_records": [...],
#   "subdomains": [...]
# }
```

### 6. Get ALL results for an asset (NEW! 🆕)

After running scans, you can get all accumulated results for an asset:

```bash
# Get ALL scan results for an asset (DNS + WHOIS + Subdomains)
# This returns data from ALL scan jobs, not just one
curl http://localhost:8080/assets/$ASSET_ID/results

# Response (combined view):
# {
#   "dns_records": [
#     {"record_type": "A", "value": "93.184.216.34", ...},
#     {"record_type": "MX", "value": "mail.example.com", ...},
#     ...
#   ],
#   "whois_records": {
#     "registrar": "IANA",
#     "expiry_date": "2024-08-13T04:00:00Z",
#     ...
#   },
#   "subdomains": [
#     {"name": "www.example.com", "is_active": true, ...},
#     {"name": "mail.example.com", "is_active": true, ...},
#     ...
#   ]
# }

# Or get specific result types:
curl http://localhost:8080/assets/$ASSET_ID/dns        # DNS records only
curl http://localhost:8080/assets/$ASSET_ID/whois      # WHOIS only
curl http://localhost:8080/assets/$ASSET_ID/subdomains # Subdomains only
```

**When to use:**

- `/scan-jobs/{job_id}/results` → Results from a specific scan job
- `/assets/{id}/results` → **All results ever collected for this asset** 🆕

---

## Architecture Overview

### High-Level Flow

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ POST /assets/{id}/scan
       ▼
┌─────────────────────────────────┐
│      HTTP Handler Layer         │
│   (scan_handler.go)             │
│   - Parse request               │
│   - Return 202 Accepted         │
└──────┬──────────────────────────┘
       │ Start async scan
       ▼
┌─────────────────────────────────┐
│      Service Layer              │
│   (scan_service.go)             │
│   - Create scan job             │
│   - Launch goroutine            │
│   - Track status                │
└──────┬──────────────────────────┘
       │ Perform scan
       ▼
┌─────────────────────────────────┐
│      Scanner Layer              │
│   - dns_scanner.go              │
│   - whois_scanner.go            │
│   - subdomain_scanner.go        │
│   - (students add more)         │
└──────┬──────────────────────────┘
       │ Save results
       ▼
┌─────────────────────────────────┐
│      Storage Layer              │
│   (postgres.go)                 │
│   - scan_jobs table             │
│   - dns_records table           │
│   - whois_records table         │
│   - subdomains table            │
└─────────────────────────────────┘
```

### Database Schema

```
assets (Session 3)
  └─► scan_jobs (which scans were run)
      ├─► dns_records (what DNS records were found)
      ├─► whois_records (registration information)
      └─► subdomains (discovered subdomains)

Foreign Keys:
  - All tables reference assets (parent)
  - All scan results reference scan_jobs (provenance)
  - CASCADE delete: delete asset → all scans gone
```

### Code Organization

```
session5-easm/
├── cmd/server/main.go              # Entry point, wiring
├── internal/
│   ├── config/config.go            # .env file loading (viper)
│   ├── model/
│   │   ├── asset.go                # Asset entity (Session 2)
│   │   ├── scan.go                 # Scan entities (NEW!)
│   │   └── errors.go               # Error definitions
│   ├── scanner/                    # Scanner implementations (NEW!)
│   │   ├── dns_scanner.go          # DNS record scanning
│   │   ├── whois_scanner.go        # WHOIS information
│   │   ├── subdomain_scanner.go    # Subdomain enumeration
│   │   └── wordlists/
│   │       └── subdomains.txt      # Embedded wordlist
│   ├── service/
│   │   ├── asset_service.go        # Asset CRUD (Session 4)
│   │   └── scan_service.go         # Scan orchestration (NEW!)
│   ├── storage/
│   │   ├── storage.go              # Storage interfaces
│   │   └── postgres/
│   │       └── postgres.go         # PostgreSQL implementation
│   ├── handler/
│   │   ├── asset_handler.go        # Asset endpoints (Session 4)
│   │   ├── scan_handler.go         # Scan endpoints (NEW!)
│   │   └── health_handler.go       # Health check
│   └── validator/
│       └── asset_validator.go      # Input validation (Session 4)
├── migrations/
│   ├── 001_create_assets.up.sql    # Asset table (Session 3)
│   ├── 001_create_assets.down.sql
│   ├── 002_create_scan_tables.up.sql   # Scan tables (NEW!)
│   └── 002_create_scan_tables.down.sql
├── SCANNING_ARCHITECTURE.md        # Detailed scanning docs (NEW!)
├── README.md                        # This file
├── docker-compose.yml
├── .env.example
└── go.mod
```

---

## Implemented Scanners

### 1. DNS Scanner (`dns_scanner.go`)

**Purpose:** Discover IP addresses and DNS configuration

**Record Types:**

- **A** - IPv4 addresses
- **AAAA** - IPv6 addresses
- **CNAME** - Canonical name aliases
- **MX** - Mail server records
- **NS** - Nameserver records
- **TXT** - Text records (SPF, DKIM, verification)

**Demo:**

```bash
curl -X POST http://localhost:8080/assets/$ASSET_ID/scan \
  -H "Content-Type: application/json" \
  -d '{"scan_type": "dns"}'
```

**Teaching Points:**

- Go's `net` package (standard library)
- DNS record types and their purposes
- IP extraction for further scanning
- Security implications (reveals infrastructure)

---

### 2. WHOIS Scanner (`whois_scanner.go`)

**Purpose:** Get domain registration information

**Discovered Information:**

- Registrar name
- Registration date
- Expiration date ⚠️ (security critical!)
- Nameservers
- Domain status
- Contact emails

**Protocol:**

- TCP connection to port 43
- Plain text query/response
- TLD-specific servers

**Demo:**

```bash
curl -X POST http://localhost:8080/assets/$ASSET_ID/scan \
  -H "Content-Type: application/json" \
  -d '{"scan_type": "whois"}'
```

**Teaching Points:**

- WHOIS protocol (RFC 3912)
- Text parsing challenges
- Date format handling
- Expiration monitoring importance

---

### 3. Subdomain Scanner (`subdomain_scanner.go`)

**Purpose:** Discover all subdomains of a domain

**Method:** DNS Bruteforce

1. Load wordlist (common subdomain names)
2. For each word: tryquery `{word}.{domain}`
3. If resolves → subdomain found!

**Concurrency:**

- 50 concurrent workers (goroutines)
- Rate limiting: 100 queries/second
- Context-based cancellation
- Timeout: 5 minutes

**Demo:**

```bash
curl -X POST http://localhost:8080/assets/$ASSET_ID/scan \
  -H "Content-Type: application/json" \
  -d '{"scan_type": "subdomain"}'

# Takes longer (trying 100+ names)
# Poll for status:
curl http://localhost:8080/scan-jobs/$JOB_ID
```

**Teaching Points:**

- Worker pool pattern
- Rate limiting (token bucket)
- Context for cancellation
- Go embed for wordlists

---

## API Endpoints Reference

### Asset Management (Session 2-4)

| Method | Endpoint       | Description                           |
| ------ | -------------- | ------------------------------------- |
| POST   | `/assets`      | Create asset                          |
| GET    | `/assets`      | List assets (with pagination/filters) |
| GET    | `/assets/{id}` | Get single asset                      |
| PUT    | `/assets/{id}` | Update asset                          |
| DELETE | `/assets/{id}` | Delete asset                          |

### Scan Operations (Session 5 - NEW!)

| Method | Endpoint                  | Description              | Query Params                               | Status Code  |
| ------ | ------------------------- | ------------------------ | ------------------------------------------ | ------------ |
| POST   | `/assets/{id}/scan`       | Start scan (async)       | -                                          | 202 Accepted |
| POST   | `/assets/{id}/scan/demo`  | 🎓 Demo: Sync vs Async comparison | -                               | 200 OK       |
| GET    | `/assets/{id}/scans`      | List all scans for asset | `page`, `page_size`, `scan_type`, `status` | 200 OK       |
| GET    | `/scan-jobs/{id}`         | Get scan job status      | -                                          | 200 OK       |
| GET    | `/scan-jobs/{id}/results` | Get scan results         | -                                          | 200 OK       |

**Demo Endpoint (Teaching Tool):**
- The `/assets/{id}/scan/demo` endpoint runs the same scans twice - once synchronously (sequential) and once asynchronously (concurrent with goroutines)
- Performance comparison is printed to server console
- Demonstrates the speed advantage of async I/O operations (typically 2-3x faster)
- See [test_sync_vs_async.md](test_sync_vs_async.md) for detailed guide

**Pagination & Filtering:**

- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 20)
- `scan_type`: Filter by scan type (optional)
- `status`: Filter by status: `pending`, `running`, `completed`, `failed` (optional)

**Example:** `GET /assets/123/scans?page=1&page_size=10&status=completed`

### Scan Results by Asset

| Method | Endpoint                  | Description                                        | Query Params                       |
| ------ | ------------------------- | -------------------------------------------------- | ---------------------------------- |
| GET    | `/assets/{id}/results`    | **All scan results** (DNS + WHOIS + Subdomains) 🆕 | -                                  |
| GET    | `/assets/{id}/dns`        | DNS records only (paginated)                       | `page`, `page_size`, `record_type` |
| GET    | `/assets/{id}/whois`      | WHOIS information only                             | -                                  |
| GET    | `/assets/{id}/subdomains` | Discovered subdomains only (paginated)             | `page`, `page_size`                |

**Pagination:**

- All list endpoints support pagination via `page` and `page_size` query parameters
- Response includes metadata: `total`, `page`, `page_size`, `total_pages`

**Filtering:**

- DNS endpoint: Filter by `record_type` (A, AAAA, MX, NS, TXT, CNAME)
- Example: `GET /assets/123/dns?page=1&page_size=20&record_type=A`

**Paginated Response Format:**

```json
{
  "data": [...],
  "total": 150,
  "page": 1,
  "page_size": 20,
  "total_pages": 8
}
```

**Note:** `/assets/{id}/results` returns combined view of all scan data for the asset, regardless of which scan jobs created them. Perfect for getting complete reconnaissance data in one request. This endpoint returns all data without pagination.

### Scan Types

**Available Scan Types:**

| Scan Type   | Category   | Description                                               | Time   |
| ----------- | ---------- | --------------------------------------------------------- | ------ |
| `all`       | 🔵 Passive | **Run all passive scans** (DNS + WHOIS + Subdomain)       | ~5-10s |
| `dns`       | 🔵 Passive | DNS record scanning (A, AAAA, MX, NS, TXT, CNAME)         | ~1s    |
| `whois`     | 🔵 Passive | WHOIS domain registration information                     | ~2s    |
| `subdomain` | 🔵 Passive | Subdomain enumeration via DNS bruteforce                  | ~3-5s  |
| `port`      | 🔴 Active  | Port scanning (⚠️ Student exercise - requires permission) | varies |
| `ssl`       | 🔴 Active  | SSL certificate probing (⚠️ Student exercise)             | varies |
| `asn`       | 🔵 Passive | ASN and IP range lookup (Student exercise)                | ~1s    |

**Recommended Usage:**

- For comprehensive reconnaissance: Use `"scan_type": "all"`
- For specific information: Use individual scan types
- For active scans: Ensure you have written authorization

---

## Teaching Flow

### Part 1: Introduction (30 minutes)

**1.1 EASM Concepts (15 min)**

- What is attack surface?
- External vs internal assets
- Why continuous discovery matters
- Real-world breach examples (subdomain takeover, etc.)

**1.2 Session 5 Goals (15 min)**

- Implement 3 scan types
- Learn async patterns
- Practice external API integration
- Students complete 3 more scan types

**Demo:** Show finished system in action
-'Start scan → poll status → get results

---

### Part 2: Async Job Pattern (45 minutes)

**2.1 Why Async? (10 min)**

Problem:

```go
// Synchronous (bad for web)
func ScanDNS(domain string) ([]*DNSRecord, error) {
    // Takes 5-10 seconds
    records := performScan(domain)  // Blocks!
    return records, nil
}
// HTTP request times out!
```

Solution:

```go
// Asynchronous (good)
func StartScan(domain string) (*ScanJob, error) {
    job := CreateJob()
    job.Status = "pending"
    SaveJob(job)

    go performScan(job)  // Background!

    return job, nil  // Returns immediately
}
// HTTP returns 202 Accepted with job ID
```

**2.2 Job/Task Pattern (15 min)**

```
Client                  Server
  │                       │
  ├─ POST /scan ─────────►│
  │                       ├─ Create job (pending)
  │◄───── 202 + job ID ───┤
  │                       ├─ Start goroutine
  │                       │   └─► Scan (running)
  │                       │
  ├─ GET /job/{id} ──────►│
  │◄───── Status: running ┤
  │                       │
  ├─ GET /job/{id} ──────►│   Scan complete!
  │◄───── Status: completed┤
  │                       │
  ├─ GET /job/{id}/results►│
  │◄───── [results] ───────┤
```

**Teaching Activity:**

- Walk through `scan_service.go`
- Trace async flow
- Show how status updates work

**2.3 Code Walkthrough: scan_service.go (20 min)**

Key methods:

```go
func (s *ScanService) StartScan(...) (*ScanJob, error) {
    // 1. Validate
    // 2. Create job
    // 3. Save to database
    // 4. Launch goroutine
    go s.performScan(asset, job)
    // 5. Return immediately
    return job, nil
}

func (s *ScanService) performScan(...) {
    // Update: pending → running
    // Perform scan
    // Update: running → completed/failed
    // Never returns anything (background)
}
```

**Demo:**

- Start scan
- Show database: status changes
- Poll endpoint
- Get results

---

### Part 3: DNS Scanner (30 minutes)

**3.1 DNS Basics (10 min)**

DNS = Internet's phonebook

- Domain → IP address
- Multiple record types
- Critical for internet function

**3.2 Code Walkthrough: dns_scanner.go (15 min)**

```go
// Go's net package
ips, err := net.LookupIP("example.com")
mxs, err := net.LookupMX("example.com")
txts, err := net.LookupTXT("example.com")

// Very simple!
// Cross-platform
// No dependencies
```

Walk through each record type:

- A/AAAA - Addresses
- MX - Mail
- NS - Nameservers
- TXT - Verification

**3.3 Live Demo (5 min)**

```bash
# DNS scan
curl -X POST .../scan -d '{"scan_type": "dns"}'

# Show results
curl .../scan-jobs/{id}/results

# Point out in database
docker exec -it mini-asm-db psql -U postgres -d mini_asm
SELECT * FROM dns_records;
```

---

### Part 4: WHOIS Scanner (30 minutes)

**4.1 WHOIS Protocol (10 min)**

- Very simple: TCP port 43
- Send domain + CRLF
- Get plain text response
- No standard format (parsing hard!)

**4.2 Code Walkthrough: whois_scanner.go (15 min)**

```go
// Connect to WHOIS server
conn, err := net.Dial("tcp", "whois.verisign-grs.com:43")

// Send query
fmt.Fprintf(conn, "example.com\r\n")

// Read response (text)
scanner := bufio.NewScanner(conn)
for scanner.Scan() {
    line := scanner.Text()
    // Parse line by line
}
```

Parsing challenges:

- No standard format
- Multiple date formats
- Have to guess based on keywords

**4.3 Live Demo (5 min)**

```bash
# WHOIS scan
curl -X POST .../scan -d '{"scan_type": "whois"}'

# Check expiry date!
curl .../assets/{id}/whois | grep expiry_date
```

---

### Part 5: Subdomain Scanner (45 minutes)

**5.1 Why Subdomain Enumeration? (10 min)**

Security implications:

- dev.example.com → development server (weak security?)
- admin.example.com → admin panel (default password?)
- old.example.com → abandoned (unpatched?)

Real breaches start here!

**5.2 DNS Bruteforce Method (15 min)**

```
Wordlist: www, api, mail, dev, ...

For each word:
    domain = word + ".example.com"
    if DNS_lookup(domain) succeeds:
        Found subdomain!
```

Why concurrent?

```
Sequential: 100 words × 1 second = 100 seconds
Concurrent (50 workers): 100 words / 50 = 2 seconds
```

**5.3 Code Walkthrough: subdomain_scanner.go (15 min)**

Key patterns:

1. **Worker pool:**

   ```go
   jobs := make(chan string, len(wordlist))

   for i := 0; i < 50; i++ {
       go func() {
           for word := range jobs {
               // Scan word.domain.com
           }
       }()
   }
   ```

2. **Rate limiting:**

   ```go
   limiter := time.NewTicker(time.Second / 100)
   <-limiter.C  // Wait for token
   ```

3. **Context cancellation:**

   ```go
   ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
   defer cancel()

   select {
   case <-ctx.Done():
       return  // Stop!
   default:
       // Continue
   }
   ```

**5.4 Live Demo (5 min)**

```bash
# Subdomain scan (slow)
curl -X POST .../scan -d '{"scan_type": "subdomain"}'

# Poll status (watch results count increase)
while true; do
    curl .../scan-jobs/{id} | jq '.status, .results'
    sleep 2
done

# Get results
curl .../assets/{id}/subdomains
```

---

### Part 6: Database Schema (30 minutes)

**6.1 Table Overview (10 min)**

```sql
assets (parent)
  ├── scan_jobs (track scans)
  ├── dns_records (DNS results)
  ├── whois_records (WHOIS results)
  └── subdomains (subdomain results)
```

**6.2 Foreign Keys & CASCADE (10 min)**

```sql
FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE
```

Demonstration:

```sql
-- Create asset
INSERT INTO assets ...;

-- Create scan
INSERT INTO scan_jobs ...;

-- Create results
INSERT INTO dns_records ...;

-- Delete asset
DELETE FROM assets WHERE id = '...';

-- All related data gone!
SELECT COUNT(*) FROM scan_jobs WHERE asset_id = '...';  -- 0
SELECT COUNT(*) FROM dns_records WHERE asset_id = '...';  -- 0
```

**6.3 UPSERT Pattern (10 min)**

```sql
INSERT INTO subdomains (...)
VALUES (...)
ON CONFLICT (asset_id, name) DO UPDATE SET
    scan_job_id = EXCLUDED.scan_job_id;
```

Why?

- Subdomain found multiple times
- Don't create duplicate
- Update with latest scan info

---

### Part 7: Demo & Testing (30 minutes)

**Full Workflow Demo:**

```bash
# 1. Create domain
ASSET=$(curl -X POST .../assets \
  -d '{"name":"example.com","type":"domain"}' \
  | jq -r '.id')

# 2. Run all scans
curl -X POST .../assets/$ASSET/scan -d '{"scan_type":"dns"}'
curl -X POST .../assets/$ASSET/scan -d '{"scan_type":"whois"}'
curl -X POST .../assets/$ASSET/scan -d '{"scan_type":"subdomain"}'

# 3. Check all scans
curl .../assets/$ASSET/scans | jq

# 4. Get all results
curl .../assets/$ASSET/dns | jq
curl .../assets/$ASSET/whois | jq
curl .../assets/$ASSET/subdomains | jq

# 5. Show in database
docker exec -it mini-asm-db psql -U postgres -d mini_asm
\dt  -- Show tables
SELECT scan_type, status, results FROM scan_jobs;
SELECT record_type, COUNT(*) FROM dns_records GROUP BY record_type;
SELECT COUNT(*) FROM subdomains WHERE is_active = true;
```

---

## Student Exercises

### Exercise 1: Port Scanner (Intermediate)

**Goal:** Discover open ports and services on IP addresses

**Implementation Hints:**

```go
// internal/scanner/port_scanner.go

type PortScanner struct {
    commonPorts []int
    timeout     time.Duration
}

func (s *PortScanner) Scan(asset *model.Asset) ([]*model.PortResult, error) {
    // asset.Type should be "ip"

    for _, port := range s.commonPorts {
        addr := fmt.Sprintf("%s:%d", asset.Name, port)

        conn, err := net.DialTimeout("tcp", addr, s.timeout)
        if err == nil {
            // Port open!
            conn.Close()

            // Try to detect service (read banner)
            service := detectService(port, readBanner(conn))

            results = append(results, &PortResult{
                Port:    port,
                Service: service,
                State:   "open",
            })
        }
    }

    return results, nil
}

func detectService(port int, banner string) string {
    // Match banner against known patterns
    if strings.Contains(banner, "SSH") {
        return "ssh"
    }
    // ... more detection logic

    // Fallback to common services
    commonServices := map[int]string{
        80:   "http",
        443:  "https",
        22:   "ssh",
        // ...
    }
    return commonServices[port]
}
```

**Database Migration:**

```sql
-- migrations/003_add_port_scan.up.sql

CREATE TABLE IF NOT EXISTS port_results (
    id UUID PRIMARY KEY,
    asset_id UUID NOT NULL,
    scan_job_id UUID NOT NULL,
    port INTEGER NOT NULL,
    service VARCHAR(50),
    state VARCHAR(20) NOT NULL,
    banner TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_port_result_asset
        FOREIGN KEY (asset_id)
        REFERENCES assets(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_port_result_scan_job
        FOREIGN KEY (scan_job_id)
        REFERENCES scan_jobs(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_port_results_asset_id ON port_results(asset_id);
CREATE INDEX idx_port_results_port ON port_results(port);
```

**Testing:**

```bash
# Create IP asset (from DNS scan results)
curl -X POST .../assets \
  -d '{"name":"93.184.216.34","type":"ip"}'

# Port scan
curl -X POST .../assets/{ip-asset-id}/scan \
  -d '{"scan_type":"port"}'
```

---

### Exercise 2: SSL/TLS Certificate Scanner (Intermediate)

**Goal:** Analyze SSL certificates for security and discovery

**Implementation Hints:**

```go
// internal/scanner/ssl_scanner.go

type SSLScanner struct {
    timeout time.Duration
}

func (s *SSLScanner) Scan(asset *model.Asset) (*model.SSLRecord, error) {
    // Connect to port 443
    conn, err := tls.Dial("tcp", asset.Name+":443", &tls.Config{
        InsecureSkipVerify: true, // For analysis, not production!
    })
    if err != nil {
        return nil, err
    }
    defer conn.Close()

    // Get certificate
    cert := conn.ConnectionState().PeerCertificates[0]

    return &model.SSLRecord{
        CommonName:   cert.Subject.CommonName,
        SANs:         cert.DNSNames,  // Subject Alternative Names
        Issuer:       cert.Issuer.CommonName,
        NotBefore:    cert.NotBefore,
        NotAfter:     cert.NotAfter,
        IsExpired:    time.Now().After(cert.NotAfter),
        IsSelfSigned: cert.Issuer.CommonName == cert.Subject.CommonName,
    }, nil
}
```

---

### Exercise 3: Certificate Transparency (Advanced)

**Goal:** Discover subdomains via CT logs

**Implementation Hints:**

```go
// internal/scanner/ct_scanner.go

func (s *CTScanner) Scan(asset *model.Asset) ([]*model.Subdomain, error) {
    // Query crt.sh API
    url := fmt.Sprintf("https://crt.sh/?q=%%.%s&output=json", asset.Name)

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var entries []struct {
        NameValue string `json:"name_value"`
    }

    json.NewDecoder(resp.Body).Decode(&entries)

    // Extract unique subdomains
    subdomains := []string{}
    seen := make(map[string]bool)

    for _, entry := range entries {
        names := strings.Split(entry.NameValue, "\n")
        for _, name := range names {
            name = strings.TrimSpace(name)
            if !seen[name] && strings.HasSuffix(name, asset.Name) {
                seen[name] = true

                // Verify it resolves
                if _, err := net.LookupIP(name); err == nil {
                    subdomains = append(subdomains, &Subdomain{
                        Name:   name,
                        Source: "certificate_transparency",
                    })
                }
            }
        }
    }

    return subdomains, nil
}
```

---

### Exercise 4: Recursive Scanning (Advanced)

**Goal:** Automatically scan discovered subdomains

**Implementation:**

```go
// internal/service/scan_service.go

func (s *ScanService) ScanRecursively(assetID string) error {
    // 1. Run DNS scan
    dnsJob := s.StartScan(assetID, ScanTypeDNS)

    // 2. Wait for completion
    s.waitForJob(dnsJob.ID)

    // 3. Extract IPs, create IP assets
    records := s.GetDNSRecordsByScan(dnsJob.ID)
    ips := extractIPs(records)

    for _, ip := range ips {
        ipAsset := createIPAsset(ip)
        // 4. Port scan each IP
        s.StartScan(ipAsset.ID, ScanTypePort)
    }

    // 5. Run subdomain scan
    subdomainJob := s.StartScan(assetID, ScanTypeSubdomain)

    // 6. Wait for subdomains
    s.waitForJob(subdomainJob.ID)

    // 7. Scan each subdomain (DNS + SSL)
    subdomains := s.GetSubdomainsByScan(subdomainJob.ID)

    for _, subdomain := range subdomains {
        subAsset := createDomainAsset(subdomain.Name)
        s.StartScan(subAsset.ID, ScanTypeDNS)
        s.StartScan(subAsset.ID, ScanTypeSSL)
    }

    return nil
}
```

---

## Troubleshooting

### Database Connection Errors

```
Error: failed to connect to database: connection refused
```

**Solution:**

```bash
# Check Docker container running
docker-compose ps

# If not running
docker-compose up -d

# Check logs
docker-compose logs db

# Verify connection
docker exec -it mini-asm-db psql -U postgres -d mini_asm -c "\dt"
```

---

### Migrations Not Applied

```
Error: relation "scan_jobs" does not exist
```

**Solution:**

```bash
# Migrations auto-run on first start
# Force re-run:
docker-compose down -v  # Delete volumes
docker-compose up -d    # Migrations run again

# Or manually:
docker exec -it mini-asm-db psql -U postgres -d mini_asm < migrations/002_create_scan_tables.up.sql
```

---

### WHOIS Server Errors

```
Error: WHOIS query failed: connection timeout
```

**Possible Causes:**

1. **Firewall blocks port 43**
   - Check corporate firewall
   - Try different network

2. **WHOIS server down**
   - Happens occasionally
   - Try different domain/TLD

3. **Rate limited**
   - WHOIS servers have strict limits
   - Wait a few minutes

**Solution:**

```go
// Use longer timeout
scanner := NewWHOISScanner()
scanner.timeout = 30 * time.Second

// Or implement fallback to RDAP (HTTP-based)
```

---

### DNS Lookup Failures

```
Error: lookup example.com: no such host
```

**Causes:**

1. Domain doesn't exist
2. DNS server slow/unreachable
3. Network issues

**Solutions:**

```bash
# Test DNS manually
nslookup example.com
dig example.com

# Try different DNS (in code)
resolver := &net.Resolver{
    PreferGo: true,
    Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
        d := net.Dialer{Timeout: 10 * time.Second}
        return d.DialContext(ctx, network, "8.8.8.8:53") // Google DNS
    },
}
```

---

### Subdomain Scan Stuck

```
Status: running for 10+ minutes
```

**Causes:**

1. **Slow DNS responses**
2. **Timeout too long**
3. **Goroutine leak**

**Solutions:**

```go
// Reduce timeout
lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)  // Was 5s

// Add max scan duration
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)  // Was 5 min

// Check goroutines
log.Printf("Active goroutines: %d", runtime.NumGoroutine())
```

---

### Too Many Open Files

```
Error: too many open files
```

**Cause:** Scanner opens many network connections

**Solutions:**

1. **Limit concurrency:**

   ```go
   maxWorkers = 10  // Was 50
   ```

2. **Increase OS limit (Linux/Mac):**

   ```bash
   ulimit -n 4096
   ```

3. **Ensure connections close:**
   ```go
   defer conn.Close()  // Always defer!
   ```

---

## Next Steps

### After Session 5

**Students should be able to:**

- ✅ Explain EASM concepts
- ✅ Implement async job patterns
- ✅ Integrate external services (DNS, WHOIS)
- ✅ Use goroutines for background work
- ✅ Design multi-table database schemas
- ✅ Build scanner abstractions

**Continue to Session 6:** Deployment & Frontend

- Docker containerization
- Frontend dashboard
- Visualization of scan results
- Deployment strategies

### Practice Projects

1. **Build Complete EASM Tool**
   - Implement all 6 scan types
   - Add scheduling (cron-like)
   - Webhook notifications
   - Web dashboard

2. **Security Scanner**
   - Known vulnerability detection
   - Weak SSL configurations
   - Open database ports
   - Default credentials checking

3. **Asset Inventory System**
   - Automatic asset discovery
   - Change detection
   - Compliance reporting
   - Integration with cloud APIs

---

## Additional Resources

### Official Documentation

- [Go net package](https://pkg.go.dev/net)
- [RFC 1035 - DNS](https://www.rfc-editor.org/rfc/rfc1035)
- [RFC 3912 - WHOIS](https://www.rfc-editor.org/rfc/rfc3912)
- [Certificate Transparency](https://certificate.transparency.dev/)

### EASM Tools for Reference

- [Amass](https://github.com/OWASP/Amass) - OWASP subdomain enumeration
- [Subfinder](https://github.com/projectdiscovery/subfinder) - Fast subdomain discovery
- [Nuclei](https://github.com/projectdiscovery/nuclei) - Vulnerability scanner
- [Nmap](https://nmap.org/) - Port scanning

### Learning Resources

- [Concurrency in Go (Book)](https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/)
- [PostgreSQL Foreign Keys](https://www.postgresql.org/docs/current/tutorial-fk.html)
- [REST API Design Best Practices](https://restfulapi.net/)

---

## Summary

Session 5 introduces real-world security concepts through practical implementation:

1. **EASM fundamentals** - Why attack surface matters
2. **Async patterns** - Background jobs with goroutines
3. **External integration** - DNS, WHOIS protocols
4. **Scanner architecture** - Extensible design
5. **Database relationships** - Foreign keys, CASCADE
6. **Production patterns** - Rate limiting, timeouts

**Key Takeaway:** Security tools are just well-designed software systems applying fundamental CS concepts (concurrency, networking, databases) to security problems!

---

**🎓 Good luck with Session 5! Start discovering that attack surface! 🔍**
