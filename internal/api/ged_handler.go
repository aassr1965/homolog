package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
	"homolog/internal/repository"
	"homolog/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GEDHandler gère les points d'entrée HTTP REST du Coffre-Fort GED Pharmaceutique
type GEDHandler struct {
	repo            *repository.GEDRepository
	storageProvider storage.StorageProvider
	auditService    *audit.Service
	pool            *pgxpool.Pool
}

// NewGEDHandler initialise le handler GED
func NewGEDHandler(repo *repository.GEDRepository, storageProvider storage.StorageProvider, auditService *audit.Service, pool *pgxpool.Pool) *GEDHandler {
	return &GEDHandler{
		repo:            repo,
		storageProvider: storageProvider,
		auditService:    auditService,
		pool:            pool,
	}
}

// GetTaxonomy renvoie l'arbre complet des 5 branches et 21 sous-dossiers avec compteurs réels
func (h *GEDHandler) GetTaxonomy(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	// Cloisonnement strict : Les consultants externes n'ont pas accès à la GED transverse d'entreprise (CDC 7.2.2)
	if user.Role == string(domain.RoleConsultantExterne) {
		writeJSONError(w, http.StatusForbidden, "Accès réglementaire refusé : Les consultants externes n'ont pas accès à la GED transverse de l'entreprise (CDC Section 7.2.2)")
		return
	}

	branches, err := h.repo.GetTaxonomyWithCounts(r.Context(), user.OrganizationID)
	if err != nil {
		log.Printf("[ERREUR] GetTaxonomy: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de charger la taxonomie GED")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"branches": branches,
		"total":    len(branches),
	})
}

// ListDocuments retourne la liste paginée et filtrée des documents
func (h *GEDHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	// Cloisonnement strict : Les consultants externes n'ont pas accès à la GED transverse d'entreprise (CDC 7.2.2)
	if user.Role == string(domain.RoleConsultantExterne) {
		writeJSONError(w, http.StatusForbidden, "Accès réglementaire refusé : Les consultants externes n'ont pas accès à la GED transverse de l'entreprise (CDC Section 7.2.2)")
		return
	}

	taxonomyPath := strings.TrimSpace(r.URL.Query().Get("taxonomy_path"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	docs, total, err := h.repo.ListDocuments(r.Context(), user.OrganizationID, taxonomyPath, status, search, limit, offset)
	if err != nil {
		log.Printf("[ERREUR] ListDocuments: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Erreur lors de la récupération des documents")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total":     total,
		"documents": docs,
		"items":     docs,
	})
}

// GetDocument retourne les détails d'un document spécifique
func (h *GEDHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de document invalide")
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), user.OrganizationID, docID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Document non trouvé")
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

// GetDocumentVersions retourne l'historique complet des versions
func (h *GEDHandler) GetDocumentVersions(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de document invalide")
		return
	}

	versions, err := h.repo.GetDocumentVersions(r.Context(), docID)
	if err != nil {
		log.Printf("[ERREUR] GetDocumentVersions: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de charger les versions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total":    len(versions),
		"versions": versions,
	})
}

