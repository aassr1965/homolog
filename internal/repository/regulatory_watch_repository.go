package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"homolog/internal/domain"
)

// RegulatoryWatchRepository gère la persistance de la veille juridique et des textes JORADP
type RegulatoryWatchRepository struct {
	pool *pgxpool.Pool
}

// NewRegulatoryWatchRepository instancie le repository de veille réglementaire
func NewRegulatoryWatchRepository(pool *pgxpool.Pool) *RegulatoryWatchRepository {
	return &RegulatoryWatchRepository{pool: pool}
}

// CreateRegulatoryWatchItem enregistre un nouveau texte de loi, décret ou arrêté
func (r *RegulatoryWatchRepository) CreateRegulatoryWatchItem(ctx context.Context, item *domain.RegulatoryWatchItem) (*domain.RegulatoryWatchItem, error) {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now

	query := `
		INSERT INTO regulatory_watch_items (
			id, organization_id, title_fr, title_ar, legal_reference, joradp_number,
			published_date, effective_date, category, impact_level, summary_fr,
			impact_analysis, compliance_status, source_url, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17
		)
	`
	_, err := r.pool.Exec(ctx, query,
		item.ID, item.OrganizationID, item.TitleFR, item.TitleAR, item.LegalReference, item.JORADPNumber,
		item.PublishedDate, item.EffectiveDate, string(item.Category), string(item.ImpactLevel), item.SummaryFR,
		item.ImpactAnalysis, string(item.ComplianceStatus), item.SourceURL, item.CreatedBy, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("échec création texte de veille réglementaire: %w", err)
	}

	return r.GetRegulatoryWatchItem(ctx, item.OrganizationID, item.ID)
}

// GetRegulatoryWatchItem retourne un texte officiel par son ID
func (r *RegulatoryWatchRepository) GetRegulatoryWatchItem(ctx context.Context, orgID, itemID uuid.UUID) (*domain.RegulatoryWatchItem, error) {
	query := `
		SELECT 
			id, organization_id, title_fr, COALESCE(title_ar, ''), legal_reference, COALESCE(joradp_number, ''),
			published_date, effective_date, category, impact_level, summary_fr,
			COALESCE(impact_analysis, ''), compliance_status, COALESCE(source_url, ''),
			created_by, created_at, updated_at
		FROM regulatory_watch_items
		WHERE organization_id = $1 AND id = $2
	`

	var item domain.RegulatoryWatchItem
	var catStr, impactStr, statusStr string

	err := r.pool.QueryRow(ctx, query, orgID, itemID).Scan(
		&item.ID, &item.OrganizationID, &item.TitleFR, &item.TitleAR, &item.LegalReference, &item.JORADPNumber,
		&item.PublishedDate, &item.EffectiveDate, &catStr, &impactStr, &item.SummaryFR,
		&item.ImpactAnalysis, &statusStr, &item.SourceURL,
		&item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("échec lecture texte de veille réglementaire: %w", err)
	}

	item.Category = domain.RegulatoryCategory(catStr)
	item.ImpactLevel = domain.ImpactLevel(impactStr)
	item.ComplianceStatus = domain.RegulatoryComplianceStatus(statusStr)
	return &item, nil
}

// ListRegulatoryWatchItems liste les textes réglementaires avec filtres
func (r *RegulatoryWatchRepository) ListRegulatoryWatchItems(ctx context.Context, orgID uuid.UUID, category, impact, status string) ([]domain.RegulatoryWatchItem, error) {
	query := `
		SELECT 
			id, organization_id, title_fr, COALESCE(title_ar, ''), legal_reference, COALESCE(joradp_number, ''),
			published_date, effective_date, category, impact_level, summary_fr,
			COALESCE(impact_analysis, ''), compliance_status, COALESCE(source_url, ''),
			created_by, created_at, updated_at
		FROM regulatory_watch_items
		WHERE organization_id = $1
	`
	args := []any{orgID}
	argIdx := 2

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}

	if impact != "" {
		query += fmt.Sprintf(" AND impact_level = $%d", argIdx)
		args = append(args, impact)
		argIdx++
	}

	if status != "" {
		query += fmt.Sprintf(" AND compliance_status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	query += " ORDER BY published_date DESC, created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("échec liste veille réglementaire: %w", err)
	}
	defer rows.Close()

	var list []domain.RegulatoryWatchItem
	for rows.Next() {
		var item domain.RegulatoryWatchItem
		var catStr, impactStr, statusStr string

		err := rows.Scan(
			&item.ID, &item.OrganizationID, &item.TitleFR, &item.TitleAR, &item.LegalReference, &item.JORADPNumber,
			&item.PublishedDate, &item.EffectiveDate, &catStr, &impactStr, &item.SummaryFR,
			&item.ImpactAnalysis, &statusStr, &item.SourceURL,
			&item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("échec scan texte de veille: %w", err)
		}

		item.Category = domain.RegulatoryCategory(catStr)
		item.ImpactLevel = domain.ImpactLevel(impactStr)
		item.ComplianceStatus = domain.RegulatoryComplianceStatus(statusStr)
		list = append(list, item)
	}

	if list == nil {
		list = []domain.RegulatoryWatchItem{}
	}
	return list, nil
}

// UpdateRegulatoryWatchAnalysis met à jour l'analyse d'impact et le statut de conformité
func (r *RegulatoryWatchRepository) UpdateRegulatoryWatchAnalysis(ctx context.Context, orgID, itemID uuid.UUID, analysis string, status domain.RegulatoryComplianceStatus) (*domain.RegulatoryWatchItem, error) {
	now := time.Now().UTC()
	query := `
		UPDATE regulatory_watch_items
		SET impact_analysis = $1,
		    compliance_status = $2,
		    updated_at = $3
		WHERE organization_id = $4 AND id = $5
	`
	cmd, err := r.pool.Exec(ctx, query, analysis, string(status), now, orgID, itemID)
	if err != nil {
		return nil, fmt.Errorf("échec mise à jour analyse veille: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	return r.GetRegulatoryWatchItem(ctx, orgID, itemID)
}
