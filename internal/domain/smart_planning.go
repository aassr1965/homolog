package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SmartPlanningHealth représente l'état de santé du chemin critique ou d'un jalon
type SmartPlanningHealth string

const (
	HealthCompleted     SmartPlanningHealth = "COMPLETED"      // Jalon accompli / validé
	HealthOnTrack       SmartPlanningHealth = "ON_TRACK"       // Dans les temps (marge >= 30j)
	HealthWarningTight  SmartPlanningHealth = "WARNING_TIGHT"  // Marge restreinte (7j à 30j)
	HealthCriticalDelay SmartPlanningHealth = "CRITICAL_DELAY" // Retard critique (-15j à 7j)
	HealthFatalBlock    SmartPlanningHealth = "FATAL_BLOCK"    // Dépassement majeur (> 15j de retard sur jalon incompressible)
)

// SmartPlanningMilestoneCode identifie de manière unique les 6 jalons incompressibles du rétroplanning
type SmartPlanningMilestoneCode string

const (
	MilestoneJ240StabilitiesZoneIVb SmartPlanningMilestoneCode = "J_240_STABILITIES_ZONE_IVB"
	MilestoneJ150QOSCheckerReview   SmartPlanningMilestoneCode = "J_150_QOS_CHECKER_REVIEW"
	MilestoneJ60ETasjilAppointment  SmartPlanningMilestoneCode = "J_60_ETASJIL_APPOINTMENT"
	MilestoneJ30LeafletAndFees      SmartPlanningMilestoneCode = "J_30_LEAFLET_AND_FEES"
	MilestoneJ15PDTVisaPackageSeal  SmartPlanningMilestoneCode = "J_15_PDT_VISA_PACKAGE_SEAL"
	MilestoneJ0ANPPDepositDeadline  SmartPlanningMilestoneCode = "J_0_ANPP_DEPOSIT_DEADLINE"
)

// SmartPlanningMilestone modélise une étape critique calculée à rebours
type SmartPlanningMilestone struct {
	Code                SmartPlanningMilestoneCode `json:"code"`
	NameFR              string                     `json:"name_fr"`
	DayOffset           int                        `json:"day_offset"` // -240, -150, -60, -30, -15, 0
	TargetDate          time.Time                  `json:"target_date"`
	DaysRemaining       int                        `json:"days_remaining"` // TargetDate - Now (en jours)
	Health              SmartPlanningHealth        `json:"health"`
	IsCompleted         bool                       `json:"is_completed"`
	CompletionNote      string                     `json:"completion_note,omitempty"`
	AssociatedItemCodes []string                   `json:"associated_item_codes"`
	DepartmentOwner     string                     `json:"department_owner"` // "RD", "QA", "RA", "PDT"
	LegalReference      string                     `json:"legal_reference"`
	ExplanationFR       string                     `json:"explanation_fr"`
	AlertMessage        string                     `json:"alert_message,omitempty"`
}

// SmartPlanningAnalysis encapsule la restitution complète du rétroplanning inversé d'un dossier
type SmartPlanningAnalysis struct {
	DossierID               uuid.UUID                `json:"dossier_id"`
	DossierCode             string                   `json:"dossier_code"`
	ProcedureCode           ProcedureCode            `json:"procedure_code"`
	IsRenewalProcedure      bool                     `json:"is_renewal_procedure"`
	TargetDepositDate       time.Time                `json:"target_deposit_date"`
	IsEstimatedTargetDate   bool                     `json:"is_estimated_target_date"`
	TargetDateOrigin        string                   `json:"target_date_origin"` // "dossier_target_date", "anpp_validity_j180", "default_estimated"
	DaysToTargetDeposit     int                      `json:"days_to_target_deposit"`
	WorkingDaysPerWeek      int                      `json:"working_days_per_week"`
	RequiredWorkingDays     int                      `json:"required_working_days"`
	HolidaysExcludedCount   int                      `json:"holidays_excluded_count"`
	RecommendedDepositDate  *time.Time               `json:"recommended_deposit_date,omitempty"`
	IsTargetDateSufficient  bool                     `json:"is_target_date_sufficient"`
	OverallHealth           SmartPlanningHealth      `json:"overall_health"`
	OverallHealthLabel      string                   `json:"overall_health_label"`
	ActiveCriticalMilestone *SmartPlanningMilestone  `json:"active_critical_milestone,omitempty"`
	Milestones              []SmartPlanningMilestone `json:"milestones"`
	CriticalWarnings        []string                 `json:"critical_warnings"`
	RecommendedActionFR     string                   `json:"recommended_action_fr"`
	EvaluatedAt             time.Time                `json:"evaluated_at"`
}

