-- +migrate Up
-- Create views for the query framework
-- These views join tables to provide a simplified interface for dashboard queries

-- Findings view: joins finding_instances with finding_definitions, assets, aliases, and intel_cve
-- This allows querying fields like severity, cve, asset_name, epss_score, is_kev from a single view
CREATE OR REPLACE VIEW findings AS
SELECT
    fi.id,
    fi.tenant_id,
    fi.asset_id,
    fi.definition_uid,
    fi.scanner_status,
    fi.effective_status,
    fi.effective_reason,
    fi.effective_revision,
    fi.first_observed_at,
    fi.last_observed_at,
    fi.evidence_json,
    fi.created_at,
    fi.updated_at,
    -- From finding_definitions (normalize severity to lowercase for consistency)
    fd.source,
    LOWER(fd.severity_default) AS severity,
    fd.title,
    -- From assets
    a.canonical_name AS asset_name,
    -- From intel_cve (left join)
    ic.epss_score,
    ic.epss_percentile,
    ic.is_kev,
    ic.kev_date_added,
    ic.kev_due_date,
    -- Single CVE as string (for simple queries)
    (SELECT fda.alias_value FROM finding_definition_aliases fda
     WHERE fda.definition_uid = fi.definition_uid
     AND fda.alias_type = 'CVE'
     LIMIT 1) AS cve,
    -- Has CVE flag (for filtering)
    EXISTS(SELECT 1 FROM finding_definition_aliases fda
           WHERE fda.definition_uid = fi.definition_uid
           AND fda.alias_type = 'CVE') AS has_cve
FROM finding_instances fi
JOIN finding_definitions fd ON fi.definition_uid = fd.definition_uid
JOIN assets a ON fi.asset_id = a.id
LEFT JOIN intel_cve ic ON (
    SELECT fda.alias_value FROM finding_definition_aliases fda
    WHERE fda.definition_uid = fi.definition_uid
    AND fda.alias_type = 'CVE'
    LIMIT 1
) = ic.cve;

-- Assets view: extends assets with identifier fields
-- This allows querying hostname_norm and shortname_norm directly
-- Note: Hostname matching prefers identifiers matching canonical_name, then most recent
CREATE OR REPLACE VIEW assets_extended AS
SELECT
    a.id,
    a.tenant_id,
    a.canonical_name,
    a.first_seen_at,
    a.last_seen_at,
    a.owner_team_id,
    a.is_active,
    a.created_at,
    a.updated_at,
    -- Prefer hostname_norm that matches canonical_name (case-insensitive), fallback to most recent
    COALESCE(
        (SELECT ai.id_value FROM asset_identifiers ai
         WHERE ai.asset_id = a.id
         AND ai.id_type = 'hostname_norm'
         AND LOWER(ai.id_value) = LOWER(a.canonical_name)
         LIMIT 1),
        (SELECT ai.id_value FROM asset_identifiers ai
         WHERE ai.asset_id = a.id
         AND ai.id_type = 'hostname_norm'
         ORDER BY ai.last_seen_at DESC
         LIMIT 1),
        LOWER(a.canonical_name)
    ) AS hostname_norm,
    -- Shortname identifier
    (SELECT ai.id_value FROM asset_identifiers ai
     WHERE ai.asset_id = a.id
     AND ai.id_type = 'shortname_norm'
     ORDER BY ai.last_seen_at DESC
     LIMIT 1) AS shortname_norm,
    -- Primary IPv4
    (SELECT ai.id_value FROM asset_identifiers ai
     WHERE ai.asset_id = a.id
     AND ai.id_type = 'ipv4'
     ORDER BY ai.last_seen_at DESC
     LIMIT 1) AS ipv4
FROM assets a;

-- Comment on views for documentation
COMMENT ON VIEW findings IS 'Joined view of finding_instances with definitions, assets, aliases, and threat intel for dashboard queries';
COMMENT ON VIEW assets_extended IS 'Extended assets view with identifier fields for dashboard queries';
