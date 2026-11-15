package api

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// SeedInitialPaymentSlips injecte des bordereaux de démonstration BV1 et BV2 (Jalon 1.7 Phase 2)
func SeedInitialPaymentSlips(ctx context.Context, pool *pgxpool.Pool, orgSlug string) {
	var orgID uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = $1", orgSlug).Scan(&orgID)
	if err != nil {
		return
	}

	// Idempotence : si des bordereaux existent déjà, ne rien faire
	var count int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM anpp_payment_slips WHERE organization_id = $1", orgID).Scan(&count)
	if count > 0 {
		return
	}

	// Récupération d'un dossier
	var dossierID uuid.UUID
	err = pool.QueryRow(ctx, "SELECT id FROM dossiers WHERE organization_id = $1 ORDER BY created_at ASC LIMIT 1", orgID).Scan(&dossierID)
	if err != nil {
		return
	}

	// Récupération du Pharmacien Directeur Technique
	var dtUserID uuid.UUID
	_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE organization_id = $1 AND role = 'directeur_technique' LIMIT 1", orgID).Scan(&dtUserID)

	// Récupération des barèmes actifs
	feeRepo := repository.NewFeeRepository(pool)
	schedGenerique, err := feeRepo.GetFeeScheduleForProcedure(ctx, "generique_importe", time.Now().UTC())
	if err != nil {
		schedGenerique, _ = feeRepo.GetFeeScheduleForProcedure(ctx, "generique_local", time.Now().UTC())
	}
	if schedGenerique == nil {
		return
	}

	schedVar, _ := feeRepo.GetFeeScheduleForProcedure(ctx, "variation_mineure", time.Now().UTC())
	if schedVar == nil {
		schedVar = schedGenerique
	}

	now := time.Now().UTC()
	paidDate := now.AddDate(0, 0, -35)
	dtUserIDStr := dtUserID.String()

	// 1. Bordereau BV1 Dépôt Initial (Payé avec quittance Trésor)
	wordsFR1, wordsAR1 := domain.ConvertAmountDZDToWords(schedGenerique.AmountDZD)
	bv1 := domain.ANPPPaymentSlip{
		ID:               uuid.New().String(),
		OrganizationID:   orgID.String(),
		DossierID:        dossierID.String(),
		FeeScheduleID:    schedGenerique.ID,
		SlipType:         domain.SlipTypeBV1,
		SlipNumber:       "BV1-2026-AMOX-001",
		PaymentPurpose:   "Redevance officielle de dépôt de dossier d'enregistrement initial (ANPP - Décret 20-325)",
		AmountDZD:        schedGenerique.AmountDZD,
		AmountWordsFR:    wordsFR1,
		AmountWordsAR:    wordsAR1,
		BankAccountRIB:   "00100624030000001234 (Compte Trésorerie Centrale ANPP Alger)",
		Status:           domain.PaymentStatusPaye,
		PaymentReference: "QUIT-TRESOR-2026-N°45812",
		PaidAt:           &paidDate,
		GeneratedBy:      &dtUserIDStr,
		GeneratedAt:      paidDate.AddDate(0, 0, -2),
	}
	_, _ = feeRepo.CreatePaymentSlip(ctx, &bv1)
	// Mise à jour de la quittance
	_ = feeRepo.UpdatePaymentSlipStatus(ctx, orgID.String(), bv1.ID, domain.PaymentStatusPaye, "QUIT-TRESOR-2026-N°45812", &paidDate, nil)

	// 2. Bordereau BV2 Compléments / Analyses LNCM (Émis en attente de règlement)
	wordsFR2, wordsAR2 := domain.ConvertAmountDZDToWords(schedVar.AmountDZD)
	bv2 := domain.ANPPPaymentSlip{
		ID:             uuid.New().String(),
		OrganizationID: orgID.String(),
		DossierID:      dossierID.String(),
		FeeScheduleID:  schedVar.ID,
		SlipType:       domain.SlipTypeBV2,
		SlipNumber:     "BV2-2026-AMOX-002",
		PaymentPurpose: "Redevance complémentaire pour expertise analytique et contrôle de lots (LNCM)",
		AmountDZD:      schedVar.AmountDZD,
		AmountWordsFR:  wordsFR2,
		AmountWordsAR:  wordsAR2,
		BankAccountRIB: "00100624030000001234 (Compte Trésorerie Centrale ANPP Alger)",
		Status:         domain.PaymentStatusEmis,
		GeneratedBy:    &dtUserIDStr,
		GeneratedAt:    now.AddDate(0, 0, -5),
	}
	_, _ = feeRepo.CreatePaymentSlip(ctx, &bv2)

	log.Printf("[SEED] Bordereaux de versement BV1 (%s) et BV2 (%s) injectés avec succès (Jalon 1.7 Phase 2).", bv1.SlipNumber, bv2.SlipNumber)
}