// CreateDocument crée un nouveau document avec son fichier initial téléversé dans le CAS
func (h *GEDHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	// Limite de taille max (ex: 50 Mo pour un document BPF standard)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Formulaire multipart invalide ou fichier trop volumineux")
		return
	}

	titleFR := strings.TrimSpace(r.FormValue("title_fr"))
	titleAR := strings.TrimSpace(r.FormValue("title_ar"))
	taxonomyPath := strings.TrimSpace(r.FormValue("taxonomy_path"))
	docCode := strings.TrimSpace(r.FormValue("document_code"))
	docType := strings.TrimSpace(r.FormValue("document_type"))
	department := strings.TrimSpace(r.FormValue("department"))
	revisionSummary := strings.TrimSpace(r.FormValue("revision_summary"))
	expiryDateStr := strings.TrimSpace(r.FormValue("expiry_date"))

	if titleFR == "" || taxonomyPath == "" || docCode == "" {
		writeJSONError(w, http.StatusBadRequest, "Le titre (FR), le code et le dossier de taxonomie sont obligatoires")
		return
	}

	if !domain.ValidateTaxonomyPath(taxonomyPath) {
		writeJSONError(w, http.StatusBadRequest, "Chemin de taxonomie non conforme au plan de classement réglementaire BPF")
		return
	}

	if revisionSummary == "" {
		revisionSummary = "Création initiale du document dans le coffre-fort GED"
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Le fichier binaire associé est obligatoire")
		return
	}
	defer file.Close()

	// 1. Stockage CAS physique avec hachage SHA-256 à la volée
	fileSHA256, sizeBytes, storagePath, err := h.storageProvider.Save(r.Context(), file)
	if err != nil {
		log.Printf("[ERREUR CAS] Échec de sauvegarde fichier : %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Échec de stockage sécurisé du document")
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	var expiryDate *time.Time
	if expiryDateStr != "" {
		if t, err := time.Parse("2006-01-02", expiryDateStr); err == nil {
			expiryDate = &t
		}
	}

	// Relations optionnelles
	var estID, siteID, medID, dmID *uuid.UUID
	if v := strings.TrimSpace(r.FormValue("establishment_id")); v != "" {
		if parsed, err := uuid.Parse(v); err == nil {
			estID = &parsed
		}
	}
	if v := strings.TrimSpace(r.FormValue("site_id")); v != "" {
		if parsed, err := uuid.Parse(v); err == nil {
			siteID = &parsed
		}
	}
	if v := strings.TrimSpace(r.FormValue("medicine_id")); v != "" {
		if parsed, err := uuid.Parse(v); err == nil {
			medID = &parsed
		}
	}
	if v := strings.TrimSpace(r.FormValue("medical_device_id")); v != "" {
		if parsed, err := uuid.Parse(v); err == nil {
			dmID = &parsed
		}
	}

	tagsRaw := strings.TrimSpace(r.FormValue("tags"))
	var tags []string
	if tagsRaw != "" {
		for _, part := range strings.Split(tagsRaw, ",") {
			p := strings.TrimSpace(part)
			if p != "" {
				tags = append(tags, p)
			}
		}
	}

	doc := &domain.GEDDocumentDetailed{
		ID:              uuid.New(),
		OrganizationID:  user.OrganizationID,
		TaxonomyPath:    taxonomyPath,
		TitleFR:         titleFR,
		TitleAR:         titleAR,
		DocumentCode:    docCode,
		DocumentType:    docType,
		Department:      department,
		Tags:            tags,
		Status:          domain.GEDStatusBrouillon,
		CurrentVersion:  "v0.1",
		ExpiryDate:      expiryDate,
		EstablishmentID: estID,
		SiteID:          siteID,
		MedicineID:      medID,
		MedicalDeviceID: dmID,
		CreatedBy:       &user.UserID,
	}

	initialVersion := &domain.GEDVersionDetail{
		ID:               uuid.New(),
		VersionLabel:     "v0.1",
		IsMajorVersion:   false,
		RevisionSummary:  revisionSummary,
		FileSHA256:       fileSHA256,
		FileSizeBytes:    sizeBytes,
		MimeType:         mimeType,
		OriginalFilename: filepath.Base(header.Filename),
		StoragePath:      storagePath,
		AuthorID:         &user.UserID,
	}

	createdDoc, err := h.repo.CreateDocument(r.Context(), doc, initialVersion)
	if err != nil {
		log.Printf("[ERREUR] CreateDocument repo: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Échec d'enregistrement en base de données")
		return
	}

	// Audit Trail ALCOA+
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDocUpload,
		EntityType:     "ged_document",
		EntityID:       createdDoc.DocumentCode,
		Details: map[string]any{
			"document_id":   createdDoc.ID.String(),
			"title_fr":      createdDoc.TitleFR,
			"taxonomy_path": createdDoc.TaxonomyPath,
			"version_label": initialVersion.VersionLabel,
			"file_sha256":   initialVersion.FileSHA256,
			"file_size":     initialVersion.FileSizeBytes,
			"original_file": initialVersion.OriginalFilename,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, createdDoc)
}

