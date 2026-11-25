package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// DB wraps the pgxpool.Pool
type DB struct {
	Pool *pgxpool.Pool
}

// Connect initializes the PostgreSQL connection pool and runs pending migrations
func Connect(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erreur de configuration PostgreSQL: %w", err)
	}

	// Performance et stabilité BPF
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("impossible de créer le pool de connexions PostgreSQL: %w", err)
	}

	// Test ping
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("impossible de joindre le serveur PostgreSQL: %w", err)
	}

	db := &DB{Pool: pool}

	// Exécution des migrations SQL
	if err := db.RunMigrations(ctx); err != nil {
		return nil, fmt.Errorf("échec de l'application des migrations: %w", err)
	}

	log.Println("[PostgreSQL] Connexion établie et schéma vérifié avec succès.")
	return db, nil
}

// RunMigrations applies embedded migration files
func (db *DB) RunMigrations(ctx context.Context) error {
	migrationFiles := []string{
		"migrations/001_initial_schema.sql",
		"migrations/002_establishments_and_products.sql",
		"migrations/003_ged_vault_alcoa.sql",
		"migrations/004_anpp_ctd_dossiers.sql",
		"migrations/005_rbac_and_dynamic_permissions.sql",
		"migrations/006_consultant_guest_cloisonnement.sql",
		"migrations/007_industrial_and_commercial_secrecy.sql",
		"migrations/008_foreign_manufacturer_and_cpp.sql",
		"migrations/009_medical_devices_traceability_and_crud.sql",
		"migrations/010_anpp_deficiency_letters_and_tracking.sql",
		"migrations/011_anpp_fee_schedules_and_payment_slips.sql",
		"migrations/012_anpp_appointments_and_regulatory_watch.sql",
		"migrations/013_maker_checker_workflow_and_reporting.sql",
		"migrations/014_api_holder_role_and_branding.sql",
		"migrations/015_realign_checklist_departments.sql",
		"migrations/016_smart_planning_working_days.sql",
		"migrations/017_algerian_holidays.sql",
		"migrations/018_algerian_holidays_aid_3_days.sql",
		"migrations/019_anpp_ich_master_templates.sql",
		"migrations/020_pharmacien_assistant_and_dt_delegations.sql",
		"migrations/021_tenant_regulatory_settings_and_fees.sql",
		"migrations/022_harden_admin_lab_need_to_know.sql",
		"migrations/023_sod_prevent_self_granting_and_admin_lab_escalation.sql",
		"migrations/024_volet_1a4_payment_slip_verification_and_fx_attestation.sql",
		"migrations/025_volet_1a5_presubmission_and_cost_consolidation.sql",
		"migrations/026_anpp_variations_engine.sql",
		"migrations/027_dm_homologation_and_regularization.sql",
		"migrations/028_dm_post_homologation_modifications.sql",
		"migrations/029_optimize_query_and_foreign_key_indexes.sql",
		"migrations/030_miph_corporate_agrements_and_tabadol.sql",
		"migrations/031_granular_corporate_checklist_items.sql",
		"migrations/032_audit_logs_location.sql",
		"migrations/033_jalon_2_7bis_decret_21_82_and_regulatory_watch.sql",
	}

	for _, file := range migrationFiles {
		sqlContent, err := migrationFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("lecture du fichier de migration %s impossible: %w", file, err)
		}

		_, err = db.Pool.Exec(ctx, string(sqlContent))
		if err != nil {
			return fmt.Errorf("exécution de la migration %s échouée: %w", file, err)
		}
		log.Printf("[Migration] %s appliquée avec succès.", file)
	}

	return nil
}

// Close closes the connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
