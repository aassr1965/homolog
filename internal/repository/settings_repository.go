package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/domain"
)

// SettingsRepository gère la persistance et le versioning temporel ALCOA+ des paramètres du tenant
type SettingsRepository struct {
	pool *pgxpool.Pool
}

// NewSettingsRepository instancie le repository des paramètres
func NewSettingsRepository(pool *pgxpool.Pool) *SettingsRepository {
	return &SettingsRepository{pool: pool}
}

// EnsureTenantDefaults vérifie et initialise les paramètres par défaut si nécessaire
func (r *SettingsRepository) EnsureTenantDefaults(ctx context.Context, orgID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "SELECT seed_tenant_default_settings($1)", orgID)
	return err
}

// GetActiveSettingsMap retourne l'ensemble des paramètres actifs sous forme de dictionnaire indexé par clé
func (r *SettingsRepository) GetActiveSettingsMap(ctx context.Context, orgID uuid.UUID) (map[string]*domain.TenantRegulatorySetting, error) {
	list, err := r.GetActiveSettingsList(ctx, orgID)
	if err != nil {
		return nil, err
	}
	res := make(map[string]*domain.TenantRegulatorySetting, len(list))
	for _, s := range list {
		res[s.SettingKey] = s
	}
	return res, nil
}

// GetActiveSettingsList retourne la liste de tous les paramètres actuellement en vigueur pour l'organisation
func (r *SettingsRepository) GetActiveSettingsList(ctx context.Context, orgID uuid.UUID) ([]*domain.TenantRegulatorySetting, error) {
	query := `
		SELECT s.id, s.organization_id, s.setting_key, s.setting_value, s.data_type, s.category,
		       s.label_fr, s.label_ar, s.description, s.legal_reference, s.valid_from, s.valid_to,
		       s.is_active, s.pending_value, s.pending_reason, s.pending_proposed_by,
		       CONCAT(u1.first_name, ' ', u1.last_name) AS pending_proposer_name,
		       s.status, s.approved_by,
		       CONCAT(u2.first_name, ' ', u2.last_name) AS approved_by_name,
		       s.approved_at, s.created_at, s.updated_at
		FROM tenant_regulatory_settings s
		LEFT JOIN users u1 ON u1.id = s.pending_proposed_by
		LEFT JOIN users u2 ON u2.id = s.approved_by
		WHERE s.organization_id = $1 AND s.is_active = TRUE AND s.valid_to IS NULL
		ORDER BY s.category, s.setting_key
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération settings actifs: %w", err)
	}
	defer rows.Close()

	var settings []*domain.TenantRegulatorySetting
	for rows.Next() {
		s := &domain.TenantRegulatorySetting{}
		err := rows.Scan(
			&s.ID, &s.OrganizationID, &s.SettingKey, &s.SettingValue, &s.DataType, &s.Category,
			&s.LabelFR, &s.LabelAR, &s.Description, &s.LegalReference, &s.ValidFrom, &s.ValidTo,
			&s.IsActive, &s.PendingValue, &s.PendingReason, &s.PendingProposedBy,
			&s.PendingProposerName,
			&s.Status, &s.ApprovedBy,
			&s.ApprovedByName,
			&s.ApprovedAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		settings = append(settings, s)
	}
	return settings, nil
}

// GetSettingByKey retourne le paramètre actif d'une clé spécifique
func (r *SettingsRepository) GetSettingByKey(ctx context.Context, orgID uuid.UUID, key string) (*domain.TenantRegulatorySetting, error) {
	query := `
		SELECT s.id, s.organization_id, s.setting_key, s.setting_value, s.data_type, s.category,
		       s.label_fr, s.label_ar, s.description, s.legal_reference, s.valid_from, s.valid_to,
		       s.is_active, s.pending_value, s.pending_reason, s.pending_proposed_by,
		       CONCAT(u1.first_name, ' ', u1.last_name) AS pending_proposer_name,
		       s.status, s.approved_by,
		       CONCAT(u2.first_name, ' ', u2.last_name) AS approved_by_name,
		       s.approved_at, s.created_at, s.updated_at
		FROM tenant_regulatory_settings s
		LEFT JOIN users u1 ON u1.id = s.pending_proposed_by
		LEFT JOIN users u2 ON u2.id = s.approved_by
		WHERE s.organization_id = $1 AND s.setting_key = $2 AND s.is_active = TRUE AND s.valid_to IS NULL
		LIMIT 1
	`
	s := &domain.TenantRegulatorySetting{}
	err := r.pool.QueryRow(ctx, query, orgID, key).Scan(
		&s.ID, &s.OrganizationID, &s.SettingKey, &s.SettingValue, &s.DataType, &s.Category,
		&s.LabelFR, &s.LabelAR, &s.Description, &s.LegalReference, &s.ValidFrom, &s.ValidTo,
		&s.IsActive, &s.PendingValue, &s.PendingReason, &s.PendingProposedBy,
		&s.PendingProposerName,
		&s.Status, &s.ApprovedBy,
		&s.ApprovedByName,
		&s.ApprovedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur récupération setting %s: %w", key, err)
	}
	return s, nil
}

// GetSettingHistory retourne l'historique complet d'un paramètre (GAMP 5 / ALCOA+ Audit Trail)
func (r *SettingsRepository) GetSettingHistory(ctx context.Context, orgID uuid.UUID, key string) ([]*domain.TenantRegulatorySetting, error) {
	query := `
		SELECT s.id, s.organization_id, s.setting_key, s.setting_value, s.data_type, s.category,
		       s.label_fr, s.label_ar, s.description, s.legal_reference, s.valid_from, s.valid_to,
		       s.is_active, s.pending_value, s.pending_reason, s.pending_proposed_by,
		       CONCAT(u1.first_name, ' ', u1.last_name) AS pending_proposer_name,
		       s.status, s.approved_by,
		       CONCAT(u2.first_name, ' ', u2.last_name) AS approved_by_name,
		       s.approved_at, s.created_at, s.updated_at
		FROM tenant_regulatory_settings s
		LEFT JOIN users u1 ON u1.id = s.pending_proposed_by
		LEFT JOIN users u2 ON u2.id = s.approved_by
		WHERE s.organization_id = $1 AND s.setting_key = $2
		ORDER BY s.valid_from DESC
	`
	rows, err := r.pool.Query(ctx, query, orgID, key)
	if err != nil {
		return nil, fmt.Errorf("erreur historique setting %s: %w", key, err)
	}
	defer rows.Close()

	var history []*domain.TenantRegulatorySetting
	for rows.Next() {
		s := &domain.TenantRegulatorySetting{}
		err := rows.Scan(
			&s.ID, &s.OrganizationID, &s.SettingKey, &s.SettingValue, &s.DataType, &s.Category,
			&s.LabelFR, &s.LabelAR, &s.Description, &s.LegalReference, &s.ValidFrom, &s.ValidTo,
			&s.IsActive, &s.PendingValue, &s.PendingReason, &s.PendingProposedBy,
			&s.PendingProposerName,
			&s.Status, &s.ApprovedBy,
			&s.ApprovedByName,
			&s.ApprovedAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan history setting: %w", err)
		}
		history = append(history, s)
	}
	return history, nil
}

// ProposeSettingUpdate enregistre une proposition de modification (Maker) nécessitant le visa DT
func (r *SettingsRepository) ProposeSettingUpdate(ctx context.Context, orgID uuid.UUID, key, newValue, reason string, proposedBy uuid.UUID) error {
	query := `
		UPDATE tenant_regulatory_settings
		SET pending_value = $3,
		    pending_reason = $4,
		    pending_proposed_by = $5,
		    status = 'PENDING_APPROVAL_DT',
		    updated_at = NOW()
		WHERE organization_id = $1 AND setting_key = $2 AND is_active = TRUE AND valid_to IS NULL
	`
	tag, err := r.pool.Exec(ctx, query, orgID, key, newValue, reason, proposedBy)
	if err != nil {
		return fmt.Errorf("erreur proposition mise à jour setting %s: %w", key, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("aucun paramètre actif trouvé pour la clé %s", key)
	}
	return nil
}

// ApproveAndApplySetting approuve la proposition en attente et crée une nouvelle version temporelle (Approver - DT)
func (r *SettingsRepository) ApproveAndApplySetting(ctx context.Context, orgID uuid.UUID, key string, approvedBy uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Récupérer le paramètre actif avec sa proposition
	var current domain.TenantRegulatorySetting
	queryCurrent := `
		SELECT id, setting_value, data_type, category, label_fr, label_ar, description, legal_reference,
		       pending_value, pending_reason
		FROM tenant_regulatory_settings
		WHERE organization_id = $1 AND setting_key = $2 AND is_active = TRUE AND valid_to IS NULL
		FOR UPDATE
	`
	err = tx.QueryRow(ctx, queryCurrent, orgID, key).Scan(
		&current.ID, &current.SettingValue, &current.DataType, &current.Category, &current.LabelFR,
		&current.LabelAR, &current.Description, &current.LegalReference, &current.PendingValue, &current.PendingReason,
	)
	if err != nil {
		return fmt.Errorf("paramètre en attente non trouvé pour %s: %w", key, err)
	}
	if current.PendingValue == nil || *current.PendingValue == "" {
		return fmt.Errorf("aucune modification en attente d'approbation pour %s", key)
	}

	now := time.Now().UTC()

	// 2. Clôturer la version précédente (valid_to = now, is_active = false, status = HISTORICAL)
	closeQuery := `
		UPDATE tenant_regulatory_settings
		SET valid_to = $3,
		    is_active = FALSE,
		    status = 'HISTORICAL',
		    pending_value = NULL,
		    pending_reason = NULL,
		    pending_proposed_by = NULL,
		    updated_at = $3
		WHERE id = $1 AND organization_id = $2
	`
	if _, err := tx.Exec(ctx, closeQuery, current.ID, orgID, now); err != nil {
		return fmt.Errorf("erreur clôture ancienne version: %w", err)
	}

	// 3. Insérer la nouvelle version en vigueur (valid_from = now, valid_to = NULL, is_active = true)
	insertQuery := `
		INSERT INTO tenant_regulatory_settings (
			organization_id, setting_key, setting_value, data_type, category,
			label_fr, label_ar, description, legal_reference, valid_from, valid_to,
			is_active, status, approved_by, approved_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, NULL,
			TRUE, 'ACTIVE', $11, $10, $10, $10
		)
	`
	_, err = tx.Exec(ctx, insertQuery,
		orgID, key, *current.PendingValue, current.DataType, current.Category,
		current.LabelFR, current.LabelAR, current.Description, current.LegalReference, now,
		approvedBy,
	)
	if err != nil {
		return fmt.Errorf("erreur insertion nouvelle version: %w", err)
	}

	return tx.Commit(ctx)
}

// DirectApplySetting applique immédiatement une nouvelle valeur avec versioning (ex: DT ou action admin directe)
func (r *SettingsRepository) DirectApplySetting(ctx context.Context, orgID uuid.UUID, key, newValue, reason string, approvedBy uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var current domain.TenantRegulatorySetting
	queryCurrent := `
		SELECT id, data_type, category, label_fr, label_ar, description, legal_reference
		FROM tenant_regulatory_settings
		WHERE organization_id = $1 AND setting_key = $2 AND is_active = TRUE AND valid_to IS NULL
		FOR UPDATE
	`
	err = tx.QueryRow(ctx, queryCurrent, orgID, key).Scan(
		&current.ID, &current.DataType, &current.Category, &current.LabelFR,
		&current.LabelAR, &current.Description, &current.LegalReference,
	)
	if err != nil {
		return fmt.Errorf("paramètre actif non trouvé pour %s: %w", key, err)
	}

	now := time.Now().UTC()

	// 1. Clôturer l'actuelle
	closeQuery := `
		UPDATE tenant_regulatory_settings
		SET valid_to = $3,
		    is_active = FALSE,
		    status = 'HISTORICAL',
		    updated_at = $3
		WHERE id = $1 AND organization_id = $2
	`
	if _, err := tx.Exec(ctx, closeQuery, current.ID, orgID, now); err != nil {
		return fmt.Errorf("erreur clôture version précédente: %w", err)
	}

	// 2. Créer la nouvelle
	insertQuery := `
		INSERT INTO tenant_regulatory_settings (
			organization_id, setting_key, setting_value, data_type, category,
			label_fr, label_ar, description, legal_reference, valid_from, valid_to,
			is_active, status, approved_by, approved_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, NULL,
			TRUE, 'ACTIVE', $11, $10, $10, $10
		)
	`
	_, err = tx.Exec(ctx, insertQuery,
		orgID, key, newValue, current.DataType, current.Category,
		current.LabelFR, current.LabelAR, current.Description, current.LegalReference, now,
		approvedBy,
	)
	if err != nil {
		return fmt.Errorf("erreur insertion nouvelle version directe: %w", err)
	}

	return tx.Commit(ctx)
}
