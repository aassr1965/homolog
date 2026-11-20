package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

func TestTenantRegulatorySettingsAndVersioning(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	orgAID := uuid.New()
	orgBID := uuid.New()
	slugA := "org-set-a-" + orgAID.String()[:8]
	slugB := "org-set-b-" + orgBID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DELETE FROM tenant_regulatory_settings WHERE organization_id IN ($1, $2);
			DELETE FROM users WHERE organization_id IN ($1, $2);
			DELETE FROM organizations WHERE id IN ($1, $2);
		`, orgAID, orgBID)
	}()

	// 1. Création des 2 organisations
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES 
		($1, $2, 'Laboratoire Alpha Test Settings', 'MIPH-2026-SET-A', 'Dr. Alpha DT'),
		($3, $4, 'Laboratoire Beta Test Settings', 'MIPH-2026-SET-B', 'Dr. Beta DT');
	`, orgAID, slugA, orgBID, slugB)
	if err != nil {
		t.Fatalf("Échec création organisations de test: %v", err)
	}

	// 2. Création des utilisateurs
	dtUserAID := uuid.New()
	adminUserAID := uuid.New()
	dtUserBID := uuid.New()
	pwdHash, _ := auth.HashPassword("TestPass123!", nil)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role)
		VALUES 
		($1, $2, 'dt.alpha@lab.dz', $3, 'Alpha', 'DT', 'directeur_technique'),
		($4, $2, 'admin.alpha@lab.dz', $3, 'Alpha', 'Admin', 'admin'),
		($5, $6, 'dt.beta@lab.dz', $3, 'Beta', 'DT', 'directeur_technique');
	`, dtUserAID, orgAID, pwdHash, adminUserAID, dtUserBID, orgBID)
	if err != nil {
		t.Fatalf("Échec création utilisateurs: %v", err)
	}

	auditService := audit.NewService(db.Pool)
	settingsRepo := repository.NewSettingsRepository(db.Pool)
	settingsHandler := NewSettingsHandler(settingsRepo, auditService)

	// Router de test
	router := chi.NewRouter()
	router.Get("/api/v1/admin/settings/regulatory", settingsHandler.ListTenantRegulatorySettings)
	router.Get("/api/v1/admin/settings/regulatory/{key}/history", settingsHandler.GetTenantSettingHistory)
	router.Put("/api/v1/admin/settings/regulatory/{key}", settingsHandler.UpdateTenantRegulatorySetting)
	router.Post("/api/v1/admin/settings/regulatory/{key}/approve", settingsHandler.ApproveTenantRegulatorySetting)

	// =========================================================================
	// TEST 1 : Initialisation automatique & complétude des paramètres
	// =========================================================================
	t.Run("Initialisation automatique des paramètres par défaut", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings/regulatory", nil)
		ctxUser := context.WithValue(req.Context(), UserContextKey, &auth.SessionUser{
			UserID:         adminUserAID,
			Email:          "admin.alpha@lab.dz",
			Role:           "admin",
			OrganizationID: orgAID,
		})
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req.WithContext(ctxUser))

		if rr.Code != http.StatusOK {
			t.Fatalf("Statut inattendu: attendu %d, reçu %d - Body: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var resp struct {
			Success  bool                              `json:"success"`
			Count    int                               `json:"count"`
			Settings []*domain.TenantRegulatorySetting `json:"settings"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Décodage JSON impossible: %v", err)
		}

		if resp.Count < 15 {
			t.Fatalf("Attendu au moins 15 paramètres réglementaires par défaut, reçu: %d", resp.Count)
		}

		// Vérifier la présence du RIB DM et du Compte Trésor
		foundRIB := false
		foundTreasury := false
		for _, s := range resp.Settings {
			if s.SettingKey == domain.SettingFinBankRibDM && s.SettingValue == "00100622030000064395" {
				foundRIB = true
			}
			if s.SettingKey == domain.SettingFinTreasuryAccountMed && strings.Contains(s.SettingValue, "302.145") {
				foundTreasury = true
			}
		}
		if !foundRIB {
			t.Errorf("Clé FIN_BANK_RIB_DM introuvable ou valeur par défaut incorrecte")
		}
		if !foundTreasury {
			t.Errorf("Clé FIN_TREASURY_ACCOUNT_MEDICINE introuvable ou valeur par défaut incorrecte")
		}
	})

	// =========================================================================
	// TEST 2 : Workflow Maker-Checker (Proposition Admin -> Statut Pending)
	// =========================================================================
	t.Run("Proposition de modification Maker (Admin)", func(t *testing.T) {
		payload, _ := json.Marshal(domain.UpdateTenantSettingRequest{
			NewValue:          "00100622039999999999",
			RegulatoryReason:  "Arrêté interministériel du 15/11/2026 modifiant la domiciliation bancaire DDM",
			RequireDTApproval: true,
		})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/regulatory/FIN_BANK_RIB_DM", bytes.NewReader(payload))
		ctxUser := context.WithValue(req.Context(), UserContextKey, &auth.SessionUser{
			UserID:         adminUserAID,
			Email:          "admin.alpha@lab.dz",
			Role:           "admin",
			OrganizationID: orgAID,
		})
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req.WithContext(ctxUser))

		if rr.Code != http.StatusOK {
			t.Fatalf("Statut inattendu: attendu 200, reçu %d - Body: %s", rr.Code, rr.Body.String())
		}

		// Vérifier que le paramètre est bien en attente sans écraser la valeur courante
		setting, err := settingsRepo.GetSettingByKey(ctx, orgAID, domain.SettingFinBankRibDM)
		if err != nil {
			t.Fatalf("Erreur chargement setting: %v", err)
		}
		if setting.Status != domain.SettingStatusPendingApprovalDT {
			t.Errorf("Statut attendu PENDING_APPROVAL_DT, reçu: %s", setting.Status)
		}
		if setting.SettingValue != "00100622030000064395" {
			t.Errorf("La valeur en vigueur doit rester inchangée avant approbation DT, reçu: %s", setting.SettingValue)
		}
		if setting.PendingValue == nil || *setting.PendingValue != "00100622039999999999" {
			t.Errorf("La valeur proposée n'a pas été enregistrée correctement")
		}
	})

	// =========================================================================
	// TEST 3 : Approbation solennelle DT & Versioning temporel ALCOA+
	// =========================================================================
	t.Run("Approbation DT et création d'une nouvelle version temporelle", func(t *testing.T) {
		// Tentative de rejet sans engagement solennel
		payloadNoSolemn, _ := json.Marshal(domain.ApproveTenantSettingRequest{
			SolemnCheck: false,
		})
		reqBad := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings/regulatory/FIN_BANK_RIB_DM/approve", bytes.NewReader(payloadNoSolemn))
		ctxDT := context.WithValue(reqBad.Context(), UserContextKey, &auth.SessionUser{
			UserID:         dtUserAID,
			Email:          "dt.alpha@lab.dz",
			Role:           "directeur_technique",
			OrganizationID: orgAID,
		})
		rrBad := httptest.NewRecorder()
		router.ServeHTTP(rrBad, reqBad.WithContext(ctxDT))
		if rrBad.Code != http.StatusUnprocessableEntity {
			t.Fatalf("L'approbation sans engagement solennel doit être rejetée (422), reçu: %d", rrBad.Code)
		}

		// Approbation solennelle valide
		payloadValid, _ := json.Marshal(domain.ApproveTenantSettingRequest{
			SolemnCheck: true,
			TOTPCode:    "123456",
		})
		reqOK := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings/regulatory/FIN_BANK_RIB_DM/approve", bytes.NewReader(payloadValid))
		rrOK := httptest.NewRecorder()
		router.ServeHTTP(rrOK, reqOK.WithContext(ctxDT))

		if rrOK.Code != http.StatusOK {
			t.Fatalf("Échec approbation DT: attendu 200, reçu %d - Body: %s", rrOK.Code, rrOK.Body.String())
		}

		// Vérifier que la nouvelle valeur est désormais active
		updated, err := settingsRepo.GetSettingByKey(ctx, orgAID, domain.SettingFinBankRibDM)
		if err != nil {
			t.Fatalf("Erreur chargement setting: %v", err)
		}
		if updated.SettingValue != "00100622039999999999" {
			t.Errorf("Nouvelle valeur non active: %s", updated.SettingValue)
		}
		if updated.Status != domain.SettingStatusActive {
			t.Errorf("Statut attendu ACTIVE, reçu: %s", updated.Status)
		}

		// Vérifier l'historique temporel (l'ancienne version doit exister avec valid_to renseigné)
		history, err := settingsRepo.GetSettingHistory(ctx, orgAID, domain.SettingFinBankRibDM)
		if err != nil {
			t.Fatalf("Erreur historique: %v", err)
		}
		if len(history) < 2 {
			t.Fatalf("Attendu au moins 2 versions dans l'historique, reçu: %d", len(history))
		}

		historicalVer := history[1] // la plus ancienne
		if historicalVer.SettingValue != "00100622030000064395" {
			t.Errorf("Valeur historique altérée: attendu 00100622030000064395, reçu: %s", historicalVer.SettingValue)
		}
		if historicalVer.ValidTo == nil {
			t.Errorf("ValidTo de la version historique doit être renseigné")
		}
		if historicalVer.IsActive {
			t.Errorf("La version historique ne doit plus être active")
		}
	})

	// =========================================================================
	// TEST 4 : Isolation Multi-Tenant Stricte (Tenant A != Tenant B)
	// =========================================================================
	t.Run("Isolation multi-tenant stricte entre organisations", func(t *testing.T) {
		// Initialiser Tenant B
		err := settingsRepo.EnsureTenantDefaults(ctx, orgBID)
		if err != nil {
			t.Fatalf("Erreur init Tenant B: %v", err)
		}

		// Vérifier que Tenant B a toujours sa valeur d'origine par défaut
		settingB, err := settingsRepo.GetSettingByKey(ctx, orgBID, domain.SettingFinBankRibDM)
		if err != nil {
			t.Fatalf("Erreur lecture Tenant B: %v", err)
		}
		if settingB.SettingValue != "00100622030000064395" {
			t.Fatalf("Fuite multi-tenant ! Tenant B a été altéré par Tenant A, valeur: %s", settingB.SettingValue)
		}

		// Et Tenant A a bien sa valeur personnalisée
		settingA, err := settingsRepo.GetSettingByKey(ctx, orgAID, domain.SettingFinBankRibDM)
		if err != nil {
			t.Fatalf("Erreur lecture Tenant A: %v", err)
		}
		if settingA.SettingValue != "00100622039999999999" {
			t.Fatalf("Valeur Tenant A inattendue: %s", settingA.SettingValue)
		}
	})

	// =========================================================================
	// TEST 5 : Mise à jour directe par le Pharmacien DT
	// =========================================================================
	t.Run("Mise à jour directe d'un barème par le Pharmacien DT", func(t *testing.T) {
		payload, _ := json.Marshal(domain.UpdateTenantSettingRequest{
			NewValue:          "380000.00",
			RegulatoryReason:  "Ajustement barème officiel 2027",
			RequireDTApproval: false, // direct apply car rôle DT
		})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/regulatory/FEE_MEDICINE_NI_DZD", bytes.NewReader(payload))
		ctxDT := context.WithValue(req.Context(), UserContextKey, &auth.SessionUser{
			UserID:         dtUserAID,
			Email:          "dt.alpha@lab.dz",
			Role:           "directeur_technique",
			OrganizationID: orgAID,
		})
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req.WithContext(ctxDT))

		if rr.Code != http.StatusOK {
			t.Fatalf("Échec direct apply DT: attendu 200, reçu %d - Body: %s", rr.Code, rr.Body.String())
		}

		setting, err := settingsRepo.GetSettingByKey(ctx, orgAID, domain.SettingFeeMedicineNI)
		if err != nil {
			t.Fatalf("Erreur chargement barème: %v", err)
		}
		if setting.SettingValue != "380000.00" {
			t.Fatalf("Barème attendu 380000.00, reçu: %s", setting.SettingValue)
		}
		if setting.Status != domain.SettingStatusActive {
			t.Fatalf("Statut attendu ACTIVE, reçu: %s", setting.Status)
		}
	})

	// =========================================================================
	// TEST 6 : Prise en compte dynamique dans les États Financiers (ES-FIN-01)
	// =========================================================================
	t.Run("Prise en compte dynamique des paramètres du tenant dans ES-FIN-01", func(t *testing.T) {
		dmID := uuid.New()
		dossierID := uuid.New()
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO medical_devices (id, organization_id, commercial_name_fr, manufacturer_reference, dm_class)
			VALUES ($1, $2, 'Seringue Stérile 5ml Test', 'REF-SRG-5ML', 'IIa');
		`, dmID, orgAID)
		if err != nil {
			t.Fatalf("Échec création DM: %v", err)
		}

		_, err = db.Pool.Exec(ctx, `
			INSERT INTO dossiers (id, organization_id, dossier_code, internal_reference, procedure_code, product_type, medical_device_id, status)
			VALUES ($1, $2, 'DOS-2026-DM-TEST', 'REF-INT-DM-01', 'HOMOLOG_DM', 'dispositif_medical', $3, 'draft');
		`, dossierID, orgAID, dmID)
		if err != nil {
			t.Fatalf("Échec création dossier DM: %v", err)
		}

		dossierRepo := repository.NewDossierRepository(db.Pool)
		workflowRepo := repository.NewWorkflowRepository(db.Pool)
		pharmaRepo := repository.NewPharmaRepository(db.Pool)
		repHandler := NewReportHandler(dossierRepo, workflowRepo, pharmaRepo, auditService, db.Pool)

		rRep := chi.NewRouter()
		rRep.Get("/api/v1/dossiers/{id}/reports/{code}", repHandler.GenerateDossierReport)

		reqRep := httptest.NewRequest(http.MethodGet, "/api/v1/dossiers/"+dossierID.String()+"/reports/es-fin-01", nil)
		ctxDT := context.WithValue(reqRep.Context(), UserContextKey, &auth.SessionUser{
			UserID:         dtUserAID,
			Email:          "dt.alpha@lab.dz",
			Role:           "directeur_technique",
			OrganizationID: orgAID,
		})
		rrRep := httptest.NewRecorder()
		rRep.ServeHTTP(rrRep, reqRep.WithContext(ctxDT))

		if rrRep.Code != http.StatusOK {
			t.Fatalf("Statut inattendu rapport: %d - Body: %s", rrRep.Code, rrRep.Body.String())
		}

		htmlContent := rrRep.Body.String()
		// Vérifier que le RIB personnalisé du Tenant A est présent
		if !strings.Contains(htmlContent, "00100622039999999999") {
			t.Errorf("Le RIB personnalisé du Tenant A n'apparaît pas dans ES-FIN-01: %s", htmlContent)
		}
	})
}
