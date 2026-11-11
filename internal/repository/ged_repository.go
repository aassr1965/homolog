package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"homolog/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GEDRepository gère la persistance et les requêtes PostgreSQL pour le Coffre-Fort GED Pharmaceutique
type GEDRepository struct {
	pool *pgxpool.Pool
}

// NewGEDRepository initialise le repository GED
func NewGEDRepository(pool *pgxpool.Pool) *GEDRepository {
	return &GEDRepository{pool: pool}
}

// GetTaxonomyWithCounts retourne l'arborescence des 5 branches et 21 sous-dossiers avec compteurs réels
func (r *GEDRepository) GetTaxonomyWithCounts(ctx context.Context, orgID uuid.UUID) ([]domain.GEDBranch, error) {
	branches := domain.GetStandardGEDTaxonomy()

	// Récupération des décomptes groupés par taxonomy_path
	query := `
		SELECT taxonomy_path, COUNT(*)
		FROM ged_documents
		WHERE organization_id = $1
		GROUP BY taxonomy_path
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("échec de lecture des compteurs de taxonomie : %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var path string
		var count int
		if err := rows.Scan(&path, &count); err == nil {
			counts[path] = count
		}
	}

	for i := range branches {
		branchTotal := 0
		for j := range branches[i].Folders {
			folderCount := counts[branches[i].Folders[j].Path]
			branches[i].Folders[j].DocumentCount = folderCount
			branchTotal += folderCount
		}
		branches[i].TotalDocs = branchTotal
	}

	return branches, nil
}

// ListDocuments retourne la liste paginée et filtrée des documents GED
func (r *GEDRepository) ListDocuments(ctx context.Context, orgID uuid.UUID, taxonomyPath, status, search string, limit, offset int) ([]domain.GEDDocumentDetailed, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var countQuery strings.Builder
	var listQuery strings.Builder
	args := []any{orgID}

	countQuery.WriteString(`SELECT COUNT(*) FROM ged_documents d WHERE d.organization_id = $1`)

	listQuery.WriteString(`
		SELECT 
			d.id, d.organization_id, d.taxonomy_path, d.title_fr,
			COALESCE(d.title_ar, ''),
			COALESCE(d.document_code, ''),
			d.document_type,
			COALESCE(d.department, ''),
			d.tags,
			d.status,
			d.current_version,
			d.expiry_date,
			d.is_locked,
			d.locked_by,
			COALESCE(ul.first_name || ' ' || ul.last_name, ul.email, ''),
			d.locked_at,
			COALESCE(d.lock_reason, ''),
			d.establishment_id,
			COALESCE(e.name_fr, ''),
			d.site_id,
			COALESCE(s.site_name_fr, ''),
			d.medicine_id,
			COALESCE(m.brand_name_fr, ''),
			d.medical_device_id,
			COALESCE(dm.commercial_name_fr, ''),
			d.created_by,
			COALESCE(uc.first_name || ' ' || uc.last_name, uc.email, ''),
			d.created_at,
			d.updated_at,
			v.id, v.version_label, v.is_major_version, v.revision_summary,
			v.file_sha256, v.file_size_bytes, v.mime_type, v.original_filename,
			v.storage_path, v.is_signed_dt, v.signed_by_dt_id,
			COALESCE(us.first_name || ' ' || us.last_name, us.email, ''),
			v.signed_at,
			COALESCE(v.signature_declaration, ''),
			COALESCE(v.signature_sha256, ''),
			COALESCE(v.signed_ip, ''),
			v.author_id,
			COALESCE(uv.first_name || ' ' || uv.last_name, uv.email, ''),
			v.created_at,
			(SELECT COUNT(*) FROM ged_versions gv WHERE gv.document_id = d.id) AS versions_count
		FROM ged_documents d
		LEFT JOIN users ul ON d.locked_by = ul.id
		LEFT JOIN users uc ON d.created_by = uc.id
		LEFT JOIN establishments e ON d.establishment_id = e.id
		LEFT JOIN manufacturing_sites s ON d.site_id = s.id
		LEFT JOIN medicines m ON d.medicine_id = m.id
		LEFT JOIN medical_devices dm ON d.medical_device_id = dm.id
		LEFT JOIN LATERAL (
			SELECT * FROM ged_versions gv
			WHERE gv.document_id = d.id
			ORDER BY gv.created_at DESC
			LIMIT 1
		) v ON TRUE
		LEFT JOIN users us ON v.signed_by_dt_id = us.id
		LEFT JOIN users uv ON v.author_id = uv.id
		WHERE d.organization_id = $1
	`)

	if taxonomyPath != "" {
		args = append(args, taxonomyPath)
		cond := fmt.Sprintf(` AND d.taxonomy_path = $%d`, len(args))
		countQuery.WriteString(cond)
		listQuery.WriteString(cond)
	}

	if status != "" {
		args = append(args, status)
		cond := fmt.Sprintf(` AND d.status = $%d`, len(args))
		countQuery.WriteString(cond)
		listQuery.WriteString(cond)
	}

	if search != "" {
		args = append(args, "%"+search+"%")
		cond := fmt.Sprintf(` AND (d.title_fr ILIKE $%d OR d.document_code ILIKE $%d OR d.title_ar ILIKE $%d)`, len(args), len(args), len(args))
		countQuery.WriteString(cond)
		listQuery.WriteString(cond)
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery.String(), args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("erreur de décompte des documents : %w", err)
	}

	args = append(args, limit, offset)
	listQuery.WriteString(fmt.Sprintf(` ORDER BY d.updated_at DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args)))

	rows, err := r.pool.Query(ctx, listQuery.String(), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur de sélection des documents : %w", err)
	}
	defer rows.Close()

	var list []domain.GEDDocumentDetailed
	for rows.Next() {
		var doc domain.GEDDocumentDetailed
		var (
			vID             *uuid.UUID
			vLabel          *string
			vIsMajor        *bool
			vSummary        *string
			vFileSHA        *string
			vFileSize       *int64
			vMime           *string
			vOrigFile       *string
			vStorage        *string
			vSignedDT       *bool
			vSignedByDTID   *uuid.UUID
			vSignedByDTName *string
			vSignedAt       *time.Time
			vSigDecl        *string
			vSigSHA         *string
			vSignedIP       *string
			vAuthorID       *uuid.UUID
			vAuthorName     *string
			vCreatedAt      *time.Time
			vCount          int
		)

		err := rows.Scan(
			&doc.ID, &doc.OrganizationID, &doc.TaxonomyPath, &doc.TitleFR,
			&doc.TitleAR, &doc.DocumentCode, &doc.DocumentType, &doc.Department,
			&doc.Tags, &doc.Status, &doc.CurrentVersion, &doc.ExpiryDate,
			&doc.IsLocked, &doc.LockedBy, &doc.LockedByName, &doc.LockedAt,
			&doc.LockReason, &doc.EstablishmentID, &doc.EstablishmentName,
			&doc.SiteID, &doc.SiteName, &doc.MedicineID, &doc.MedicineBrandName,
			&doc.MedicalDeviceID, &doc.MedicalDeviceName, &doc.CreatedBy,
			&doc.CreatedByName, &doc.CreatedAt, &doc.UpdatedAt,
			&vID, &vLabel, &vIsMajor, &vSummary,
			&vFileSHA, &vFileSize, &vMime, &vOrigFile,
			&vStorage, &vSignedDT, &vSignedByDTID,
			&vSignedByDTName, &vSignedAt, &vSigDecl,
			&vSigSHA, &vSignedIP, &vAuthorID,
			&vAuthorName, &vCreatedAt, &vCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("erreur de scan document : %w", err)
		}

		doc.VersionsCount = vCount
		if vID != nil && *vID != uuid.Nil {
			doc.LatestVersion = &domain.GEDVersionDetail{
				ID:                   *vID,
				DocumentID:           doc.ID,
				VersionLabel:         derefStr(vLabel),
				IsMajorVersion:       derefBool(vIsMajor),
				RevisionSummary:      derefStr(vSummary),
				FileSHA256:           derefStr(vFileSHA),
				FileSizeBytes:        derefInt64(vFileSize),
				MimeType:             derefStr(vMime),
				OriginalFilename:     derefStr(vOrigFile),
				StoragePath:          derefStr(vStorage),
				IsSignedDT:           derefBool(vSignedDT),
				SignedByDTID:         vSignedByDTID,
				SignedByDTName:       derefStr(vSignedByDTName),
				SignedAt:             vSignedAt,
				SignatureDeclaration: derefStr(vSigDecl),
				SignatureSHA256:      derefStr(vSigSHA),
				SignedIP:             derefStr(vSignedIP),
				AuthorID:             vAuthorID,
				AuthorName:           derefStr(vAuthorName),
				CreatedAt:            derefTime(vCreatedAt),
			}
		}

		list = append(list, doc)
	}

	return list, total, nil
}