// MilestoneTemplate définit les métadonnées statiques des 6 jalons incompressibles du CDC
type MilestoneTemplate struct {
	Code            SmartPlanningMilestoneCode
	NameFR          string
	DayOffset       int
	DepartmentOwner string
	AssociatedItems []string
	LegalReference  string
	ExplanationFR   string
}

// OfficialSmartPlanningTemplates contient les spécifications officielles du CDC (Section 7.4.3.6)
var OfficialSmartPlanningTemplates = []MilestoneTemplate{
	{
		Code:            MilestoneJ240StabilitiesZoneIVb,
		NameFR:          "Lancement étuves Stabilités Zone IVb (M6 légal)",
		DayOffset:       -240,
		DepartmentOwner: "RD",
		AssociatedItems: []string{"3.2.P.8", "3.2.S.7"},
		LegalReference:  "Guide CTD ANPP Juillet 2025 Module 3.2.P.8 & ICH Q1A(R2) Zone IVb",
		ExplanationFR:   "Délai d'épreuve physique incompressible minimal de 6 mois (180j) à 30°C/75% HR avec 60 jours d'analyse avant dépôt.",
	},
	{
		Code:            MilestoneJ150QOSCheckerReview,
		NameFR:          "Clôture Synthèses Qualité (QOS 2.3) & Revue Checker AQ",
		DayOffset:       -150,
		DepartmentOwner: "QA",
		AssociatedItems: []string{"2.3", "3.2.P.3.5", "3.2.P.5"},
		LegalReference:  "Guide CTD ANPP Module 2.3 & Bonnes Pratiques de Fabrication (BPF)",
		ExplanationFR:   "Agrégation et relecture indépendante par l'Assurance Qualité (Checker) de l'ensemble des synthèses chimiques et analytiques du Module 3.",
	},
	{
		Code:            MilestoneJ60ETasjilAppointment,
		NameFR:          "Réservation Créneau RDV E-TASJIL (rdvpp/rdvdm.anpp.dz)",
		DayOffset:       -60,
		DepartmentOwner: "RA",
		AssociatedItems: []string{"1.A.1", "1.A.2"},
		LegalReference:  "Plateforme E-Services ANPP & Calendrier d'audience réglementaire",
		ExplanationFR:   "Réservation anticipée obligatoire de la plage d'audience sur le portail officiel de l'ANPP avant engorgement des créneaux.",
	},
	{
		Code:            MilestoneJ30LeafletAndFees,
		NameFR:          "Double Visa Notice Bilingue FR/AR & Bordereaux BV1/BV2",
		DayOffset:       -30,
		DepartmentOwner: "RA",
		AssociatedItems: []string{"1.P.1", "1.P.2.1", "1.P.2.2", "1.P.4"},
		LegalReference:  "Décret exécutif n° 20-325 Art. 12 & Décret 21-82 (Redevances Trésor/BNA)",
		ExplanationFR:   "Contrôle strict de parité bilingue français-arabe de la notice patient (1.P.2), validation du RCP (1.P.1) et liquidation des frais officiels d'enregistrement.",
	},
	{
		Code:            MilestoneJ15PDTVisaPackageSeal,
		NameFR:          "Visa Légal 2FA Pharmacien DT & Scellement Package E-TASJIL",
		DayOffset:       -15,
		DepartmentOwner: "PDT",
		AssociatedItems: []string{"1.A.7", "1.P.3", "1.P.8", "1.P.9"},
		LegalReference:  "Décret exécutif n° 21-82 Art. 27 & 21 CFR Part 11 (Signature Scellée SHA-256)",
		ExplanationFR:   "Apposition de la signature électronique qualifiée du Pharmacien Directeur Technique et génération du Submission Package scellé.",
	},
	{
		Code:            MilestoneJ0ANPPDepositDeadline,
		NameFR:          "Dépôt Légal Officiel ANPP (Guichet E-TASJIL / Physique)",
		DayOffset:       0,
		DepartmentOwner: "RA",
		AssociatedItems: []string{"1.A.3"},
		LegalReference:  "Décret 20-325 Art. 12 (Délai légal impératif J-180 avant caducité AMM)",
		ExplanationFR:   "Date butoir impérative d'enregistrement de la demande et délivrance du récépissé de dépôt officiel de l'ANPP.",
	},
}

