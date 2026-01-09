package ingest

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// UpsertAssetResult contains the result of an asset upsert operation
type UpsertAssetResult struct {
	Asset    *repository.Asset
	NewAsset bool
	Reason   MatchReason
}

// UpsertAsset creates or updates an asset based on matching rules
func UpsertAsset(ctx context.Context, db *sqlx.DB, tenantID int64, source string, vmAsset *VMAsset, scannedAt time.Time) (*UpsertAssetResult, error) {
	// Create matcher
	matcher := NewAssetMatcher(db, tenantID, source)

	// Try to match existing asset
	matchResult, err := matcher.MatchAsset(ctx, vmAsset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to match asset")
		return nil, err
	}

	assetRepo := repository.NewAssetRepository(db)
	var asset *repository.Asset
	newAsset := matchResult.NewAsset

	if newAsset {
		// Create new asset
		asset, err = createNewAsset(ctx, assetRepo, tenantID, vmAsset, scannedAt)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create new asset")
			return nil, err
		}
		log.Info().
			Int64("asset_id", asset.ID).
			Str("canonical_name", asset.CanonicalName).
			Msg("Created new asset")
	} else {
		// Update existing asset
		asset = matchResult.Asset
		err = updateExistingAsset(ctx, assetRepo, asset, vmAsset, source, scannedAt)
		if err != nil {
			log.Error().Err(err).Int64("asset_id", asset.ID).Msg("Failed to update asset")
			return nil, err
		}
		log.Debug().
			Int64("asset_id", asset.ID).
			Str("canonical_name", asset.CanonicalName).
			Str("match_reason", string(matchResult.Reason)).
			Msg("Updated existing asset")
	}

	// Upsert identifiers
	err = upsertIdentifiers(ctx, assetRepo, tenantID, asset.ID, vmAsset, source, scannedAt)
	if err != nil {
		log.Error().Err(err).Int64("asset_id", asset.ID).Msg("Failed to upsert identifiers")
		return nil, err
	}

	return &UpsertAssetResult{
		Asset:    asset,
		NewAsset: newAsset,
		Reason:   matchResult.Reason,
	}, nil
}

// createNewAsset creates a new asset from VM asset data
func createNewAsset(ctx context.Context, repo *repository.AssetRepository, tenantID int64, vmAsset *VMAsset, scannedAt time.Time) (*repository.Asset, error) {
	// Determine canonical name (use hostname if available, otherwise generate one)
	canonicalName := ""
	if vmAsset.Hostname != "" {
		canonicalName = NormalizeHostname(vmAsset.Hostname)
	} else if len(vmAsset.IPAddresses) > 0 {
		// Use first IP as canonical name if no hostname
		canonicalName = NormalizeIP(vmAsset.IPAddresses[0])
	} else if len(vmAsset.ExternalIDs) > 0 {
		// Use first external ID as fallback
		for _, v := range vmAsset.ExternalIDs {
			canonicalName = NormalizeExternalID(v)
			break
		}
	}

	if canonicalName == "" {
		return nil, ValidationError{Field: "asset", Message: "unable to determine canonical name"}
	}

	asset := &repository.Asset{
		TenantID:      tenantID,
		CanonicalName: canonicalName,
		FirstSeenAt:   scannedAt,
		LastSeenAt:    scannedAt,
		IsActive:      true,
	}

	err := repo.Create(ctx, asset)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

// updateExistingAsset updates an existing asset with new data
func updateExistingAsset(ctx context.Context, repo *repository.AssetRepository, asset *repository.Asset, vmAsset *VMAsset, source string, scannedAt time.Time) error {
	// Update last_seen_at if newer
	if scannedAt.After(asset.LastSeenAt) {
		err := repo.UpdateLastSeen(ctx, asset.ID, scannedAt)
		if err != nil {
			return err
		}
		asset.LastSeenAt = scannedAt
	}

	// Update canonical_name if hostname changed and is newer/better
	if vmAsset.Hostname != "" {
		newCanonicalName := NormalizeHostname(vmAsset.Hostname)
		if newCanonicalName != asset.CanonicalName {
			// Update canonical name
			err := repo.UpdateCanonicalName(ctx, asset.ID, newCanonicalName)
			if err != nil {
				return err
			}
			asset.CanonicalName = newCanonicalName
		}
	}

	return nil
}

// upsertIdentifiers creates or updates identifiers for an asset
func upsertIdentifiers(ctx context.Context, repo *repository.AssetRepository, tenantID, assetID int64, vmAsset *VMAsset, source string, scannedAt time.Time) error {
	// 1. Hostname identifier
	if vmAsset.Hostname != "" {
		if err := upsertHostnameIdentifiers(ctx, repo, tenantID, assetID, vmAsset.Hostname, source, scannedAt); err != nil {
			return err
		}
	}

	// 2. IP address identifiers
	for _, ip := range vmAsset.IPAddresses {
		normalizedIP := NormalizeIP(ip)
		if normalizedIP == "" {
			continue
		}
		if err := repo.UpsertIdentifier(ctx, tenantID, assetID, "ipv4", normalizedIP, source, scannedAt, scannedAt); err != nil {
			return err
		}
	}

	// 3. External ID identifiers
	for idType, idValue := range vmAsset.ExternalIDs {
		normalizedID := NormalizeExternalID(idValue)
		if normalizedID == "" {
			continue
		}
		// Namespace the external ID type
		namespacedType := "external_id:" + idType
		if err := repo.UpsertIdentifier(ctx, tenantID, assetID, namespacedType, normalizedID, source, scannedAt, scannedAt); err != nil {
			return err
		}
	}

	return nil
}

// upsertHostnameIdentifiers upserts hostname and shortname identifiers
func upsertHostnameIdentifiers(ctx context.Context, repo *repository.AssetRepository, tenantID, assetID int64, hostname, source string, scannedAt time.Time) error {
	normalizedHostname := NormalizeHostname(hostname)
	if err := repo.UpsertIdentifier(ctx, tenantID, assetID, "hostname_norm", normalizedHostname, source, scannedAt, scannedAt); err != nil {
		return err
	}

	// Shortname identifier (only if different from hostname)
	shortname := NormalizeShortname(hostname)
	if shortname != "" && shortname != normalizedHostname {
		if err := repo.UpsertIdentifier(ctx, tenantID, assetID, "shortname_norm", shortname, source, scannedAt, scannedAt); err != nil {
			return err
		}
	}

	return nil
}
