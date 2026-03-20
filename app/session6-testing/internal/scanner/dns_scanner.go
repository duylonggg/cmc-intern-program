package scanner

import (
	"fmt"
	"net"
	"strings"

	"mini-asm/internal/model"
)

// DNSScanner performs DNS lookups on domains/subdomains
//
// SCAN CATEGORY: 🔵 PASSIVE / OSINT

type DNSScanner struct{}

// NewDNSScanner creates a new DNS scanner instance
func NewDNSScanner() *DNSScanner {
	return &DNSScanner{}
}

// Type returns the scan type identifier
func (s *DNSScanner) Type() model.ScanType {
	return model.ScanTypeDNS
}

// Scan performs DNS lookups for various record types
func (s *DNSScanner) Scan(asset *model.Asset) ([]*model.DNSRecord, error) {
	if asset.Type != model.TypeDomain {
		return nil, fmt.Errorf("DNS scan only works on domain assets, got: %s", asset.Type)
	}

	domain := asset.Name
	records := []*model.DNSRecord{}

	// Query A records (IPv4)
	aRecords, err := s.lookupA(domain)
	if err == nil {
		records = append(records, aRecords...)
	}

	// Query AAAA records (IPv6)
	aaaaRecords, err := s.lookupAAAA(domain)
	if err == nil {
		records = append(records, aaaaRecords...)
	}

	// Query CNAME records
	cnameRecord, err := s.lookupCNAME(domain)
	if err == nil && cnameRecord != nil {
		records = append(records, cnameRecord)
	}

	// Query MX records (mail servers)
	mxRecords, err := s.lookupMX(domain)
	if err == nil {
		records = append(records, mxRecords...)
	}

	// Query NS records (name servers)
	nsRecords, err := s.lookupNS(domain)
	if err == nil {
		records = append(records, nsRecords...)
	}

	// Query TXT records
	txtRecords, err := s.lookupTXT(domain)
	if err == nil {
		records = append(records, txtRecords...)
	}

	// Note: SOA lookup skipped for simplicity (students can add)

	return records, nil
}

// lookupA queries A records (IPv4 addresses)
func (s *DNSScanner) lookupA(domain string) ([]*model.DNSRecord, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}

	records := []*model.DNSRecord{}
	for _, ip := range ips {
		// Filter for IPv4 only
		if ip.To4() != nil {
			records = append(records, &model.DNSRecord{
				RecordType: "A",
				Name:       domain,
				Value:      ip.String(),
				TTL:        0, // Go's net package doesn't expose TTL easily
			})
		}
	}

	return records, nil
}

// lookupAAAA queries AAAA records (IPv6 addresses)
func (s *DNSScanner) lookupAAAA(domain string) ([]*model.DNSRecord, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}

	records := []*model.DNSRecord{}
	for _, ip := range ips {
		// Filter for IPv6 only
		if ip.To4() == nil {
			records = append(records, &model.DNSRecord{
				RecordType: "AAAA",
				Name:       domain,
				Value:      ip.String(),
				TTL:        0,
			})
		}
	}

	return records, nil
}

// lookupCNAME queries CNAME record (canonical name alias)
func (s *DNSScanner) lookupCNAME(domain string) (*model.DNSRecord, error) {
	cname, err := net.LookupCNAME(domain)
	if err != nil {
		return nil, err
	}

	// LookupCNAME returns domain itself if no CNAME exists
	if cname == domain || cname == domain+"." {
		return nil, fmt.Errorf("no CNAME record")
	}

	return &model.DNSRecord{
		RecordType: "CNAME",
		Name:       domain,
		Value:      strings.TrimSuffix(cname, "."), // Remove trailing dot
		TTL:        0,
	}, nil
}

// lookupMX queries MX records (mail servers)
func (s *DNSScanner) lookupMX(domain string) ([]*model.DNSRecord, error) {
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return nil, err
	}

	records := []*model.DNSRecord{}
	for _, mx := range mxRecords {
		records = append(records, &model.DNSRecord{
			RecordType: "MX",
			Name:       domain,
			// Format: "priority hostname"
			Value: fmt.Sprintf("%d %s", mx.Pref, strings.TrimSuffix(mx.Host, ".")),
			TTL:   0,
		})
	}

	return records, nil
}

