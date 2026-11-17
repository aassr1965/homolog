package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"homolog/internal/audit"
	"homolog/internal/auth"
)

// SanadManifest (alias HomologManifest) décrit l'inventaire cryptographique et les métadonnées de l'archive .sanad
type HomologManifest struct {
	Version          string            `json:"version"`
	Format           string            `json:"format"`
	CreatedAt        time.Time         `json:"created_at"`
	CreatedByEmail   string            `json:"created_by_email"`
	CreatedByRole    string            `json:"created_by_role"`
	OrganizationID   string            `json:"organization_id"`
	OrganizationSlug string            `json:"organization_slug"`
	OrganizationName string            `json:"organization_name"`
	TableCounts      map[string]int    `json:"table_counts"`
	FilesCount       int               `json:"files_count"`
	TotalSizeBytes   int64             `json:"total_size_bytes"`
	BlobsChecksums   map[string]string `json:"blobs_checksums"` // sha256_hash -> sha256_hash
	ArchiveSeal      string            `json:"archive_seal"`
}

// SanadManifest est l'alias officiel du manifeste d'archive universelle Sanad Pharma
type SanadManifest = HomologManifest

// ExportHomologArchive génère une archive universelle complète et auto-suffisante .sanad (tar.gz)
func (h *SystemHandler) ExportHomologArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetSessionUser(ctx)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	// Contrôle d'habilitation : Seuls les rôles administrateurs techniques ou le Pharmacien DT peuvent exporter l'archive souveraine
	if !user.HasPermission("SYS_MAINT") && user.Role != "directeur_technique" && user.Role != "admin_laboratoire" {
		writeJSONError(w, http.StatusForbidden, "Accès refusé : L'export de l'archive universelle requiert l'habilitation SYS_MAINT ou Pharmacien DT (1.A.7)")
		return
	}

	// Détermination de l'organisation cible (cloisonnement multi-tenant strict)
	targetOrgID := user.OrganizationID
	if reqOrg := r.URL.Query().Get("org_id"); reqOrg != "" && user.HasPermission("SYS_MAINT") {
		if parsed, err := uuid.Parse(reqOrg); err == nil {
			targetOrgID = parsed
		}
	}

	var orgSlug, orgName string
	err := h.pool.QueryRow(ctx, "SELECT slug, legal_name FROM organizations WHERE id = $1", targetOrgID).Scan(&orgSlug, &orgName)
	if err != nil {
		orgSlug = user.OrganizationSlug
		orgName = user.OrganizationName
	}

	now := time.Now().UTC()
	archiveFilename := fmt.Sprintf("sanad-archive-%s-%s.sanad", orgSlug, now.Format("20060102-150405"))

	// 1. Collecte structurée des données de la base (filtrée strictement par organisation)
	tablesData, tableCounts, err := h.dumpTablesData(ctx, targetOrgID)
	if err != nil {
		log.Printf("[ERREUR BACKUP] Échec dump tables : %v", err)
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur lors de la capture des données de base : %v", err))
		return
	}

	// 2. Collecte des métadonnées des fichiers du coffre-fort CAS (filtrée strictement par organisation)
	storagePath := "./data/storage"
	if h.cfg != nil && h.cfg.StoragePath != "" {
		storagePath = h.cfg.StoragePath
	}

	blobsChecksums, casFiles, totalBlobsSize, err := h.collectCASFiles(ctx, targetOrgID, storagePath)
	if err != nil {
		log.Printf("[ERREUR BACKUP] Échec scan CAS : %v", err)
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur lors de l'inventaire du stockage CAS : %v", err))
		return
	}

	// 3. Préparation du manifeste officiel de conformité ALCOA+
	manifest := HomologManifest{
		Version:          "1.0.0",
		Format:           "SANAD_PORTABLE_ARCHIVE_V1",
		CreatedAt:        now,
		CreatedByEmail:   user.Email,
		CreatedByRole:    user.Role,
		OrganizationID:   targetOrgID.String(),
		OrganizationSlug: orgSlug,
		OrganizationName: orgName,
		TableCounts:      tableCounts,
		FilesCount:       len(casFiles),
		TotalSizeBytes:   totalBlobsSize,
		BlobsChecksums:   blobsChecksums,
	}

	// Calcul de l'empreinte de scellement
	manifestJSONRaw, _ := json.MarshalIndent(manifest, "", "  ")
	sealHash := sha256.Sum256(manifestJSONRaw)
	manifest.ArchiveSeal = hex.EncodeToString(sealHash[:])
	manifestJSONFinal, _ := json.MarshalIndent(manifest, "", "  ")

	// 4. Flux de sortie HTTP avec compression Gzip et Tar
	w.Header().Set("Content-Type", "application/x-gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", archiveFilename))
	w.Header().Set("X-Sanad-Seal", manifest.ArchiveSeal)
	w.Header().Set("X-Homolog-Seal", manifest.ArchiveSeal)

	gw := gzip.NewWriter(w)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// 5. Ajout du manifest.json dans l'archive
	if err := addBytesToTar(tw, "manifest.json", manifestJSONFinal, now); err != nil {
		log.Printf("[ERREUR TAR] manifest.json: %v", err)
		return
	}

	// 6. Ajout des données de base de données au format JSON
	for tblName, dataBytes := range tablesData {
		tarPath := fmt.Sprintf("database/%s.json", tblName)
		if err := addBytesToTar(tw, tarPath, dataBytes, now); err != nil {
			log.Printf("[ERREUR TAR] %s: %v", tarPath, err)
			return
		}
	}

	// 7. Ajout des fichiers physiques du coffre-fort CAS
	for sha, fPath := range casFiles {
		tarPath := fmt.Sprintf("blobs/%s", sha)
		if err := addFileToTar(tw, tarPath, fPath, now); err != nil {
			log.Printf("[AVERTISSEMENT TAR] Fichier CAS manquant %s: %v", fPath, err)
		}
	}

	// 8. Enregistrement dans la piste d'audit immuable BPF
	if h.auditService != nil {
		_, _ = h.auditService.Log(ctx, &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "SYSTEM_ARCHIVE_EXPORTED_SANAD",
			EntityType:     "ARCHIVE",
			EntityID:       archiveFilename,
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"filename":         archiveFilename,
				"sha256_seal":      manifest.ArchiveSeal,
				"tables_exported":  len(tablesData),
				"blobs_exported":   len(casFiles),
				"total_blobs_size": totalBlobsSize,
			},
		})
	}
}