// DetermineTargetDepositDate détermine la date de dépôt visée selon la hiérarchie légale
func DetermineTargetDepositDate(d *Dossier, now time.Time) (time.Time, string, bool) {
	// 1. Si target_deposit_date est explicitement définie sur le dossier
	if d.TargetDepositDate != nil && !d.TargetDepositDate.IsZero() {
		return *d.TargetDepositDate, "dossier_target_date", false
	}

	// 2. Si procédure de renouvellement avec date d'expiration AMM (J-180 légal selon décret 20-325)
	isRenewal := d.ProcedureCode == ProcedureRI || d.ProcedureCode == ProcedureRF
	if isRenewal && d.ANPPValidityEndDate != nil && !d.ANPPValidityEndDate.IsZero() {
		// Le dépôt doit avoir lieu impérativement au moins 180 jours avant la fin de validité
		legalRenewalDeadline := d.ANPPValidityEndDate.AddDate(0, 0, -180)
		return legalRenewalDeadline, "anpp_validity_j180", false
	}

	// 3. Date estimée par défaut : Création + 9 mois (270 jours)
	baseTime := d.CreatedAt
	if baseTime.IsZero() {
		baseTime = now
	}
	estimated := baseTime.AddDate(0, 0, 270)
	return estimated, "default_estimated", true
}

