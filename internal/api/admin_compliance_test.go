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
)

func TestAdmin_CreateUser_PasswordCompliance(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	auditSvc := audit.NewService(db.Pool)
	sm := auth.NewSessionManager(db.Pool, 60, 15)
	adminHandler := NewAdminHandler(db.Pool, sm, auditSvc)

	// Création d'une organisation de test
	orgID := uuid.New()
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Labo Test Admin', 'MIPH-ADMIN-01', 'Dr. Admin')
		ON CONFLICT (id) DO NOTHING
	`, orgID, "org-admin-"+uuid.New().String()[:8])

	// Admin exécutant
	adminUser := &auth.SessionUser{
		UserID:         uuid.New(),
		Email:          "admin@labo-test.dz",
		Role:           string(domain.RoleAdminSystemeLocal),
		OrganizationID: orgID,
		Permissions:    []string{string(domain.PermUserManage)},
	}

	// 1. Mot de passe trop faible (< 12 caractères) -> Doit être rejeté (400)
	weakPayload := CreateUserRequest{
		Email:     "user1@labo-test.dz",
		FirstName: "Nabil",
		LastName:  "Test",
		Role:      string(domain.RoleResponsableAffairesReg),
		Password:  "Court1!",
	}
	body, _ := json.Marshal(weakPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), UserContextKey, adminUser))
	w := httptest.NewRecorder()

	adminHandler.CreateUser(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Attendu HTTP 400 pour mot de passe trop court, obtenu: %d", w.Code)
	}

	// 2. Consultant externe sans date d'expiration -> Doit être rejeté (400)
	consultantNoExp := CreateUserRequest{
		Email:     "consultant@labo-test.dz",
		FirstName: "Karim",
		LastName:  "Consultant",
		Role:      string(domain.RoleConsultantExterne),
		Password:  "SuperPassword2026!#@$",
	}
	body, _ = json.Marshal(consultantNoExp)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), UserContextKey, adminUser))
	w = httptest.NewRecorder()

	adminHandler.CreateUser(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Attendu HTTP 400 pour consultant externe sans contrat d'expiration, obtenu: %d", w.Code)
	}

	// 3. Création valide avec mot de passe robuste Argon2id
	futureDate := time.Now().AddDate(1, 0, 0)
	validPayload := CreateUserRequest{
		Email:             "pharmacien.cq@labo-test.dz",
		FirstName:         "Yasmine",
		LastName:          "Benali",
		Role:              string(domain.RoleContributeurQualiteBPF),
		Password:          "QualitePharma2026!#@$",
		ContractExpiresAt: &futureDate,
	}
	body, _ = json.Marshal(validPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), UserContextKey, adminUser))
	w = httptest.NewRecorder()

	adminHandler.CreateUser(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Attendu HTTP 201 pour utilisateur valide, obtenu: %d (%s)", w.Code, w.Body.String())
	}

	var res map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	newUserIDStr, ok := res["user_id"].(string)
	if !ok || newUserIDStr == "" {
		t.Fatal("Identifiant utilisateur manquant dans la réponse de création")
	}

	// Vérifier le hash Argon2id en base
	var passwordHash string
	err := db.Pool.QueryRow(ctx, "SELECT password_hash FROM users WHERE id = $1", newUserIDStr).Scan(&passwordHash)
	if err != nil {
		t.Fatalf("Utilisateur non trouvé en base: %v", err)
	}
	match, err := auth.VerifyPassword("QualitePharma2026!#@$", passwordHash)
	if err != nil || !match {
		t.Fatalf("Le mot de passe haché Argon2id n'est pas vérifié: %v", err)
	}

	// 4. Tentative de recréation avec le même email -> Rejet 409 Conflict
	reqDup := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(body))
	reqDup = reqDup.WithContext(context.WithValue(reqDup.Context(), UserContextKey, adminUser))
	wDup := httptest.NewRecorder()

	adminHandler.CreateUser(wDup, reqDup)
	if wDup.Code != http.StatusConflict {
		t.Fatalf("Attendu HTTP 409 pour doublon d'email, obtenu: %d", wDup.Code)
	}
}

func TestAdmin_UpdateUserStatus_And_DisallowDelete(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	auditSvc := audit.NewService(db.Pool)
	sm := auth.NewSessionManager(db.Pool, 60, 15)
	adminHandler := NewAdminHandler(db.Pool, sm, auditSvc)

	orgID := uuid.New()
	adminID := uuid.New()
	targetUserID := uuid.New()

	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Labo Test Statut', 'MIPH-STATUT-01', 'Dr. Statut')
		ON CONFLICT (id) DO NOTHING
	`, orgID, "org-statut-"+uuid.New().String()[:8])

	// Création de l'utilisateur cible
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active, permissions_version)
		VALUES ($1, $2, 'cible@labo-test.dz', 'dummy-hash', 'Cible', 'Test', 'contributeur_rd_production', TRUE, 1)
	`, targetUserID, orgID)
	if err != nil {
		t.Fatalf("Erreur création utilisateur cible: %v", err)
	}

	adminUser := &auth.SessionUser{
		UserID:         adminID,
		Email:          "admin@labo-test.dz",
		Role:           string(domain.RoleAdminSystemeLocal),
		OrganizationID: orgID,
		Permissions:    []string{string(domain.PermUserManage)},
	}

	// 1. Auto-désactivation interdite
	selfPayload := UpdateUserStatusRequest{
		IsActive: false,
		Reason:   "Tentative auto-désactivation",
	}
	body, _ := json.Marshal(selfPayload)
	reqSelf := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/"+adminID.String()+"/status", bytes.NewReader(body))
	reqSelf = reqSelf.WithContext(context.WithValue(reqSelf.Context(), UserContextKey, adminUser))
	rctxSelf := chi.NewRouteContext()
	rctxSelf.URLParams.Add("id", adminID.String())
	reqSelf = reqSelf.WithContext(context.WithValue(reqSelf.Context(), chi.RouteCtxKey, rctxSelf))
	wSelf := httptest.NewRecorder()

	adminHandler.UpdateUserStatus(wSelf, reqSelf)
	if wSelf.Code != http.StatusBadRequest {
		t.Fatalf("Attendu HTTP 400 pour auto-désactivation, obtenu: %d", wSelf.Code)
	}

	// 2. Désactivation valide d'un collaborateur
	deactPayload := UpdateUserStatusRequest{
		IsActive: false,
		Reason:   "Départ du collaborateur vers un autre laboratoire (21 CFR Part 11)",
	}
	body, _ = json.Marshal(deactPayload)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/"+targetUserID.String()+"/status", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), UserContextKey, adminUser))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", targetUserID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	adminHandler.UpdateUserStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour désactivation, obtenu: %d (%s)", w.Code, w.Body.String())
	}

	// Vérifier l'état en base
	var isActive bool
	_ = db.Pool.QueryRow(ctx, "SELECT is_active FROM users WHERE id = $1", targetUserID).Scan(&isActive)
	if isActive {
		t.Fatal("L'utilisateur cible devrait être désactivé (is_active = FALSE)")
	}

	// 3. Rejet formel du DELETE physique (21 CFR Part 11)
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/"+targetUserID.String(), nil)
	reqDel = reqDel.WithContext(context.WithValue(reqDel.Context(), UserContextKey, adminUser))
	rctxDel := chi.NewRouteContext()
	rctxDel.URLParams.Add("id", targetUserID.String())
	reqDel = reqDel.WithContext(context.WithValue(reqDel.Context(), chi.RouteCtxKey, rctxDel))
	wDel := httptest.NewRecorder()

	adminHandler.DisallowDeleteUser(wDel, reqDel)
	if wDel.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Attendu HTTP 405 Method Not Allowed pour suppression physique, obtenu: %d", wDel.Code)
	}
}

func TestAdmin_PasswordReset_And_CustomPermissions(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	auditSvc := audit.NewService(db.Pool)
	sm := auth.NewSessionManager(db.Pool, 60, 15)
	adminHandler := NewAdminHandler(db.Pool, sm, auditSvc)

	orgID := uuid.New()
	adminID := uuid.New()
	targetUserID := uuid.New()

	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Labo Test Perms', 'MIPH-PERMS-01', 'Dr. Perms')
		ON CONFLICT (id) DO NOTHING
	`, orgID, "org-perms-"+uuid.New().String()[:8])

	// Admin en base (nécessaire pour la contrainte foreign key assigned_by)
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active, permissions_version)
		VALUES ($1, $2, 'admin@labo-test.dz', 'admin-hash', 'Admin', 'Local', 'admin_systeme_local', TRUE, 1)
	`, adminID, orgID)

	_, err := db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active, permissions_version)
		VALUES ($1, $2, 'perms.user@labo-test.dz', 'old-hash', 'Mohamed', 'Larbi', 'charge_enregistrement', TRUE, 1)
	`, targetUserID, orgID)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}

	adminUser := &auth.SessionUser{
		UserID:         adminID,
		Email:          "admin@labo-test.dz",
		Role:           string(domain.RoleAdminSystemeLocal),
		OrganizationID: orgID,
		Permissions:    []string{string(domain.PermUserManage)},
	}

	// 1. Réinitialisation mot de passe
	resetPayload := ResetPasswordRequest{
		NewPassword: "NewSuperPassword2026!#@",
		Reason:      "Perte des identifiants déclarée par le collaborateur",
	}
	body, _ := json.Marshal(resetPayload)
	reqReset := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+targetUserID.String()+"/reset-password", bytes.NewReader(body))
	reqReset = reqReset.WithContext(context.WithValue(reqReset.Context(), UserContextKey, adminUser))
	rctxReset := chi.NewRouteContext()
	rctxReset.URLParams.Add("id", targetUserID.String())
	reqReset = reqReset.WithContext(context.WithValue(reqReset.Context(), chi.RouteCtxKey, rctxReset))
	wReset := httptest.NewRecorder()

	adminHandler.ResetUserPassword(wReset, reqReset)
	if wReset.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour réinitialisation mot de passe, obtenu: %d (%s)", wReset.Code, wReset.Body.String())
	}

	// 2. Attribution d'une surcharge de permission (ex: REF_EDIT accordé à un chargé d'enregistrement)
	trueVal := true
	permPayload := UpdateUserPermissionsRequest{
		Overrides: []OverridePermissionItem{
			{PermissionID: "REF_EDIT", Granted: &trueVal},
		},
		Reason: "Dérogation temporaire pour enrichir le référentiel des DM",
	}
	body, _ = json.Marshal(permPayload)
	reqPerm := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/"+targetUserID.String()+"/permissions", bytes.NewReader(body))
	reqPerm = reqPerm.WithContext(context.WithValue(reqPerm.Context(), UserContextKey, adminUser))
	rctxPerm := chi.NewRouteContext()
	rctxPerm.URLParams.Add("id", targetUserID.String())
	reqPerm = reqPerm.WithContext(context.WithValue(reqPerm.Context(), chi.RouteCtxKey, rctxPerm))
	wPerm := httptest.NewRecorder()

	adminHandler.UpdateUserPermissions(wPerm, reqPerm)
	if wPerm.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour mise à jour permissions, obtenu: %d (%s)", wPerm.Code, wPerm.Body.String())
	}

	// Vérifier l'insertion dans user_permissions
	var granted bool
	err = db.Pool.QueryRow(ctx, "SELECT granted FROM user_permissions WHERE user_id = $1 AND permission_id = 'REF_EDIT'", targetUserID).Scan(&granted)
	if err != nil || !granted {
		t.Fatalf("La permission surchargée REF_EDIT n'a pas été trouvée en base: %v", err)
	}

	// Vérifier que permissions_version a été incrémenté (> 1)
	var permVer int
	_ = db.Pool.QueryRow(ctx, "SELECT permissions_version FROM users WHERE id = $1", targetUserID).Scan(&permVer)
	if permVer <= 1 {
		t.Fatalf("permissions_version devrait avoir été incrémenté, actuel: %d", permVer)
	}
}