// UploadNewVersion téléverse une nouvelle version (Check-in) pour un document
func (h *GEDHandler) UploadNewVersion(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Formulaire multipart invalide")
		return
	}

	revisionSummary := strings.TrimSpace(r.FormValue("revision_summary"))
	if revisionSummary == "" {
		writeJSONError(w, http.StatusBadRequest, "Le motif de révision est obligatoire selon la norme BPF")
		return
	}

	isMajor := r.FormValue("is_major") == "true" || r.FormValue("is_major_version") == "true"

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Le fichier binaire associé est obligatoire")
		return
	}
	defer file.Close()

	fileSHA256, sizeBytes, storagePath, err := h.storageProvider.Save(r.Context(), file)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Échec de stockage physique dans le CAS")
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	v := &domain.GEDVersionDetail{
		ID:               uuid.New(),
		IsMajorVersion:   isMajor,
		RevisionSummary:  revisionSummary,
		FileSHA256:       fileSHA256,
		FileSizeBytes:    sizeBytes,
		MimeType:         mimeType,
		OriginalFilename: filepath.Base(header.Filename),
		StoragePath:      storagePath,
		AuthorID:         &user.UserID,
	}

	createdVersion, err := h.repo.CreateNewVersion(r.Context(), user.OrganizationID, docID, v)
	if err != nil {
		log.Printf("[ERREUR] UploadNewVersion: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDocUpload,
		EntityType:     "ged_version",
		EntityID:       createdVersion.VersionLabel,
		Details: map[string]any{
			"document_id":      docID.String(),
			"version_label":    createdVersion.VersionLabel,
			"is_major_version": createdVersion.IsMajorVersion,
			"revision_summary": createdVersion.RevisionSummary,
			"file_sha256":      createdVersion.FileSHA256,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, createdVersion)
}

// DownloadVersion streame le fichier binaire de façon sécurisée depuis le CAS avec son empreinte SHA-256
func (h *GEDHandler) DownloadVersion(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	// Garde-fou BPF Zero-Trust : Interdiction d'accès aux fichiers pour administrateurs techniques
	if user.Role == string(domain.RoleAdminSystemeLocal) || user.Role == string(domain.RoleSuperAdminEditeur) {
		if h.auditService != nil {
			_, _ = h.auditService.Log(r.Context(), &audit.Entry{
				OrganizationID: user.OrganizationID,
				UserID:         &user.UserID,
				UserEmail:      user.Email,
				UserRole:       user.Role,
				Action:         "AUDIT_SECURITY_ZEROTRUST_BLOCKED",
				EntityType:     "DOCUMENT_CONFIDENTIAL",
				EntityID:       docID.String(),
				IPAddress:      getClientIP(r),
				UserAgent:      r.UserAgent(),
				Details: map[string]any{
					"reason": "Règle Zero-Trust : Les administrateurs techniques n'ont pas accès aux données de santé et pièces pharmaceutiques",
					"role":   user.Role,
				},
			})
		}
		writeJSONError(w, http.StatusForbidden, "Accès Zero-Trust refusé : Les administrateurs techniques n'ont pas accès au contenu des pièces pharmaceutiques")
		return
	}

	// Protection Secret Commercial, Secret Industriel et Cloisonnement "Besoin d'en connaître" (Need-to-Know)
	if h.pool != nil {
		var secClass string
		var itemCode string
		checkSecQuery := `
			SELECT COALESCE(dci.security_classification, 'STANDARD'), dci.item_code
			FROM dossier_checklist_items dci
			WHERE dci.ged_document_id = $1
			LIMIT 1
		`
		_ = h.pool.QueryRow(r.Context(), checkSecQuery, docID).Scan(&secClass, &itemCode)

		// 1. Secret commercial : Seul le PDT ou RA (ou habilitation can_view_financials) peut télécharger
		if secClass == "SECRET_COMMERCIAL" || itemCode == "1.A.4" || itemCode == "1.H.4" || itemCode == "1.H.5" {
			isDirectorOrRA := user.Role == string(domain.RoleDirecteurTechnique) || user.Role == string(domain.RoleResponsableAffairesReg)
			if !isDirectorOrRA {
				var canFin bool
				_ = h.pool.QueryRow(r.Context(), `
					SELECT EXISTS(
						SELECT 1 FROM dossier_checklist_items dci
						JOIN dossier_assignments da ON da.dossier_id = dci.dossier_id
						WHERE dci.ged_document_id = $1 AND da.user_id = $2 AND da.can_view_financials = TRUE
					)
				`, docID, user.UserID).Scan(&canFin)
				if !canFin {
					writeJSONError(w, http.StatusForbidden, "Accès refusé : Document classifié SECRET COMMERCIAL réservé à la direction du laboratoire (Section 7.2.2)")
					return
				}
			}
		}

		// 2. Secret Industriel : Prestataire externe interdit sauf habilitation can_view_industrial_secret
		if secClass == "SECRET_INDUSTRIEL" && user.Role == string(domain.RoleConsultantExterne) {
			var canInd bool
			_ = h.pool.QueryRow(r.Context(), `
				SELECT EXISTS(
					SELECT 1 FROM dossier_checklist_items dci
					JOIN dossier_assignments da ON da.dossier_id = dci.dossier_id
					WHERE dci.ged_document_id = $1 AND da.user_id = $2 AND da.can_view_industrial_secret = TRUE
				)
			`, docID, user.UserID).Scan(&canInd)
			if !canInd {
				writeJSONError(w, http.StatusForbidden, "Accès refusé : Document classifié SECRET INDUSTRIEL réservé aux équipes internes du laboratoire (Section 7.2.2)")
				return
			}
		}

		// 3. Cloisonnement par affectation : Contributeurs opérationnels & consultants
		isOperational := user.Role == string(domain.RoleConsultantExterne) ||
			user.Role == string(domain.RoleContributeurQualiteBPF) ||
			user.Role == string(domain.RoleContributeurRDProd)

		if isOperational {
			var isAssigned bool
			checkAssignedQuery := `
				SELECT EXISTS(
					SELECT 1 FROM dossier_checklist_items dci
					WHERE dci.ged_document_id = $1
					  AND (
						dci.dossier_id IN (SELECT dossier_id FROM dossier_assignments WHERE user_id = $2 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP))
						OR
						dci.dossier_id IN (SELECT dossier_id FROM dossier_consultants WHERE user_id = $2 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP))
					  )
				)
			`
			_ = h.pool.QueryRow(r.Context(), checkAssignedQuery, docID, user.UserID).Scan(&isAssigned)
			if !isAssigned {
				writeJSONError(w, http.StatusForbidden, "Accès réglementaire refusé : Cette pièce n'appartient à aucun dossier auquel vous êtes affecté (Principe du Secret & Besoin d'en connaître)")
				return
			}
		}
	}

	versionIDStr := r.URL.Query().Get("version_id")
	var v *domain.GEDVersionDetail

	if versionIDStr != "" {
		vID, err := uuid.Parse(versionIDStr)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Identifiant de version invalide")
			return
		}
		v, err = h.repo.GetVersionByID(r.Context(), user.OrganizationID, vID)
		if err != nil {
			writeJSONError(w, http.StatusNotFound, "Version non trouvée")
			return
		}
		if v.DocumentID != docID {
			writeJSONError(w, http.StatusNotFound, "Cette version n'appartient pas au document spécifié")
			return
		}
	} else {
		doc, err := h.repo.GetDocumentByID(r.Context(), user.OrganizationID, docID)
		if err != nil || doc.LatestVersion == nil {
			writeJSONError(w, http.StatusNotFound, "Aucune version disponible pour ce document")
			return
		}
		v = doc.LatestVersion
	}

	rc, err := h.storageProvider.Open(r.Context(), v.StoragePath)
	if err != nil {
		log.Printf("[ERREUR] Téléchargement CAS: %v", err)
		writeJSONError(w, http.StatusNotFound, "Fichier physique non accessible dans le stockage CAS")
		return
	}
	defer rc.Close()

	// Protection contre Stored XSS :
	// Seuls les documents PDF et les images matricielles (PNG, JPEG) peuvent être servis inline.
	// Tous les autres formats (HTML, SVG, binaires) sont forcés en attachment.
	isSafeInline := v.MimeType == "application/pdf" || v.MimeType == "image/png" || v.MimeType == "image/jpeg"
	disposition := "attachment"
	if isSafeInline {
		disposition = "inline"
	}
	safeFilename := strings.ReplaceAll(strings.ReplaceAll(v.OriginalFilename, `"`, ""), "\n", "")
	safeFilename = strings.ReplaceAll(safeFilename, "\r", "")

	w.Header().Set("Content-Type", v.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`%s; filename=%q`, disposition, safeFilename))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Length", strconv.FormatInt(v.FileSizeBytes, 10))
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, v.FileSHA256))
	w.Header().Set("X-Content-SHA256", v.FileSHA256)

	_, _ = io.Copy(w, rc)

	// Traçabilité ALCOA+ de consultation/téléchargement
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDocView,
		EntityType:     "ged_document",
		EntityID:       docID.String(),
		Details: map[string]any{
			"version_label": v.VersionLabel,
			"file_sha256":   v.FileSHA256,
			"filename":      v.OriginalFilename,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})
}

