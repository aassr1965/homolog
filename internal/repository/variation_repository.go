package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/domain"
)

var (
	ErrVariationNotFound = errors.New("demande de variation post-AMM introuvable")
	ErrItemNotFound      = errors.New("rubrique de variation introuvable")
	ErrDocumentNotFound  = errors.New("pièce justificative de variation introuvable")
)

// VariationRepository gère la persistance des variations post-AMM selon LD-DEPP07/V01
type VariationRepository struct {
	pool *pgxpool.Pool
}

// NewVariationRepository instancie le repository des variations
func NewVariationRepository(pool *pgxpool.Pool) *VariationRepository {
	return &VariationRepository{pool: pool}
}

// CreateVariationRequest crée une nouvelle demande de modification
func (r *VariationRepository) CreateVariationRequest(ctx context.Context, req *domain.VariationRequest) error {
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	now := time.Now().UTC()
	req.CreatedAt = now
	req.UpdatedAt = now

	if req.OverallClassification == "" {
		req.OverallClassification = domain.ClassMmin
	}
	if req.NominalClassification == "" {
		req.NominalClassification = domain.ClassMmin
	}
	if req.Status == "" {
		req.Status = "brouillon"
	}
	if req.FeeAmountDZD == 0 {
		req.FeeAmountDZD = domain.CalculateVariationFee(req.OverallClassification)
	}

	query := `
		INSERT INTO anpp_variation_requests (
			id, organization_id, medicine_id, dossier_id, tracking_number,
			title, description, is_grouping, grouping_justification,
			overall_classification, nominal_classification, has_escalation, escalation_summary,
			current_approved_situation, proposed_situation, technical_justification,
			non_modification_attested, non_modification_details, fee_amount_dzd,
			fee_slip_id, status, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16,
			$17, $18, $19,
			$20, $21, $22, $23, $24
		)
	`
	_, err := r.pool.Exec(ctx, query,
		req.ID, req.OrganizationID, req.MedicineID, req.DossierID, req.TrackingNumber,
		req.Title, req.Description, req.IsGrouping, req.GroupingJustification,
		string(req.OverallClassification), string(req.NominalClassification), req.HasEscalation, req.EscalationSummary,
		req.CurrentApprovedSituation, req.ProposedSituation, req.TechnicalJustification,
		req.NonModificationAttested, req.NonModificationDetails, req.FeeAmountDZD,
		req.FeeSlipID, req.Status, req.CreatedBy, req.CreatedAt, req.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insertion variation request échouée: %w", err)
	}

	// Initialiser automatiquement les 12 pièces obligatoires de la LD-DEPP07
	return r.InitializeVariationDocuments(ctx, req.ID)
}

