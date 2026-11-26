package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/reporting"
)

// ANPPTemplateHandler gère la consultation, le téléchargement et la surcharge en base des gabarits officiels ANPP.
type ANPPTemplateHandler struct {
	pool         *pgxpool.Pool
	auditService *audit.Service
}

// NewANPPTemplateHandler initialise le handler et enregistre le résolveur dynamique dans le package reporting.
func NewANPPTemplateHandler(pool *pgxpool.Pool, auditService *audit.Service) *ANPPTemplateHandler {
	h := &ANPPTemplateHandler{pool: pool, auditService: auditService}
	reporting.RegisterTemplateOverrideResolver(h.resolveOverride)
	return h
}

func (h *ANPPTemplateHandler) resolveOverride(orgID, filename string) ([]byte, error) {
	if h == nil || h.pool == nil {
		return nil, nil
	}
	clean := path.Base(filename)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var data []byte
	if parsedOrg, err := uuid.Parse(strings.TrimSpace(orgID)); err == nil && parsedOrg != uuid.Nil {
		err = h.pool.QueryRow(ctx, `
			SELECT content_bytes FROM anpp_template_overrides
			WHERE template_name = $1 AND (organization_id = $2 OR organization_id IS NULL)
			ORDER BY (organization_id IS NOT NULL) DESC
			LIMIT 1
		`, clean, parsedOrg).Scan(&data)
		if err == nil && len(data) > 0 {
			return data, nil
		}
		return nil, nil
	}

	err := h.pool.QueryRow(ctx, `
		SELECT content_bytes FROM anpp_template_overrides
		WHERE template_name = $1 AND organization_id IS NULL
		LIMIT 1
	`, clean).Scan(&data)
	if err == nil && len(data) > 0 {
		return data, nil
	}
	return nil, nil
}

// ANPPTemplateItem décrit l'état d'un gabarit officiel ANPP (embarqué ou surchargé).
type ANPPTemplateItem struct {
	Name           string  `json:"name"`
	Source         string  `json:"source"` // "embedded", "organization", "global"
	Version        int     `json:"version"`
	SHA256         string  `json:"sha256"`
	SizeBytes      int     `json:"size_bytes"`
	UploadedByName string  `json:"uploaded_by_name,omitempty"`
	Notes          string  `json:"notes,omitempty"`
	UpdatedAt      *string `json:"updated_at,omitempty"`
}

func isValidKnownANPPTemplate(name string) bool {
	for _, n := range reporting.ListANPPTemplates() {
		if n == name {
			return true
		}
	}
	return false
}

func canManageOrgTemplates(role string) bool {
	switch role {
	case "directeur_technique", "responsable_affaires_reglementaires", "admin_systeme_local", "super_admin_editeur":
		return true
	default:
		return false
	}
}

func canManageGlobalTemplates(role string) bool {
	return role == "super_admin_editeur" || role == "admin_systeme_local"
}

// ListTemplates retourne la liste des 28 gabarits ANPP avec leur source active (embarqué vs surchargé).
func (h *ANPPTemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*auth.SessionUser)
	if !ok || user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	type overrideMeta struct {
		scope          string
		version        int
		sha256         string
		sizeBytes      int
		uploadedByName string
		notes          string
		updatedAt      time.Time
	}

	overrides := make(map[string]overrideMeta)
	rows, err := h.pool.Query(r.Context(), `
		SELECT template_name,
		       CASE WHEN organization_id IS NOT NULL THEN 'organization' ELSE 'global' END AS scope,
		       version, sha256_hash, OCTET_LENGTH(content_bytes), uploaded_by_name, notes, updated_at
		FROM anpp_template_overrides
		WHERE organization_id = $1 OR organization_id IS NULL
		ORDER BY (organization_id IS NOT NULL) ASC
	`, user.OrganizationID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name, scope, sha, byName, notes string
			var ver, size int
			var upd time.Time
			if scanErr := rows.Scan(&name, &scope, &ver, &sha, &size, &byName, &notes, &upd); scanErr == nil {
				// organization vient après global (ASC), donc écrase global si présent
				overrides[name] = overrideMeta{
					scope:          scope,
					version:        ver,
					sha256:         sha,
					sizeBytes:      size,
					uploadedByName: byName,
					notes:          notes,
					updatedAt:      upd,
				}
			}
		}
	}

	names := reporting.ListANPPTemplates()
	items := make([]ANPPTemplateItem, 0, len(names))
	for _, name := range names {
		if ov, exists := overrides[name]; exists {
			ts := ov.updatedAt.Format(time.RFC3339)
			items = append(items, ANPPTemplateItem{
				Name:           name,
				Source:         ov.scope,
				Version:        ov.version,
				SHA256:         ov.sha256,
				SizeBytes:      ov.sizeBytes,
				UploadedByName: ov.uploadedByName,
				Notes:          ov.notes,
				UpdatedAt:      &ts,
			})
			continue
		}
		raw, _ := reporting.GetANPPTemplate(name)
		sum := sha256.Sum256(raw)
		items = append(items, ANPPTemplateItem{
			Name:      name,
			Source:    "embedded",
			Version:   0,
			SHA256:    hex.EncodeToString(sum[:]),
			SizeBytes: len(raw),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"templates":         items,
		"can_upload_org":    canManageOrgTemplates(user.Role),
		"can_upload_global": canManageGlobalTemplates(user.Role),
	})
}

