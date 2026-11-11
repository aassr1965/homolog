package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// PharmaHandler gère les requêtes REST sur le référentiel pharmaceutique ANPP
type PharmaHandler struct {
	repo         *repository.PharmaRepository
	auditService *audit.Service
}

// NewPharmaHandler crée une instance de PharmaHandler
func NewPharmaHandler(repo *repository.PharmaRepository, auditService *audit.Service) *PharmaHandler {
	return &PharmaHandler{
		repo:         repo,
		auditService: auditService,
	}
}

// ============================================================================
// ÉTABLISSEMENTS PHARMACEUTIQUES
// ============================================================================

// CreateEstablishment enregistre un nouvel établissement agréé MIPH
func (h *PharmaHandler) CreateEstablishment(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var est domain.Establishment
	if err := json.NewDecoder(r.Body).Decode(&est); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	est.OrganizationID = user.OrganizationID
	if err := h.repo.CreateEstablishment(r.Context(), &est); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Journalisation BPF (21 CFR Part 11)
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         "ESTABLISHMENT_CREATE",
		EntityType:     "establishment",
		EntityID:       est.ID.String(),
		Details: map[string]any{
			"name_fr":                 est.NameFR,
			"miph_license_number":     est.MIPHLicenseNumber,
			"technical_director_name": est.TechnicalDirectorName,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, est)
}

// GetEstablishment retourne le détail d'un établissement
func (h *PharmaHandler) GetEstablishment(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	est, err := h.repo.GetEstablishmentByID(r.Context(), user.OrganizationID, id)
	if err != nil {
		if err == repository.ErrNotFound {
			writeJSONError(w, http.StatusNotFound, "Établissement non trouvé")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur lors de la récupération")
		return
	}

	writeJSON(w, http.StatusOK, est)
}

// ListEstablishments retourne la liste paginée des établissements
func (h *PharmaHandler) ListEstablishments(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	list, total, err := h.repo.ListEstablishments(r.Context(), user.OrganizationID, limit, offset)
	if err != nil {
		log.Printf("[ERREUR] ListEstablishments: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Erreur de base de données")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total":          total,
		"items":          list,
		"establishments": list,
	})
}

