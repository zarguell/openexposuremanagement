package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

// FindingDefinition represents a finding definition
type FindingDefinition struct {
	DefinitionUID      string    `db:"definition_uid"`
	Source             string    `db:"source"`
	SourceDefinitionID string    `db:"source_definition_id"`
	Title              string    `db:"title"`
	SeverityDefault    string    `db:"severity_default"`
	ReferencesJSON     []string  `db:"references_json"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

// FindingDefinitionAlias represents an alias for a finding definition (e.g., CVE)
type FindingDefinitionAlias struct {
	ID           int64     `db:"id"`
	DefinitionUID string   `db:"definition_uid"`
	AliasType    string    `db:"alias_type"`
	AliasValue   string    `db:"alias_value"`
	CreatedAt    time.Time `db:"created_at"`
}

// DefinitionRepository handles finding definitions and aliases
type DefinitionRepository struct {
	db *sqlx.DB
}

// NewDefinitionRepository creates a new definition repository
func NewDefinitionRepository(db *sqlx.DB) *DefinitionRepository {
	return &DefinitionRepository{db: db}
}

// UpsertDefinition creates or updates a finding definition
func (r *DefinitionRepository) UpsertDefinition(ctx context.Context, def *FindingDefinition) error {
	query := `
		INSERT INTO finding_definitions (
			definition_uid, source, source_definition_id, title,
			severity_default, references_json, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (source, source_definition_id)
		DO UPDATE SET
			title = EXCLUDED.title,
			severity_default = EXCLUDED.severity_default,
			references_json = EXCLUDED.references_json,
			updated_at = NOW()
		RETURNING definition_uid, source, source_definition_id, title, severity_default,
				  references_json, created_at, updated_at
	`

	var referencesJSON []byte
	if def.ReferencesJSON != nil {
		var err error
		referencesJSON, err = json.Marshal(def.ReferencesJSON)
		if err != nil {
			return err
		}
	}

	err := r.db.QueryRowContext(ctx, query,
		def.DefinitionUID, def.Source, def.SourceDefinitionID,
		def.Title, def.SeverityDefault, referencesJSON,
	).Scan(
		&def.DefinitionUID, &def.Source, &def.SourceDefinitionID,
		&def.Title, &def.SeverityDefault, &referencesJSON,
		&def.CreatedAt, &def.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// Unmarshal references back
	if referencesJSON != nil {
		json.Unmarshal(referencesJSON, &def.ReferencesJSON)
	}

	return nil
}

// GetDefinition retrieves a definition by its UID
func (r *DefinitionRepository) GetDefinition(ctx context.Context, definitionUID string) (*FindingDefinition, error) {
	query := `
		SELECT definition_uid, source, source_definition_id, title,
		       severity_default, references_json, created_at, updated_at
		FROM finding_definitions
		WHERE definition_uid = $1
	`

	var def FindingDefinition
	var referencesJSON []byte

	err := r.db.QueryRowContext(ctx, query, definitionUID).Scan(
		&def.DefinitionUID, &def.Source, &def.SourceDefinitionID,
		&def.Title, &def.SeverityDefault, &referencesJSON,
		&def.CreatedAt, &def.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal references
	if referencesJSON != nil {
		json.Unmarshal(referencesJSON, &def.ReferencesJSON)
	}

	return &def, nil
}

// GetBySourceAndID retrieves a definition by source and source definition ID
func (r *DefinitionRepository) GetBySourceAndID(ctx context.Context, source, sourceDefID string) (*FindingDefinition, error) {
	query := `
		SELECT definition_uid, source, source_definition_id, title,
		       severity_default, references_json, created_at, updated_at
		FROM finding_definitions
		WHERE source = $1 AND source_definition_id = $2
	`

	var def FindingDefinition
	var referencesJSON []byte

	err := r.db.QueryRowContext(ctx, query, source, sourceDefID).Scan(
		&def.DefinitionUID, &def.Source, &def.SourceDefinitionID,
		&def.Title, &def.SeverityDefault, &referencesJSON,
		&def.CreatedAt, &def.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal references
	if referencesJSON != nil {
		json.Unmarshal(referencesJSON, &def.ReferencesJSON)
	}

	return &def, nil
}

// UpsertAlias creates or updates an alias for a definition
func (r *DefinitionRepository) UpsertAlias(ctx context.Context, alias *FindingDefinitionAlias) error {
	query := `
		INSERT INTO finding_definition_aliases (definition_uid, alias_type, alias_value, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (definition_uid, alias_type, alias_value) DO NOTHING
		RETURNING id, definition_uid, alias_type, alias_value, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		alias.DefinitionUID, alias.AliasType, alias.AliasValue,
	).Scan(
		&alias.ID, &alias.DefinitionUID, &alias.AliasType,
		&alias.AliasValue, &alias.CreatedAt,
	)

	if err == sql.ErrNoRows {
		// Conflict occurred, fetch existing record
		return r.getExistingAlias(ctx, alias)
	}

	return err
}

// getExistingAlias fetches an existing alias
func (r *DefinitionRepository) getExistingAlias(ctx context.Context, alias *FindingDefinitionAlias) error {
	query := `
		SELECT id, definition_uid, alias_type, alias_value, created_at
		FROM finding_definition_aliases
		WHERE definition_uid = $1 AND alias_type = $2 AND alias_value = $3
	`

	return r.db.QueryRowContext(ctx, query,
		alias.DefinitionUID, alias.AliasType, alias.AliasValue,
	).Scan(
		&alias.ID, &alias.DefinitionUID, &alias.AliasType,
		&alias.AliasValue, &alias.CreatedAt,
	)
}

// GetAliasesForDefinition retrieves all aliases for a definition
func (r *DefinitionRepository) GetAliasesForDefinition(ctx context.Context, definitionUID string) ([]FindingDefinitionAlias, error) {
	query := `
		SELECT id, definition_uid, alias_type, alias_value, created_at
		FROM finding_definition_aliases
		WHERE definition_uid = $1
		ORDER BY alias_type, alias_value
	`

	rows, err := r.db.QueryContext(ctx, query, definitionUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aliases []FindingDefinitionAlias
	for rows.Next() {
		var alias FindingDefinitionAlias
		if err := rows.Scan(
			&alias.ID, &alias.DefinitionUID, &alias.AliasType,
			&alias.AliasValue, &alias.CreatedAt,
		); err != nil {
			return nil, err
		}
		aliases = append(aliases, alias)
	}

	return aliases, nil
}

// DeleteAliasesForDefinition removes all aliases for a definition
func (r *DefinitionRepository) DeleteAliasesForDefinition(ctx context.Context, definitionUID string) error {
	query := `DELETE FROM finding_definition_aliases WHERE definition_uid = $1`
	_, err := r.db.ExecContext(ctx, query, definitionUID)
	return err
}
