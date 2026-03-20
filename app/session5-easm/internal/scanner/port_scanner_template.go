package scanner

import (
	"fmt"
	"log"
	"net"
	"time"

	"mini-asm/internal/model"
)

/*
⚠️⚠️⚠️ WARNING: ACTIVE SCANNING - READ THIS CAREFULLY ⚠️⚠️⚠️

SCAN CATEGORY: 🔴 ACTIVE / INTRUSIVE

This is an ACTIVE scan that directly probes target systems by attempting
TCP/UDP connections to ports. This is FUNDAMENTALLY DIFFERENT from passive scans.

LEGAL CONSIDERATIONS:
================================================================================
🚨 UNAUTHORIZED PORT SCANNING MAY BE ILLEGAL 🚨

Laws that may apply:
- 🇺🇸 Computer Fraud and Abuse Act (CFAA) - 18 U.S.C. § 1030
  "Whoever intentionally accesses a computer without authorization..."
  Penalties: Up to 10 years in prison, fines

- 🇬🇧 Computer Misuse Act 1990
  "Unauthorized access to computer material"
  Penalties: Up to 2 years imprisonment

- 🇪🇺 NIS Directive & Local Laws
  Many EU countries have specific cybercrime legislation

- Other Countries: Similar laws exist worldwide

REAL CONSEQUENCES:
- FBI investigation
- Criminal charges
- Civil lawsuits
- Job loss
- Difficulty finding employment in tech

REAL CASES:
- Security researcher Aaron Swartz - Prosecuted under CFAA
- Various penetration testers arrested for scanning without proper authorization
- Bug bounty hunters banned for exceeding scope

================================================================================

WHEN PORT SCANNING IS LEGAL:
✅ You own the systems being scanned
✅ Written authorization from system owner
✅ Penetration testing contract with clear scope
✅ Bug bounty program (within defined scope)
✅ Authorized security assessment

WHEN PORT SCANNING IS ILLEGAL:
❌ No permission from target owner
❌ "Just testing" or "learning" on third-party systems
❌ Scanning your employer/university/ISP without approval
❌ Beyond authorized scope in pentesting
❌ Treating bug bounty as blanket permission

================================================================================

FOR THIS TRAINING EXERCISE:
✅ ONLY scan: 127.0.0.1 (localhost)
✅ ONLY scan: VMs you created specifically for this purpose
❌ NEVER scan: Real websites or services
❌ NEVER scan: Your school/company network without approval
❌ NEVER scan: IP addresses you don't own

The implementation below includes safety checks to prevent unauthorized scanning.
These checks are for learning purposes. In real penetration testing, authorization
is proven through signed documents, not code checks.

================================================================================

TECHNICAL DETAILS:

Why this is active:
- Sends TCP SYN packets to target ports (initiates connections)
- Target systems log all connection attempts
- Firewalls and IDS/IPS will detect and may block
- Can trigger security alerts and incident response
- May consume target resources

What target sees:
1. Firewall logs: "SYN from YOUR_IP to target:port"
2. IDS alerts: "Port scan detected from YOUR_IP"
3. Service logs: Multiple connection attempts
4. Incident response: Investigation, potential legal action

Detection example:
```
# Snort IDS rule that will catch this
alert tcp any any -> $HOME_NET any (msg:"SCAN Port scan detected";
  flags:S; threshold:type both, track by_src, count 10, seconds 60;)
```

================================================================================
*/

// PortScanner performs port scanning on IP addresses or hostnames
//
// SCAN CATEGORY: 🔴 ACTIVE - REQUIRES AUTHORIZATION
//
// ⚠️ IMPLEMENTATION NOTE FOR STUDENTS:
// This is a TEMPLATE file showing the structure and safety considerations.
// Students should:
// 1. Read and understand all warnings above
// 2. Sign authorization form before implementing
// 3. Only test on localhost or authorized VMs
// 4. Implement safety checks below
type PortScanner struct {
	timeout       time.Duration
	maxWorkers    int
	commonPorts   []int
	authorizedIPs map[string]bool // Whitelist of IPs allowed to scan
}

