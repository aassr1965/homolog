package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/domain"
)

// Actions standardisées pour l'Audit Trail BPF / 21 CFR Part 11
const (
	ActionAuthLogin         = "AUTH_LOGIN"
	ActionAuthLoginFailed   = "AUTH_LOGIN_FAILED"
	ActionAuthLogout        = "AUTH_LOGOUT"
	Action2FASetup          = "2FA_SETUP"
	Action2FAVerify         = "2FA_VERIFY"
	ActionDocView           = "DOC_VIEW"
	ActionDocUpload         = "DOC_UPLOAD"
	ActionDocSignDT         = "DOC_SIGN_DT"
	ActionDocStatusChange   = "DOC_STATUS_CHANGE"
	ActionDossierCreate     = "DOSSIER_CREATE"
	ActionDossierLockDT     = "DOSSIER_LOCK_DT"
	ActionDossierExport     = "DOSSIER_EXPORT"
	ActionDossierItemAttach = "DOSSIER_ITEM_ATTACH"
	ActionDossierItemDetach = "DOSSIER_ITEM_DETACH"
	ActionDossierItemStatus = "DOSSIER_ITEM_STATUS"
	ActionDossierPDTSign    = "DOSSIER_PDT_SIGN"
	ActionDossierStatus     = "DOSSIER_STATUS_CHANGE"
	ActionSysBackup         = "SYS_BACKUP"
	ActionRoleChange        = "ROLE_CHANGE"
	ActionUserCreated       = "USER_CREATED"
	ActionDeficiencyLetterCreate = "DEFICIENCY_LETTER_CREATE"
	ActionDeficiencyTaskUpdate   = "DEFICIENCY_TASK_UPDATE"
	ActionDeficiencyTaskSignDT   = "DEFICIENCY_TASK_SIGN_DT"
	ActionFeeSlipGenerated       = "FEE_SLIP_GENERATED"
	ActionFeeSlipStatusUpdate     = "FEE_SLIP_STATUS_UPDATE"
	ActionFeeScheduleUpdate      = "FEE_SCHEDULE_UPDATE"
	ActionPreSubmissionAudit     = "PRE_SUBMISSION_AUDIT"
	ActionPreSubmissionLock      = "PRE_SUBMISSION_LOCK"
	ActionManifestExport         = "ESERVICES_MANIFEST_EXPORT"
	ActionAppointmentCreate      = "APPOINTMENT_CREATE"
	ActionAppointmentStatus      = "APPOINTMENT_STATUS_UPDATE"
	ActionRegulatoryWatchCreate  = "REGULATORY_WATCH_CREATE"
	ActionRegulatoryWatchUpdate  = "REGULATORY_WATCH_UPDATE"
)

// Entry représente les données à consigner dans l'Audit Trail
type Entry struct {
	OrganizationID uuid.UUID
	UserID         *uuid.UUID
	UserEmail      string
	UserRole       string
	Action         string
	EntityType     string
	EntityID       string
	Details        map[string]any
	IPAddress      string
	UserAgent      string
}

// Service gère l'enregistrement et la restitution de la piste d'audit ALCOA+
type Service struct {
	pool *pgxpool.Pool
}

// NewService instancie le service d'Audit Trail
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Log consigne une action de manière immuable avec calcul d'empreinte SHA-256
func (s *Service) Log(ctx context.Context, entry *Entry) (*domain.AuditLog, error) {
	now := time.Now().UTC()

	detailsJSON, err := json.Marshal(entry.Details)
	if err != nil {
		detailsJSON = []byte("{}")
	}

	// Calcul de l'empreinte cryptographique SHA-256 scellée
	hasher := sha256.New()
	hasher.Write([]byte(now.Format(time.RFC3339Nano)))
	hasher.Write([]byte(entry.OrganizationID.String()))
	if entry.UserID != nil {
		hasher.Write([]byte(entry.UserID.String()))
	}
	hasher.Write([]byte(entry.Action))
	hasher.Write([]byte(entry.EntityType))
	hasher.Write([]byte(entry.EntityID))
	hasher.Write(detailsJSON)
	hasher.Write([]byte(entry.IPAddress))
	sha256Hash := hex.EncodeToString(hasher.Sum(nil))

	query := `
		INSERT INTO audit_logs (
			organization_id, user_id, user_email, user_role, action, 
			entity_type, entity_id, details, ip_address, user_agent, 
			sha256_hash, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	var id int64
	err = s.pool.QueryRow(ctx, query,
		entry.OrganizationID,
		entry.UserID,
		entry.UserEmail,
		entry.UserRole,
		entry.Action,
		entry.EntityType,
		entry.EntityID,
		detailsJSON,
		entry.IPAddress,
		entry.UserAgent,
		sha256Hash,
		now,
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("impossible d'écrire dans l'Audit Trail immuable: %w", err)
	}

	return &domain.AuditLog{
		ID:             id,
		OrganizationID: entry.OrganizationID,
		UserID:         entry.UserID,
		UserEmail:      entry.UserEmail,
		UserRole:       entry.UserRole,
		Action:         entry.Action,
		EntityType:     entry.EntityType,
		EntityID:       entry.EntityID,
		Details:        string(detailsJSON),
		IPAddress:      entry.IPAddress,
		UserAgent:      entry.UserAgent,
		SHA256Hash:     sha256Hash,
		CreatedAt:      now,
	}, nil
}

// List récupère une page de l'Audit Trail avec tri chronologique inverse (pour inspections)
func (s *Service) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.AuditLog, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1", orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			id, organization_id, user_id, user_email, user_role, action,
			entity_type, entity_id, details, ip_address, user_agent,
			sha256_hash, created_at
		FROM audit_logs
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		var rawDetails []byte

		if err := rows.Scan(
			&l.ID, &l.OrganizationID, &l.UserID, &l.UserEmail, &l.UserRole, &l.Action,
			&l.EntityType, &l.EntityID, &rawDetails, &l.IPAddress, &l.UserAgent,
			&l.SHA256Hash, &l.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		l.Details = string(rawDetails)
		logs = append(logs, &l)
	}

	return logs, total, nil
}
