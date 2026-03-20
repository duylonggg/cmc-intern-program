package scanner

import (
	"context"
	"embed"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"mini-asm/internal/model"
)

//go:embed wordlists/subdomains.txt
var wordlistFS embed.FS

// SubdomainScanner discovers subdomains using various methods
//
// SCAN CATEGORY: 🔵 PASSIVE / OSINT
// - Uses DNS bruteforce (just DNS lookups)
// - Can also use Certificate Transparency logs (passive)
// - No direct interaction with target web servers
// - Legal: No authorization required
// - Stealth: Only DNS queries, no HTTP requests to target
//
// LEGAL STATUS: ✅ Safe
// While more aggressive than single DNS queries, subdomain enumeration
// only performs DNS lookups. DNS is a public, distributed database.
// Rate limiting ensures respectful behavior toward DNS servers.
//
// METHODS:
// 1. DNS Bruteforce - Query DNS for common subdomain names
// 2. Certificate Transparency - Query public CT logs (crt.sh)
// 3. Search Engine Scraping - Parse Google/Bing results (passive)
type SubdomainScanner struct {
	wordlist       []string
	timeout        time.Duration
	maxWorkers     int
	requestsPerSec int
}

// NewSubdomainScanner creates a new subdomain scanner instance
func NewSubdomainScanner() (*SubdomainScanner, error) {
	scanner := &SubdomainScanner{
		timeout:        5 * time.Second,
		maxWorkers:     50,  // 50 concurrent DNS queries
		requestsPerSec: 100, // Rate limit: 100 queries/sec
	}

	// Load wordlist
	if err := scanner.loadWordlist(); err != nil {
		return nil, fmt.Errorf("failed to load wordlist: %w", err)
	}

	return scanner, nil
}

// Type returns the scan type identifier
func (s *SubdomainScanner) Type() model.ScanType {
	return model.ScanTypeSubdomain
}

// Scan discovers subdomains using DNS bruteforce
func (s *SubdomainScanner) Scan(asset *model.Asset, ctx context.Context) ([]*model.Subdomain, error) {
	if asset.Type != model.TypeDomain {
		return nil, fmt.Errorf("subdomain scan only works on domain assets, got: %s", asset.Type)
	}

	domain := asset.Name

	// Method 1: DNS Bruteforce (implemented)
	subdomains := s.bruteforce(domain, ctx)

	// Method 2: Certificate Transparency (student exercise)
	// Method 3: Web Scraping (student exercise)
	// Method 4: Search Engine (student exercise)

	return subdomains, nil
}

// bruteforce discovers subdomains by trying common names
func (s *SubdomainScanner) bruteforce(domain string, ctx context.Context) []*model.Subdomain {
	results := make([]*model.Subdomain, 0)
	resultsMutex := sync.Mutex{}

	// Create worker pool
	jobs := make(chan string, len(s.wordlist))

	// Rate limiter (fixed-rate limiting using ticker)
	// Allows 1 request every (1/requestsPerSec) seconds
	// Example: 100 req/sec = 1 request per 10ms
	// Note: This is NOT a token bucket (no burst capacity)
	limiter := time.NewTicker(time.Second / time.Duration(s.requestsPerSec))
	defer limiter.Stop()

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < s.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for word := range jobs {
				// Check context cancellation
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Rate limiting
				<-limiter.C

				// Try this subdomain
				subdomain := fmt.Sprintf("%s.%s", word, domain)

				// DNS lookup with timeout
				lookupCtx, cancel := context.WithTimeout(ctx, s.timeout)
				ips, err := net.DefaultResolver.LookupIP(lookupCtx, "ip", subdomain)
				cancel()

				if err == nil && len(ips) > 0 {
					// Subdomain exists!
					result := &model.Subdomain{
						Name:     subdomain,
						Source:   "dns_bruteforce",
						IsActive: true, // We just resolved it, so it's active
					}

					resultsMutex.Lock()
					results = append(results, result)
					resultsMutex.Unlock()
				}
			}
		}()
	}

	// Send jobs to workers
	for _, word := range s.wordlist {
		select {
		case jobs <- word:
		case <-ctx.Done():
			break
		}
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()

	return results
}

// loadWordlist loads the subdomain wordlist from embedded file
func (s *SubdomainScanner) loadWordlist() error {
	// Read embedded wordlist
	data, err := wordlistFS.ReadFile("wordlists/subdomains.txt")
	if err != nil {
		// Fallback to default wordlist if file not found
		s.wordlist = s.defaultWordlist()
		return nil
	}

	// Parse lines
	lines := strings.Split(string(data), "\n")
	s.wordlist = make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		s.wordlist = append(s.wordlist, line)
	}

	if len(s.wordlist) == 0 {
		s.wordlist = s.defaultWordlist()
	}

	return nil
}