// EvaluateSmartPlanning exécute l'analyse du rétroplanning dynamique inversé
func EvaluateSmartPlanning(d *Dossier, items []ChecklistItem, now time.Time) *SmartPlanningAnalysis {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	targetDate, origin, isEstimated := DetermineTargetDepositDate(d, now)
	isRenewal := d.ProcedureCode == ProcedureRI || d.ProcedureCode == ProcedureRF

	// Normaliser les dates à minuit UTC pour comparaison en jours entiers
	nowMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	targetMidnight := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)

	daysToTarget := int(targetMidnight.Sub(nowMidnight).Hours() / 24)

	// Index des items par code pour consultation rapide
	itemsByCode := make(map[string]ChecklistItem)
	for _, it := range items {
		itemsByCode[it.ItemCode] = it
	}

	analysis := &SmartPlanningAnalysis{
		DossierID:             d.ID,
		DossierCode:           d.DossierCode,
		ProcedureCode:         d.ProcedureCode,
		IsRenewalProcedure:    isRenewal,
		TargetDepositDate:     targetMidnight,
		IsEstimatedTargetDate: isEstimated,
		TargetDateOrigin:      origin,
		DaysToTargetDeposit:   daysToTarget,
		OverallHealth:         HealthOnTrack,
		Milestones:            make([]SmartPlanningMilestone, 0, len(OfficialSmartPlanningTemplates)),
		CriticalWarnings:      make([]string, 0),
		EvaluatedAt:           now,
	}

	worstHealthRank := 0
	healthRanking := map[SmartPlanningHealth]int{
		HealthCompleted:     1,
		HealthOnTrack:       2,
		HealthWarningTight:  3,
		HealthCriticalDelay: 4,
		HealthFatalBlock:    5,
	}

	var firstPendingMilestone *SmartPlanningMilestone

	for _, tmpl := range OfficialSmartPlanningTemplates {
		// Date théorique du jalon = TargetDate + DayOffset
		milestoneDate := targetMidnight.AddDate(0, 0, tmpl.DayOffset)
		daysRemaining := int(milestoneDate.Sub(nowMidnight).Hours() / 24)

		isCompleted, completionNote := checkMilestoneCompletion(tmpl.Code, d, itemsByCode)

		var health SmartPlanningHealth
		var alertMsg string

		if isCompleted {
			health = HealthCompleted
		} else {
			if daysRemaining >= 30 {
				health = HealthOnTrack
			} else if daysRemaining >= 7 {
				health = HealthWarningTight
				alertMsg = fmt.Sprintf("Jalon critique à échéance proche (%d jours restants).", daysRemaining)
			} else if daysRemaining >= -15 {
				health = HealthCriticalDelay
				if daysRemaining >= 0 {
					alertMsg = fmt.Sprintf("URGENCE : Jalon critique dans %d jour(s) ! Risque immédiat de dérive sur le chemin critique.", daysRemaining)
				} else {
					alertMsg = fmt.Sprintf("DÉPASSÉ de %d jour(s) ! Le chemin critique du dossier est en dérive active.", -daysRemaining)
				}
			} else {
				// Retard > 15 jours sur jalon non réalisé
				health = HealthFatalBlock
				alertMsg = fmt.Sprintf("RUPTURE CALENDAIRE : Dérive majeure de %d jours sur un délai réglementaire incompressible (%s). La date de dépôt cible doit être réévaluée.", -daysRemaining, tmpl.NameFR)
			}
		}

		ms := SmartPlanningMilestone{
			Code:                tmpl.Code,
			NameFR:              tmpl.NameFR,
			DayOffset:           tmpl.DayOffset,
			TargetDate:          milestoneDate,
			DaysRemaining:       daysRemaining,
			Health:              health,
			IsCompleted:         isCompleted,
			CompletionNote:      completionNote,
			AssociatedItemCodes: tmpl.AssociatedItems,
			DepartmentOwner:     tmpl.DepartmentOwner,
			LegalReference:      tmpl.LegalReference,
			ExplanationFR:       tmpl.ExplanationFR,
			AlertMessage:        alertMsg,
		}

		analysis.Milestones = append(analysis.Milestones, ms)

		if alertMsg != "" {
			analysis.CriticalWarnings = append(analysis.CriticalWarnings, alertMsg)
		}

		// Déterminer le premier jalon en attente
		if !isCompleted && firstPendingMilestone == nil {
			firstPendingMilestone = &analysis.Milestones[len(analysis.Milestones)-1]
		}

		// Mise à jour de la sévérité maximale
		rank := healthRanking[health]
		if rank > worstHealthRank {
			worstHealthRank = rank
			analysis.OverallHealth = health
		}
	}

	analysis.ActiveCriticalMilestone = firstPendingMilestone

	// Prise en compte du paramètre configurable de jours ouvrables par semaine
	workingDaysPerWeek := d.WorkingDaysPerWeek
	if workingDaysPerWeek != 6 {
		workingDaysPerWeek = 5 // Standard Algérie ANPP / MIPH par défaut
	}
	analysis.WorkingDaysPerWeek = workingDaysPerWeek

	requiredWorkDays := CalculateRequiredWorkingDays(analysis.Milestones)
	analysis.RequiredWorkingDays = requiredWorkDays

	if requiredWorkDays > 0 {
		recDate, holidaysCount := AddWorkingDaysWithHolidayCount(nowMidnight, requiredWorkDays, workingDaysPerWeek)
		analysis.RecommendedDepositDate = &recDate
		analysis.HolidaysExcludedCount = holidaysCount
		analysis.IsTargetDateSufficient = !targetMidnight.Before(recDate)
	} else {
		analysis.RecommendedDepositDate = &targetMidnight
		analysis.HolidaysExcludedCount = 0
		analysis.IsTargetDateSufficient = true
	}

	recDateFormatted := ""
	if analysis.RecommendedDepositDate != nil {
		recDateFormatted = analysis.RecommendedDepositDate.Format("02/01/2006")
	}

	holidaysSuffix := ""
	if analysis.HolidaysExcludedCount > 0 {
		holidaysSuffix = fmt.Sprintf(" et %d jours fériés ANPP déduits", analysis.HolidaysExcludedCount)
	}

	// Labels synthétiques et recommandation réglementaire avec date recommandée basée sur les jours ouvrables et fériés
	switch analysis.OverallHealth {
	case HealthCompleted:
		analysis.OverallHealthLabel = "Toutes les étapes critiques sont validées"
		analysis.RecommendedActionFR = "Dossier prêt pour la télé-soumission officielle à l'ANPP."
	case HealthOnTrack:
		analysis.OverallHealthLabel = "Chemin critique dans les temps (Conforme)"
		if firstPendingMilestone != nil {
			analysis.RecommendedActionFR = fmt.Sprintf("Poursuivre la préparation : %s (Échéance dans %d jours). Date butoir recommandée : %s (basée sur %d jours ouvrables/semaine%s, %d jours ouvrés restants).", firstPendingMilestone.NameFR, firstPendingMilestone.DaysRemaining, recDateFormatted, workingDaysPerWeek, holidaysSuffix, requiredWorkDays)
		}
	case HealthWarningTight:
		analysis.OverallHealthLabel = "Attention : Marge résiduelle faible sur les jalons critiques"
		if firstPendingMilestone != nil {
			analysis.RecommendedActionFR = fmt.Sprintf("Accélérer la finalisation de l'étape '%s' (Pôle %s). Date butoir recommandée : %s (calculée sur une base de %d jours ouvrables/semaine%s).", firstPendingMilestone.NameFR, firstPendingMilestone.DepartmentOwner, recDateFormatted, workingDaysPerWeek, holidaysSuffix)
		}
	case HealthCriticalDelay:
		analysis.OverallHealthLabel = "Alerte : Dérive active sur le chemin critique du dossier"
		if firstPendingMilestone != nil {
			analysis.RecommendedActionFR = fmt.Sprintf("Arbitrage requis par le Pharmacien DT sur l'étape '%s'. Date minimale recommandée : %s (basée sur %d jours ouvrables/semaine%s, soit %d jours ouvrés incompressibles).", firstPendingMilestone.NameFR, recDateFormatted, workingDaysPerWeek, holidaysSuffix, requiredWorkDays)
		}
	case HealthFatalBlock:
		analysis.OverallHealthLabel = "Rupture réglementaire : Délais incompressibles dépassés"
		if firstPendingMilestone != nil {
			analysis.RecommendedActionFR = fmt.Sprintf("Pour rattraper la dérive calendaire sur '%s', la date butoir minimale recommandée est le %s (basée sur %d jours ouvrables par semaine%s, soit %d jours ouvrés incompressibles).", firstPendingMilestone.NameFR, recDateFormatted, workingDaysPerWeek, holidaysSuffix, requiredWorkDays)
		} else {
			analysis.RecommendedActionFR = fmt.Sprintf("Date butoir minimale recommandée : %s (basée sur %d jours ouvrables par semaine%s).", recDateFormatted, workingDaysPerWeek, holidaysSuffix)
		}
	}

	return analysis
}