// CheckOut verrouille un document pour modification exclusive
func (h *GEDHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Verrouillage pour révision BPF"
	}

	if err := h.repo.CheckOutDocument(r.Context(), user.OrganizationID, docID, user.UserID, req.Reason); err != nil {
		writeJSONError(w, http.StatusConflict, err.Error())
		return
	}

	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         "GED_DOC_CHECKOUT",
		EntityType:     "ged_document",
		EntityID:       docID.String(),
		Details: map[string]any{
			"reason": req.Reason,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Document verrouillé avec succès (Check-out actif)",
	})
}

// CancelCheckOut annule le verrouillage d'un document
func (h *GEDHandler) CancelCheckOut(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	isAdmin := user.Role == string(domain.RoleAdminLaboratoire) || user.Role == string(domain.RoleAdminSystemeLocal)
	if err := h.repo.CancelCheckOut(r.Context(), user.OrganizationID, docID, user.UserID, isAdmin); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         "GED_DOC_CANCEL_CHECKOUT",
		EntityType:     "ged_document",
		EntityID:       docID.String(),
		IPAddress:      getClientIP(r),
		UserAgent:      r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Verrouillage annulé avec succès",
	})
}

// UpdateStatus modifie le statut du cycle de vie documentaire après validation des règles
func (h *GEDHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	var req struct {
		TargetStatus domain.GEDDocumentStatus `json:"target_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Format JSON invalide")
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), user.OrganizationID, docID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Document non trouvé")
		return
	}

	allowed, reason := doc.CanTransitionTo(req.TargetStatus, domain.UserRole(user.Role))
	if !allowed {
		writeJSONError(w, http.StatusForbidden, fmt.Sprintf("Transition de statut non autorisée : %s", reason))
		return
	}

	if err := h.repo.UpdateDocumentStatus(r.Context(), user.OrganizationID, docID, req.TargetStatus); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur de mise à jour du statut")
		return
	}

	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDocStatusChange,
		EntityType:     "ged_document",
		EntityID:       docID.String(),
		Details: map[string]any{
			"from_status": doc.Status,
			"to_status":   req.TargetStatus,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "ok",
		"target_status": req.TargetStatus,
	})
}

// SignDocumentDT réalise la signature électronique légale et le scellement SHA-256 par le Directeur Technique
func (h *GEDHandler) SignDocumentDT(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	// Seul le Pharmacien Directeur Technique titulaire de la décision 1.A.7 a ce pouvoir légal
	if user.Role != string(domain.RoleDirecteurTechnique) {
		writeJSONError(w, http.StatusForbidden, "Seul le Pharmacien Directeur Technique titulaire de la décision ministérielle (1.A.7) est habilité à apposer ce visa légal")
		return
	}

	idStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de document invalide")
		return
	}

	var req struct {
		VersionID   string `json:"version_id"`
		Password    string `json:"password"`
		Declaration string `json:"declaration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide")
		return
	}

	if req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "La re-saisie du mot de passe est obligatoire pour signer (21 CFR Part 11)")
		return
	}

	// Re-authentification : vérification du mot de passe
	var passwordHash string
	err = h.pool.QueryRow(r.Context(), `SELECT password_hash FROM users WHERE id = $1`, user.UserID).Scan(&passwordHash)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur lors de la vérification des accréditations")
		return
	}

	valid, err := auth.VerifyPassword(req.Password, passwordHash)
	if err != nil || !valid {
		writeJSONError(w, http.StatusUnauthorized, "Mot de passe incorrect : signature légale rejetée")
		return
	}

	var versionID uuid.UUID
	if req.VersionID != "" {
		versionID, err = uuid.Parse(req.VersionID)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Identifiant de version invalide")
			return
		}
	} else {
		doc, err := h.repo.GetDocumentByID(r.Context(), user.OrganizationID, docID)
		if err != nil || doc.LatestVersion == nil {
			writeJSONError(w, http.StatusNotFound, "Aucune version trouvée à signer")
			return
		}
		versionID = doc.LatestVersion.ID
	}

	v, err := h.repo.GetVersionByID(r.Context(), user.OrganizationID, versionID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Version introuvable")
		return
	}
	if v.DocumentID != docID {
		writeJSONError(w, http.StatusNotFound, "Cette version n'appartient pas au document spécifié")
		return
	}

	// Garde-fou BPF SoD : Auto-approbation formellement interdite (Maker != Approver)
	if v.AuthorID != nil && *v.AuthorID == user.UserID {
		if h.auditService != nil {
			_, _ = h.auditService.Log(r.Context(), &audit.Entry{
				OrganizationID: user.OrganizationID,
				UserID:         &user.UserID,
				UserEmail:      user.Email,
				UserRole:       user.Role,
				Action:         "AUDIT_SECURITY_SOD_BLOCKED",
				EntityType:     "GED_DOCUMENT",
				EntityID:       docID.String(),
				IPAddress:      getClientIP(r),
				UserAgent:      r.UserAgent(),
				Details: map[string]any{
					"reason":     "Tentative d'auto-approbation BPF (SoD) : Le Pharmacien Directeur Technique est également l'auteur/rédacteur de cette version",
					"version_id": versionID.String(),
				},
			})
		}
		writeJSONError(w, http.StatusForbidden, "ALERTE BPF / SoD : Auto-approbation interdite. Le Pharmacien Directeur Technique ne peut pas approuver un document dont il est lui-même le rédacteur (Maker) sans visa intermédiaire indépendant (Checker).")
		return
	}

	declaration := req.Declaration
	if declaration == "" {
		declaration = domain.SolemnPDTDeclarationGED
	}

	now := time.Now().UTC()
	clientIP := getClientIP(r)

	// Scellement cryptographique immuable
	signatureSeal := domain.ComputeSignatureSeal(v.FileSHA256, user.UserID, user.Role, now, clientIP)

	if err := h.repo.SignDocumentVersionByDT(r.Context(), user.OrganizationID, docID, versionID, user.UserID, declaration, signatureSeal, clientIP); err != nil {
		log.Printf("[ERREUR] SignDocumentVersionByDT: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Échec de l'enregistrement de la signature")
		return
	}

	// Audit Trail ALCOA+
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDocSignDT,
		EntityType:     "ged_document",
		EntityID:       docID.String(),
		Details: map[string]any{
			"version_id":     versionID.String(),
			"version_label":  v.VersionLabel,
			"file_sha256":    v.FileSHA256,
			"signature_seal": signatureSeal,
			"declaration":    declaration,
			"legal_standard": "BPF / 21 CFR Part 11 / Guide ANPP",
		},
		IPAddress: clientIP,
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"status":         "ok",
		"message":        "Document scellé et signé électroniquement avec succès par le Pharmacien Directeur Technique",
		"signature_seal": signatureSeal,
		"signed_at":      now.Format(time.RFC3339),
		"signatory":      user.FirstName + " " + user.LastName,
	})
}