// defaultWordlist returns a default list of common subdomain names
func (s *SubdomainScanner) defaultWordlist() []string {
	return []string{
		// Web servers
		"www", "www1", "www2", "www3",
		"web", "web1", "web2",

		// APIs and applications
		"api", "api1", "api2", "api-prod", "api-dev",
		"app", "app1", "app2",
		"portal", "dashboard", "admin", "panel",

		// Development and staging
		"dev", "development", "develop",
		"test", "testing", "qa",
		"stage", "staging", "stg",
		"uat", "preprod", "pre-prod",
		"demo", "sandbox",

		// Communication
		"mail", "email", "smtp", "imap", "pop", "pop3",
		"webmail", "exchange",

		// Servers and infrastructure
		"ftp", "sftp", "ssh",
		"vpn", "remote", "rdp",
		"ns", "ns1", "ns2", "ns3", "ns4",
		"dns", "dns1", "dns2",

		// Databases
		"db", "database", "mysql", "postgres", "pgsql",
		"mongo", "redis", "elastic", "elasticsearch",

		// Content and media
		"blog", "shop", "store", "forum",
		"cdn", "static", "assets", "media", "images", "img",
		"files", "download", "downloads", "upload",

		// Mobile
		"m", "mobile", "touch", "wap",
		"android", "ios",

		// Security
		"secure", "ssl", "tls",
		"auth", "login", "sso",

		// Monitoring and tools
		"status", "health", "monitor", "monitoring",
		"metrics", "logs", "logging",
		"grafana", "kibana", "prometheus",

		// Cloud providers
		"aws", "azure", "gcp",
		"s3", "ec2",

		// Old/backup
		"old", "backup", "bak",
		"archive", "www-old", "app-old",
		"v1", "v2", "v3",

		// Regional
		"us", "eu", "asia", "apac",
		"uk", "de", "fr", "jp", "cn",

		// Common words
		"help", "support", "docs", "documentation",
		"news", "blog", "about", "careers", "jobs",
	}
}