// DefaultAlgerianHolidays répertorie les jours fériés légaux et chômés en Algérie (Loi n° 63-278 modifiée par la Loi n° 23-07 du 3 mai 2023 - 3 jours pour l'Aïd)
var DefaultAlgerianHolidays = map[string]string{
	// Fêtes nationales fixes (format MM-DD)
	"01-01": "Jour de l'An",
	"01-12": "Yennayer (Nouvel An Amazigh)",
	"05-01": "Fête du Travail",
	"07-05": "Fête de l'Indépendance et de la Jeunesse",
	"11-01": "Fête de la Révolution (1er Novembre)",

	// Fêtes religieuses chômées payées 2026 (format YYYY-MM-DD - Loi 23-07 : 3 jours chômés pour Aïd El-Fitr et Aïd El-Adha)
	"2026-03-20": "Aïd El-Fitr (Jour 1)",
	"2026-03-21": "Aïd El-Fitr (Jour 2)",
	"2026-03-22": "Aïd El-Fitr (Jour 3)",
	"2026-05-27": "Aïd El-Adha (Jour 1)",
	"2026-05-28": "Aïd El-Adha (Jour 2)",
	"2026-05-29": "Aïd El-Adha (Jour 3)",
	"2026-06-16": "1er Moharram (Nouvel An de l'Hégire)",
	"2026-06-25": "Achoura (10 Moharram)",
	"2026-08-25": "El Mawlid Ennabawi Echarif",

	// Fêtes religieuses chômées payées 2027 (format YYYY-MM-DD - Loi 23-07 : 3 jours chômés pour Aïd El-Fitr et Aïd El-Adha)
	"2027-03-10": "Aïd El-Fitr (Jour 1)",
	"2027-03-11": "Aïd El-Fitr (Jour 2)",
	"2027-03-12": "Aïd El-Fitr (Jour 3)",
	"2027-05-17": "Aïd El-Adha (Jour 1)",
	"2027-05-18": "Aïd El-Adha (Jour 2)",
	"2027-05-19": "Aïd El-Adha (Jour 3)",
	"2027-06-06": "1er Moharram (Nouvel An de l'Hégire)",
	"2027-06-15": "Achoura (10 Moharram)",
	"2027-08-15": "El Mawlid Ennabawi Echarif",

	// Fêtes religieuses chômées payées 2028 (format YYYY-MM-DD - Loi 23-07 : 3 jours chômés pour Aïd El-Fitr et Aïd El-Adha)
	"2028-02-27": "Aïd El-Fitr (Jour 1)",
	"2028-02-28": "Aïd El-Fitr (Jour 2)",
	"2028-02-29": "Aïd El-Fitr (Jour 3)",
	"2028-05-05": "Aïd El-Adha (Jour 1)",
	"2028-05-06": "Aïd El-Adha (Jour 2)",
	"2028-05-07": "Aïd El-Adha (Jour 3)",
	"2028-05-25": "1er Moharram (Nouvel An de l'Hégire)",
	"2028-06-03": "Achoura (10 Moharram)",
	"2028-08-03": "El Mawlid Ennabawi Echarif",
}

