package genrefreshcontracts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
)

type authoredDomainInventory struct {
	RefreshPayloadType *string `json:"refreshPayloadType"`
}

type authoredDomainRegistration struct {
	Domain string `json:"domain"`
}

func loadContractDomains() ([]domainSpec, error) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("locate refresh domain contract: caller path unavailable")
	}
	contractPath := filepath.Join(filepath.Dir(sourceFile), "..", "..", "refresh", "domain", "refresh-domain-contract.json")
	contents, err := os.ReadFile(contractPath)
	if err != nil {
		return nil, fmt.Errorf("read refresh domain contract: %w", err)
	}
	var authored struct {
		DomainInventory map[string]authoredDomainInventory `json:"domainInventory"`
		Domains         []authoredDomainRegistration       `json:"domains"`
	}
	if err := json.Unmarshal(contents, &authored); err != nil {
		return nil, fmt.Errorf("decode refresh domain contract: %w", err)
	}

	result := make([]domainSpec, 0, len(authored.Domains))
	seen := make(map[string]struct{}, len(authored.Domains))
	for _, registration := range authored.Domains {
		domain, err := contractDomainSpec(registration, authored.DomainInventory, seen)
		if err != nil {
			return nil, err
		}
		result = append(result, domain)
	}
	if len(seen) != len(authored.DomainInventory) {
		return nil, fmt.Errorf("refresh domain contract has %d inventory entries but %d registrations", len(authored.DomainInventory), len(seen))
	}
	return result, nil
}

func contractDomainSpec(
	registration authoredDomainRegistration,
	inventoryByDomain map[string]authoredDomainInventory,
	seen map[string]struct{},
) (domainSpec, error) {
	if registration.Domain == "" {
		return domainSpec{}, fmt.Errorf("refresh domain contract contains an empty domain registration")
	}
	if _, duplicate := seen[registration.Domain]; duplicate {
		return domainSpec{}, fmt.Errorf("refresh domain contract registers %q more than once", registration.Domain)
	}
	inventory, ok := inventoryByDomain[registration.Domain]
	if !ok {
		return domainSpec{}, fmt.Errorf("refresh domain %q is missing from domainInventory", registration.Domain)
	}
	seen[registration.Domain] = struct{}{}
	if inventory.RefreshPayloadType == nil {
		return domainSpec{domain: registration.Domain, frontendOwned: true}, nil
	}
	if *inventory.RefreshPayloadType == "" {
		return domainSpec{}, fmt.Errorf("refresh domain %q has an empty refreshPayloadType", registration.Domain)
	}
	return domainSpec{domain: registration.Domain, payload: *inventory.RefreshPayloadType}, nil
}

func validateDomainPayloadTypes(domains []domainSpec, typesByName map[string]reflect.Type) error {
	for _, domain := range domains {
		if domain.frontendOwned {
			continue
		}
		if _, ok := typesByName[domain.payload]; !ok {
			return fmt.Errorf("refresh domain %q references unregistered payload type %q", domain.domain, domain.payload)
		}
	}
	return nil
}