/*
🎓 NOTES - Subdomain Scanner

=== WHAT IS SUBDOMAIN ENUMERATION? ===

Subdomain: A domain that is part of a larger domain

Example:
  example.com → root domain
  www.example.com → subdomain
  api.example.com → subdomain
  mail.example.com → subdomain

Structure: {subdomain}.{domain}.{tld}

Why enumerate?
- Find all entry points to organization's infrastructure
- Discover forgotten/unprotected services
- Map attack surface
- Asset inventory

=== SUBDOMAIN DISCOVERY METHODS ===

1. **DNS Bruteforce** (implemented):
   - Try common names (www, api, mail, etc.)
   - Query DNS for each
   - If resolves → subdomain exists
   - Pros: Simple, reliable
   - Cons: Only finds common names, slow

2. **Certificate Transparency** (student exercise):
   - Query CT logs (public certificate database)
   - Certificates list Subject Alternative Names (SANs)
   - Find: *.example.com, api.example.com
   - Pros: Finds many subdomains
   - Cons: Only SSL-enabled subdomains
   - API: https://crt.sh/?q=%.example.com&output=json

3. **Search Engine Discovery**:
   - Google: site:example.com
   - Bing: site:example.com
   - Extract unique subdomains from results
   - Pros: Finds indexed subdomains
   - Cons: Rate limiting, incomplete

4. **Web Scraping**:
   - Crawl main domain
   - Extract links to subdomains
   - Follow internal links
   - Pros: Finds linked subdomains
   - Cons: Slow, might miss unlinked

5. **DNS Zone Transfer** (rare):
   - Request full DNS zone (AXFR)
   - Usually disabled (security)
   - Pros: Complete list if available
   - Cons: Almost never works

6. **Reverse DNS**:
   - Find IP ranges owned by org
   - Reverse lookup each IP
   - Find associated domains
   - Pros: Finds IPs → domains
   - Cons: Need IP ranges first

=== DNS BRUTEFORCE IMPLEMENTATION ===

Core algorithm:

```go
wordlist := ["www", "api", "mail", ...]
domain := "example.com"

for each word in wordlist:
    subdomain := word + "." + domain
    if DNS_lookup(subdomain) succeeds:
        save subdomain
```

Example:
```
www.example.com → resolves to 93.184.216.34 ✅ Found!
api.example.com → resolves to 192.168.1.1 ✅ Found!
xyz.example.com → NXDOMAIN ❌ Not found
```

=== WORDLIST STRATEGY ===

Our default wordlist includes:

1. **Common Web**: www, api, app, portal
2. **Environments**: dev, staging, test, prod
3. **Infrastructure**: mail, ftp, vpn, dns
4. **Databases**: db, mysql, postgres, mongo
5. **CDN/Static**: cdn, static, assets, media
6. **Regional**: us, eu, uk, asia
7. **Old/Backup**: old, backup, archive

Size tradeoff:
- Small wordlist (100 items): Fast, might miss subdomains
- Large wordlist (10,000+ items): Slow, more thorough
- SecLists has comprehensive wordlists

For teaching: ~100 items (fast demos)
For production: Use SecLists or DNSdumpster wordlists

=== CONCURRENT SCANNING ===

Sequential vs Parallel:

Sequential (slow):
```go
for _, word := range wordlist {
    lookup(word + "." + domain)  // 1ms each
}
// 1000 words * 1ms = 1 second
```

Parallel (fast):
```go
// 50 workers
for i := 0; i < 50; i++ {
    go worker()  // Each processes jobs concurrently
}
// 1000 words / 50 workers = 20ms (if 1ms each)
```

Our implementation:
- 50 concurrent workers (goroutines)
- Job queue pattern

Why not unlimited concurrency?
- Too many goroutines → memory usage
- Too many DNS queries → server might throttle/block
- Optimal: 50-100 workers

=== RATE LIMITING ===

Problem: 50 workers * instant = 50 queries/ms = 50,000 queries/sec
Result: DNS server blocks you!

Solution: Rate limiter (token bucket)

```go
limiter := time.NewTicker(time.Second / 100)  // 100 req/sec

for word := range jobs {
    <-limiter.C  // Wait for token
    lookup(word + "." + domain)
}
```

How it works:
- Ticker generates token every 10ms (100/sec)
- Worker waits for token (<-limiter.C)
- Gets token → makes request
- Fair distribution across workers

Rate limits to consider:
- Google DNS (8.8.8.8): ~1000 req/sec
- Local DNS: varies
- WHOIS: ~1-10 req/sec (much stricter!)

Our default: 100 req/sec (conservative)

=== CONTEXT AND CANCELLATION ===

Problem: Scan might take minutes, user wants to stop

Solution: Context cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

go func() {
    time.Sleep(5 * time.Minute)
    cancel()  // Stop after 5 minutes
}()

scanner.Scan(asset, ctx)
```

In worker:
```go
select {
case <-ctx.Done():
    return  // Stop immediately
default:
    // Continue working
}
```

Benefits:
- Graceful shutdown
- Timeout support
- User can cancel
- Prevents runaway goroutines

=== GO EMBED FOR WORDLISTS ===

Embedding files in binary:

```go
//go:embed wordlists/subdomains.txt
var wordlistFS embed.FS

data, err := wordlistFS.ReadFile("wordlists/subdomains.txt")
```

Why embed?
- ✅ Single binary (no external files)
- ✅ Can't forget to deploy wordlist
- ✅ Consistent across environments
- ❌ Larger binary size

Alternative: Load from file
```go
data, err := os.ReadFile("subdomains.txt")
```

For this project: We try embedded first, fallback to default

=== WORKER POOL PATTERN ===

Pattern for concurrent processing:

```go
// 1. Create job channel
jobs := make(chan string, len(wordlist))

// 2. Start workers
for i := 0; i < numWorkers; i++ {
    go worker(jobs, results)
}

// 3. Send jobs
for _, word := range wordlist {
    jobs <- word
}
close(jobs)  // Signal: no more jobs

// 4. Wait for completion
wg.Wait()
```

Flow:
```
Main goroutine: Send jobs → [jobs channel]
                                ↓
                            Worker 1 ──┐
                            Worker 2 ──┤→ [results]
                            Worker 3 ──┘
```

Benefits:
- Bounded concurrency (not unlimited)
- Fair work distribution
- Easy to add/remove workers
- Graceful completion

=== THREAD-SAFE RESULTS COLLECTION ===

Problem: Multiple goroutines writing to same slice

```go
// ❌ RACE CONDITION
results := []*Subdomain{}
for i := 0; i < 50; i++ {
    go func() {
        results = append(results, subdomain)  // UNSAFE!
    }()
}
```

Solution: Mutex

```go
// ✅ THREAD-SAFE
results := []*Subdomain{}
mutex := sync.Mutex{}

for i := 0; i < 50; i++ {
    go func() {
        mutex.Lock()
        results = append(results, subdomain)
        mutex.Unlock()
    }()
}
```

Alternative: Channel
```go
resultChan := make(chan *Subdomain, 100)
// Workers send to channel
// Main goroutine reads from channel
```

We use mutex (simpler for this case)

=== DNS LOOKUP WITH TIMEOUT ===

```go
lookupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
ips, err := net.DefaultResolver.LookupIP(lookupCtx, "ip", subdomain)
cancel()
```

Why timeout?
- Some domains slow to resolve
- Prevents hanging forever
- Keep scan moving

5 second timeout:
- Generous (most resolve < 1s)
- Prevents dead queries blocking

=== SECURITY IMPLICATIONS ===

What attackers learn:

1. **dev.example.com** → Development environment
   - Might have weaker security
   - Debug mode enabled?
   - Credentials in config files?

2. **admin.example.com** → Admin panel
   - Default credentials?
   - Brute force login?

3. **backup.example.com** → Backup files
   - Publicly accessible?
   - Sensitive data?

4. **old.example.com** → Abandoned site
   - Unpatched vulnerabilities?

Real-world: Many breaches start with subdomain discovery!

=== PERFORMANCE OPTIMIZATION ===

Factors affecting speed:

1. **Workers**: More = faster (up to a point)
   - 1 worker: 1000 words = 1000s (sequential)
   - 50 workers: 1000 words = 20s (parallel)
   - 1000 workers: Diminishing returns (DNS limits)

2. **Rate Limit**: Lower = faster (but risky)
   - 10 req/sec: 1000 words = 100s
   - 100 req/sec: 1000 words = 10s
   - 1000 req/sec: Might get blocked!

3. **Timeout**: Lower = faster (but might miss)
   - 1s timeout: Fast but might miss slow domains
   - 10s timeout: Thorough but slower

4. **Wordlist**: Smaller = faster
   - 100 words: ~1 second
   - 10,000 words: ~100 seconds
   - Million words: Hours!

Our defaults (balanced):
- 50 workers
- 100 req/sec
- 5s timeout
- ~100 word default list

=== INTEGRATION WITH SCANNING FLOW ===

```
1. User has domain: example.com
2. Subdomain scan triggered
3. Bruteforce 100 common names
4. Found: www, api, mail
5. Create Subdomain records
6. Create Asset for each subdomain
7. Trigger DNS scan on each subdomain
8. Recursive discovery continues...
```

Recursive example:
```
example.com (root)
├─ www.example.com → DNS scan → 93.184.216.34
├─ api.example.com → DNS scan → 192.168.1.1
│  └─ api-v2.example.com (found via cert transparency)
└─ mail.example.com → DNS scan → 192.168.1.2
```

=== STUDENT EXERCISES ===

1. **Add Certificate Transparency**:
   ```go
   func (s *SubdomainScanner) certTransparency(domain string) ([]*Subdomain, error) {
       url := "https://crt.sh/?q=%." + domain + "&output=json"
       // HTTP GET request
       // Parse JSON response
       // Extract unique names
   }
   ```

2. **Add Permutation Scanner**:
   ```
   api.example.com exists
   Try variations:
   - api-v2.example.com
   - api2.example.com
   - api-prod.example.com
   ```

3. **Add DNS Recursion**:
   ```
   Query: NS records for example.com
   Query each nameserver: AXFR (zone transfer)
   If succeeds (rare): Get all subdomains!
   ```

4. **Add Historical Data**:
   ```
   Query: SecurityTrails API
   or: DNSdumpster
   Get: Previously seen subdomains
   ```

5. **Add Wildcard Detection**:
   ```
   Problem: *.example.com resolves to 1.2.3.4
   random123.example.com → 1.2.3.4 (fake)

   Solution:
   - Query random subdomain first
   - If resolves: wildcard detected
   - Filter out false positives
   ```

=== REAL-WORLD TOOLS ===

Professional subdomain scanners:

- **Amass** (OWASP) - Comprehensive, uses multiple methods
- **Subfinder** - Fast, uses multiple sources
- **Sublist3r** - Python, popular
- **DNSdumpster** - Web interface
- **SecurityTrails** - Commercial API

Our implementation: Educational, solid foundation

Compare:
- Amass: 10+ methods, very thorough, slow
- Ours: 1 method (bruteforce), fast, good for demos

=== DEMO SCENARIOS ===

1. **Small domain**:
   ```
   example.com
   Found: www, api (2 subdomains)
   Time: 1 second
   ```

2. **Large organization**:
   ```
   google.com
   Found: www, mail, drive, docs, ... (50+ subdomains)
   Time: 5 seconds
   ```

3. **No subdomains**:
   ```
   obscure-domain.com
   Found: None
   Time: 2 seconds (tried all 100 words)
   ```

4. **Wildcard domain**:
   ```
   *.example.com → 1.2.3.4
   Found: All words (false positives!)
   Need wildcard detection
   ```

=== KEY TAKEAWAYS ===

1. Subdomain enumeration finds hidden assets
2. Multiple methods exist (bruteforce, CT, scraping)
3. Concurrency = speed (workers + job queue)
4. Rate limiting prevents blocking
5. Context enables cancellation/timeout
6. Thread-safe result collection (mutex)
7. Security goldmine for attackers
8. Foundation for recursive scanning

Subdomain enum is a gateway to full asset discovery!
*/