// IsAlgerianHoliday vérifie si une date donnée correspond à un jour férié chômé officiel algérien
func IsAlgerianHoliday(t time.Time, customHolidays ...map[string]string) (bool, string) {
	var holidays map[string]string
	if len(customHolidays) > 0 && customHolidays[0] != nil {
		holidays = customHolidays[0]
	} else {
		holidays = DefaultAlgerianHolidays
	}

	// 1. Recherche par date complète "YYYY-MM-DD" (notamment fêtes religieuses)
	fullDate := t.Format("2006-01-02")
	if name, exists := holidays[fullDate]; exists {
		return true, name
	}

	// 2. Recherche par jour fixe annuel "MM-DD" (fêtes nationales)
	fixedDay := t.Format("01-02")
	if name, exists := holidays[fixedDay]; exists {
		return true, name
	}

	return false, ""
}

// AddWorkingDays calcule la date résultante en ajoutant un nombre de jours ouvrés réels,
// en excluant les jours de repos hebdomadaires (week-end algérien 5j ou 6j) ainsi que les jours fériés légaux algériens.
func AddWorkingDays(start time.Time, workingDays int, daysPerWeek int, customHolidays ...map[string]string) time.Time {
	res, _ := AddWorkingDaysWithHolidayCount(start, workingDays, daysPerWeek, customHolidays...)
	return res
}