// GetVariationRequestByID récupère la demande avec tous ses items et ses 12 pièces
func (r *VariationRepository) GetVariationRequestByID(ctx context.Context, id uuid.UUID) (*domain.VariationRequest, error) {
	query := `
		SELECT 
			v.id, v.organization_id, v.medicine_id, v.dossier_id, v.tracking_number,
			v.title, COALESCE(v.description, ''), v.is_grouping, COALESCE(v.grouping_justification, ''),
			v.overall_classification, v.nominal_classification, v.has_escalation, COALESCE(v.escalation_summary, ''),
			v.current_approved_situation, v.proposed_situation, v.technical_justification,
			v.non_modification_attested, COALESCE(v.non_modification_details, ''), v.fee_amount_dzd,
			v.fee_slip_id, v.status, v.dt_verified_by, v.dt_verified_at,
			COALESCE(v.dt_signature_seal, ''), COALESCE(v.dt_verification_notes, ''),
			v.created_by, v.created_at, v.updated_at,
			COALESCE(m.brand_name_fr, ''), COALESCE(m.dci_fr, ''), COALESCE(m.dosage_fr, ''),
			COALESCE(m.pharmaceutical_form_fr, ''), COALESCE(m.registration_decision_number, '')
		FROM anpp_variation_requests v
		LEFT JOIN medicines m ON v.medicine_id = m.id
		WHERE v.id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var req domain.VariationRequest
	var overallClass, nominalClass string
	err := row.Scan(
		&req.ID, &req.OrganizationID, &req.MedicineID, &req.DossierID, &req.TrackingNumber,
		&req.Title, &req.Description, &req.IsGrouping, &req.GroupingJustification,
		&overallClass, &nominalClass, &req.HasEscalation, &req.EscalationSummary,
		&req.CurrentApprovedSituation, &req.ProposedSituation, &req.TechnicalJustification,
		&req.NonModificationAttested, &req.NonModificationDetails, &req.FeeAmountDZD,
		&req.FeeSlipID, &req.Status, &req.DTVerifiedBy, &req.DTVerifiedAt,
		&req.DTSignatureSeal, &req.DTVerificationNotes,
		&req.CreatedBy, &req.CreatedAt, &req.UpdatedAt,
		&req.MedicineName, &req.MedicineDCI, &req.MedicineDosage,
		&req.MedicineForm, &req.RegistrationNumber,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVariationNotFound
		}
		return nil, fmt.Errorf("lecture variation request échouée: %w", err)
	}
	req.OverallClassification = domain.VariationClassification(overallClass)
	req.NominalClassification = domain.VariationClassification(nominalClass)

	// Charger les items
	items, err := r.GetVariationItems(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	req.Items = items

	// Charger les 12 documents
	docs, err := r.GetVariationDocuments(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	req.Documents = docs

	// Construire l'arbre d'impact
	rubriqueCodes := make([]string, 0, len(items))
	for _, it := range items {
		rubriqueCodes = append(rubriqueCodes, it.RubriqueCode)
	}
	impactTree := domain.GenerateChangeImpactTree(rubriqueCodes)
	req.ChangeImpactTree = &impactTree

	return &req, nil
}

// ListVariationRequests liste les variations par organisation et produit optionnel
func (r *VariationRepository) ListVariationRequests(ctx context.Context, orgID uuid.UUID, medicineID *uuid.UUID, status string) ([]domain.VariationRequest, error) {
	query := `
		SELECT 
			v.id, v.organization_id, v.medicine_id, v.dossier_id, v.tracking_number,
			v.title, COALESCE(v.description, ''), v.is_grouping, COALESCE(v.grouping_justification, ''),
			v.overall_classification, v.nominal_classification, v.has_escalation, COALESCE(v.escalation_summary, ''),
			v.current_approved_situation, v.proposed_situation, v.technical_justification,
			v.non_modification_attested, COALESCE(v.non_modification_details, ''), v.fee_amount_dzd,
			v.fee_slip_id, v.status, v.dt_verified_by, v.dt_verified_at,
			COALESCE(v.dt_signature_seal, ''), COALESCE(v.dt_verification_notes, ''),
			v.created_by, v.created_at, v.updated_at,
			COALESCE(m.brand_name_fr, ''), COALESCE(m.dci_fr, ''), COALESCE(m.dosage_fr, ''),
			COALESCE(m.pharmaceutical_form_fr, ''), COALESCE(m.registration_decision_number, '')
		FROM anpp_variation_requests v
		LEFT JOIN medicines m ON v.medicine_id = m.id
		WHERE v.organization_id = $1
	`
	args := []interface{}{orgID}
	idx := 2

	if medicineID != nil && *medicineID != uuid.Nil {
		query += fmt.Sprintf(" AND v.medicine_id = $%d", idx)
		args = append(args, *medicineID)
		idx++
	}

	if status != "" {
		query += fmt.Sprintf(" AND v.status = $%d", idx)
		args = append(args, status)
		idx++
	}

	query += " ORDER BY v.created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("requête liste variations échouée: %w", err)
	}
	defer rows.Close()

	list := make([]domain.VariationRequest, 0)
	for rows.Next() {
		var req domain.VariationRequest
		var overallClass, nominalClass string
		err := rows.Scan(
			&req.ID, &req.OrganizationID, &req.MedicineID, &req.DossierID, &req.TrackingNumber,
			&req.Title, &req.Description, &req.IsGrouping, &req.GroupingJustification,
			&overallClass, &nominalClass, &req.HasEscalation, &req.EscalationSummary,
			&req.CurrentApprovedSituation, &req.ProposedSituation, &req.TechnicalJustification,
			&req.NonModificationAttested, &req.NonModificationDetails, &req.FeeAmountDZD,
			&req.FeeSlipID, &req.Status, &req.DTVerifiedBy, &req.DTVerifiedAt,
			&req.DTSignatureSeal, &req.DTVerificationNotes,
			&req.CreatedBy, &req.CreatedAt, &req.UpdatedAt,
			&req.MedicineName, &req.MedicineDCI, &req.MedicineDosage,
			&req.MedicineForm, &req.RegistrationNumber,
		)
		if err != nil {
			return nil, fmt.Errorf("scan variation request échoué: %w", err)
		}
		req.OverallClassification = domain.VariationClassification(overallClass)
		req.NominalClassification = domain.VariationClassification(nominalClass)
		list = append(list, req)
	}

	return list, nil
}

// AddVariationItem ajoute une rubrique dans une demande et recalcule la criticité globale
func (r *VariationRepository) AddVariationItem(ctx context.Context, item *domain.VariationItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now

	// Évaluation de l'escalade
	nominal, effective, isEsc, reason := domain.EvaluateRubriqueItem(item.RubriqueCode, item.ConditionsEvaluated)
	item.NominalClassification = nominal
	item.EffectiveClassification = effective
	item.IsEscalated = isEsc
	item.EscalationReason = reason
	item.AllConditionsMet = !isEsc

	rubDef := domain.FindRubriqueDefinition(item.RubriqueCode)
	if rubDef != nil {
		item.RubriqueTitle = rubDef.Title
		item.Category = rubDef.Category
	}

	condJSON, err := json.Marshal(item.ConditionsEvaluated)
	if err != nil {
		condJSON = []byte("{}")
	}

	query := `
		INSERT INTO anpp_variation_items (
			id, variation_request_id, rubrique_code, rubrique_title, category,
			nominal_classification, effective_classification, is_escalated, escalation_reason,
			conditions_evaluated, all_conditions_met, current_state, proposed_state, notes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14,
			$15, $16
		)
	`
	_, err = r.pool.Exec(ctx, query,
		item.ID, item.VariationRequestID, item.RubriqueCode, item.RubriqueTitle, item.Category,
		string(item.NominalClassification), string(item.EffectiveClassification), item.IsEscalated, item.EscalationReason,
		condJSON, item.AllConditionsMet, item.CurrentState, item.ProposedState, item.Notes,
		item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insertion variation item échouée: %w", err)
	}

	// Recalculer l'état global de la demande
	return r.recalculateRequestOverall(ctx, item.VariationRequestID)
}

// UpdateVariationItem met à jour les conditions et recalcule l'escalade
func (r *VariationRepository) UpdateVariationItem(ctx context.Context, item *domain.VariationItem) error {
	item.UpdatedAt = time.Now().UTC()

	nominal, effective, isEsc, reason := domain.EvaluateRubriqueItem(item.RubriqueCode, item.ConditionsEvaluated)
	item.NominalClassification = nominal
	item.EffectiveClassification = effective
	item.IsEscalated = isEsc
	item.EscalationReason = reason
	item.AllConditionsMet = !isEsc

	condJSON, err := json.Marshal(item.ConditionsEvaluated)
	if err != nil {
		condJSON = []byte("{}")
	}

	query := `
		UPDATE anpp_variation_items SET
			nominal_classification = $1,
			effective_classification = $2,
			is_escalated = $3,
			escalation_reason = $4,
			conditions_evaluated = $5,
			all_conditions_met = $6,
			current_state = $7,
			proposed_state = $8,
			notes = $9,
			updated_at = $10
		WHERE id = $11 AND variation_request_id = $12
	`
	res, err := r.pool.Exec(ctx, query,
		string(item.NominalClassification), string(item.EffectiveClassification), item.IsEscalated, item.EscalationReason,
		condJSON, item.AllConditionsMet, item.CurrentState, item.ProposedState, item.Notes,
		item.UpdatedAt, item.ID, item.VariationRequestID,
	)
	if err != nil {
		return fmt.Errorf("mise à jour variation item échouée: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrItemNotFound
	}

	return r.recalculateRequestOverall(ctx, item.VariationRequestID)
}

// DeleteVariationItem retire une rubrique et recalcule la sévérité
func (r *VariationRepository) DeleteVariationItem(ctx context.Context, itemID uuid.UUID, variationRequestID uuid.UUID) error {
	query := `DELETE FROM anpp_variation_items WHERE id = $1 AND variation_request_id = $2`
	res, err := r.pool.Exec(ctx, query, itemID, variationRequestID)
	if err != nil {
		return fmt.Errorf("suppression variation item échouée: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrItemNotFound
	}

	return r.recalculateRequestOverall(ctx, variationRequestID)
}

// GetVariationItems récupère la liste des rubriques d'une demande
func (r *VariationRepository) GetVariationItems(ctx context.Context, variationRequestID uuid.UUID) ([]domain.VariationItem, error) {
	query := `
		SELECT 
			id, variation_request_id, rubrique_code, rubrique_title, category,
			nominal_classification, effective_classification, is_escalated, COALESCE(escalation_reason, ''),
			conditions_evaluated, all_conditions_met, COALESCE(current_state, ''), COALESCE(proposed_state, ''),
			COALESCE(notes, ''), created_at, updated_at
		FROM anpp_variation_items
		WHERE variation_request_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, variationRequestID)
	if err != nil {
		return nil, fmt.Errorf("lecture variation items échouée: %w", err)
	}
	defer rows.Close()

	items := make([]domain.VariationItem, 0)
	for rows.Next() {
		var it domain.VariationItem
		var nomClass, effClass string
		var condRaw []byte
		err := rows.Scan(
			&it.ID, &it.VariationRequestID, &it.RubriqueCode, &it.RubriqueTitle, &it.Category,
			&nomClass, &effClass, &it.IsEscalated, &it.EscalationReason,
			&condRaw, &it.AllConditionsMet, &it.CurrentState, &it.ProposedState,
			&it.Notes, &it.CreatedAt, &it.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan variation item échoué: %w", err)
		}
		it.NominalClassification = domain.VariationClassification(nomClass)
		it.EffectiveClassification = domain.VariationClassification(effClass)
		it.ConditionsEvaluated = make(map[string]bool)
		if len(condRaw) > 0 {
			_ = json.Unmarshal(condRaw, &it.ConditionsEvaluated)
		}
		items = append(items, it)
	}
	return items, nil
}