// UpdateEstablishment met à jour les informations d'un établissement
func (h *PharmaHandler) UpdateEstablishment(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	var est domain.Establishment
	if err := json.NewDecoder(r.Body).Decode(&est); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	est.ID = id
	est.OrganizationID = user.OrganizationID

	if err := h.repo.UpdateEstablishment(r.Context(), &est); err != nil {
		if err == repository.ErrNotFound {
			writeJSONError(w, http.StatusNotFound, "Établissement non trouvé")
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit trail
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         "ESTABLISHMENT_UPDATE",
		EntityType:     "establishment",
		EntityID:       est.ID.String(),
		Details: map[string]any{
			"name_fr": est.NameFR,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, est)
}

// ============================================================================
// SITES DE FABRICATION ET DE CONTRÔLE (1.H.8 & 1.M)
// ============================================================================

// CreateManufacturingSite ajoute un nouveau site intervenant
func (h *PharmaHandler) CreateManufacturingSite(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var site domain.ManufacturingSite
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	site.OrganizationID = user.OrganizationID
	if err := h.repo.CreateManufacturingSite(r.Context(), &site); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         "MANUFACTURING_SITE_CREATE",
		EntityType:     "manufacturing_site",
		EntityID:       site.ID.String(),
		Details: map[string]any{
			"site_name_fr": site.SiteNameFR,
			"country_code": site.CountryCode,
			"site_role":    string(site.SiteRole),
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, site)
}

// ListManufacturingSites retourne les sites filtrés par rôle
func (h *PharmaHandler) ListManufacturingSites(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	roleFilter := strings.TrimSpace(r.URL.Query().Get("role"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	list, total, err := h.repo.ListManufacturingSites(r.Context(), user.OrganizationID, roleFilter, limit, offset)
	if err != nil {
		log.Printf("[ERREUR] ListManufacturingSites: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Erreur base de données")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total": total,
		"items": list,
		"sites": list,
	})
}

// GetManufacturingSite retourne un site par son identifiant
func (h *PharmaHandler) GetManufacturingSite(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	site, err := h.repo.GetManufacturingSiteByID(r.Context(), user.OrganizationID, id)
	if err != nil {
		if err == repository.ErrNotFound {
			writeJSONError(w, http.StatusNotFound, "Site de fabrication non trouvé")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur lors de la récupération")
		return
	}

	writeJSON(w, http.StatusOK, site)
}

// ============================================================================
// MÉDICAMENTS (Guide CTD Juillet 2025)
// ============================================================================

// MedicineCreateRequest format de la requête de création d'un médicament
type MedicineCreateRequest struct {
	domain.Medicine
	AssociatedSiteIDs []uuid.UUID `json:"associated_site_ids,omitempty"`
}

// CreateMedicine enregistre un nouveau médicament et crée son produit parent
func (h *PharmaHandler) CreateMedicine(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var req MedicineCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.Medicine.OrganizationID = user.OrganizationID
	if err := h.repo.CreateMedicine(r.Context(), &req.Medicine, req.AssociatedSiteIDs); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         "MEDICINE_CREATE",
		EntityType:     "medicine",
		EntityID:       req.Medicine.ID.String(),
		Details: map[string]any{
			"brand_name_fr": req.Medicine.BrandNameFR,
			"dci_fr":        req.Medicine.DCIFR,
			"origin_type":   string(req.Medicine.OriginType),
			"atc_code":      req.Medicine.ATCCode,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, req.Medicine)
}

// GetMedicine retourne un médicament avec ses sites intervenants (1.H.8)
func (h *PharmaHandler) GetMedicine(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	med, err := h.repo.GetMedicineByID(r.Context(), user.OrganizationID, id)
	if err != nil {
		if err == repository.ErrNotFound {
			writeJSONError(w, http.StatusNotFound, "Médicament non trouvé")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur base de données")
		return
	}

	writeJSON(w, http.StatusOK, med)
}

// ListMedicines retourne les médicaments filtrés par type d'origine ou recherche
func (h *PharmaHandler) ListMedicines(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	origin := strings.TrimSpace(r.URL.Query().Get("origin"))
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	list, total, err := h.repo.ListMedicines(r.Context(), user.OrganizationID, origin, search, limit, offset)
	if err != nil {
		log.Printf("[ERREUR] ListMedicines: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Erreur base de données")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total":     total,
		"items":     list,
		"medicines": list,
	})
}

// ============================================================================
// DISPOSITIFS MÉDICAUX (Décret 20-324 & Arrêté 10/05/2021)
// ============================================================================

// CreateMedicalDevice enregistre un nouveau DM
func (h *PharmaHandler) CreateMedicalDevice(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var dm domain.MedicalDevice
	if err := json.NewDecoder(r.Body).Decode(&dm); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	dm.OrganizationID = user.OrganizationID
	if err := h.repo.CreateMedicalDevice(r.Context(), &dm); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         "MEDICAL_DEVICE_CREATE",
		EntityType:     "medical_device",
		EntityID:       dm.ID.String(),
		Details: map[string]any{
			"commercial_name_fr":     dm.CommercialNameFR,
			"manufacturer_reference": dm.ManufacturerReference,
			"dm_class":               dm.DMClass,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, dm)
}

// GetMedicalDevice retourne un DM par son identifiant
func (h *PharmaHandler) GetMedicalDevice(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	dm, err := h.repo.GetMedicalDeviceByID(r.Context(), user.OrganizationID, id)
	if err != nil {
		if err == repository.ErrNotFound {
			writeJSONError(w, http.StatusNotFound, "Dispositif médical non trouvé")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur base de données")
		return
	}

	writeJSON(w, http.StatusOK, dm)
}

// ListMedicalDevices retourne les dispositifs médicaux paginés
func (h *PharmaHandler) ListMedicalDevices(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	classFilter := strings.TrimSpace(r.URL.Query().Get("class"))
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	list, total, err := h.repo.ListMedicalDevices(r.Context(), user.OrganizationID, classFilter, search, limit, offset)
	if err != nil {
		log.Printf("[ERREUR] ListMedicalDevices: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Erreur base de données")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total":   total,
		"items":   list,
		"devices": list,
	})
}

// Helper writeJSON
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