// lookupNS queries NS records (name servers)
func (s *DNSScanner) lookupNS(domain string) ([]*model.DNSRecord, error) {
	nsRecords, err := net.LookupNS(domain)
	if err != nil {
		return nil, err
	}

	records := []*model.DNSRecord{}
	for _, ns := range nsRecords {
		records = append(records, &model.DNSRecord{
			RecordType: "NS",
			Name:       domain,
			Value:      strings.TrimSuffix(ns.Host, "."),
			TTL:        0,
		})
	}

	return records, nil
}

// lookupTXT queries TXT records (text records for verification, SPF, DKIM, etc.)
func (s *DNSScanner) lookupTXT(domain string) ([]*model.DNSRecord, error) {
	txtRecords, err := net.LookupTXT(domain)
	if err != nil {
		return nil, err
	}

	records := []*model.DNSRecord{}
	for _, txt := range txtRecords {
		records = append(records, &model.DNSRecord{
			RecordType: "TXT",
			Name:       domain,
			Value:      txt,
			TTL:        0,
		})
	}

	return records, nil
}

// ExtractIPs extracts unique IP addresses from DNS scan results
func (s *DNSScanner) ExtractIPs(records []*model.DNSRecord) []string {
	ipSet := make(map[string]bool)
	ips := []string{}

	for _, record := range records {
		if record.RecordType == "A" || record.RecordType == "AAAA" {
			if !ipSet[record.Value] {
				ipSet[record.Value] = true
				ips = append(ips, record.Value)
			}
		}
	}

	return ips
}