// InitializeVariationDocuments insère les 12 pièces obligatoires de la LD-DEPP07 Section 5.3
func (r *VariationRepository) InitializeVariationDocuments(ctx context.Context, variationRequestID uuid.UUID) error {
	now := time.Now().UTC()
	for _, spec := range domain.MasterVariationDocumentsList {
		query := `
			INSERT INTO anpp_variation_documents (
				id, variation_request_id, doc_index, doc_code, title, description,
				is_mandatory, is_attached, verified_status, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, FALSE, 'en_attente', $8, $9
			) ON CONFLICT (variation_request_id, doc_index) DO NOTHING
		`
		_, err := r.pool.Exec(ctx, query,
			uuid.New(), variationRequestID, spec.Index, spec.Code, spec.Title, spec.Description,
			spec.IsMandatory, now, now,
		)
		if err != nil {
			return fmt.Errorf("initialisation document %d échouée: %w", spec.Index, err)
		}
	}
	return nil
}

// GetVariationDocuments retourne les 12 pièces d'une demande
func (r *VariationRepository) GetVariationDocuments(ctx context.Context, variationRequestID uuid.UUID) ([]domain.VariationDocument, error) {
	query := `
		SELECT 
			d.id, d.variation_request_id, d.doc_index, d.doc_code, d.title,
			COALESCE(d.description, ''), d.is_mandatory, d.is_attached,
			d.ged_document_id, d.verified_status, d.verified_by, d.verified_at,
			COALESCE(d.notes, ''), d.created_at, d.updated_at,
			COALESCE(g.title_fr, ''), COALESCE(v.file_sha256, '')
		FROM anpp_variation_documents d
		LEFT JOIN ged_documents g ON d.ged_document_id = g.id
		LEFT JOIN LATERAL (
			SELECT file_sha256 FROM ged_versions WHERE document_id = g.id ORDER BY created_at DESC LIMIT 1
		) v ON true
		WHERE d.variation_request_id = $1
		ORDER BY d.doc_index ASC
	`
	rows, err := r.pool.Query(ctx, query, variationRequestID)
	if err != nil {
		return nil, fmt.Errorf("lecture variation docs échouée: %w", err)
	}
	defer rows.Close()

	docs := make([]domain.VariationDocument, 0)
	for rows.Next() {
		var d domain.VariationDocument
		err := rows.Scan(
			&d.ID, &d.VariationRequestID, &d.DocIndex, &d.DocCode, &d.Title,
			&d.Description, &d.IsMandatory, &d.IsAttached,
			&d.GEDDocumentID, &d.VerifiedStatus, &d.VerifiedBy, &d.VerifiedAt,
			&d.Notes, &d.CreatedAt, &d.UpdatedAt,
			&d.GEDFileName, &d.GEDHashSHA256,
		)
		if err != nil {
			return nil, fmt.Errorf("scan variation doc échoué: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, nil
}

// AttachDocument rattache un fichier GED à une pièce requise
func (r *VariationRepository) AttachDocument(ctx context.Context, docID uuid.UUID, gedDocID uuid.UUID) error {
	now := time.Now().UTC()
	query := `
		UPDATE anpp_variation_documents SET
			ged_document_id = $1,
			is_attached = TRUE,
			verified_status = 'conforme',
			updated_at = $2
		WHERE id = $3
	`
	res, err := r.pool.Exec(ctx, query, gedDocID, now, docID)
	if err != nil {
		return fmt.Errorf("rattachement document échoué: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrDocumentNotFound
	}
	return nil
}

// SignAndSealVariationDT enregistre la signature 21 CFR Part 11 du Pharmacien DT
func (r *VariationRepository) SignAndSealVariationDT(ctx context.Context, variationID uuid.UUID, dtUserID uuid.UUID, seal string, notes string) error {
	now := time.Now().UTC()
	query := `
		UPDATE anpp_variation_requests SET
			status = 'scelle_dt',
			dt_verified_by = $1,
			dt_verified_at = $2,
			dt_signature_seal = $3,
			dt_verification_notes = $4,
			updated_at = $5
		WHERE id = $6
	`
	res, err := r.pool.Exec(ctx, query, dtUserID, now, seal, notes, now, variationID)
	if err != nil {
		return fmt.Errorf("scellement DT variation échoué: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrVariationNotFound
	}
	return nil
}

// recalculateRequestOverall recalcule la criticité globale et la redevance applicable
func (r *VariationRepository) recalculateRequestOverall(ctx context.Context, variationRequestID uuid.UUID) error {
	items, err := r.GetVariationItems(ctx, variationRequestID)
	if err != nil {
		return err
	}

	overall, hasEsc, highest := domain.CalculateOverallClassification(items)
	feeAmount := domain.CalculateVariationFee(overall)
	isGrouping := len(items) > 1

	escalationSummary := ""
	if hasEsc {
		escalationSummary = fmt.Sprintf("Reclassement d'au moins une rubrique vers un niveau supérieur (Rubrique déterminante: %s).", highest)
	}

	now := time.Now().UTC()
	query := `
		UPDATE anpp_variation_requests SET
			overall_classification = $1,
			has_escalation = $2,
			escalation_summary = $3,
			is_grouping = $4,
			fee_amount_dzd = $5,
			updated_at = $6
		WHERE id = $7
	`
	_, err = r.pool.Exec(ctx, query,
		string(overall), hasEsc, escalationSummary, isGrouping, feeAmount, now, variationRequestID,
	)
	if err != nil {
		return fmt.Errorf("recalcul variation overall échoué: %w", err)
	}
	return nil
}
