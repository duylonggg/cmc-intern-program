package scanner

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"mini-asm/internal/model"
)

// WHOISScanner performs WHOIS lookups on domains
type WHOISScanner struct {
	timeout time.Duration
}

// NewWHOISScanner creates a new WHOIS scanner instance
func NewWHOISScanner() *WHOISScanner {
	return &WHOISScanner{
		timeout: 10 * time.Second, // 10 second timeout
	}
}

// Type returns the scan type identifier
func (s *WHOISScanner) Type() model.ScanType {
	return model.ScanTypeWHOIS
}

// Scan performs WHOIS lookup for a domain
func (s *WHOISScanner) Scan(asset *model.Asset) (*model.WHOISRecord, error) {
	if asset.Type != model.TypeDomain {
		return nil, fmt.Errorf("WHOIS scan only works on domain assets, got: %s", asset.Type)
	}

	domain := asset.Name

	// Query WHOIS server
	rawData, err := s.queryWHOIS(domain)
	if err != nil {
		return nil, fmt.Errorf("WHOIS query failed: %w", err)
	}

	// Parse WHOIS response
	record := &model.WHOISRecord{
		RawData: rawData,
	}

	s.parseWHOIS(rawData, record)

	return record, nil
}

// queryWHOIS connects to WHOIS server and retrieves data
func (s *WHOISScanner) queryWHOIS(domain string) (string, error) {
	// Determine WHOIS server based on TLD
	whoisServer := s.getWHOISServer(domain)

	// Connect to WHOIS server (port 43)
	conn, err := net.DialTimeout("tcp", whoisServer+":43", s.timeout)
	if err != nil {
		return "", fmt.Errorf("failed to connect to WHOIS server %s: %w", whoisServer, err)
	}
	defer conn.Close()

	// Set read deadline
	conn.SetDeadline(time.Now().Add(s.timeout))

	// Send domain query
	_, err = fmt.Fprintf(conn, "%s\r\n", domain)
	if err != nil {
		return "", fmt.Errorf("failed to send query: %w", err)
	}

	// Read response
	response := strings.Builder{}
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		response.WriteString(scanner.Text())
		response.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return response.String(), nil
}

// getWHOISServer determines the appropriate WHOIS server for a domain
func (s *WHOISScanner) getWHOISServer(domain string) string {
	// Extract TLD (top-level domain)
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return "whois.iana.org" // Fallback
	}

	tld := parts[len(parts)-1]

	// Map common TLDs to WHOIS servers
	servers := map[string]string{
		"com":  "whois.verisign-grs.com",
		"net":  "whois.verisign-grs.com",
		"org":  "whois.pir.org",
		"edu":  "whois.educause.edu",
		"gov":  "whois.dotgov.gov",
		"io":   "whois.nic.io",
		"dev":  "whois.nic.google",
		"app":  "whois.nic.google",
		"xyz":  "whois.nic.xyz",
		"tech": "whois.nic.tech",
		"info": "whois.afilias.net",
		"biz":  "whois.biz",
	}

	if server, ok := servers[tld]; ok {
		return server
	}

	// Fallback: try TLD-specific server
	return fmt.Sprintf("whois.nic.%s", tld)
}

// parseWHOIS extracts structured data from WHOIS response
func (s *WHOISScanner) parseWHOIS(rawData string, record *model.WHOISRecord) {
	lines := strings.Split(rawData, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on colon (common separator)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		// Extract fields based on key
		switch {
		// Registrar
		case strings.Contains(key, "registrar") && !strings.Contains(key, "url") && !strings.Contains(key, "whois"):
			if record.Registrar == "" {
				record.Registrar = value
			}

		// Creation date
		case strings.Contains(key, "creation") || strings.Contains(key, "created"):
			if record.CreatedDate == nil {
				if date := s.parseDate(value); date != nil {
					record.CreatedDate = date
				}
			}

		// Expiry date
		case strings.Contains(key, "expir") || strings.Contains(key, "expiration"):
			if record.ExpiryDate == nil {
				if date := s.parseDate(value); date != nil {
					record.ExpiryDate = date
				}
			}

		// Name servers
		case strings.Contains(key, "name server") || strings.Contains(key, "nserver"):
			// Append to name servers list (we'll collect all)
			if record.NameServers == "" {
				record.NameServers = value
			} else {
				record.NameServers += "," + value
			}

		// Status
		case strings.Contains(key, "status") || strings.Contains(key, "domain status"):
			if record.Status == "" {
				record.Status = value
			} else {
				record.Status += "," + value
			}

		// Emails
		case strings.Contains(key, "email") || strings.Contains(key, "e-mail"):
			if record.Emails == "" {
				record.Emails = value
			} else {
				record.Emails += "," + value
			}
		}
	}
}