// GetDocumentByID récupère un document spécifique avec sa dernière version
func (r *GEDRepository) GetDocumentByID(ctx context.Context, orgID, docID uuid.UUID) (*domain.GEDDocumentDetailed, error) {
	docs, _, err := r.ListDocuments(ctx, orgID, "", "", "", 1, 0)
	if err != nil {
		return nil, err
	}
	// Recherche ciblée par docID
	query := `
		SELECT 
			d.id, d.organization_id, d.taxonomy_path, d.title_fr,
			COALESCE(d.title_ar, ''),
			COALESCE(d.document_code, ''),
			d.document_type,
			COALESCE(d.department, ''),
			d.tags,
			d.status,
			d.current_version,
			d.expiry_date,
			d.is_locked,
			d.locked_by,
			COALESCE(ul.first_name || ' ' || ul.last_name, ul.email, ''),
			d.locked_at,
			COALESCE(d.lock_reason, ''),
			d.establishment_id,
			COALESCE(e.name_fr, ''),
			d.site_id,
			COALESCE(s.site_name_fr, ''),
			d.medicine_id,
			COALESCE(m.brand_name_fr, ''),
			d.medical_device_id,
			COALESCE(dm.commercial_name_fr, ''),
			d.created_by,
			COALESCE(uc.first_name || ' ' || uc.last_name, uc.email, ''),
			d.created_at,
			d.updated_at,
			v.id, v.version_label, v.is_major_version, v.revision_summary,
			v.file_sha256, v.file_size_bytes, v.mime_type, v.original_filename,
			v.storage_path, v.is_signed_dt, v.signed_by_dt_id,
			COALESCE(us.first_name || ' ' || us.last_name, us.email, ''),
			v.signed_at,
			COALESCE(v.signature_declaration, ''),
			COALESCE(v.signature_sha256, ''),
			COALESCE(v.signed_ip, ''),
			v.author_id,
			COALESCE(uv.first_name || ' ' || uv.last_name, uv.email, ''),
			v.created_at,
			(SELECT COUNT(*) FROM ged_versions gv WHERE gv.document_id = d.id) AS versions_count
		FROM ged_documents d
		LEFT JOIN users ul ON d.locked_by = ul.id
		LEFT JOIN users uc ON d.created_by = uc.id
		LEFT JOIN establishments e ON d.establishment_id = e.id
		LEFT JOIN manufacturing_sites s ON d.site_id = s.id
		LEFT JOIN medicines m ON d.medicine_id = m.id
		LEFT JOIN medical_devices dm ON d.medical_device_id = dm.id
		LEFT JOIN LATERAL (
			SELECT * FROM ged_versions gv
			WHERE gv.document_id = d.id
			ORDER BY gv.created_at DESC
			LIMIT 1
		) v ON TRUE
		LEFT JOIN users us ON v.signed_by_dt_id = us.id
		LEFT JOIN users uv ON v.author_id = uv.id
		WHERE d.organization_id = $1 AND d.id = $2
	`
	_ = docs // avoid unused
	var doc domain.GEDDocumentDetailed
	var (
		vID             *uuid.UUID
		vLabel          *string
		vIsMajor        *bool
		vSummary        *string
		vFileSHA        *string
		vFileSize       *int64
		vMime           *string
		vOrigFile       *string
		vStorage        *string
		vSignedDT       *bool
		vSignedByDTID   *uuid.UUID
		vSignedByDTName *string
		vSignedAt       *time.Time
		vSigDecl        *string
		vSigSHA         *string
		vSignedIP       *string
		vAuthorID       *uuid.UUID
		vAuthorName     *string
		vCreatedAt      *time.Time
		vCount          int
	)

	err = r.pool.QueryRow(ctx, query, orgID, docID).Scan(
		&doc.ID, &doc.OrganizationID, &doc.TaxonomyPath, &doc.TitleFR,
		&doc.TitleAR, &doc.DocumentCode, &doc.DocumentType, &doc.Department,
		&doc.Tags, &doc.Status, &doc.CurrentVersion, &doc.ExpiryDate,
		&doc.IsLocked, &doc.LockedBy, &doc.LockedByName, &doc.LockedAt,
		&doc.LockReason, &doc.EstablishmentID, &doc.EstablishmentName,
		&doc.SiteID, &doc.SiteName, &doc.MedicineID, &doc.MedicineBrandName,
		&doc.MedicalDeviceID, &doc.MedicalDeviceName, &doc.CreatedBy,
		&doc.CreatedByName, &doc.CreatedAt, &doc.UpdatedAt,
		&vID, &vLabel, &vIsMajor, &vSummary,
		&vFileSHA, &vFileSize, &vMime, &vOrigFile,
		&vStorage, &vSignedDT, &vSignedByDTID,
		&vSignedByDTName, &vSignedAt, &vSigDecl,
		&vSigSHA, &vSignedIP, &vAuthorID,
		&vAuthorName, &vCreatedAt, &vCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("document non trouvé")
		}
		return nil, fmt.Errorf("erreur de recherche document : %w", err)
	}

	doc.VersionsCount = vCount
	if vID != nil && *vID != uuid.Nil {
		doc.LatestVersion = &domain.GEDVersionDetail{
			ID:                   *vID,
			DocumentID:           doc.ID,
			VersionLabel:         derefStr(vLabel),
			IsMajorVersion:       derefBool(vIsMajor),
			RevisionSummary:      derefStr(vSummary),
			FileSHA256:           derefStr(vFileSHA),
			FileSizeBytes:        derefInt64(vFileSize),
			MimeType:             derefStr(vMime),
			OriginalFilename:     derefStr(vOrigFile),
			StoragePath:          derefStr(vStorage),
			IsSignedDT:           derefBool(vSignedDT),
			SignedByDTID:         vSignedByDTID,
			SignedByDTName:       derefStr(vSignedByDTName),
			SignedAt:             vSignedAt,
			SignatureDeclaration: derefStr(vSigDecl),
			SignatureSHA256:      derefStr(vSigSHA),
			SignedIP:             derefStr(vSignedIP),
			AuthorID:             vAuthorID,
			AuthorName:           derefStr(vAuthorName),
			CreatedAt:            derefTime(vCreatedAt),
		}
	}

	return &doc, nil
}

