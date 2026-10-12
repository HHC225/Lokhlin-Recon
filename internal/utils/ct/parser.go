// ct-----------------------------------------
// @file      : parser.go
// @time      : 2025/01/24 21:00
// -------------------------------------------

// ============================================

package ct

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/google/certificate-transparency-go/x509"
	"recon/internal/utils/ct/models"
)

type CertificateParser struct {
	maxDomains int
}

func NewCertificateParser() *CertificateParser {
	return &CertificateParser{
		maxDomains: 100,
	}
}

func (p *CertificateParser) ParseCertificateStream(rawData []byte, index int64, timestamp time.Time, logURL string) (*models.CertInfo, error) {
	if len(rawData) == 0 {
		return nil, fmt.Errorf("Certificate data is empty")
	}

	certInfo := &models.CertInfo{
		LogURL:    logURL,
		Index:     index,
		Timestamp: timestamp,
	}

	domains, err := p.ExtractDomainsFast(rawData)
	if err != nil {

		domains = []string{}
	}

	if len(domains) > p.maxDomains {
		domains = domains[:p.maxDomains]
	}

	certInfo.AllDomains = domains

	if len(domains) > 0 {
		certInfo.CommonName = domains[0]
	}

	return certInfo, nil
}

func (p *CertificateParser) ExtractDomainsFast(certData []byte) ([]string, error) {
	domains := make([]string, 0, 10)

	cert, err := x509.ParseCertificate(certData)
	if err != nil {

		return p.extractDomainsFromPatterns(certData), nil
	}

	if cert.DNSNames != nil {
		domains = append(domains, cert.DNSNames...)
	}

	if cert.Subject.CommonName != "" && !cert.IsCA {

		cnAlreadyAdded := false
		for _, domain := range domains {
			if domain == cert.Subject.CommonName {
				cnAlreadyAdded = true
				break
			}
		}

		if !cnAlreadyAdded {
			domains = append(domains, cert.Subject.CommonName)
		}
	}

	validDomains := make([]string, 0, len(domains))
	for _, domain := range domains {
		if p.isValidDomain(domain) {
			validDomains = append(validDomains, strings.ToLower(strings.TrimSpace(domain)))
		}
	}

	validDomains = p.deduplicateDomains(validDomains)

	return validDomains, nil
}

func (p *CertificateParser) extractDomainsFromPatterns(data []byte) []string {
	domains := make([]string, 0, 10)
	dataStr := string(data)

	separators := []string{"\x00", "\n", "\r", "\t", " ", ",", ";", "|"}

	for _, sep := range separators {
		parts := strings.Split(dataStr, sep)
		for _, part := range parts {
			if p.isValidDomain(part) {
				domains = append(domains, strings.ToLower(strings.TrimSpace(part)))
			}
		}
	}

	domains = p.deduplicateDomains(domains)

	return domains
}

func (p *CertificateParser) isValidDomain(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 4 || len(s) > 253 {
		return false
	}

	if !strings.Contains(s, ".") {
		return false
	}

	for _, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') &&
			!(r >= '0' && r <= '9') && r != '.' && r != '-' {
			return false
		}
	}

	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return false
	}

	tld := parts[len(parts)-1]
	if len(tld) < 2 || len(tld) > 6 {
		return false
	}

	return true
}

func (p *CertificateParser) deduplicateDomains(domains []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(domains))

	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain != "" && !seen[domain] {
			seen[domain] = true
			result = append(result, domain)
		}
	}

	return result
}

func (p *CertificateParser) CalculateCertHash(certData []byte) string {
	hash := sha256.Sum256(certData)
	return fmt.Sprintf("%x", hash)
}

func (p *CertificateParser) ValidateCertificateData(data []byte) error {
	if len(data) < 100 {
		return fmt.Errorf("Certificate data is too small")
	}

	if len(data) > 100*1024 {
		return fmt.Errorf("Certificate data is too large")
	}

	if data[0] != 0x30 {
		return fmt.Errorf("Invalid DER-encoded certificate")
	}

	return nil
}

func (p *CertificateParser) GetParserStats() *ParserStats {
	return &ParserStats{
		MaxDomains: p.maxDomains,
	}
}

type ParserStats struct {
	MaxDomains      int   `json:"max_domains"`
	TotalParsed     int64 `json:"total_parsed"`
	ParseErrors     int64 `json:"parse_errors"`
	DomainExtracted int64 `json:"domain_extracted"`
}

type StreamProcessor struct {
	parser   *CertificateParser
	matcher  *DomainMatcher
	eventBus *models.EventBus

	stats struct {
		totalProcessed int64
		totalMatched   int64
		totalErrors    int64
	}
}

func NewStreamProcessor(matcher *DomainMatcher, eventBus *models.EventBus) *StreamProcessor {
	return &StreamProcessor{
		parser:   NewCertificateParser(),
		matcher:  matcher,
		eventBus: eventBus,
	}
}

func (sp *StreamProcessor) ProcessCertificate(rawData []byte, index int64, timestamp time.Time, logURL string) error {

	certInfo, err := sp.parser.ParseCertificateStream(rawData, index, timestamp, logURL)
	if err != nil {
		sp.stats.totalErrors++
		return fmt.Errorf("Failed to parse certificate: %w", err)
	}

	sp.stats.totalProcessed++

	certEvent := &models.CertificateProcessedEvent{
		BaseEvent: models.BaseEvent{
			Type:      models.EventTypeCertificateProcessed,
			ID:        models.GenerateEventID(),
			Timestamp: time.Now(),
		},
		Certificate:    certInfo,
		ProcessingTime: 0,
		Success:        true,
		Error:          "",
	}

	go func(evt *models.CertificateProcessedEvent) {
		sp.eventBus.Publish(evt)
	}(certEvent)

	if len(certInfo.AllDomains) == 0 {
		return nil
	}

	matches := sp.matcher.MatchSubdomains(certInfo.AllDomains)
	if len(matches) == 0 {
		return nil
	}

	sp.stats.totalMatched++

	for _, match := range matches {
		match.Certificate = certInfo

		event := &models.SubdomainDiscoveredEvent{
			BaseEvent: models.BaseEvent{
				Type:      models.EventTypeSubdomainDiscovered,
				ID:        models.GenerateEventID(),
				Timestamp: time.Now(),
			},
			SubdomainEvent: models.SubdomainEvent{
				Subdomain:    match.Subdomain,
				RootDomain:   match.RootDomain,
				Certificate:  certInfo,
				DiscoveredAt: match.MatchedAt,
				EventID:      models.GenerateEventID(),
				IsNew:        true,
			},
		}

		go func(evt *models.SubdomainDiscoveredEvent) {
			sp.eventBus.Publish(evt)
		}(event)
	}

	return nil
}

func (sp *StreamProcessor) GetStats() *StreamProcessorStats {
	return &StreamProcessorStats{
		TotalProcessed: sp.stats.totalProcessed,
		TotalMatched:   sp.stats.totalMatched,
		TotalErrors:    sp.stats.totalErrors,
	}
}

type StreamProcessorStats struct {
	TotalProcessed int64 `json:"total_processed"`
	TotalMatched   int64 `json:"total_matched"`
	TotalErrors    int64 `json:"total_errors"`
}