// parseDate attempts to parse various date formats
func (s *WHOISScanner) parseDate(value string) *time.Time {
	// Clean up value
	value = strings.TrimSpace(value)

	// Common date formats in WHOIS responses
	formats := []string{
		"2006-01-02T15:04:05Z",        // ISO 8601
		"2006-01-02 15:04:05",         // Common format
		"2006-01-02",                  // Date only
		"02-Jan-2006",                 // Some registrars
		"2006/01/02",                  // Alternative separator
		"Mon Jan 2 15:04:05 MST 2006", // Some use this
	}

	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return &t
		}
	}

	// Try regex-based extraction for dates like "2023-01-15T00:00:00Z (before text)"
	re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z)`)
	if matches := re.FindStringSubmatch(value); len(matches) > 1 {
		if t, err := time.Parse("2006-01-02T15:04:05Z", matches[1]); err == nil {
			return &t
		}
	}

	// Couldn't parse
	return nil
}

/*
🎓 NOTES - WHOIS Scanner

=== WHAT IS WHOIS? ===

WHOIS is a protocol (RFC 3912) for querying database of:
- Domain registrations
- IP address allocations
- Autonomous system numbers

For domains, WHOIS tells us:
- Who owns it (registrant)
- Who manages it (registrar)
- When registered/expires
- Name servers
- Contact information

=== WHOIS PROTOCOL ===

Very simple protocol:
1. Connect to WHOIS server on port 43 (TCP)
2. Send domain name + CRLF (\r\n)
3. Server sends back text response
4. Close connection

Example session:
```
Client → Server: example.com\r\n
Server → Client:
   Domain Name: EXAMPLE.COM
   Registrar: Example Registrar Inc.
   Creation Date: 1995-08-14T04:00:00Z
   Expiration Date: 2024-08-13T04:00:00Z
   ...
```

No authentication, no encryption (plain text)!

=== WHOIS SERVERS ===

Different WHOIS servers for different TLDs:

- .com, .net → whois.verisign-grs.com
- .org → whois.pir.org
- .io → whois.nic.io
- .edu → whois.educause.edu

General pattern: whois.nic.{tld}

IANA registry: whois.iana.org (can redirect to specific servers)

=== THIN vs THICK WHOIS ===

Thin WHOIS:
- Registry only has basic info
- Refers to registrar's WHOIS for details
- Example: .com (Verisign)

Thick WHOIS:
- Registry has complete info
- No need to query registrar
- Example: .org, .info

Our implementation: Query registry only (thin approach)
Students can enhance: Follow referrals to registrar

=== PARSING CHALLENGES ===

WHOIS responses are NOT standardized:

Example 1 (Verisign):
```
Domain Name: EXAMPLE.COM
Registrar: Example Inc.
Creation Date: 2023-01-15T00:00:00Z
```

Example 2 (Alternative):
```
domain:       example.com
registrar:    Example Inc.
created:      2023-01-15
```

Example 3 (Another format):
```
Domain name: example.com
Registrar: Example Inc.
Created on: 15-Jan-2023
```

Our strategy:
- Case-insensitive matching
- Flexible key matching (contains, not equals)
- Multiple date format parsing
- Extract what we can, ignore rest

=== IMPLEMENTATION DETAILS ===

1. Connection with Timeout:
```go
conn, err := net.DialTimeout("tcp", server+":43", 10*time.Second)
```
- Port 43 (WHOIS standard)
- Timeout prevents hanging forever
- Some servers slow/unresponsive

2. Setting Deadline:
```go
conn.SetDeadline(time.Now().Add(timeout))
```
- Limits total time for read/write
- Prevents reading forever
- Returns error after timeout

3. Sending Query:
```go
fmt.Fprintf(conn, "%s\r\n", domain)
```
- CRLF line ending (\r\n) required by protocol
- Just \n might not work with some servers

4. Reading Response:
```go
scanner := bufio.NewScanner(conn)
for scanner.Scan() {
    // Process line by line
}
```
- Line-by-line reading
- Efficient for text protocol
- Automatically handles line endings

=== PARSING STRATEGY ===

1. Split into lines
2. Skip comments (%, #) and empty lines
3. Split on colon (:)
4. Match keys flexibly:
   - "registrar" matches "Registrar:", "Registrar Name:", etc.
   - Case-insensitive
   - Contains, not equals

5. Extract values
6. Parse dates with multiple formats
7. Aggregate multi-value fields (name servers)

Example:
```
Input:
   Name Server: ns1.example.com
   Name Server: ns2.example.com

Output:
   NameServers: "ns1.example.com,ns2.example.com"