/*
🎓 NOTES - DNS Scanner

=== DNS (Domain Name System) ===

Purpose: Translate domain names to IP addresses and other info

Example:
  User types: www.example.com
  DNS resolves: 93.184.216.34
  Browser connects to IP

=== DNS RECORD TYPES ===

1. A Record (Address):
   - Maps domain to IPv4 address
   - Example: example.com → 93.184.216.34
   - Most common record type
   - One domain can have multiple A records (load balancing)

2. AAAA Record (IPv6 Address):
   - Maps domain to IPv6 address
   - Example: example.com → 2606:2800:220:1:248:1893:25c8:1946
   - Same as A but for newer IPv6
   - Pronounced "quad-A"

3. CNAME Record (Canonical Name):
   - Alias from one domain to another
   - Example: www.example.com → example.com
   - Cannot coexist with other records for same name
   - Common for CDN: example.com → example.cdn.com

4. MX Record (Mail Exchange):
   - Specifies mail servers for domain
   - Example: example.com → 10 mail.example.com
   - Number is priority (lower = higher priority)
   - Can have multiple MX records for redundancy

5. NS Record (Name Server):
   - Specifies authoritative name servers
   - Example: example.com → ns1.example.com
   - Usually 2-4 name servers for redundancy
   - Critical for domain resolution

6. TXT Record (Text):
   - Arbitrary text data
   - Uses:
     * SPF: Email sender authorization
       "v=spf1 include:_spf.google.com ~all"
     * DKIM: Email signature verification
     * Domain verification (Google, etc.)
     * DMARC: Email authentication policy
   - Can be very long (up to 255 chars per string, multiple strings)

7. SOA Record (Start of Authority):
   - Administrative information about zone
   - Primary name server, admin email, serial number
   - Refresh/retry timings
   - Not implemented (students can add)

=== GO'S net PACKAGE ===

Standard library provides convenient DNS functions:

```go
// Lookup IP addresses (both A and AAAA)
ips, err := net.LookupIP("example.com")
// Returns: []net.IP

// Lookup CNAME
cname, err := net.LookupCNAME("www.example.com")
// Returns: "example.com."

// Lookup MX records
mxs, err := net.LookupMX("example.com")
// Returns: []*net.MX{Pref: 10, Host: "mail.example.com."}

// Lookup NS records
nss, err := net.LookupNS("example.com")
// Returns: []*net.NS{Host: "ns1.example.com."}

// Lookup TXT records
txts, err := net.LookupTXT("example.com")
// Returns: []string{"v=spf1 ...", "google-site-verification=..."}
```

Advantages:
- ✅ Built-in (no dependencies)
- ✅ Cross-platform
- ✅ Simple API
- ✅ Uses system DNS resolver

Limitations:
- ❌ No TTL access
- ❌ No low-level control
- ❌ Can't specify DNS server

For advanced needs: github.com/miekg/dns

=== IMPLEMENTATION NOTES ===

1. Error Handling:
   - DNS lookup fails → not fatal
   - Domain might not have all record types
   - Example: No MX records → not an error, just no email
   - We collect what's available

2. IPv4 vs IPv6 Filtering:
   ```go
   if ip.To4() != nil {
       // This is IPv4
   } else {
       // This is IPv6
   }
   ```

3. Trailing Dots:
   - DNS returns: "example.com."
   - We store: "example.com"
   - Trailing dot = fully qualified domain name (FQDN)
   - We trim for consistency

4. CNAME Special Case:
   - If no CNAME exists, returns domain itself
   - We check: if cname == domain → no CNAME
   - Only return if different

=== SECURITY INSIGHTS ===

What attackers learn from DNS:

1. A/AAAA Records:
   - Find IP addresses to target
   - Identify hosting provider
   - Check for CDN usage

2. MX Records:
   - Email infrastructure
   - Potential phishing targets
   - Find mail server IPs

3. NS Records:
   - DNS infrastructure
   - Potential DNS hijacking targets
   - Find managed DNS providers

4. TXT Records:
   - SPF → Email sending IPs
   - Third-party services (Google, AWS)
   - Verification tokens (sometimes leak info)

=== IP EXTRACTION ===

From DNS results, we extract IPs:

```go
func ExtractIPs(records []*DNSRecord) []string {
    // Use map to deduplicate
    ipSet := make(map[string]bool)

    for _, record := range records {
        if record.Type == "A" || record.Type == "AAAA" {
            ipSet[record.Value] = true
        }
    }

    // Convert to slice
    ips := []string{}
    for ip := range ipSet {
        ips = append(ips, ip)
    }

    return ips
}
```

These IPs become new IP assets → can be port scanned!

=== DEMO EXAMPLES ===

1. Single Domain:
```bash
# Scan example.com
POST /assets/{id}/scan
{
  "scan_type": "dns"
}

# Results:
A      example.com      93.184.216.34
AAAA   example.com      2606:2800:220:1:...
NS     example.com      a.iana-servers.net
NS     example.com      b.iana-servers.net
TXT    example.com      v=spf1 -all
```

2. Subdomain:
```bash
# Scan mail.google.com
A      mail.google.com  142.250.185.133
MX     google.com       10 smtp.google.com
```

3. No Records:
```bash
# Scan nonexistent.example.com
Error: no such host
```

=== COMPARISON WITH COMMAND LINE ===

Our scanner vs dig/nslookup:

```bash
# Command line
dig example.com A
nslookup example.com

# Our scanner (programmatic)
scanner.Scan(asset)
```

Advantages of programmatic:
- Automated
- Store in database
- Process results
- Monitor changes over time

=== TTL LIMITATION ===

Go's net package doesn't expose TTL (Time To Live):

```go
TTL: 0  // We can't get this easily
```

TTL tells how long to cache the result.

To get TTL, need lower-level DNS library:
```go
import "github.com/miekg/dns"
// More complex but full control
```

For teaching: TTL=0 is fine (students can enhance)

=== INTEGRATION WITH SCANNING FLOW ===

DNS scan fits in the flow:

```
1. User creates domain asset: example.com
2. DNS scan triggered
3. Scanner queries A, AAAA, MX, NS, TXT, CNAME
4. Results stored in dns_records table
5. IPs extracted from A/AAAA records
6. New IP assets created (automatic or manual)
7. Port scan triggered on new IPs
```

Each DNS record is a clue for further discovery!

=== STUDENT EXERCISES ===

1. Add SOA Record Lookup:
   - Research SOA format
   - Implement lookupSOA()
   - Parse: mname, rname, serial, refresh, retry, expire

2. Add Low-Level DNS:
   - Use github.com/miekg/dns
   - Get TTL values
   - Specify custom DNS server

3. Add DNS History Tracking:
   - Compare current DNS vs previous scan
   - Detect changes (IP changed, record added)
   - Alert on suspicious changes

4. Add Reverse DNS:
   - From IP → domain name
   - Useful for IP asset enrichment

=== REAL-WORLD USAGE ===

DNS scanning is foundational for:
- 🔍 Reconnaissance (pentesting, bug bounty)
- 📊 Asset inventory (know all your IPs)
- 🔐 Security monitoring (detect unauthorized changes)
- 📧 Email security (validate SPF/DKIM/DMARC)
- 🌍 CDN detection (CNAME patterns)
- ⚠️ Incident response (compromised DNS?)

DNS is the phonebook of the internet - essential to understand!
*/