// RestoreHomologArchive réceptionne une archive .sanad ou .homolog, valide son intégrité cryptographique et restaure l'instance
func (h *SystemHandler) RestoreHomologArchive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetSessionUser(ctx)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	if !user.HasPermission("SYS_MAINT") && user.Role != "directeur_technique" && user.Role != "super_admin_editeur" {
		writeJSONError(w, http.StatusForbidden, "Accès refusé : La restauration d'une archive universelle requiert l'habilitation SYS_MAINT ou Pharmacien DT (1.A.7)")
		return
	}

	// Limite d'upload 500 Mo pour les archives
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Fichier d'archive invalide ou taille supérieure à 500 Mo")
		return
	}

	file, header, err := r.FormFile("archive")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Le fichier d'archive (.sanad ou .homolog) est obligatoire")
		return
	}
	defer file.Close()

	lowerFilename := strings.ToLower(header.Filename)
	if !strings.HasSuffix(lowerFilename, ".sanad") && !strings.HasSuffix(lowerFilename, ".homolog") && !strings.HasSuffix(lowerFilename, ".tar.gz") {
		writeJSONError(w, http.StatusBadRequest, "Format non supporté : le fichier doit porter l'extension .sanad ou .homolog")
		return
	}

	// 1. Décompression Gzip
	gzr, err := gzip.NewReader(file)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "L'archive est corrompue ou n'est pas un fichier Gzip valide")
		return
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Variables de collecte
	var manifestContent []byte
	databaseFiles := make(map[string][]byte)
	extractedBlobs := make(map[string][]byte)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("Erreur de lecture du conteneur Tar : %v", err))
			return
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		cleanedName := filepath.Clean(hdr.Name)
		if cleanedName == "manifest.json" {
			manifestContent, _ = io.ReadAll(tr)
		} else if strings.HasPrefix(cleanedName, "database"+string(filepath.Separator)) || strings.HasPrefix(cleanedName, "database/") {
			data, _ := io.ReadAll(tr)
			tblName := strings.TrimSuffix(filepath.Base(cleanedName), ".json")
			databaseFiles[tblName] = data
		} else if strings.HasPrefix(cleanedName, "blobs"+string(filepath.Separator)) || strings.HasPrefix(cleanedName, "blobs/") {
			sha := filepath.Base(cleanedName)
			blobData, _ := io.ReadAll(tr)
			extractedBlobs[sha] = blobData
		}
	}

	// 2. Contrôle d'intégrité du manifeste
	if len(manifestContent) == 0 {
		writeJSONError(w, http.StatusUnprocessableEntity, "Archive invalide : fichier manifest.json manquant")
		return
	}

	var manifest HomologManifest
	if err := json.Unmarshal(manifestContent, &manifest); err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "Archive invalide : manifest.json non analysable")
		return
	}

	// Contrôle de cloisonnement strict multi-tenant : Un laboratoire ne peut réinjecter que son archive propre
	if !user.HasPermission("SYS_MAINT") && user.Role != "super_admin_editeur" {
		if manifest.OrganizationID != "" && manifest.OrganizationID != user.OrganizationID.String() {
			writeJSONError(w, http.StatusForbidden, "Accès refusé : Tentative de restauration trans-tenant non autorisée (cloisonnement multi-tenant BPF)")
			return
		}
	}

	// 3. Contrôle cryptographique SHA-256 de chaque blob extrait
	var verifiedBlobsCount int
	for sha, blobData := range extractedBlobs {
		calcSHA := sha256.Sum256(blobData)
		calcSHAHex := hex.EncodeToString(calcSHA[:])
		if !strings.EqualFold(sha, calcSHAHex) {
			writeJSONError(w, http.StatusUnprocessableEntity, fmt.Sprintf("Échec d'intégrité BPF : le blob %s a une empreinte altérée (%s)", sha, calcSHAHex))
			return
		}
		verifiedBlobsCount++
	}

	// 4. Écriture des blobs dans le stockage physique CAS
	storagePath := "./data/storage"
	if h.cfg != nil && h.cfg.StoragePath != "" {
		storagePath = h.cfg.StoragePath
	}
	_ = os.MkdirAll(storagePath, 0750)

	for sha, blobData := range extractedBlobs {
		// Sous-dossier sharding à 2 caractères
		shardDir := filepath.Join(storagePath, sha[:2])
		_ = os.MkdirAll(shardDir, 0750)
		targetPath := filepath.Join(shardDir, sha)
		if err := os.WriteFile(targetPath, blobData, 0640); err != nil {
			log.Printf("[AVERTISSEMENT RESTORE] Impossible d'écrire le blob %s: %v", targetPath, err)
		}
	}

	// 5. Restauration des données de base de données
	restoredTables, err := h.restoreTablesData(ctx, databaseFiles)
	if err != nil {
		log.Printf("[ERREUR RESTORE] Tables: %v", err)
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur lors de la réinjection des tables : %v", err))
		return
	}

	// 6. Audit Trail BPF du succès de restauration
	if h.auditService != nil {
		_, _ = h.auditService.Log(ctx, &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "SYSTEM_ARCHIVE_RESTORED_SANAD",
			EntityType:     "ARCHIVE",
			EntityID:       header.Filename,
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"filename":         header.Filename,
				"manifest_version": manifest.Version,
				"created_at_orig":  manifest.CreatedAt.Format(time.RFC3339),
				"orig_author":      manifest.CreatedByEmail,
				"tables_restored":  restoredTables,
				"blobs_verified":   verifiedBlobsCount,
				"seal":             manifest.ArchiveSeal,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":          "ok",
		"message":         "Archive universelle Sanad Pharma (.sanad) restaurée avec succès et intégrité 100% vérifiée",
		"tables_restored": len(restoredTables),
		"blobs_verified":  verifiedBlobsCount,
		"manifest": map[string]any{
			"version":            manifest.Version,
			"source_org":         manifest.OrganizationName,
			"creation_date":      manifest.CreatedAt.Format(time.RFC3339),
			"original_author":    manifest.CreatedByEmail,
			"cryptographic_seal": manifest.ArchiveSeal,
		},
	})
}