// DownloadTemplate télécharge le fichier brut actif (.docx ou .xlsx) pour inspection ou adaptation locale.
func (h *ANPPTemplateHandler) DownloadTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*auth.SessionUser)
	if !ok || user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}
	name := path.Base(chi.URLParam(r, "name"))
	if !isValidKnownANPPTemplate(name) {
		writeJSONError(w, http.StatusNotFound, "Gabarit ANPP inconnu")
		return
	}
	data, err := reporting.GetANPPTemplateForOrg(user.OrganizationID.String(), name)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	ct := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	if strings.HasSuffix(strings.ToLower(name), ".xlsx") {
		ct = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	_, _ = w.Write(data)
}

// PreviewTemplate retourne un aperçu structuré JSON (paragraphes, tableaux Word ou feuilles Excel) pour affichage en modale.
func (h *ANPPTemplateHandler) PreviewTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*auth.SessionUser)
	if !ok || user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}
	name := path.Base(chi.URLParam(r, "name"))
	if !isValidKnownANPPTemplate(name) {
		writeJSONError(w, http.StatusNotFound, "Gabarit ANPP inconnu")
		return
	}
	data, err := reporting.GetANPPTemplateForOrg(user.OrganizationID.String(), name)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	preview, err := reporting.ExtractANPPTemplatePreview(name, data)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur lecture aperçu gabarit: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(preview)
}

