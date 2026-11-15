package api

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedInitialRegulatoryWatchAndAppointments injecte le corpus juridique algérien officiel et des créneaux ANPP
func SeedInitialRegulatoryWatchAndAppointments(ctx context.Context, pool *pgxpool.Pool, orgSlug string) {
	var orgID uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = $1", orgSlug).Scan(&orgID)
	if err != nil {
		return
	}

	// 1. Injection des Textes Légaux Officiels de Référence (Corpus Pharmaceutique JORADP)
	texts := []struct {
		TitleFR        string
		TitleAR        string
		LegalRef       string
		JORADP         string
		PubDate        string
		EffDate        string
		Category       string
		ImpactLevel    string
		SummaryFR      string
		ImpactAnalysis string
		Compliance     string
	}{
		{
			TitleFR:        "Modalités d'Enregistrement des Produits Pharmaceutiques",
			TitleAR:        "كيفيات تسجيل المواد الصيدلانية للاستعمال البشري",
			LegalRef:       "Décret Exécutif N° 20-325",
			JORADP:         "JORADP N° 78 du 28 Novembre 2020",
			PubDate:        "2020-11-28",
			EffDate:        "2020-11-28",
			Category:       "enregistrement_amm",
			ImpactLevel:    "critique",
			SummaryFR:      "Fixe les règles et conditions de soumission des demandes d'AMM, délais légaux de traitement par l'ANPP (90 jours à 180 jours), redevances et obligations du Pharmacien DT.",
			ImpactAnalysis: "Texte socle régissant l'ensemble des modules CTD (1 à 5). Sanad Pharma applique la conformité stricte dès la phase d'audit pré-dépôt.",
			Compliance:     "conforme",
		},
		{
			TitleFR:        "Modalités d'Homologation des Dispositifs Médicaux",
			TitleAR:        "كيفيات المصادقة على المستلزمات الطبية",
			LegalRef:       "Décret Exécutif N° 20-324",
			JORADP:         "JORADP N° 78 du 28 Novembre 2020",
			PubDate:        "2020-11-28",
			EffDate:        "2020-11-28",
			Category:       "dispositifs_medicaux",
			ImpactLevel:    "critique",
			SummaryFR:      "Définit la classification des DM (Classes I, IIa, IIb, III, DMDIV), les exigences d'essais cliniques, biocompatibilité et déclaration de conformité ISO 13485.",
			ImpactAnalysis: "Mise à jour des dossiers techniques selon les 16 rubriques réglementaires et Arrêté du 10 mai 2021.",
			Compliance:     "conforme",
		},
		{
			TitleFR:        "Composition du Dossier Technique d'Homologation des Dispositifs Médicaux",
			TitleAR:        "تحديد محتوى الملف التقني للمستلزمات الطبية",
			LegalRef:       "Arrêté Ministériel du 10 Mai 2021",
			JORADP:         "JORADP N° 45 du 12 Juin 2021",
			PubDate:        "2021-05-10",
			EffDate:        "2021-06-12",
			Category:       "dispositifs_medicaux",
			ImpactLevel:    "majeur",
			SummaryFR:      "Précise les pièces obligatoires [E] : certificat CE/FDA, rapport de gestion des risques ISO 14971, notice et étiquetage bilingue arabe/français.",
			ImpactAnalysis: "Matrice de checklist 16 rubriques opérationnelle dans le module Dispositifs DM.",
			Compliance:     "conforme",
		},
		{
			TitleFR:        "Guide Algérien des Bonnes Pratiques de Fabrication (BPF)",
			TitleAR:        "دليل الممارسات الحسنة للصنع للمواد الصيدلانية",
			LegalRef:       "Arrêté Ministériel du 16 Février 2021",
			JORADP:         "JORADP N° 19 du 14 Mars 2021",
			PubDate:        "2021-02-16",
			EffDate:        "2021-03-14",
			Category:       "bpf_inspection",
			ImpactLevel:    "critique",
			SummaryFR:      "Adopte les principes PIC/S pour l'assurance qualité, validation de nettoyage, contrôle de contamination croisée et intégrité des données électroniques (21 CFR Part 11).",
			ImpactAnalysis: "Coffre-fort GED Sanad Pharma certifié immuable (Audit Trail SHA-256) pour satisfaire aux inspections de conformité BPF.",
			Compliance:     "conforme",
		},
		{
			TitleFR:        "Barème des Redevances de Dépôt, d'Expertise et d'Homologation ANPP",
			TitleAR:        "تحديد الإتاوات المستحقة للوكالة الوطنية للمواد الصيدلانية",
			LegalRef:       "Arrêté Interministériel Finances / MIPH",
			JORADP:         "JORADP N° 34 du 15 Mai 2024",
			PubDate:        "2024-05-15",
			EffDate:        "2024-06-01",
			Category:       "redevances_tarifs",
			ImpactLevel:    "majeur",
			SummaryFR:      "Fixe les montants des bordereaux BV1 (dépôt initial) et BV2 (expertises/variations) : 150 000 DZD pour génériques locaux, 350 000 DZD pour princeps.",
			ImpactAnalysis: "Intégré dans le calculateur automatique des bordereaux avec conversion légale en toutes lettres.",
			Compliance:     "conforme",
		},
	}

	for _, t := range texts {
		var exists int
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM regulatory_watch_items WHERE organization_id = $1 AND legal_reference = $2", orgID, t.LegalRef).Scan(&exists)
		if exists == 0 {
			_, err := pool.Exec(ctx, `
				INSERT INTO regulatory_watch_items (
					organization_id, title_fr, title_ar, legal_reference, joradp_number,
					published_date, effective_date, category, impact_level, summary_fr,
					impact_analysis, compliance_status
				) VALUES (
					$1, $2, $3, $4, $5,
					$6::date, $7::date, $8, $9, $10,
					$11, $12
				)
			`, orgID, t.TitleFR, t.TitleAR, t.LegalRef, t.JORADP, t.PubDate, t.EffDate, t.Category, t.ImpactLevel, t.SummaryFR, t.ImpactAnalysis, t.Compliance)
			if err != nil {
				log.Printf("[WARN] Seed regulatory text %s: %v", t.LegalRef, err)
			}
		}
	}

	// 2. Injection de rendez-vous de démonstration (Séances ANPP avec compte à rebours)
	var dossierID *uuid.UUID
	var dtUserID *uuid.UUID
	var dID uuid.UUID
	var uID uuid.UUID

	if err := pool.QueryRow(ctx, "SELECT id FROM dossiers WHERE organization_id = $1 ORDER BY created_at ASC LIMIT 1", orgID).Scan(&dID); err == nil {
		dossierID = &dID
	}
	if err := pool.QueryRow(ctx, "SELECT id FROM users WHERE organization_id = $1 AND role = 'directeur_technique' LIMIT 1", orgID).Scan(&uID); err == nil {
		dtUserID = &uID
	}

	now := time.Now().UTC()
	appointments := []struct {
		AppType   string
		DaysAhead int
		Location  string
		SlotRef   string
		Officer   string
		Status    string
		Notes     string
	}{
		{
			AppType:   "seance_commission",
			DaysAhead: 5, // Alerte J <= 7
			Location:  "ANPP - Direction de l'Enregistrement, Salle des Commissions N°2, Dely Ibrahim",
			SlotRef:   "ANPP/CE/2026-N°089",
			Officer:   "Secrétariat de la Commission d'Enregistrement",
			Status:    "confirme_anpp",
			Notes:     "Audition contradictoire et présentation de la réponse aux compléments analytiques Module 3.",
		},
		{
			AppType:   "expertise_lncm",
			DaysAhead: 12,
			Location:  "Laboratoire National de Contrôle des Produits Pharmaceutiques (LNCM), Annexe Sidi Abdellah",
			SlotRef:   "LNCM/EXP/2026-N°0342",
			Officer:   "Département de Contrôle Physico-Chimique",
			Status:    "planifie",
			Notes:     "Remise de 3 lots industriels pilotes sous température dirigée (20-25°C) pour essais de dissolution comparative.",
		},
	}

	for _, a := range appointments {
		var exists int
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM anpp_deposit_appointments WHERE organization_id = $1 AND slot_reference = $2", orgID, a.SlotRef).Scan(&exists)
		if exists == 0 {
			sched := now.Add(time.Duration(a.DaysAhead) * 24 * time.Hour)
			_, err := pool.Exec(ctx, `
				INSERT INTO anpp_deposit_appointments (
					organization_id, dossier_id, appointment_type, scheduled_at,
					location, slot_reference, contact_officer, assigned_delegate_id,
					status, notes
				) VALUES (
					$1, $2, $3, $4,
					$5, $6, $7, $8,
					$9, $10
				)
			`, orgID, dossierID, a.AppType, sched, a.Location, a.SlotRef, a.Officer, dtUserID, a.Status, a.Notes)
			if err != nil {
				log.Printf("[WARN] Seed appointment %s: %v", a.SlotRef, err)
			}
		}
	}
}