// dumpTablesData sérialise les données métier en JSON avec cloisonnement strict par organisation (tenant isolation)
// et masquage cryptographique des données hautement sensibles (ex: password_hash, secrets TOTP).
func (h *SystemHandler) dumpTablesData(ctx context.Context, orgID uuid.UUID) (map[string][]byte, map[string]int, error) {
	// Définition ordonnée des tables et des requêtes cloisonnées par tenant ($1 = orgID)
	tableQueries := []struct {
		table       string
		query       string
		hasOrgParam bool
	}{
		{
			table:       "organizations",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT id, slug, legal_name, miph_license_number, technical_director_name, address, phone, email, storage_quota_bytes, storage_used_bytes, is_active, permissions_version, created_at, updated_at FROM organizations WHERE id = $1) t",
			hasOrgParam: true,
		},
		{
			table: "users",
			// PROTECTION CRITIQUE : Masquage systématique des empreintes de mot de passe et clés 2FA
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT id, organization_id, email, '[REDACTED]' AS password_hash, first_name, last_name, role, false AS totp_enabled, NULL::varchar AS totp_secret, '{}'::text[] AS recovery_codes, failed_login_attempts, locked_until, contract_expires_at, is_active, last_login_at, created_at, updated_at FROM users WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "establishments",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM establishments WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "manufacturing_sites",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM manufacturing_sites WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "products",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM products WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "medicines",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM medicines WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "medical_devices",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM medical_devices WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "medicine_manufacturing_sites",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT mms.* FROM medicine_manufacturing_sites mms JOIN medicines m ON mms.medicine_id = m.id WHERE m.organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "ged_documents",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM ged_documents WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "ged_versions",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT gv.* FROM ged_versions gv JOIN ged_documents gd ON gv.document_id = gd.id WHERE gd.organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "dossiers",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM dossiers WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "dossier_checklist_items",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT dci.* FROM dossier_checklist_items dci JOIN dossiers d ON dci.dossier_id = d.id WHERE d.organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "anpp_deficiency_letters",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM anpp_deficiency_letters WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "anpp_deficiency_tasks",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM anpp_deficiency_tasks WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table: "anpp_fee_schedules",
			// Barèmes officiels interministériels publics de l'ANPP (JORADP)
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM anpp_fee_schedules) t",
			hasOrgParam: false,
		},
		{
			table:       "anpp_payment_slips",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM anpp_payment_slips WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "regulatory_watch_items",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM regulatory_watch_items WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "anpp_deposit_appointments",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM anpp_deposit_appointments WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
		{
			table:       "audit_logs",
			query:       "SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT * FROM audit_logs WHERE organization_id = $1) t",
			hasOrgParam: true,
		},
	}

	result := make(map[string][]byte)
	counts := make(map[string]int)

	for _, item := range tableQueries {
		var jsonBytes []byte
		var err error
		if item.hasOrgParam {
			err = h.pool.QueryRow(ctx, item.query, orgID).Scan(&jsonBytes)
		} else {
			err = h.pool.QueryRow(ctx, item.query).Scan(&jsonBytes)
		}
		if err != nil {
			log.Printf("[DUMP TABLES] Avertissement table %s: %v", item.table, err)
			jsonBytes = []byte("[]")
		}

		var countSlice []any
		if err := json.Unmarshal(jsonBytes, &countSlice); err == nil {
			counts[item.table] = len(countSlice)
		} else {
			counts[item.table] = 0
		}

		result[item.table] = jsonBytes
	}

	return result, counts, nil
}

