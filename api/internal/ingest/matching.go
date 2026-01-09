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

const (
	MatchReasonExternalID   MatchReason = "external_id"
	MatchReasonHostname     MatchReason = "hostname"
	MatchReasonShortname    MatchReason = "shortname"
	MatchReasonIPAndHost    MatchReason = "ip_and_hostname"
	MatchReasonStaticIP     MatchReason = "static_ip"
	MatchReasonNoMatch      MatchReason = "no_match"
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
	useShortnames bool // Enable shortname matching (default: false)
}

// NewAssetMatcher creates a new asset matcher
func NewAssetMatcher(db *sqlx.DB, tenantID int64, source string) *AssetMatcher {
	return &AssetMatcher{
		db:       db,
		tenantID: tenantID,
		source:   source,
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
			if matchedAsset, err := m.findByExternalID(ctx, idType, normalizedID); err == nil && matchedAsset != nil {
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
		if matchedAsset, err := m.findByHostname(ctx, normalizedHostname); err == nil && matchedAsset != nil {
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
		if matchedAsset, err := m.findByShortname(ctx, shortname); err == nil && matchedAsset != nil {
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
			if matchedAsset, err := m.findByIPAndHostname(ctx, normalizedIP, normalizedHostname); err == nil && matchedAsset != nil {
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

	// 5. Try static IP (only if asset.StaticIP is true)
	if asset.StaticIP && len(asset.IPAddresses) > 0 {
		for _, ip := range asset.IPAddresses {
			normalizedIP := NormalizeIP(ip)
			if matchedAsset, err := m.findByStaticIP(ctx, normalizedIP); err == nil && matchedAsset != nil {
				log.Debug().
					Str("reason", "static_ip").
					Str("ip", normalizedIP).
					Msg("Asset matched by static IP")
				return &AssetMatchResult{
					Asset:     matchedAsset,
					Reason:    MatchReasonStaticIP,
					MatchedID: normalizedIP,
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
	// Query: select asset by joining with asset_identifiers
	// WHERE tenant_id = $1 AND id_type = $2 AND id_value = $3
	// TODO: Implement query
	return nil, fmt.Errorf("not yet implemented")
}

// findByHostname finds an asset by canonical name
func (m *AssetMatcher) findByHostname(ctx context.Context, hostname string) (*repository.Asset, error) {
	// Query: SELECT * FROM assets WHERE tenant_id = $1 AND canonical_name = $2
	// TODO: Implement query
	return nil, fmt.Errorf("not yet implemented")
}

// findByShortname finds an asset by shortname identifier
func (m *AssetMatcher) findByShortname(ctx context.Context, shortname string) (*repository.Asset, error) {
	// Query: join with asset_identifiers WHERE id_type = 'shortname_norm' AND id_value = $1
	// TODO: Implement query
	return nil, fmt.Errorf("not yet implemented")
}

// findByIPAndHostname finds an asset by IP and hostname combination
func (m *AssetMatcher) findByIPAndHostname(ctx context.Context, ip, hostname string) (*repository.Asset, error) {
	// Query: join asset with asset_identifiers twice:
	// - once for ipv4 identifier
	// - once for hostname_norm identifier
	// Both must point to the same asset
	// TODO: Implement query
	return nil, fmt.Errorf("not yet implemented")
}

// findByStaticIP finds an asset by static IP within an inactivity window
func (m *AssetMatcher) findByStaticIP(ctx context.Context, ip string) (*repository.Asset, error) {
	// Query: join with asset_identifiers WHERE id_type = 'ipv4' AND id_value = $1
	// Also check that the asset was recently seen (inactivity window)
	// TODO: Implement query
	return nil, fmt.Errorf("not yet implemented")
}

// EnableShortnames enables shortname matching (disabled by default)
func (m *AssetMatcher) EnableShortnames() {
	m.useShortnames = true
}