// AddWorkingDaysWithHolidayCount effectue le calcul en retournant également le nombre de jours fériés rencontrés et exclus.
func AddWorkingDaysWithHolidayCount(start time.Time, workingDays int, daysPerWeek int, customHolidays ...map[string]string) (time.Time, int) {
	if workingDays <= 0 {
		return start, 0
	}
	if daysPerWeek != 6 {
		daysPerWeek = 5 // Standard Algérie ANPP / MIPH (Dimanche à Jeudi)
	}

	curr := start
	added := 0
	holidaysCount := 0

	for added < workingDays {
		curr = curr.AddDate(0, 0, 1)
		weekday := curr.Weekday() // 0 = Sunday, 1 = Monday, ..., 5 = Friday, 6 = Saturday

		// 1. Contrôle des week-ends
		isWeekend := false
		if daysPerWeek == 5 {
			// Dimanche-Jeudi : Vendredi (5) et Samedi (6) chômés
			isWeekend = (weekday == time.Friday || weekday == time.Saturday)
		} else if daysPerWeek == 6 {
			// Samedi-Jeudi : Vendredi (5) seul chômé
			isWeekend = (weekday == time.Friday)
		}

		if isWeekend {
			continue
		}

		// 2. Contrôle des jours fériés légaux algériens
		isHoliday, _ := IsAlgerianHoliday(curr, customHolidays...)
		if isHoliday {
			holidaysCount++
			continue
		}

		added++
	}
	return curr, holidaysCount
}

// CalculateRequiredWorkingDays estime le volume minimal de jours ouvrés nécessaires pour franchir les étapes en attente
func CalculateRequiredWorkingDays(milestones []SmartPlanningMilestone) int {
	maxWorkDays := 0
	workDaysByCode := map[SmartPlanningMilestoneCode]int{
		MilestoneJ240StabilitiesZoneIVb: 172, // ~240j calendaires * 5/7 (6 mois incubation continue + 60j analyse)
		MilestoneJ150QOSCheckerReview:   108, // ~150j calendaires * 5/7
		MilestoneJ60ETasjilAppointment:  43,  // ~60j calendaires * 5/7
		MilestoneJ30LeafletAndFees:      22,  // ~30j calendaires * 5/7
		MilestoneJ15PDTVisaPackageSeal:  11,  // ~15j calendaires * 5/7
		MilestoneJ0ANPPDepositDeadline:  0,
	}

	for _, m := range milestones {
		if !m.IsCompleted {
			if wd, ok := workDaysByCode[m.Code]; ok && wd > maxWorkDays {
				maxWorkDays = wd
			}
		}
	}
	return maxWorkDays
}

// checkMilestoneCompletion évalue si un jalon du rétroplanning a été concrètement franchi
func checkMilestoneCompletion(code SmartPlanningMilestoneCode, d *Dossier, itemsByCode map[string]ChecklistItem) (bool, string) {
	switch code {
	case MilestoneJ240StabilitiesZoneIVb:
		// Stabilités 3.2.P.8 au moins en revue ou scellées
		it, exists := itemsByCode["3.2.P.8"]
		if exists && isWorkflowStatusAtLeast(it.WorkflowStatus, WorkflowStatusUnderReview) {
			return true, fmt.Sprintf("Stabilités 3.2.P.8 au statut '%s'", it.WorkflowStatus)
		}
		return false, ""

	case MilestoneJ150QOSCheckerReview:
		// QOS 2.3 au moins en revue Checker
		it, exists := itemsByCode["2.3"]
		if exists && isWorkflowStatusAtLeast(it.WorkflowStatus, WorkflowStatusUnderReview) {
			return true, fmt.Sprintf("QOS 2.3 au statut '%s'", it.WorkflowStatus)
		}
		return false, ""

	case MilestoneJ60ETasjilAppointment:
		// Renseigné si référence de rendez-vous ou dépôt présent
		if d.DepositReceiptRef != nil && *d.DepositReceiptRef != "" {
			return true, fmt.Sprintf("Créneau réservé / Réf : %s", *d.DepositReceiptRef)
		}
		if d.TrackingNumber != nil && *d.TrackingNumber != "" {
			return true, fmt.Sprintf("Numéro de tracking E-TASJIL : %s", *d.TrackingNumber)
		}
		return false, ""

	case MilestoneJ30LeafletAndFees:
		// Notice FR 1.P.2.1 et Notice AR 1.P.2.2 au moins en revue Checker ou validées
		itFR, hasFR := itemsByCode["1.P.2.1"]
		itAR, hasAR := itemsByCode["1.P.2.2"]
		if hasFR && hasAR &&
			isWorkflowStatusAtLeast(itFR.WorkflowStatus, WorkflowStatusPendingApproval) &&
			isWorkflowStatusAtLeast(itAR.WorkflowStatus, WorkflowStatusPendingApproval) {
			return true, "Notice bilingue FR/AR validée par le Checker"
		}
		return false, ""

	case MilestoneJ15PDTVisaPackageSeal:
		// Visa PDT apposé ou dossier scellé pour dépôt
		if d.PDTSignedAt != nil && !d.PDTSignedAt.IsZero() {
			return true, fmt.Sprintf("Visa Pharmacien DT apposé le %s", d.PDTSignedAt.Format("02/01/2006"))
		}
		if d.IsLockedForDeposit {
			return true, "Dossier formellement verrouillé pour dépôt ANPP"
		}
		return false, ""

	case MilestoneJ0ANPPDepositDeadline:
		// Dossier déposé ou approuvé
		if d.Status == DossierStatusSubmittedANPP || d.Status == DossierStatusApprovedANPP {
			return true, fmt.Sprintf("Dossier au statut officiel '%s'", d.Status)
		}
		if d.ANPPDepositDate != nil && !d.ANPPDepositDate.IsZero() {
			return true, fmt.Sprintf("Dépôt effectif enregistré le %s", d.ANPPDepositDate.Format("02/01/2006"))
		}
		return false, ""
	}

	return false, ""
}