// UploadTemplate permet au PDT, RA ou Admin Système de remplacer un gabarit officiel ANPP sans recompiler.
func (h *ANPPTemplateHandler) UploadTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*auth.SessionUser)
	if !ok || user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}
	if !canManageOrgTemplates(user.Role) {
		writeJSONError(w, http.StatusForbidden, "Seuls le PDT, le Responsable Affaires Réglementaires ou l'Admin Système peuvent mettre à jour les gabarits ANPP")
		return
	}

	name := path.Base(chi.URLParam(r, "name"))
	if !isValidKnownANPPTemplate(name) {
		writeJSONError(w, http.StatusBadRequest, "Nom de gabarit ANPP non reconnu")
		return
	}

	if err := r.ParseMultipartForm(16 << 20); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Requête multipart invalide (max 16 Mo)")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Fichier 'file' manquant")
		return
	}
	defer file.Close()

	expectedExt := strings.ToLower(path.Ext(name))
	uploadedExt := strings.ToLower(path.Ext(header.Filename))
	if uploadedExt != expectedExt {
		writeJSONError(w, http.StatusBadRequest, "Extension invalide : attendu "+expectedExt)
		return
	}

	content, err := io.ReadAll(io.LimitReader(file, 16<<20))
	if err != nil || len(content) < 100 || content[0] != 'P' || content[1] != 'K' {
		writeJSONError(w, http.StatusBadRequest, "Fichier Office (.docx/.xlsx) invalide ou corrompu")
		return
	}

	notes := strings.TrimSpace(r.FormValue("notes"))
	scope := strings.TrimSpace(r.FormValue("scope"))
	isGlobal := scope == "global"
	if isGlobal && !canManageGlobalTemplates(user.Role) {
		writeJSONError(w, http.StatusForbidden, "Surcharge globale réservée à l'Admin Système ou Super Admin")
		return
	}

	sum := sha256.Sum256(content)
	shaHex := hex.EncodeToString(sum[:])

	var version int
	if isGlobal {
		err = h.pool.QueryRow(r.Context(), `
			INSERT INTO anpp_template_overrides (organization_id, template_name, content_bytes, sha256_hash, uploaded_by, uploaded_by_name, version, notes)
			VALUES (NULL, $1, $2, $3, $4, $5, 1, $6)
			ON CONFLICT (template_name) WHERE organization_id IS NULL
			DO UPDATE SET
				content_bytes = EXCLUDED.content_bytes,
				sha256_hash = EXCLUDED.sha256_hash,
				uploaded_by = EXCLUDED.uploaded_by,
				uploaded_by_name = EXCLUDED.uploaded_by_name,
				version = anpp_template_overrides.version + 1,
				notes = EXCLUDED.notes,
				updated_at = NOW()
			RETURNING version
		`, name, content, shaHex, user.UserID, user.Email, notes).Scan(&version)
	} else {
		err = h.pool.QueryRow(r.Context(), `
			INSERT INTO anpp_template_overrides (organization_id, template_name, content_bytes, sha256_hash, uploaded_by, uploaded_by_name, version, notes)
			VALUES ($1, $2, $3, $4, $5, $6, 1, $7)
			ON CONFLICT (organization_id, template_name) WHERE organization_id IS NOT NULL
			DO UPDATE SET
				content_bytes = EXCLUDED.content_bytes,
				sha256_hash = EXCLUDED.sha256_hash,
				uploaded_by = EXCLUDED.uploaded_by,
				uploaded_by_name = EXCLUDED.uploaded_by_name,
				version = anpp_template_overrides.version + 1,
				notes = EXCLUDED.notes,
				updated_at = NOW()
			RETURNING version
		`, user.OrganizationID, name, content, shaHex, user.UserID, user.Email, notes).Scan(&version)
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur enregistrement gabarit ANPP")
		return
	}

	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "ANPP_TEMPLATE_UPLOADED",
			EntityType:     "anpp_template",
			EntityID:       name,
			Details: map[string]any{
				"template_name": name,
				"scope":         map[bool]string{true: "global", false: "organization"}[isGlobal],
				"version":       version,
				"sha256":        shaHex,
				"notes":         notes,
			},
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"template_name": name,
		"source":        map[bool]string{true: "global", false: "organization"}[isGlobal],
		"version":       version,
		"sha256":        shaHex,
	})
}

// DeleteOverride supprime la surcharge pour restaurer le gabarit embarqué par défaut.
func (h *ANPPTemplateHandler) DeleteOverride(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*auth.SessionUser)
	if !ok || user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}
	if !canManageOrgTemplates(user.Role) {
		writeJSONError(w, http.StatusForbidden, "Permission insuffisante")
		return
	}

	name := path.Base(chi.URLParam(r, "name"))
	if !isValidKnownANPPTemplate(name) {
		writeJSONError(w, http.StatusBadRequest, "Gabarit ANPP inconnu")
		return
	}

	scope := strings.TrimSpace(r.URL.Query().Get("scope"))
	if scope == "global" {
		if !canManageGlobalTemplates(user.Role) {
			writeJSONError(w, http.StatusForbidden, "Suppression globale réservée à l'Admin Système ou Super Admin")
			return
		}
		_, _ = h.pool.Exec(r.Context(), `DELETE FROM anpp_template_overrides WHERE template_name = $1 AND organization_id IS NULL`, name)
	} else {
		_, _ = h.pool.Exec(r.Context(), `DELETE FROM anpp_template_overrides WHERE template_name = $1 AND organization_id = $2`, name, user.OrganizationID)
	}

	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "ANPP_TEMPLATE_REVERTED",
			EntityType:     "anpp_template",
			EntityID:       name,
			Details:        map[string]any{"template_name": name, "scope": scope},
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "reverted", "template_name": name})
}
