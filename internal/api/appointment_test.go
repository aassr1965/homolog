package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

func TestAppointmentAndRegulatoryWatchAPI(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-appreg-" + testOrgID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DELETE FROM anpp_deposit_appointments WHERE organization_id = $1;
			DELETE FROM regulatory_watch_items WHERE organization_id = $1;
			DELETE FROM users WHERE organization_id = $1;
			DELETE FROM organizations WHERE id = $1;
		`, testOrgID)
	}()

	// 1. Organisation de test
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Laboratoires Veille & Séances ANPP', 'MIPH-2026-APPREG', 'Dr. Yacine Benkhalid')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation: %v", err)
	}

	// 2. Utilisateur test
	dtUserID := uuid.New()
	pwdHash, _ := auth.HashPassword("TestPass123!#", nil)
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, 'dt-appreg@test.dz', $3, 'Yacine', 'B.', 'directeur_technique', TRUE)
	`, dtUserID, testOrgID, pwdHash)
	if err != nil {
		t.Fatalf("échec création utilisateur: %v", err)
	}

	auditService := audit.NewService(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	appRepo := repository.NewAppointmentRepository(db.Pool)
	regRepo := repository.NewRegulatoryWatchRepository(db.Pool)

	appHandler := NewAppointmentHandler(appRepo, dossierRepo, auditService)
	regHandler := NewRegulatoryWatchHandler(regRepo, auditService)

	dtUser := &auth.SessionUser{
		UserID:         dtUserID,
		OrganizationID: testOrgID,
		Email:          "dt-appreg@test.dz",
		FirstName:      "Yacine",
		LastName:       "B.",
		Role:           "directeur_technique",
		Permissions:    []string{string(domain.PermDossierView), string(domain.PermDossierCreate), string(domain.PermRefView), string(domain.PermRefEdit)},
	}

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), UserContextKey, dtUser)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})

	r.Get("/api/v1/appointments", appHandler.ListAppointments)
	r.Post("/api/v1/appointments", appHandler.CreateAppointment)
	r.Get("/api/v1/appointments/{appId}", appHandler.GetAppointment)
	r.Put("/api/v1/appointments/{appId}/status", appHandler.UpdateAppointmentStatus)

	r.Get("/api/v1/regulatory-watch", regHandler.ListRegulatoryWatch)
	r.Post("/api/v1/regulatory-watch", regHandler.CreateRegulatoryWatchItem)
	r.Get("/api/v1/regulatory-watch/{itemId}", regHandler.GetRegulatoryWatchItem)
	r.Put("/api/v1/regulatory-watch/{itemId}/analysis", regHandler.UpdateRegulatoryWatchAnalysis)

	// =========================================================================
	// Test 1 : Création d'un rendez-vous de séance ANPP
	// =========================================================================
	schedDate := time.Now().UTC().Add(4 * 24 * time.Hour).Format(time.RFC3339)
	createAppPayload := map[string]any{
		"appointment_type": "seance_commission",
		"scheduled_at":     schedDate,
		"location":         "ANPP - Salle des Commissions N°1, Dely Ibrahim",
		"slot_reference":   "RDV-ANPP-2026-N°0892",
		"contact_officer":  "Dr. Ziani (Président de Commission)",
		"notes":            "Audition contradictoire sur les données de bioéquivalence.",
	}
	bodyBytes, _ := json.Marshal(createAppPayload)

	req := httptest.NewRequest("POST", "/api/v1/appointments", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Test 1: statut attendu 201, reçu %d: %s", rr.Code, rr.Body.String())
	}

	var createdApp domain.ANPPAppointment
	json.Unmarshal(rr.Body.Bytes(), &createdApp)
	if createdApp.SlotReference != "RDV-ANPP-2026-N°0892" {
		t.Errorf("Test 1: slot reference mismatch: %s", createdApp.SlotReference)
	}
	t.Logf("Test 1 PASS: Rendez-vous ANPP créé avec succès (ID: %s)", createdApp.ID)

	// =========================================================================
	// Test 2 : Liste des rendez-vous et calcul d'urgence J <= 7
	// =========================================================================
	req = httptest.NewRequest("GET", "/api/v1/appointments?upcoming=true", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Test 2: statut attendu 200, reçu %d: %s", rr.Code, rr.Body.String())
	}

	var listAppResp struct {
		Appointments []domain.ANPPAppointment `json:"appointments"`
		Total        int                      `json:"total"`
		UrgentCount  int                      `json:"urgent_count"`
	}
	json.Unmarshal(rr.Body.Bytes(), &listAppResp)

	if listAppResp.Total == 0 || listAppResp.UrgentCount == 0 {
		t.Errorf("Test 2: attendu au moins 1 rendez-vous urgent (dans 4 jours), reçu %d (total: %d)", listAppResp.UrgentCount, listAppResp.Total)
	}
	t.Logf("Test 2 PASS: Liste des rendez-vous ANPP (Total: %d, Urgents J<=7: %d)", listAppResp.Total, listAppResp.UrgentCount)

	// =========================================================================
	// Test 3 : Changement de statut du rendez-vous vers 'confirme_anpp'
	// =========================================================================
	statusPayload := `{"status": "confirme_anpp", "notes": "Convocation officielle reçue par courriel le 14/11"}`
	req = httptest.NewRequest("PUT", "/api/v1/appointments/"+createdApp.ID.String()+"/status", bytes.NewBufferString(statusPayload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Test 3: statut attendu 200, reçu %d: %s", rr.Code, rr.Body.String())
	}
	var updatedApp domain.ANPPAppointment
	json.Unmarshal(rr.Body.Bytes(), &updatedApp)
	if updatedApp.Status != domain.AppointmentStatusConfirmeANPP {
		t.Errorf("Test 3: statut attendu confirme_anpp, reçu %s", updatedApp.Status)
	}
	t.Logf("Test 3 PASS: Statut du rendez-vous mis à jour vers confirme_anpp")

	// =========================================================================
	// Test 4 : Création d'un texte de Veille Réglementaire JORADP
	// =========================================================================
	regPayload := map[string]any{
		"title_fr":          "Nouvelles Normes d'Essais de Dissolution In Vitro",
		"title_ar":          "معايير جديدة لاختبارات الانحلال المخبري",
		"legal_reference":   "Circulaire ANPP N° 04/2026",
		"joradp_number":     "Bulletin Officiel ANPP N° 12",
		"published_date":    "2026-06-01",
		"effective_date":    "2026-07-01",
		"category":          "enregistrement_amm",
		"impact_level":      "critique",
		"summary_fr":        "Harmonisation des profils de dissolution comparative à pH 1.2, 4.5 et 6.8 selon ICH Q4B.",
		"compliance_status": "plan_action_requis",
	}
	bodyBytes, _ = json.Marshal(regPayload)

	req = httptest.NewRequest("POST", "/api/v1/regulatory-watch", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Test 4: statut attendu 201, reçu %d: %s", rr.Code, rr.Body.String())
	}
	var createdReg domain.RegulatoryWatchItem
	json.Unmarshal(rr.Body.Bytes(), &createdReg)
	if createdReg.LegalReference != "Circulaire ANPP N° 04/2026" {
		t.Errorf("Test 4: legal reference mismatch")
	}
	t.Logf("Test 4 PASS: Texte de veille réglementaire enregistré (ID: %s)", createdReg.ID)

	// =========================================================================
	// Test 5 : Mise à jour de l'analyse d'impact et conformité
	// =========================================================================
	analysisPayload := `{"impact_analysis": "Études de dissolution complémentaires réalisées au laboratoire de contrôle; conformité atteinte.", "compliance_status": "conforme"}`
	req = httptest.NewRequest("PUT", "/api/v1/regulatory-watch/"+createdReg.ID.String()+"/analysis", bytes.NewBufferString(analysisPayload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Test 5: statut attendu 200, reçu %d: %s", rr.Code, rr.Body.String())
	}
	var updatedReg domain.RegulatoryWatchItem
	json.Unmarshal(rr.Body.Bytes(), &updatedReg)
	if updatedReg.ComplianceStatus != domain.ComplianceStatusConforme {
		t.Errorf("Test 5: statut attendu conforme, reçu %s", updatedReg.ComplianceStatus)
	}
	t.Logf("Test 5 PASS: Analyse d'impact mise à jour et statut passé à conforme")
}