// CalculateSmartDueDates propage les dates cibles calculées aux items de la checklist CTD
// Permet d'aligner automatiquement chaque département sur son jalon critique
func CalculateSmartDueDates(analysis *SmartPlanningAnalysis, items []ChecklistItem) map[string]time.Time {
	dueDates := make(map[string]time.Time)

	// Dictionnaire des dates par code de jalon
	milestoneDateByCode := make(map[SmartPlanningMilestoneCode]time.Time)
	for _, m := range analysis.Milestones {
		milestoneDateByCode[m.Code] = m.TargetDate
	}

	dateJ240 := milestoneDateByCode[MilestoneJ240StabilitiesZoneIVb]
	dateJ150 := milestoneDateByCode[MilestoneJ150QOSCheckerReview]
	dateJ60 := milestoneDateByCode[MilestoneJ60ETasjilAppointment]
	dateJ30 := milestoneDateByCode[MilestoneJ30LeafletAndFees]
	dateJ15 := milestoneDateByCode[MilestoneJ15PDTVisaPackageSeal]

	for _, it := range items {
		code := it.ItemCode

		// Règle 1 : Module 3 Qualité (Stabilités, Méthodes, Spécifications) -> Échéance J-240
		if strings.HasPrefix(code, "3.") || strings.HasPrefix(code, "5.") {
			dueDates[code] = dateJ240
		} else if strings.HasPrefix(code, "2.") {
			// Règle 2 : Module 2 Synthèses Qualité / Cliniques (QOS) -> Échéance J-150
			dueDates[code] = dateJ150
		} else if strings.HasPrefix(code, "1.A") {
			// Règle 3 : Module 1 Administratif (Formulaires, Rendez-vous E-TASJIL) -> Échéance J-60
			if code == "1.A.7" {
				dueDates[code] = dateJ15 // Visa PDT -> J-15
			} else {
				dueDates[code] = dateJ60
			}
		} else if strings.HasPrefix(code, "1.P") {
			// Règle 4 : Module 1.P (Notices, RCP, Étuis, BV1/BV2) -> Échéance J-30
			if code == "1.P.3" || code == "1.P.8" || code == "1.P.9" {
				dueDates[code] = dateJ15 // Engagements officiels -> J-15
			} else {
				dueDates[code] = dateJ30
			}
		} else {
			// Par défaut : J-30
			dueDates[code] = dateJ30
		}
	}

	return dueDates
}