// GetDocumentVersions retourne l'historique complet des versions d'un document
func (r *GEDRepository) GetDocumentVersions(ctx context.Context, docID uuid.UUID) ([]domain.GEDVersionDetail, error) {
	query := `
		SELECT 
			v.id, v.document_id, v.version_label, v.is_major_version, v.revision_summary,
			v.file_sha256, v.file_size_bytes, v.mime_type, v.original_filename,
			v.storage_path, v.is_signed_dt, v.signed_by_dt_id,
			COALESCE(us.first_name || ' ' || us.last_name, us.email, ''),
			v.signed_at,
			COALESCE(v.signature_declaration, ''),
			COALESCE(v.signature_sha256, ''),
			COALESCE(v.signed_ip, ''),
			v.author_id,
			COALESCE(ua.first_name || ' ' || ua.last_name, ua.email, ''),
			v.created_at
		FROM ged_versions v
		LEFT JOIN users us ON v.signed_by_dt_id = us.id
		LEFT JOIN users ua ON v.author_id = ua.id
		WHERE v.document_id = $1
		ORDER BY v.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, docID)
	if err != nil {
		return nil, fmt.Errorf("échec de lecture des versions : %w", err)
	}
	defer rows.Close()

	var versions []domain.GEDVersionDetail
	for rows.Next() {
		var v domain.GEDVersionDetail
		err := rows.Scan(
			&v.ID, &v.DocumentID, &v.VersionLabel, &v.IsMajorVersion, &v.RevisionSummary,
			&v.FileSHA256, &v.FileSizeBytes, &v.MimeType, &v.OriginalFilename,
			&v.StoragePath, &v.IsSignedDT, &v.SignedByDTID, &v.SignedByDTName,
			&v.SignedAt, &v.SignatureDeclaration, &v.SignatureSHA256, &v.SignedIP,
			&v.AuthorID, &v.AuthorName, &v.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur de scan version : %w", err)
		}
		versions = append(versions, v)
	}

	return versions, nil
}

// GetVersionByID récupère une version spécifique pour téléchargement
func (r *GEDRepository) GetVersionByID(ctx context.Context, versionID uuid.UUID) (*domain.GEDVersionDetail, error) {
	query := `
		SELECT 
			v.id, v.document_id, v.version_label, v.is_major_version, v.revision_summary,
			v.file_sha256, v.file_size_bytes, v.mime_type, v.original_filename,
			v.storage_path, v.is_signed_dt, v.signed_by_dt_id,
			COALESCE(us.first_name || ' ' || us.last_name, us.email, ''),
			v.signed_at,
			COALESCE(v.signature_declaration, ''),
			COALESCE(v.signature_sha256, ''),
			COALESCE(v.signed_ip, ''),
			v.author_id,
			COALESCE(ua.first_name || ' ' || ua.last_name, ua.email, ''),
			v.created_at
		FROM ged_versions v
		LEFT JOIN users us ON v.signed_by_dt_id = us.id
		LEFT JOIN users ua ON v.author_id = ua.id
		WHERE v.id = $1
	`
	var v domain.GEDVersionDetail
	err := r.pool.QueryRow(ctx, query, versionID).Scan(
		&v.ID, &v.DocumentID, &v.VersionLabel, &v.IsMajorVersion, &v.RevisionSummary,
		&v.FileSHA256, &v.FileSizeBytes, &v.MimeType, &v.OriginalFilename,
		&v.StoragePath, &v.IsSignedDT, &v.SignedByDTID, &v.SignedByDTName,
		&v.SignedAt, &v.SignatureDeclaration, &v.SignatureSHA256, &v.SignedIP,
		&v.AuthorID, &v.AuthorName, &v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("version de document non trouvée")
		}
		return nil, fmt.Errorf("erreur de sélection version : %w", err)
	}
	return &v, nil
}

// CreateDocument insère un document et sa version initiale au sein d'une transaction PostgreSQL
func (r *GEDRepository) CreateDocument(ctx context.Context, doc *domain.GEDDocumentDetailed, v *domain.GEDVersionDetail) (*domain.GEDDocumentDetailed, error) {
	if err := doc.Validate(); err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("impossible de démarrer la transaction : %w", err)
	}
	defer tx.Rollback(ctx)

	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}
	if doc.CurrentVersion == "" {
		doc.CurrentVersion = "v0.1"
	}
	if doc.Status == "" {
		doc.Status = domain.GEDStatusBrouillon
	}

	docQuery := `
		INSERT INTO ged_documents (
			id, organization_id, taxonomy_path, title_fr, title_ar,
			document_code, document_type, department, tags, status,
			current_version, expiry_date, is_locked, locked_by,
			establishment_id, site_id, medicine_id, medical_device_id,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, FALSE, NULL,
			$13, $14, $15, $16,
			$17, NOW(), NOW()
		)
		RETURNING created_at, updated_at
	`
	err = tx.QueryRow(ctx, docQuery,
		doc.ID, doc.OrganizationID, doc.TaxonomyPath, doc.TitleFR, doc.TitleAR,
		doc.DocumentCode, doc.DocumentType, doc.Department, doc.Tags, doc.Status,
		doc.CurrentVersion, doc.ExpiryDate,
		doc.EstablishmentID, doc.SiteID, doc.MedicineID, doc.MedicalDeviceID,
		doc.CreatedBy,
	).Scan(&doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("échec d'insertion du document : %w", err)
	}

	// Insertion de la première version
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	v.DocumentID = doc.ID
	v.VersionLabel = doc.CurrentVersion

	versionQuery := `
		INSERT INTO ged_versions (
			id, document_id, version_label, is_major_version, revision_summary,
			file_sha256, file_size_bytes, mime_type, original_filename,
			storage_path, is_signed_dt, signed_by_dt_id, signed_at,
			signature_declaration, signature_sha256, signed_ip,
			author_id, created_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16,
			$17, NOW()
		)
		RETURNING created_at
	`
	err = tx.QueryRow(ctx, versionQuery,
		v.ID, v.DocumentID, v.VersionLabel, v.IsMajorVersion, v.RevisionSummary,
		v.FileSHA256, v.FileSizeBytes, v.MimeType, v.OriginalFilename,
		v.StoragePath, v.IsSignedDT, v.SignedByDTID, v.SignedAt,
		v.SignatureDeclaration, v.SignatureSHA256, v.SignedIP,
		v.AuthorID,
	).Scan(&v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("échec d'insertion de la version initiale : %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("échec de validation de la transaction : %w", err)
	}

	doc.LatestVersion = v
	doc.VersionsCount = 1
	return doc, nil
}

// CreateNewVersion insère une nouvelle version dans l'historique et met à jour le document
func (r *GEDRepository) CreateNewVersion(ctx context.Context, orgID, docID uuid.UUID, v *domain.GEDVersionDetail) (*domain.GEDVersionDetail, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("impossible de démarrer la transaction : %w", err)
	}
	defer tx.Rollback(ctx)

	// Vérifier l'existence et l'état de verrouillage du document
	var currentVersion string
	var isLocked bool
	var lockedBy *uuid.UUID

	err = tx.QueryRow(ctx, `
		SELECT current_version, is_locked, locked_by
		FROM ged_documents
		WHERE id = $1 AND organization_id = $2
		FOR UPDATE
	`, docID, orgID).Scan(&currentVersion, &isLocked, &lockedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("document introuvable")
		}
		return nil, err
	}

	// Si verrouillé par un autre utilisateur, refuser
	if isLocked && lockedBy != nil && v.AuthorID != nil && *lockedBy != *v.AuthorID {
		return nil, errors.New("ce document est actuellement verrouillé par un autre utilisateur pour modification")
	}

	// Calcul du nouveau label de version
	newLabel := domain.GenerateNextVersionLabel(currentVersion, v.IsMajorVersion, false)
	v.VersionLabel = newLabel
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	v.DocumentID = docID

	versionQuery := `
		INSERT INTO ged_versions (
			id, document_id, version_label, is_major_version, revision_summary,
			file_sha256, file_size_bytes, mime_type, original_filename,
			storage_path, is_signed_dt, signed_by_dt_id, signed_at,
			signature_declaration, signature_sha256, signed_ip,
			author_id, created_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16,
			$17, NOW()
		)
		RETURNING created_at
	`
	err = tx.QueryRow(ctx, versionQuery,
		v.ID, v.DocumentID, v.VersionLabel, v.IsMajorVersion, v.RevisionSummary,
		v.FileSHA256, v.FileSizeBytes, v.MimeType, v.OriginalFilename,
		v.StoragePath, v.IsSignedDT, v.SignedByDTID, v.SignedAt,
		v.SignatureDeclaration, v.SignatureSHA256, v.SignedIP,
		v.AuthorID,
	).Scan(&v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("échec d'insertion de la version : %w", err)
	}

	// Mise à jour du document : nouveau current_version, updated_at, et déverrouillage automatique (Check-in)
	updateDocQuery := `
		UPDATE ged_documents
		SET current_version = $1,
		    is_locked = FALSE,
		    locked_by = NULL,
		    locked_at = NULL,
		    lock_reason = NULL,
		    updated_at = NOW()
		WHERE id = $2
	`
	_, err = tx.Exec(ctx, updateDocQuery, newLabel, docID)
	if err != nil {
		return nil, fmt.Errorf("échec de mise à jour du document : %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("échec du commit : %w", err)
	}

	return v, nil
}

// CheckOutDocument verrouille un document pour édition exclusive
func (r *GEDRepository) CheckOutDocument(ctx context.Context, orgID, docID, userID uuid.UUID, reason string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var isLocked bool
	var lockedBy *uuid.UUID

	err = tx.QueryRow(ctx, `
		SELECT is_locked, locked_by
		FROM ged_documents
		WHERE id = $1 AND organization_id = $2
		FOR UPDATE
	`, docID, orgID).Scan(&isLocked, &lockedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("document introuvable")
		}
		return err
	}

	if isLocked {
		if lockedBy != nil && *lockedBy == userID {
			return errors.New("vous avez déjà verrouillé ce document")
		}
		return errors.New("ce document est déjà en cours de modification (verrouillé par un autre utilisateur)")
	}

	_, err = tx.Exec(ctx, `
		UPDATE ged_documents
		SET is_locked = TRUE,
		    locked_by = $1,
		    locked_at = NOW(),
		    lock_reason = $2,
		    updated_at = NOW()
		WHERE id = $3
	`, userID, reason, docID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// CancelCheckOut annule le verrouillage d'un document
func (r *GEDRepository) CancelCheckOut(ctx context.Context, orgID, docID, userID uuid.UUID, isAdmin bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var isLocked bool
	var lockedBy *uuid.UUID

	err = tx.QueryRow(ctx, `
		SELECT is_locked, locked_by
		FROM ged_documents
		WHERE id = $1 AND organization_id = $2
		FOR UPDATE
	`, docID, orgID).Scan(&isLocked, &lockedBy)
	if err != nil {
		return err
	}

	if !isLocked {
		return errors.New("ce document n'est pas verrouillé")
	}

	// Seul l'auteur du verrou ou un administrateur peut annuler
	if !isAdmin && (lockedBy == nil || *lockedBy != userID) {
		return errors.New("vous n'êtes pas autorisé à annuler le verrouillage posé par un autre utilisateur")
	}

	_, err = tx.Exec(ctx, `
		UPDATE ged_documents
		SET is_locked = FALSE,
		    locked_by = NULL,
		    locked_at = NULL,
		    lock_reason = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`, docID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// UpdateDocumentStatus modifie le statut du cycle de vie documentaire
func (r *GEDRepository) UpdateDocumentStatus(ctx context.Context, orgID, docID uuid.UUID, targetStatus domain.GEDDocumentStatus) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE ged_documents
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND organization_id = $3
	`, targetStatus, docID, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("document non trouvé")
	}
	return nil
}