// NewPortScanner creates a new port scanner with safety restrictions
func NewPortScanner() *PortScanner {
	// Log warnings every time scanner is created
	log.Println("⚠️⚠️⚠️ PORT SCANNING IS ACTIVE RECONNAISSANCE ⚠️⚠️⚠️")
	log.Println("⚠️ Port scanning without authorization may be ILLEGAL")
	log.Println("⚠️ Only scan systems you own or have written permission to scan")
	log.Println("⚠️ Unauthorized port scanning can result in legal action")
	log.Println("⚠️ For this exercise: ONLY scan 127.0.0.1 (localhost)")

	return &PortScanner{
		timeout:    5 * time.Second,
		maxWorkers: 100,
		// Common ports for initial learning
		// Students can expand this list
		commonPorts: []int{
			21,   // FTP
			22,   // SSH
			23,   // Telnet
			25,   // SMTP
			53,   // DNS
			80,   // HTTP
			110,  // POP3
			143,  // IMAP
			443,  // HTTPS
			445,  // SMB
			3306, // MySQL
			3389, // RDP
			5432, // PostgreSQL
			5900, // VNC
			8080, // HTTP Alt
			8443, // HTTPS Alt
		},
		// SAFETY: Whitelist of allowed targets
		// In training, only allow localhost
		authorizedIPs: map[string]bool{
			"127.0.0.1": true,
			"localhost": true,
			"::1":       true, // IPv6 localhost
		},
	}
}

// Type returns the scan type identifier
func (s *PortScanner) Type() model.ScanType {
	return model.ScanTypePort
}

// Scan performs port scanning on a target IP address
//
// SAFETY CHECKS:
// 1. Verify target is in authorized list
// 2. Log all scan attempts
// 3. Rate limiting to avoid DoS
// 4. Timeout to avoid hanging
func (s *PortScanner) Scan(asset *model.Asset) ([]PortScanResult, error) {
	if asset.Type != model.TypeIP && asset.Type != model.TypeDomain {
		return nil, fmt.Errorf("port scan requires IP or domain asset, got: %s", asset.Type)
	}

	target := asset.Name

	// CRITICAL SAFETY CHECK #1: Authorization
	if !s.isAuthorized(target) {
		return nil, fmt.Errorf(
			"⚠️ UNAUTHORIZED PORT SCAN BLOCKED ⚠️\n"+
				"Target: %s\n"+
				"Port scanning requires explicit authorization.\n"+
				"For this training exercise, only localhost scanning is permitted.\n"+
				"Unauthorized port scanning may violate:\n"+
				"  - Computer Fraud and Abuse Act (CFAA) - US\n"+
				"  - Computer Misuse Act - UK\n"+
				"  - Local cybercrime laws\n"+
				"To scan external systems:\n"+
				"  1. Obtain written permission from system owner\n"+
				"  2. Verify scope and permitted techniques\n"+
				"  3. Document authorization before proceeding\n"+
				"For educational purposes: Scan 127.0.0.1 only",
			target,
		)
	}

	// CRITICAL SAFETY CHECK #2: Logging
	log.Printf("🔴 ACTIVE SCAN INITIATED")
	log.Printf("   Target: %s", target)
	log.Printf("   Scan Type: Port Scan (TCP Connect)")
	log.Printf("   Authorization: Training Exercise (localhost only)")
	log.Printf("   Timestamp: %s", time.Now().Format(time.RFC3339))

	// TODO for students: Implement port scanning logic here
	//
	// Steps to implement:
	// 1. Iterate through commonPorts
	// 2. For each port, attempt TCP connection
	// 3. Use net.DialTimeout() to avoid hanging
	// 4. If connection succeeds → port is open
	// 5. Attempt service detection (banner grabbing)
	// 6. Store results
	//
	// Example structure:
	//
	// results := []PortScanResult{}
	// for _, port := range s.commonPorts {
	//     if isOpen := s.scanPort(target, port); isOpen {
	//         service := s.detectService(target, port)
	//         results = append(results, PortScanResult{
	//             Port:    port,
	//             State:   "open",
	//             Service: service,
	//         })
	//     }
	// }
	// return results, nil

	return nil, fmt.Errorf("not implemented - student exercise")
}