```

=== DATE PARSING ===

Multiple formats attempted:

```go
formats := []string{
    "2006-01-02T15:04:05Z",        // ISO 8601
    "2006-01-02",                  // Date only
    "02-Jan-2006",                 // Month name
}
```

Go's time format is special:
- Uses reference date: Mon Jan 2 15:04:05 MST 2006
- Not "YYYY-MM-DD", but "2006-01-02"
- Easy to remember: 1/2 3:4:5 2006 -7

If parsing fails → date is nil (nullable field)

=== SECURITY CONSIDERATIONS ===

1. PII in WHOIS:
   - Email addresses (spam, phishing risk)
   - Names, addresses (privacy concern)
   - GDPR impact: Many registrars redact this now

2. Reconnaissance Value:
   - Expiration dates → domain hijacking risk
   - Email addresses → social engineering targets
   - Name servers → infrastructure mapping

3. Rate Limiting:
   - WHOIS servers have limits
   - Abuse → IP blocked
   - Our implementation: No rate limiting (add if needed)

4. Legal Considerations:
   - Terms of service vary by registrar
   - Some prohibit automated queries
   - Commercial use may require license

=== EXPIRATION MONITORING ===

Critical security feature:

```go
if record.ExpiryDate != nil {
    daysUntilExpiry := time.Until(*record.ExpiryDate).Hours() / 24

    if daysUntilExpiry < 30 {
        // ALERT! Domain expiring soon
        // Risk: Domain could be taken over
    }
}
```

Real-world incidents:
- Domain expires → attacker registers it
- Email/website taken over
- Phishing attacks using legitimate domain

=== COMPARISON WITH TOOLS ===

Our scanner vs command line:

```bash
# Command line
whois example.com

# Our scanner (programmatic)
scanner := NewWHOISScanner()
record, err := scanner.Scan(asset)
```

Advantages:
- Automated scanning
- Store in database
- Monitor changes over time
- Alert on expiration
- Integrate with other scans

=== ERROR HANDLING ===

Common errors:

1. Connection timeout:
   - Server down/slow
   - Network issues
   - Firewall blocking port 43

2. Domain not found:
   - Unregistered domain
   - Invalid TLD

3. Rate limiting:
   - Too many queries
   - IP blocked temporarily

4. Parsing failures:
   - Unrecognized format
   - Non-standard response

Our approach: Return error, don't crash

=== INTEGRATION WITH SCANNING FLOW ===

```
1. User has domain asset: example.com
2. WHOIS scan triggered
3. Query whois.verisign-grs.com:43
4. Parse response
5. Store in whois_records table
6. Check expiration date
7. Alert if expiring soon
```

=== STUDENT EXERCISES ===

1. Add RDAP Support (modern replacement for WHOIS):
   - RESTful API (HTTP/JSON)
   - Standardized format
   - Better for automation
   - https://rdap.org

2. Follow Referrals:
   - Parse "Registrar WHOIS Server:" line
   - Query registrar for full details
   - Two-stage lookup (thick WHOIS)

3. Add Change Detection:
   - Compare with previous WHOIS scan
   - Detect: Registrar change, name server change
   - Alert on suspicious changes

4. Add Bulk Query:
   - Query multiple domains
   - Rate limiting (e.g., 1 query per second)
   - Parallel with semaphore

5. Better Date Parsing:
   - Handle timezone offsets
   - Support more formats
   - Use dateparse library

=== REAL-WORLD TOOLS ===

Professional tools using WHOIS:

- **SecurityTrails** - Monitor WHOIS changes
- **DomainTools** - WHOIS history and analysis
- **WhoisXML API** - WHOIS data as a service
- **Amass** - OSINT framework (includes WHOIS)

Our implementation: Educational, good foundation

=== DEMO POINTS ===

1. Query well-known domain (example.com)
2. Show raw WHOIS response
3. Show parsed fields
4. Demonstrate expiration checking
5. Handle query failure (invalid domain)
6. Compare WHOIS for different TLDs (.com vs .org)

=== WHOIS ALTERNATIVES ===

Modern alternatives:

1. RDAP (Registration Data Access Protocol):
   - JSON over HTTPS
   - Standardized
   - Better for automation

2. DNS-based:
   - Some info in TXT records
   - Less complete

3. Paid APIs:
   - WhoisXML API
   - SecurityTrails API
   - More reliable, better parsing

For production: Consider paid API or RDAP
For learning: Our WHOIS implementation perfect!

=== KEY TAKEAWAYS ===

1. WHOIS is simple protocol (port 43, text)
2. Parsing is hard (no standard format)
3. Critical for domain monitoring
4. Expiration dates = security risk
5. PII concerns (emails, names)
6. Rate limiting important
7. Foundation for asset inventory

WHOIS is old but still essential!
*/