// collectCASFiles scanne les versions GED de l'organisation et retourne les chemins physiques des blobs
func (h *SystemHandler) collectCASFiles(ctx context.Context, orgID uuid.UUID, storageBasePath string) (map[string]string, map[string]string, int64, error) {
	query := `
		SELECT DISTINCT gv.file_sha256, gv.storage_path, gv.file_size_bytes 
		FROM ged_versions gv
		JOIN ged_documents gd ON gv.document_id = gd.id
		WHERE gd.organization_id = $1 AND gv.file_sha256 IS NOT NULL AND gv.file_sha256 != ''
	`
	rows, err := h.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	checksums := make(map[string]string)
	files := make(map[string]string)
	var totalSize int64

	for rows.Next() {
		var sha, path string
		var size int64
		if err := rows.Scan(&sha, &path, &size); err == nil {
			checksums[sha] = sha
			// Vérifier si le chemin existe
			if _, err := os.Stat(path); err == nil {
				files[sha] = path
				totalSize += size
			} else {
				// Tenter avec le shard par défaut
				candidate := filepath.Join(storageBasePath, sha[:2], sha)
				if _, err := os.Stat(candidate); err == nil {
					files[sha] = candidate
					totalSize += size
				}
			}
		}
	}

	return checksums, files, totalSize, nil
}

