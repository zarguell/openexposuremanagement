package ingest

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// MatchReason describes how an asset was matched
type MatchReason string

// Match reason constants
const (
	MatchReasonExternalID MatchReason = "external_id"     // Asset matched by external ID
	MatchReasonHostname   MatchReason = "hostname"        // Asset matched by hostname
	MatchReasonShortname  MatchReason = "shortname"       // Asset matched by shortname
	MatchReasonIPAndHost  MatchReason = "ip_and_hostname" // Asset matched by IP and hostname
	MatchReasonStaticIP   MatchReason = "static_ip"       // Asset matched by static IP
	MatchReasonNoMatch    MatchReason = "no_match"        // No matching asset found
)

// AssetMatchResult contains the result of asset matching
type AssetMatchResult struct {
	Asset     *repository.Asset
	Reason    MatchReason
	MatchedID string // The identifier value that was matched
	NewAsset  bool   // True if a new asset was created
}

// AssetMatcher performs deterministic asset matching
type AssetMatcher struct {
	db            *sqlx.DB
	tenantID      int64
	source        string
	repo          *repository.AssetRepository
	useShortnames bool // Enable shortname matching (default: false)
}

// NewAssetMatcher creates a new asset matcher
func NewAssetMatcher(db *sqlx.DB, tenantID int64, source string) *AssetMatcher {
	return &AssetMatcher{
		db:       db,
		tenantID: tenantID,
		source:   source,
		repo:     repository.NewAssetRepository(db),
	}
}

// MatchAsset matches an asset using the deterministic algorithm
// Order: external_ids → hostname → shortname (optional) → IP+hostname or static IP
func (m *AssetMatcher) MatchAsset(ctx context.Context, asset *VMAsset) (*AssetMatchResult, error) {
	log.Debug().
		Str("source", m.source).
		Str("hostname", asset.Hostname).
		Msg("Matching asset")

	// 1. Try external IDs (strongest match)
	if len(asset.ExternalIDs) > 0 {
		for idType, idValue := range asset.ExternalIDs {
			normalizedID := NormalizeExternalID(idValue)
			matchedAsset, err := m.findByExternalID(ctx, idType, normalizedID)
			if err == nil && matchedAsset != nil {
				log.Debug().
					Str("reason", "external_id").
					Str("id_type", idType).
					Str("id_value", normalizedID).
					Msg("Asset matched by external ID")
				return &AssetMatchResult{
					Asset:     matchedAsset,
					Reason:    MatchReasonExternalID,
					MatchedID: normalizedID,
					NewAsset:  false,
				}, nil
			}
		}
	}

	// 2. Try hostname
	if asset.Hostname != "" {
		normalizedHostname := NormalizeHostname(asset.Hostname)
		matchedAsset, err := m.findByHostname(ctx, normalizedHostname)
		if err == nil && matchedAsset != nil {
			log.Debug().
				Str("reason", "hostname").
				Str("hostname", normalizedHostname).
				Msg("Asset matched by hostname")
			return &AssetMatchResult{
				Asset:     matchedAsset,
				Reason:    MatchReasonHostname,
				MatchedID: normalizedHostname,
				NewAsset:  false,
			}, nil
		}
	}

	// 3. Try shortname (optional)
	if m.useShortnames && asset.Hostname != "" {
		shortname := NormalizeShortname(asset.Hostname)
		matchedAsset, err := m.findByShortname(ctx, shortname)
		if err == nil && matchedAsset != nil {
			log.Debug().
				Str("reason", "shortname").
				Str("shortname", shortname).
				Msg("Asset matched by shortname")
			return &AssetMatchResult{
				Asset:     matchedAsset,
				Reason:    MatchReasonShortname,
				MatchedID: shortname,
				NewAsset:  false,
			}, nil
		}
	}

	// 4. Try IP + hostname (conditional matching)
	if len(asset.IPAddresses) > 0 && asset.Hostname != "" {
		normalizedHostname := NormalizeHostname(asset.Hostname)
		for _, ip := range asset.IPAddresses {
			normalizedIP := NormalizeIP(ip)
			matchedAsset, err := m.findByIPAndHostname(ctx, normalizedIP, normalizedHostname)
			if err == nil && matchedAsset != nil {
				log.Debug().
					Str("reason", "ip_and_hostname").
					Str("ip", normalizedIP).
					Str("hostname", normalizedHostname).
					Msg("Asset matched by IP + hostname")
				return &AssetMatchResult{
					Asset:     matchedAsset,
					Reason:    MatchReasonIPAndHost,
					MatchedID: fmt.Sprintf("%s+%s", normalizedIP, normalizedHostname),
					NewAsset:  false,
				}, nil
			}
		}
	}

	// No match found
	log.Debug().Msg("No asset match found, will create new asset")
	return &AssetMatchResult{
		Asset:    nil,
		Reason:   MatchReasonNoMatch,
		NewAsset: true,
	}, nil
}

// findByExternalID finds an asset by external ID
func (m *AssetMatcher) findByExternalID(ctx context.Context, idType, idValue string) (*repository.Asset, error) {
	// Prefix external_id type with namespacing
	idType = "external_id:" + idType
	return m.repo.FindByIdentifier(ctx, m.tenantID, idType, idValue)
}

// findByHostname finds an asset by canonical name
func (m *AssetMatcher) findByHostname(ctx context.Context, hostname string) (*repository.Asset, error) {
	return m.repo.GetByCanonicalName(ctx, m.tenantID, hostname)
}

// findByShortname finds an asset by shortname identifier
func (m *AssetMatcher) findByShortname(ctx context.Context, shortname string) (*repository.Asset, error) {
	return m.repo.FindByIdentifier(ctx, m.tenantID, "shortname_norm", shortname)
}

// findByIPAndHostname finds an asset by IP and hostname combination
func (m *AssetMatcher) findByIPAndHostname(ctx context.Context, ip, hostname string) (*repository.Asset, error) {
	return m.repo.FindByIPAndHostname(ctx, m.tenantID, ip, hostname)
}

// EnableShortnames enables shortname matching (disabled by default)
func (m *AssetMatcher) EnableShortnames() {
	m.useShortnames = true
}