// isAuthorized checks if target is in the authorized list
func (s *PortScanner) isAuthorized(target string) bool {
	// Check direct match
	if s.authorizedIPs[target] {
		return true
	}

	// Resolve domain to IP and check
	ips, err := net.LookupIP(target)
	if err != nil {
		return false
	}

	for _, ip := range ips {
		if s.authorizedIPs[ip.String()] {
			return true
		}
	}

	return false
}

// scanPort attempts to connect to a single port
// TODO for students: Implement this method
func (s *PortScanner) scanPort(target string, port int) bool {
	// Example implementation:
	// address := fmt.Sprintf("%s:%d", target, port)
	// conn, err := net.DialTimeout("tcp", address, s.timeout)
	// if err != nil {
	//     return false // Port closed or filtered
	// }
	// conn.Close()
	// return true // Port open

	return false // Placeholder
}

// detectService attempts to identify the service running on an open port
// TODO for students: Implement service detection (banner grabbing)
func (s *PortScanner) detectService(target string, port int) string {
	// Example implementation:
	// 1. Connect to port
	// 2. Read initial bytes (banner)
	// 3. Match against known service signatures
	// 4. Return service name
	//
	// Known patterns:
	// - SSH: "SSH-2.0-"
	// - HTTP: "HTTP/1.1" or HTML content
	// - FTP: "220 "
	// - SMTP: "220 " with mail server info

	return "unknown" // Placeholder
}

// PortScanResult represents the result of scanning a single port
type PortScanResult struct {
	Port         int    `json:"port"`
	State        string `json:"state"`   // "open", "closed", "filtered"
	Service      string `json:"service"` // "ssh", "http", "unknown"
	Version      string `json:"version"` // e.g., "OpenSSH 8.2"
	Banner       string `json:"banner"`  // Raw banner from service
	ResponseTime int    `json:"response_time_ms"`
}

/*
================================================================================
STUDENT EXERCISE INSTRUCTIONS:
================================================================================

BEFORE STARTING:
1. Read all warnings above
2. Understand legal implications
3. Sign authorization form (instructor will provide)
4. Confirm you will ONLY scan 127.0.0.1

STEP 1: Setup Test Environment
```bash
# Start local services to scan
# Option A: Use Python's HTTP server
python3 -m http.server 8000

# Option B: Use docker to run test services
docker run -d -p 22:22 -p 80:80 -p 3306:3306 test-container
```

STEP 2: Implement scanPort() method
- Use net.DialTimeout() to attempt connection
- Handle timeout (closed) vs. success (open)
- Consider rate limiting

STEP 3: Implement detectService() method
- Read first N bytes from open port
- Match against known service signatures
- Return service name string

STEP 4: Test Implementation
```bash
# Create localhost IP asset
curl -X POST http://localhost:8080/assets \
  -d '{"name": "127.0.0.1", "type": "ip"}'

# Run port scan
curl -X POST http://localhost:8080/assets/{id}/scan \
  -d '{"scan_type": "port"}'

# Check results
curl http://localhost:8080/scan-jobs/{job_id}/results
```

STEP 5: Verify Safety
- Attempt to scan non-localhost IP
- Should see authorization error
- Verify logging is working

BONUS CHALLENGES:
1. Implement UDP scanning
2. Add OS fingerprinting
3. Integrate with nmap library
4. Add stealth scan techniques (SYN scan)
5. Implement service version detection

LEARNING OUTCOMES:
- Understand TCP/IP networking
- Experience active reconnaissance
- Learn authorization importance
- Practice ethical hacking principles
- Understand legal boundaries

================================================================================
*/