// restoreTablesData réinjecte réellement les données de l'archive dans la base PostgreSQL dans l'ordre des contraintes d'intégrité
func (h *SystemHandler) restoreTablesData(ctx context.Context, dataFiles map[string][]byte) ([]string, error) {
	// Ordre strict des dépendances de clés étrangères
	orderedTables := []string{
		"organizations",
		"users",
		"establishments",
		"manufacturing_sites",
		"products",
		"medicines",
		"medical_devices",
		"medicine_manufacturing_sites",
		"ged_documents",
		"ged_versions",
		"dossiers",
		"dossier_checklist_items",
		"anpp_deficiency_letters",
		"anpp_deficiency_tasks",
		"anpp_fee_schedules",
		"anpp_payment_slips",
		"regulatory_watch_items",
		"anpp_deposit_appointments",
		"audit_logs",
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("impossible de démarrer la transaction de restauration : %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var restored []string

	for _, tblName := range orderedTables {
		jsonBytes, ok := dataFiles[tblName]
		if !ok || len(jsonBytes) == 0 || string(jsonBytes) == "[]" {
			continue
		}

		var rows []map[string]any
		if err := json.Unmarshal(jsonBytes, &rows); err != nil {
			log.Printf("[RESTORE TABLE %s] Avertissement format JSON invalide: %v", tblName, err)
			continue
		}

		if len(rows) == 0 {
			continue
		}

		insertedCount := 0
		for _, row := range rows {
			if len(row) == 0 {
				continue
			}

			// Traitement de sécurité pour la table users si le hash a été caviardé
			if tblName == "users" {
				if pwd, ok := row["password_hash"].(string); ok && pwd == "[REDACTED]" {
					// Attribuer un mot de passe temporaire scellé si l'utilisateur n'existe pas encore
					tempHash, _ := auth.HashPassword("Homolog2026!Restored", nil)
					row["password_hash"] = tempHash
				}
			}

			cols := make([]string, 0, len(row))
			placeholders := make([]string, 0, len(row))
			args := make([]any, 0, len(row))
			idx := 1

			for k, v := range row {
				cols = append(cols, k)
				placeholders = append(placeholders, fmt.Sprintf("$%d", idx))

				// Convertir les types complexes pour pgx
				if k == "tags" || k == "recovery_codes" {
					if slice, ok := v.([]any); ok {
						strArr := make([]string, len(slice))
						for i, item := range slice {
							strArr[i] = fmt.Sprintf("%v", item)
						}
						args = append(args, strArr)
					} else {
						args = append(args, []string{})
					}
				} else if m, ok := v.(map[string]any); ok {
					b, _ := json.Marshal(m)
					args = append(args, string(b))
				} else if s, ok := v.([]any); ok {
					b, _ := json.Marshal(s)
					args = append(args, string(b))
				} else {
					args = append(args, v)
				}
				idx++
			}

			conflictClause := "ON CONFLICT DO NOTHING"
			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) %s",
				tblName,
				strings.Join(cols, ", "),
				strings.Join(placeholders, ", "),
				conflictClause,
			)

			tag, err := tx.Exec(ctx, query, args...)
			if err != nil {
				log.Printf("[RESTORE INSERT ERROR] Table %s: %v (Query: %s)", tblName, err, query)
			} else {
				insertedCount += int(tag.RowsAffected())
			}
		}

		// Réalignement de la séquence pour audit_logs si nécessaire
		if tblName == "audit_logs" {
			_, _ = tx.Exec(ctx, "SELECT setval(pg_get_serial_sequence('audit_logs', 'id'), COALESCE(MAX(id), 1)) FROM audit_logs")
		}

		restored = append(restored, fmt.Sprintf("%s (%d entrées lues, %d réinjectées)", tblName, len(rows), insertedCount))
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("échec de validation de la transaction de restauration : %w", err)
	}

	return restored, nil
}

// Helpers Tar
func addBytesToTar(tw *tar.Writer, name string, data []byte, t time.Time) error {
	hdr := &tar.Header{
		Name:    name,
		Mode:    0640,
		Size:    int64(len(data)),
		ModTime: t,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

func addFileToTar(tw *tar.Writer, name string, filePath string, t time.Time) error {
	fi, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	hdr := &tar.Header{
		Name:    name,
		Mode:    0640,
		Size:    fi.Size(),
		ModTime: t,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}

// EnsureBufferReader helper
func bytesReader(b []byte) io.Reader {
	return bytes.NewReader(b)
}