// SignDocumentVersionByDT effectue le scellement officiel par le Pharmacien Directeur Technique
func (r *GEDRepository) SignDocumentVersionByDT(ctx context.Context, orgID, docID, versionID, dtUserID uuid.UUID, declaration, signatureSeal, clientIP string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Mise à jour de la version
	res, err := tx.Exec(ctx, `
		UPDATE ged_versions
		SET is_signed_dt = TRUE,
		    signed_by_dt_id = $1,
		    signed_at = NOW(),
		    signature_declaration = $2,
		    signature_sha256 = $3,
		    signed_ip = $4
		WHERE id = $5 AND document_id = $6
	`, dtUserID, declaration, signatureSeal, clientIP, versionID, docID)
	if err != nil {
		return fmt.Errorf("erreur de signature de version : %w", err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("version non trouvée pour ce document")
	}

	// Mise à jour du document : passage en vigueur s'il était en approbation
	_, err = tx.Exec(ctx, `
		UPDATE ged_documents
		SET status = 'en_vigueur',
		    updated_at = NOW()
		WHERE id = $1 AND organization_id = $2 AND status = 'en_approbation_dt'
	`, docID, orgID)
	if err != nil {
		return fmt.Errorf("erreur de mise à jour du statut du document : %w", err)
	}

	return tx.Commit(ctx)
}

// Helper de déréférencement sécurisé
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func derefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
