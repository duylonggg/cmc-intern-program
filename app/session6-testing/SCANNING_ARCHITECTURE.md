# 🔍 EASM Scanning Architecture

## High-Level Scanning Flows

This document describes the complete scanning flows for External Attack Surface Management (EASM). These flows show how to discover and enumerate assets from a starting domain.

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Passive vs Active Reconnaissance](#passive-vs-active-reconnaissance)
3. [Complete Discovery Flow](#complete-discovery-flow)
4. [Individual Scan Types](#individual-scan-types)
5. [Implementation Status](#implementation-status)
6. [Architecture Patterns](#architecture-patterns)

---

## Overview

### What is EASM?

External Attack Surface Management helps organizations discover and monitor their external-facing assets:

- **Domains & Subdomains** - What domains do we own?
- **DNS Records** - Where do they point?
- **WHOIS Information** - Who owns them? When do they expire?
- **IP Addresses** - What IPs are associated?
- **Open Ports** - What services are running?
- **SSL Certificates** - Are they valid and secure?
- **ASN Information** - What networks do we own?

### Why Important?

1. **Security Visibility** - Know what attackers can see
2. **Shadow IT Discovery** - Find forgotten/unknown assets
3. **Compliance** - Maintain asset inventory
4. **Incident Response** - Quickly understand exposure

---

## Passive vs Active Reconnaissance

### 🔵 Passive Scanning (OSINT)

**Definition:** Intelligence gathering using publicly available information without directly interacting with target systems.

**Characteristics:**

- ✅ **Legal** - No authorization required, uses public data
- ✅ **Stealthy** - Target is unaware of reconnaissance
- ✅ **Safe** - Cannot harm or disrupt target systems
- ✅ **Continuous** - Can run 24/7 without concerns

**Passive Scans in This Project:**

| Scan Type                    | Data Source                                   | Risk Level |
| ---------------------------- | --------------------------------------------- | ---------- |
| **WHOIS**                    | Public WHOIS database via TCP port 43         | 🟢 None    |
| **DNS**                      | Public DNS servers (Google, Cloudflare, etc.) | 🟢 None    |
| **Subdomain Enumeration**    | DNS queries + Certificate Transparency        | 🟢 None    |
| **Certificate Transparency** | Public CT logs (crt.sh API)                   | 🟢 None    |
| **ASN Lookup**               | Public WHOIS/BGP routing databases            | 🟢 None    |

**Technical Details:**

```
WHOIS Query Flow:
1. Query WHOIS server (passive)
   └─► whois.iana.org:43 or TLD-specific server
2. Receive publicly available registration data
3. No direct contact with target systems

DNS Query Flow:
1. Query public DNS resolver (passive)
   └─► 8.8.8.8 (Google) or 1.1.1.1 (Cloudflare)
2. Resolver queries authoritative nameservers
3. Target sees normal DNS traffic (happens constantly)
4. No direct probing of target infrastructure
```

**Why Passive is Safe:**

- DNS queries happen billions of times per day (normal traffic)
- WHOIS lookups are designed to be public
- Certificate Transparency logs are intentionally public
- Cannot crash services or trigger alerts

---

### 🔴 Active Scanning (Direct Probing)

**Definition:** Directly interacting with target systems by sending network packets, connection attempts, or probes.

**Characteristics:**

- ⚠️ **Legal Risk** - May require authorization (CFAA, Computer Misuse Act)
- ⚠️ **Detectable** - Target will log your IP and activity
- ⚠️ **Potentially Disruptive** - Can crash services or trigger alerts
- ⚠️ **Compliance** - Requires approval, scheduling, notification

**Active Scans in This Project:**

| Scan Type                   | Target System  | Risk Level | Why It's Active                    |
| --------------------------- | -------------- | ---------- | ---------------------------------- |
| **Port Scanning**           | Target IP/host | 🔴 High    | Direct TCP/UDP connection attempts |
| **Service Detection**       | Open ports     | 🔴 High    | Banner grabbing, version probing   |
| **SSL Certificate Probing** | HTTPS services | 🟡 Medium  | TLS handshake with target server   |
| **Vulnerability Scanning**  | Services/apps  | 🔴 High    | Testing for exploitable weaknesses |

**Technical Details:**

```
Port Scan Flow:
1. Send SYN packet to target:port (active probe)
   └─► 192.168.1.1:80, 192.168.1.1:443, ...
2. Receive SYN-ACK (port open) or RST (closed)
3. Target logs connection attempt:
   └─► Firewall logs
   └─► IDS/IPS alerts
   └─► Service access logs

Service Detection Flow:
1. Connect to open port (active probe)
2. Send service-specific probes:
   └─► HTTP: GET / HTTP/1.1
   └─► SSH: Read banner
   └─► SMTP: HELO command
3. Parse response for version/product
4. Target logs all activity
```

**Legal Risks:**

```
❌ Computer Fraud and Abuse Act (CFAA) - US
   "Accessing a computer without authorization"
   → Port scanning without permission may qualify

❌ Computer Misuse Act - UK
   "Unauthorized access to computer material"
   → Includes reconnaissance activities

❌ Local Laws
   → Many countries have similar legislation
```

**When Active Scanning is Legal:**

- ✅ You own the target systems
- ✅ Written authorization from owner
- ✅ Penetration testing contract
- ✅ Bug bounty program (within scope)
- ✅ Authorized security assessment

**When Active Scanning is Illegal:**

- ❌ No permission from target owner
- ❌ "Just testing" third-party systems
- ❌ Beyond authorized scope
- ❌ No documented approval

---

### 📊 Scan Classification Summary

```
┌─────────────────────────────────────────────────────────────┐
│                    PASSIVE SCANS                             │
│                  (Safe, Legal, OSINT)                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐  ┌─────────┐ │
│  │  WHOIS   │  │   DNS    │  │  Subdomain   │  │ Cert CT │ │
│  │  Lookup  │  │  Query   │  │ Enumeration  │  │  Logs   │ │
│  └──────────┘  └──────────┘  └──────────────┘  └─────────┘ │
│                                                              │
│  • Public databases                                          │
│  • No target interaction                                     │
│  • Cannot harm systems                                       │
│  • Run anytime                                              │
│                                                              │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    ACTIVE SCANS                              │
│              (⚠️ Requires Permission)                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐  ┌─────────┐ │
│  │   Port   │  │ Service  │  │     SSL      │  │  Vuln   │ │
│  │   Scan   │  │ Detect   │  │   Probing    │  │  Scan   │ │
│  └──────────┘  └──────────┘  └──────────────┘  └─────────┘ │
│                                                              │
│  • Direct system probing                                     │
│  • Logged by target                                          │
│  • Can trigger alerts                                        │
│  • Requires authorization                                    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Complete Discovery Flow

### Starting Point: Domain Asset

```
Input: example.com (domain asset)
Goal: Discover all related assets and information
```

### Step-by-Step Flow

```
┌─────────────────────────────────────────────────────────────┐
│ 1. DOMAIN ASSET (example.com)                               │
│    - User creates or selects existing domain asset          │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ├──────────────────────────────────────┐
                   │                                      │
                   ▼                                      ▼
┌──────────────────────────────────┐    ┌──────────────────────────────────┐
│ 2. WHOIS SCAN                    │    │ 3. DNS SCAN                      │
│                                  │    │                                  │
│ Discovers:                       │    │ Queries:                         │
│ ✓ Registrar                      │    │ ✓ A records → IP addresses       │
│ ✓ Registration date              │    │ ✓ AAAA records → IPv6 addresses  │
│ ✓ Expiration date ⚠️              │    │ ✓ MX records → Mail servers      │
│ ✓ Name servers                   │    │ ✓ NS records → Name servers      │
│ ✓ Status                         │    │ ✓ TXT records → Verification     │
│ ✓ Contact emails                 │    │ ✓ CNAME records → Aliases        │
│                                  │    │ ✓ SOA record → Zone info         │
│ Creates: WHOISRecord             │    │                                  │
│                                  │    │ Creates: Multiple DNSRecord      │
└──────────────────────────────────┘    └──────────┬───────────────────────┘
                                                    │
                                                    │ Extract IPs from A/AAAA
                                                    ▼
                                    ┌──────────────────────────────────┐
                                    │ 4. CREATE IP ASSETS              │
                                    │    (Automatic or Manual)         │
                                    │                                  │
                                    │ For each unique IP:              │
                                    │ - Create new Asset (type: ip)    │
                                    │ - Link to parent domain          │
                                    └──────────┬───────────────────────┘
                                               │
        ┌──────────────────────────────────────┼────────────────────┐
        │                                      │                    │
        ▼                                      ▼                    ▼
┌─────────────────┐              ┌──────────────────────┐  ┌──────────────────┐
│ 5. SUBDOMAIN    │              │ 6. PORT SCAN         │  │ 7. ASN LOOKUP    │
│    ENUMERATION  │              │    (Per IP)          │  │    (Per IP)      │
│                 │              │                      │  │                  │
│ Methods:        │              │ Discovers:           │  │ Discovers:       │
│ ✓ DNS Bruteforce│              │ ✓ Open ports         │  │ ✓ AS Number      │
│ ✓ Certificate   │              │ ✓ Services           │  │ ✓ Organization   │
│   Transparency  │              │ ✓ Service versions   │  │ ✓ IP range       │
│ ✓ Web Scraping  │              │ ✓ Banners            │  │ ✓ Network info   │
│ ✓ Archive.org   │              │                      │  │                  │
│                 │              │ For each open port:  │  │ Creates:         │
│ Creates:        │              │ - Create Asset       │  │ - ASN records    │
│ - Subdomain     │              │   (type: service)    │  │ - IP ranges      │
│   records       │              │                      │  │                  │
│                 │              │ Creates:             │  │                  │
│ Each subdomain  │              │ - Service assets     │  │                  │
│ becomes new     │              │ - Port records       │  │                  │
│ Asset → recurse │              └──────────────────────┘  └──────────────────┘
└────────┬────────┘
         │
         │ For each subdomain:
         ▼
┌─────────────────────────────┐
│ 8. RECURSIVE SCAN           │
│                             │
│ For subdomain (e.g.,        │
│ api.example.com):           │
│                             │
│ 1. Create Asset             │
│ 2. Run DNS Scan → IPs       │
│ 3. Run Port Scan on IPs     │
│ 4. Check SSL Certificate    │
│                             │
│ Continue until no new       │
│ assets discovered           │
└─────────────────────────────┘
```

### Data Flow Summary

```
Domain (example.com)
  │
  ├─► WHOIS Scan ─────► WHOISRecord (1)
  │
  ├─► DNS Scan ───────► DNSRecord (multiple)
  │                         │
  │                         └─► Extract IPs ─► IP Assets (multiple)
  │                                                │
  │                                                ├─► Port Scan ─► Service Assets
  │                                                │
  │                                                └─► ASN Lookup ─► ASN Records
  │
  └─► Subdomain Scan ─► Subdomain (multiple)
                            │
                            └─► Recursive: DNS → Port → SSL for each
```

---

## Individual Scan Types

### 1. WHOIS Scan

**Purpose:** Get domain registration information

**Input:** Domain name (e.g., "example.com")

**Process:**

```
1. Query WHOIS server (port 43)
2. Parse response:
   - Registrar name
   - Important dates (created, expires)
   - Name servers
   - Contact information
   - Domain status
3. Store parsed data + raw response
```

**Output:** WHOISRecord

**Key Use Cases:**

- ⚠️ **Expiration monitoring** - Alert when domain expires soon
- 🔍 **Infrastructure discovery** - Find name servers
- 📧 **Contact discovery** - Find admin emails (careful: PII)
- 📊 **Asset ownership** - Confirm registration details

**Implementation:**

- Simple: TCP connection to whois.iana.org or TLD-specific server
- Medium: Parse structured fields (regex patterns)
- Advanced: Handle multiple WHOIS formats (thin vs thick)

---

### 2. DNS Scan

**Purpose:** Discover IP addresses and DNS configuration

**Input:** Domain or subdomain name

**Process:**

```
1. Query DNS server for each record type:

   A Record:
   - Query: example.com
   - Response: 93.184.216.34
   - Creates: DNSRecord (type: A, value: IP)

   AAAA Record:
   - Query: example.com
   - Response: 2606:2800:220:1:248:1893:25c8:1946
   - Creates: DNSRecord (type: AAAA, value: IPv6)

   MX Record:
   - Query: example.com
   - Response: 10 mail.example.com
   - Finds: Mail servers

   NS Record:
   - Query: example.com
   - Response: ns1.example.com
   - Finds: Authoritative name servers

   TXT Record:
   - Query: example.com
   - Response: "v=spf1 include:_spf.google.com ~all"
   - Finds: SPF, DKIM, verification records

   CNAME Record:
   - Query: www.example.com
   - Response: example.com
   - Finds: Aliases

   SOA Record:
   - Query: example.com
   - Response: ns1.example.com hostmaster.example.com ...
   - Finds: Zone information

2. For each response:
   - Create DNSRecord
   - Extract IPs (from A/AAAA)
   - Extract additional domains (from MX/NS/CNAME)

3. Create IP assets for discovered IPs
```

**Output:** Multiple DNSRecord entries

**Key Use Cases:**

- 🌐 **IP discovery** - Find all IPs serving the domain
- 📧 **Mail server discovery** - Find email infrastructure
- 🔐 **SPF/DKIM verification** - Check email authentication
- 🌍 **CDN detection** - Identify CDN usage (CNAME patterns)
- 🔄 **Infrastructure changes** - Monitor DNS changes over time

**Implementation:**

- Simple: Use Go's `net` package (LookupIP, LookupMX, etc.)
- Medium: Query all record types, parse responses
- Advanced: Follow CNAME chains, detect DNS security (DNSSEC)

---

### 3. Subdomain Enumeration

**Purpose:** Discover all subdomains of a domain

**Input:** Domain name (e.g., "example.com")

**Methods:**

#### Method 1: DNS Bruteforce

```
1. Load wordlist:
   common_subdomains.txt:
   - www
   - api
   - mail
   - dev
   - staging
   - admin
   ...

2. For each word:
   - Construct: {word}.example.com
   - Query DNS (A record)
   - If exists → found subdomain!

3. Parallel processing:
   - Use goroutines for speed
   - Rate limiting (respect DNS servers)
   - Timeout handling

Result: Subdomains with IPs
```

#### Method 2: Certificate Transparency

```
1. Query CT logs (crt.sh, Facebook CT)
   - URL: https://crt.sh/?q=%.example.com&output=json
   - Returns: All certificates with *.example.com

2. Parse certificates:
   - Extract Subject Alternative Names (SANs)
   - Find: *.example.com, api.example.com, etc.

3. Verify subdomains:
   - DNS lookup to confirm active
   - Store source: "cert_transparency"

Result: Subdomains from certificates
```

#### Method 3: Web Scraping

```
1. Crawl main domain
2. Find links: <a href="https://subdomain.example.com">
3. Extract unique subdomains
4. Verify with DNS

Result: Subdomains found in HTML
```

#### Method 4: Search Engine Discovery

```
1. Query search engines:
   - Google: site:example.com
   - Bing: site:example.com

2. Parse search results
3. Extract unique subdomains

Result: Indexed subdomains
```

**Output:** Multiple Subdomain records

**Key Use Cases:**

- 🕵️ **Shadow IT discovery** - Find forgotten subdomains
- 🔓 **Attack surface expansion** - More assets = more risk
- 📊 **Asset inventory** - Complete subdomain catalog
- ⚠️ **Abandoned asset detection** - Find unmaintained subdomains

**Implementation Priority:**

- ✅ **Bruteforce** - Simple, reliable, implement first
- ✅ **Certificate Transparency** - High value, HTTP API calls
- 🔄 **Web Scraping** - Medium value, parsing complexity
- 🔄 **Search Engine** - Low priority, rate limiting issues

---

### 4. Port Scanning

**Purpose:** Discover open ports and services on IP addresses

**Input:** IP address

**Process:**

```
1. Scan common ports:
   Common ports:
   - 80 (HTTP)
   - 443 (HTTPS)
   - 22 (SSH)
   - 21 (FTP)
   - 25 (SMTP)
   - 3306 (MySQL)
   - 5432 (PostgreSQL)
   - 6379 (Redis)
   - 27017 (MongoDB)
   ... (top 1000 ports)

2. For each port:
   - Attempt TCP connection
   - Set timeout (e.g., 5 seconds)
   - If succeeds → port is open

3. Service detection:
   - Read banner (first bytes)
   - Match against known patterns
   - Example: "SSH-2.0-OpenSSH_8.2" → SSH service

4. Version detection:
   - Parse banner for version
   - Store service + version

5. Create Service asset:
   - Name: "192.168.1.1:443"
   - Type: "service"
   - Store: port, service, version
```

**Output:** Service assets, Port records

**Key Use Cases:**

- 🔓 **Exposed services** - Find unintended public services
- 🐛 **Vulnerable services** - Check for outdated versions
- 🔐 **Security audit** - Ensure only necessary ports open
- 📊 **Service inventory** - Know what's running where

**Implementation:**

- Simple: Basic TCP connect scan
- Medium: Use existing tools (nmap wrapper)
- Advanced: Service fingerprinting, OS detection

**⚠️ Legal Warning:**

- Port scanning can be considered hostile
- Only scan assets you own or have permission
- Respect rate limits to avoid DoS
- Consider legal implications

---

### 5. SSL/TLS Certificate Analysis

**Purpose:** Analyze SSL certificates for security and discovery

**Input:** Domain or IP + port (443)

**Process:**

```
1. Initiate TLS handshake
2. Retrieve certificate chain
3. Parse certificate:
   - Issuer (CA)
   - Subject (domain)
   - Valid from/to dates
   - SANs (alternative names)
   - Public key algorithm
   - Signature algorithm

4. Security checks:
   - ⚠️ Expired certificate
   - ⚠️ Self-signed certificate
   - ⚠️ Weak encryption (< 2048 bit)
   - ⚠️ Deprecated algorithms (MD5, SHA1)
   - ✅ Valid chain to trusted CA

5. Discovery:
   - Extract SANs → new subdomains
   - Check certificate transparency logs
```

**Output:** SSL records, potentially new subdomains

**Key Use Cases:**

- 🔐 **Security validation** - Ensure valid certificates
- ⚠️ **Expiration monitoring** - Alert before expiry
- 🕵️ **Subdomain discovery** - SANs reveal related domains
- 🏢 **Organization verification** - Confirm certificate owner

**Implementation:**

- Simple: Go's `tls` package to retrieve certificate
- Medium: Parse and validate certificate fields
- Advanced: Check CT logs, OCSP validation

---

### 6. ASN Lookup

**Purpose:** Discover network ownership and IP ranges

**Input:** IP address

**Process:**

```
1. Query WHOIS for IP:
   - Use ARIN/RIPE/APNIC database
   - Example: whois 93.184.216.34

2. Parse response:
   - AS Number (e.g., AS15133)
   - Organization name
   - IP range (CIDR)
   - Country/region

3. Query AS Number:
   - Find all IP ranges owned by this AS
   - Example: AS15133 owns 93.184.216.0/24

4. Store:
   - ASN record
   - IP ranges
   - Organization info
```

**Output:** ASN records, IP range data

**Key Use Cases:**

- 🌍 **Network ownership** - Find all IPs owned by organization
- 📊 **Asset grouping** - Group IPs by network
- 🔍 **Infrastructure discovery** - Find entire IP ranges
- 🏢 **Organization mapping** - Map assets to companies

**Implementation:**

- Simple: WHOIS query for IP
- Medium: Parse ASN data
- Advanced: Enumerate entire AS ranges

---

## Implementation Status

### ✅ Implemented (Session 5A - Passive Scanning)

**Category:** 🔵 Passive / OSINT  
**Safety:** ✅ Safe to practice on any domain  
**Authorization:** ❌ Not required

1. **WHOIS Scan** (Passive)
   - Basic WHOIS query (TCP port 43)
   - Response parsing (registrar, dates, name servers)
   - Error handling
   - **Why passive:** Queries public WHOIS database

2. **DNS Scan** (Passive)
   - Query A, AAAA, MX, NS, TXT, CNAME, SOA records
   - Use Go's `net` package
   - Store all record types
   - **Why passive:** Standard DNS queries via public resolvers

3. **Subdomain Enumeration** (Passive)
   - DNS bruteforce with wordlist
   - Concurrent scanning (goroutines)
   - Rate limiting
   - **Why passive:** Only DNS lookups, no direct target contact

### 🔄 Exercises for Students (Session 5B - Active Scanning)

**Category:** 🔴 Active / Intrusive  
**Safety:** ⚠️ ONLY scan localhost or authorized systems  
**Authorization:** ✅ REQUIRED for external systems

4. **Port Scanning** (🔴 Active - ⚠️ Legal Risk)
   - Implement TCP connect scan
   - Top 100 ports
   - Service detection from banners
   - Integration with nmap (bonus)
   - **Why active:** Direct TCP/UDP connection attempts to target
   - **⚠️ REQUIRES PERMISSION:** Scan localhost only in exercises

5. **SSL Certificate Analysis** (🟡 Semi-Active)
   - TLS handshake and certificate retrieval
   - Parse certificate fields
   - Expiration checking
   - SAN extraction for subdomain discovery
   - **Why semi-active:** Establishes TLS connection, but for public cert
   - **Note:** Can also use passive CT logs instead

6. **ASN Lookup** (🔵 Passive)
   - IP WHOIS query
   - ASN parsing
   - IP range discovery
   - Organization mapping
   - **Why passive:** Queries public WHOIS/BGP databases

### 🎯 Learning Objectives

Students will learn:

- **Passive vs Active reconnaissance** - Critical security concept
- **Legal and ethical considerations** - When authorization is needed
- How to structure complex scanning systems
- Async job processing patterns
- External API/service integration
- Data modeling for security tools
- Error handling in distributed operations
- Rate limiting and resource management

### ⚠️ Safety Guidelines for Student Exercises

**BEFORE starting active scanning exercises:**

1. **Legal Education:**
   - Explain Computer Fraud and Abuse Act (CFAA)
   - Show real-world legal consequences
   - Discuss ethical hacking principles

2. **Technical Restrictions:**
   - Require students to scan ONLY `127.0.0.1` (localhost)
   - Alternatively: Provide isolated VM environment
   - Block outbound scans at firewall level
   - Monitor network activity

3. **Documentation:**
   - Signed acknowledgment form
   - Clear scope definition
   - Incident response plan

**Example Safe Practice:**

```go
// Port scanner with safety checks
func (s *PortScanner) Scan(asset *Asset) error {
    // Safety check: Only allow localhost
    if asset.Name != "127.0.0.1" && asset.Name != "localhost" {
        return fmt.Errorf("⚠️ Port scanning requires authorization. " +
            "Only localhost scanning allowed in training environment. " +
            "Unauthorized port scanning may violate laws.")
    }

    log.Warn("⚠️ Active scan initiated on: %s", asset.Name)
    // ... perform scan
}
```

---

## Architecture Patterns

### 1. Job/Task Pattern

```go
// Async scanning pattern
type ScanJob struct {
    ID        string
    AssetID   string
    ScanType  ScanType
    Status    ScanStatus  // pending → running → completed/failed
    Results   int
}

// Usage:
job := CreateScanJob(assetID, ScanTypeDNS)
job.Status = StatusRunning
go performDNSScan(job)  // Background goroutine
return job.ID           // Return immediately
```

**Benefits:**

- Non-blocking HTTP responses
- Progress tracking
- Error isolation
- Retry capability

### 2. Scanner Interface

```go
type Scanner interface {
    Scan(asset *Asset, job *ScanJob) error
    Type() ScanType
}

// Implementations:
type WHOISScanner struct{}
type DNSScanner struct{}
type SubdomainScanner struct{}

// Usage (polymorphism):
var scanners = map[ScanType]Scanner{
    ScanTypeWHOIS:     &WHOISScanner{},
    ScanTypeDNS:       &DNSScanner{},
    ScanTypeSubdomain: &SubdomainScanner{},
}

scanner := scanners[job.ScanType]
err := scanner.Scan(asset, job)
```

**Benefits:**

- Easy to add new scan types
- Testable (mock scanners)
- Consistent interface
- Separation of concerns

### 3. Result Storage Pattern

```go
// Generic result storage
type ScanResult interface {
    GetAssetID() string
    GetScanJobID() string
}

// Specific implementations:
type Subdomain implements ScanResult
type DNSRecord implements ScanResult
type WHOISRecord implements ScanResult

// Storage:
func (s *Storage) SaveScanResult(result ScanResult) error
```

**Benefits:**

- Flexible result types
- Consistent storage interface
- Easy to query
- Supports different result structures

### 4. Rate Limiting

```go
// Prevent overwhelming external services
type RateLimiter struct {
    tokens chan struct{}
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
    limiter := &RateLimiter{
        tokens: make(chan struct{}, requestsPerSecond),
    }

    // Refill tokens
    go func() {
        ticker := time.NewTicker(time.Second)
        for range ticker.C {
            for i := 0; i < requestsPerSecond; i++ {
                select {
                case limiter.tokens <- struct{}{}:
                default:
                }
            }
        }
    }()

    return limiter
}

func (r *RateLimiter) Wait() {
    <-r.tokens  // Block until token available
}

// Usage:
limiter := NewRateLimiter(10)  // 10 req/sec
for _, subdomain := range wordlist {
    limiter.Wait()
    go queryDNS(subdomain)
}
```

**Benefits:**

- Respect service limits
- Avoid DoS accusations
- Predictable performance
- Resource management

### 5. Error Handling

```go
// Graceful error handling
func (s *Scanner) Scan(asset *Asset, job *ScanJob) error {
    defer func() {
        if r := recover(); r != nil {
            job.Status = StatusFailed
            job.Error = fmt.Sprintf("panic: %v", r)
            s.storage.UpdateScanJob(job)
        }
    }()

    // Attempt scan
    results, err := s.performScan(asset)
    if err != nil {
        job.Status = StatusFailed
        job.Error = err.Error()
        return err
    }

    // Partial results?
    if len(results) == 0 {
        job.Status = StatusPartial
        job.Error = "no results found"
    } else {
        job.Status = StatusCompleted
    }

    job.Results = len(results)
    return s.storage.UpdateScanJob(job)
}
```

**Benefits:**

- Always update job status
- Distinguish failures
- Preserve partial results
- Debug information

---

## Next Steps

1. **Study this architecture** - Understand the complete flow
2. **Implement basics** - WHOIS, DNS, Subdomain (provided)
3. **Complete exercises** - Port scan, SSL, ASN (student work)
4. **Test thoroughly** - Each scan type independently
5. **Integrate** - Connect all scans for full discovery
6. **Optimize** - Add caching, better rate limiting
7. **Monitor** - Dashboard showing scan progress

---

## References

- RFC 1035: Domain Names
- RFC 3912: WHOIS Protocol
- Certificate Transparency: https://crt.sh
- Common ports: IANA Service Name Registry
- WHOIS servers: IANA WHOIS Service
