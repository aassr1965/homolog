package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestANPPAppointment_Helpers(t *testing.T) {
	now := time.Now().UTC()
	inThreeDays := now.Add(3 * 24 * time.Hour)
	inTenDays := now.Add(10 * 24 * time.Hour)
	pastDate := now.Add(-2 * 24 * time.Hour)

	appUrgent := &ANPPAppointment{
		ID:              uuid.New(),
		AppointmentType: AppointmentTypeDepotInitial,
		ScheduledAt:     inThreeDays,
		Status:          AppointmentStatusPlanifie,
	}

	if !appUrgent.IsUrgent() {
		t.Errorf("Le rendez-vous dans 3 jours devrait être considéré comme urgent (J <= 7)")
	}
	if appUrgent.IsPast() {
		t.Errorf("Le rendez-vous futur ne devrait pas être passé")
	}

	appNotUrgent := &ANPPAppointment{
		ID:              uuid.New(),
		AppointmentType: AppointmentTypeCommission,
		ScheduledAt:     inTenDays,
		Status:          AppointmentStatusPlanifie,
	}

	if appNotUrgent.IsUrgent() {
		t.Errorf("Le rendez-vous dans 10 jours ne devrait pas être urgent")
	}

	appPast := &ANPPAppointment{
		ID:              uuid.New(),
		AppointmentType: AppointmentTypeExpertiseLNCM,
		ScheduledAt:     pastDate,
		Status:          AppointmentStatusEffectue,
	}

	if !appPast.IsPast() {
		t.Errorf("Le rendez-vous passé devrait renvoyer true pour IsPast")
	}
	if appPast.IsUrgent() {
		t.Errorf("Le rendez-vous passé ne devrait pas être urgent")
	}
}

func TestRegulatoryWatchItem_Helpers(t *testing.T) {
	now := time.Now().UTC()
	effectiveDate := now.Add(-30 * 24 * time.Hour)
	futureDate := now.Add(60 * 24 * time.Hour)

	itemEffective := &RegulatoryWatchItem{
		ID:               uuid.New(),
		TitleFR:          "Décret Exécutif N° 20-325",
		Category:         RegCatEnregistrementAMM,
		ImpactLevel:      ImpactLevelCritique,
		EffectiveDate:    effectiveDate,
		ComplianceStatus: ComplianceStatusConforme,
	}

	if !itemEffective.IsEffective() {
		t.Errorf("Le texte avec date d'effet passée devrait être considéré en vigueur")
	}
	if itemEffective.RequiresActionPlan() {
		t.Errorf("Le texte conforme ne requiert pas de plan d'action d'urgence")
	}

	itemActionRequired := &RegulatoryWatchItem{
		ID:               uuid.New(),
		TitleFR:          "Nouvelle Décision relative aux études de stabilité",
		Category:         RegCatEnregistrementAMM,
		ImpactLevel:      ImpactLevelCritique,
		EffectiveDate:    futureDate,
		ComplianceStatus: ComplianceStatusPlanActionRequis,
	}

	if !itemActionRequired.RequiresActionPlan() {
		t.Errorf("Le texte critique non conforme doit requérir un plan d'action")
	}
}
