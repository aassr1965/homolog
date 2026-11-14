package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"homolog/internal/api"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/config"
	"homolog/internal/database"
	"homolog/internal/domain"
	"homolog/internal/repository"
	"homolog/internal/storage"
	"homolog/web"
)

func main() {
	log.Println("==================================================================")
	log.Println("  SANAD PHARMA - Plateforme Réglementaire ANPP & GED Pharmaceutique")
	log.Println("  Version 0.7 - Hébergement Hybride (On-Premise & Cloud SaaS)     ")
	log.Println("==================================================================")

	cfg := config.Load()
	log.Printf("[Config] Mode de déploiement : %s", cfg.DeploymentMode)
	log.Printf("[Config] Port d'écoute : %d", cfg.Port)

	// 1. Initialisation du stockage de fichiers (CAS dédupliqué SHA-256)
	storageProvider, err := storage.NewLocalCASStorage(cfg.StoragePath)
	if err != nil {
		log.Fatalf("[FATAL] Échec d'initialisation du stockage : %v", err)
	}
	log.Printf("[Stockage] Répertoire CAS local actif : %s", cfg.StoragePath)
	_ = storageProvider // sera injecté dans les services GED

	// 2. Connexion à PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var db *database.DB
	var sessionManager *auth.SessionManager
	var auditService *audit.Service
	var authHandler *api.AuthHandler
	var adminHandler *api.AdminHandler
	var systemHandler *api.SystemHandler
	var pharmaHandler *api.PharmaHandler
	var gedHandler *api.GEDHandler
	var dossierHandler *api.DossierHandler
	var deficiencyHandler *api.DeficiencyHandler

	db, err = database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("[AVERTISSEMENT] Connexion PostgreSQL non disponible au démarrage : %v", err)
		log.Println("[INFO] Le serveur démarre en attente de la base de données (Setup / Diagnostics).")
	} else {
		defer db.Close()

		// Initialisation des données de démarrage (Labo principal + comptes DT/Admin)
		if err := api.SeedInitialData(ctx, db.Pool, cfg.OnPremOrgSlug, cfg.OnPremOrgName); err != nil {
			log.Printf("[AVERTISSEMENT] Erreur lors de l'initialisation des comptes : %v", err)
		}

		// Initialisation des documents modèles du Coffre-Fort GED (Jalon 1.4)
		api.SeedInitialGEDDocuments(ctx, db.Pool, storageProvider, cfg.OnPremOrgSlug)

		// Initialisation des dossiers réglementaires ANPP de démonstration (Jalon 1.5)
		api.SeedInitialDossiers(ctx, db.Pool, cfg.OnPremOrgSlug)

		// Initialisation du Studio de Compléments ANPP & Traçabilité Délais (Jalon 1.7)
		api.SeedInitialDeficiencies(ctx, db.Pool, cfg.OnPremOrgSlug)

		sessionManager = auth.NewSessionManager(db.Pool, cfg.SessionExpiryMinutes, cfg.IdleTimeoutMinutes)
		auditService = audit.NewService(db.Pool)
		authHandler = api.NewAuthHandler(db.Pool, sessionManager, auditService, cfg)
		adminHandler = api.NewAdminHandler(db.Pool, sessionManager, auditService, cfg)
		systemHandler = api.NewSystemHandler(db.Pool, sessionManager, auditService, cfg)

		pharmaRepo := repository.NewPharmaRepository(db.Pool)
		pharmaHandler = api.NewPharmaHandler(pharmaRepo, auditService)

		gedRepo := repository.NewGEDRepository(db.Pool)
		gedHandler = api.NewGEDHandler(gedRepo, storageProvider, auditService, db.Pool)

		dossierRepo := repository.NewDossierRepository(db.Pool)
		dossierHandler = api.NewDossierHandler(dossierRepo, pharmaRepo, gedRepo, auditService, db.Pool)

		deficiencyRepo := repository.NewDeficiencyRepository(db.Pool)
		deficiencyHandler = api.NewDeficiencyHandler(deficiencyRepo, dossierRepo, auditService, db.Pool)
	}

	// 3. Configuration du Routeur HTTP
	r := chi.NewRouter()

	// Middlewares standards
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS sécurisé
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Interface Web Haute Densité 'Pharma Explorer' (embarquée dans le binaire)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		content, err := web.Files.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Interface web non trouvée", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	})

	// Présentation Commerciale Haute Fidélité (Pitch Deck 10 Diapositives)
	r.Get("/presentation", func(w http.ResponseWriter, r *http.Request) {
		content, err := web.Files.ReadFile("presentation.html")
		if err != nil {
			http.Error(w, "Support de présentation introuvable", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	})
	r.Get("/pitch", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/presentation", http.StatusMovedPermanently)
	})

	// Favicon Caducée officiel Sanad Pharma
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		ico, err := web.Files.ReadFile("favicon.ico")
		if err == nil {
			w.Header().Set("Content-Type", "image/x-icon")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Write(ico)
			return
		}
		svgIco, err := web.Files.ReadFile("img/favicon-caducee.svg")
		if err == nil {
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Write(svgIco)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Servir les assets statiques embarqués (css, js, img)
	staticFS := http.FileServer(http.FS(web.Files))
	noCacheWrapper := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			h.ServeHTTP(w, r)
		})
	}
	r.Handle("/css/*", noCacheWrapper(staticFS))
	r.Handle("/js/*", noCacheWrapper(staticFS))
	r.Handle("/img/*", staticFS)

	// Routes de diagnostic et d'état de santé
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		dbStatus := "connected"
		if db == nil {
			dbStatus = "disconnected"
		}
		jsonResp := fmt.Sprintf(`{"status":"ok","mode":"%s","environment":"%s","dev_switcher_enabled":%t,"database":"%s","time_utc":"%s","version":"0.7.0"}`,
			cfg.DeploymentMode, cfg.Environment, cfg.DevSwitcherEnabled(), dbStatus, time.Now().UTC().Format(time.RFC3339))
		w.Write([]byte(jsonResp))
	})

	r.Get("/api/v1/system/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := fmt.Sprintf(`{
			"app": "Sanad Pharma",
			"version": "0.7.0",
			"deployment_mode": "%s",
			"environment": "%s",
			"dev_switcher_enabled": %t,
			"org_slug": "%s",
			"org_name": "%s",
			"idle_timeout_min": %d,
			"max_upload_bytes": %d,
			"timestamp_utc": "%s"
		}`, cfg.DeploymentMode, cfg.Environment, cfg.DevSwitcherEnabled(), cfg.OnPremOrgSlug, cfg.OnPremOrgName, cfg.IdleTimeoutMinutes, cfg.MaxUploadSizeBytes, time.Now().UTC().Format(time.RFC3339))
		w.Write([]byte(resp))
	})

	// Routes API Authentification, 2FA, Signature DT et Audit Trail
	if authHandler != nil {
		r.Post("/api/v1/auth/login", authHandler.Login)
		r.Post("/api/v1/auth/logout", authHandler.Logout)

		// Role Switcher de développement (Environnement non-production / Dev rapide)
		// Strictement gardé : ne s'enregistre que si cfg.DevSwitcherEnabled() est vrai (désactivé d'office en production)
		if adminHandler != nil && cfg.DevSwitcherEnabled() {
			r.Post("/api/v1/auth/dev-switch", adminHandler.DevSwitchRole)
		}

		// Routes protégées par session HttpOnly et RBAC (CDC Section 7.2 / Jalon 1.2)
		r.Group(func(protected chi.Router) {
			protected.Use(api.AuthMiddleware(sessionManager))
			protected.Get("/api/v1/auth/me", authHandler.Me)
			protected.Post("/api/v1/auth/reauth-sign", authHandler.ReauthSign)

			// Piste d'audit immuable (AUDIT_VIEW)
			protected.With(api.RequirePermission(auditService, domain.PermAuditView)).Get("/api/v1/audit/logs", authHandler.GetAuditLogs)

			// Console Dédiée d'Administration Laboratoire & Matrice RBAC (USER_MANAGE)
			if adminHandler != nil {
				protected.With(api.RequirePermission(auditService, domain.PermUserManage)).Get("/api/v1/admin/users", adminHandler.ListUsers)
				protected.With(api.RequirePermission(auditService, domain.PermUserManage)).Post("/api/v1/admin/users", adminHandler.CreateUser)
				protected.With(api.RequirePermission(auditService, domain.PermUserManage)).Put("/api/v1/admin/users/{id}/role", adminHandler.UpdateUserRole)
				protected.With(api.RequirePermission(auditService, domain.PermUserManage)).Put("/api/v1/admin/users/{id}/status", adminHandler.UpdateUserStatus)
				protected.With(api.RequirePermission(auditService, domain.PermUserManage)).Post("/api/v1/admin/users/{id}/reset-password", adminHandler.ResetUserPassword)
				protected.With(api.RequirePermission(auditService, domain.PermUserManage)).Post("/api/v1/admin/users/{id}/reset-totp", adminHandler.ResetUserTOTP)
				protected.With(api.RequirePermission(auditService, domain.PermUserManage)).Put("/api/v1/admin/users/{id}/permissions", adminHandler.UpdateUserPermissions)
				protected.With(api.RequirePermission(auditService, domain.PermUserManage)).Delete("/api/v1/admin/users/{id}", adminHandler.DisallowDeleteUser)
				protected.With(api.RequirePermission(auditService, domain.PermUserManage)).Get("/api/v1/admin/roles", adminHandler.ListRoles)
				protected.With(api.RequirePermission(auditService, domain.PermUserManage)).Get("/api/v1/admin/permissions", adminHandler.ListPermissions)
			}

			// Console Dédiée Administrateur Système & Maintenance IT (SYS_MAINT / PSSI)
			if systemHandler != nil {
				protected.With(api.RequirePermission(auditService, domain.PermSysMaint)).Get("/api/v1/system/metrics", systemHandler.GetSystemMetrics)
				protected.With(api.RequirePermission(auditService, domain.PermSysMaint)).Get("/api/v1/system/sessions", systemHandler.ListActiveSessions)
				protected.With(api.RequirePermission(auditService, domain.PermSysMaint)).Delete("/api/v1/system/sessions/{id}", systemHandler.RevokeSession)
				protected.With(api.RequirePermission(auditService, domain.PermSysMaint)).Post("/api/v1/system/sessions/emergency-lockdown", systemHandler.EmergencyLockdown)
				protected.With(api.RequirePermission(auditService, domain.PermSysMaint)).Post("/api/v1/system/backup", systemHandler.TriggerBackup)
				protected.With(api.RequirePermission(auditService, domain.PermSysMaint)).Post("/api/v1/system/verify-integrity", systemHandler.VerifyCASIntegrity)
			}

			// Référentiel Établissements, Sites & Produits (Jalon 1.3 - Guide ANPP 2025)
			if pharmaHandler != nil {
				// Établissements
				protected.With(api.RequirePermission(auditService, domain.PermRefView)).Get("/api/v1/establishments", pharmaHandler.ListEstablishments)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Post("/api/v1/establishments", pharmaHandler.CreateEstablishment)
				protected.With(api.RequirePermission(auditService, domain.PermRefView)).Get("/api/v1/establishments/{id}", pharmaHandler.GetEstablishment)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Put("/api/v1/establishments/{id}", pharmaHandler.UpdateEstablishment)

				// Sites intervenants (1.H.8)
				protected.With(api.RequirePermission(auditService, domain.PermRefView)).Get("/api/v1/manufacturing-sites", pharmaHandler.ListManufacturingSites)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Post("/api/v1/manufacturing-sites", pharmaHandler.CreateManufacturingSite)
				protected.With(api.RequirePermission(auditService, domain.PermRefView)).Get("/api/v1/manufacturing-sites/{id}", pharmaHandler.GetManufacturingSite)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Put("/api/v1/manufacturing-sites/{id}", pharmaHandler.UpdateManufacturingSite)

				// Médicaments (Princeps, Générique importé, Générique local)
				protected.With(api.RequirePermission(auditService, domain.PermRefView)).Get("/api/v1/medicines", pharmaHandler.ListMedicines)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Post("/api/v1/medicines", pharmaHandler.CreateMedicine)
				protected.With(api.RequirePermission(auditService, domain.PermRefView)).Get("/api/v1/medicines/{id}", pharmaHandler.GetMedicine)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Put("/api/v1/medicines/{id}", pharmaHandler.UpdateMedicine)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Post("/api/v1/medicines/{id}/archive", pharmaHandler.ArchiveMedicine)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Delete("/api/v1/medicines/{id}", pharmaHandler.ArchiveMedicine)

				// Dispositifs Médicaux (Décret 20-324 & Arrêté 10/05/2021)
				protected.With(api.RequirePermission(auditService, domain.PermRefView)).Get("/api/v1/medical-devices", pharmaHandler.ListMedicalDevices)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Post("/api/v1/medical-devices", pharmaHandler.CreateMedicalDevice)
				protected.With(api.RequirePermission(auditService, domain.PermRefView)).Get("/api/v1/medical-devices/{id}", pharmaHandler.GetMedicalDevice)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Put("/api/v1/medical-devices/{id}", pharmaHandler.UpdateMedicalDevice)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Post("/api/v1/medical-devices/{id}/archive", pharmaHandler.ArchiveMedicalDevice)
				protected.With(api.RequirePermission(auditService, domain.PermRefEdit)).Delete("/api/v1/medical-devices/{id}", pharmaHandler.ArchiveMedicalDevice)
			}

			// Coffre-Fort GED Pharmaceutique (Jalon 1.4 - BPF & ALCOA+)
			if gedHandler != nil {
				// Taxonomie en 5 branches maîtresses
				protected.With(api.RequirePermission(auditService, domain.PermDocRead)).Get("/api/v1/ged/taxonomy", gedHandler.GetTaxonomy)

				// Documents & Versions
				protected.With(api.RequirePermission(auditService, domain.PermDocRead)).Get("/api/v1/ged/documents", gedHandler.ListDocuments)
				protected.With(api.RequirePermission(auditService, domain.PermDocDraftMaker)).Post("/api/v1/ged/documents", gedHandler.CreateDocument)
				protected.With(api.RequirePermission(auditService, domain.PermDocRead)).Get("/api/v1/ged/documents/{id}", gedHandler.GetDocument)
				protected.With(api.RequirePermission(auditService, domain.PermDocRead)).Get("/api/v1/ged/documents/{id}/versions", gedHandler.GetDocumentVersions)
				protected.With(api.RequirePermission(auditService, domain.PermDocDraftMaker)).Post("/api/v1/ged/documents/{id}/versions", gedHandler.UploadNewVersion)
				protected.With(api.RequirePermission(auditService, domain.PermDocRead)).Get("/api/v1/ged/documents/{id}/download", gedHandler.DownloadVersion)

				// Cycle de vie & Verrouillage
				protected.With(api.RequirePermission(auditService, domain.PermDocLockCheckout)).Post("/api/v1/ged/documents/{id}/checkout", gedHandler.CheckOut)
				protected.With(api.RequirePermission(auditService, domain.PermDocLockCheckout)).Post("/api/v1/ged/documents/{id}/cancel-checkout", gedHandler.CancelCheckOut)
				protected.With(api.RequireAnyPermission(auditService, domain.PermDocDraftMaker, domain.PermDocCheckReview, domain.PermDocArchive)).Post("/api/v1/ged/documents/{id}/status", gedHandler.UpdateStatus)

				// Signature électronique officielle du Pharmacien Directeur Technique (PDT - 1.A.7)
				protected.With(api.RequirePermission(auditService, domain.PermDocApprovePDT)).Post("/api/v1/ged/documents/{id}/sign-dt", gedHandler.SignDocumentDT)
			}

			// Moteur de Dossiers ANPP & Checklists Dynamiques (Jalon 1.5 - Guide CTD Juillet 2025)
			if dossierHandler != nil {
				protected.With(api.RequirePermission(auditService, domain.PermDossierView)).Get("/api/v1/dossiers", dossierHandler.ListDossiers)
				protected.With(api.RequirePermission(auditService, domain.PermDossierCreate)).Post("/api/v1/dossiers", dossierHandler.CreateDossier)
				protected.With(api.RequirePermission(auditService, domain.PermDossierView)).Get("/api/v1/dossiers/{id}", dossierHandler.GetDossier)
				protected.With(api.RequirePermission(auditService, domain.PermDocDraftMaker)).Post("/api/v1/dossiers/{id}/items/{itemId}/attach", dossierHandler.AttachDocument)
				protected.With(api.RequirePermission(auditService, domain.PermDocDraftMaker)).Post("/api/v1/dossiers/{id}/items/{itemId}/detach", dossierHandler.DetachDocument)
				protected.With(api.RequireAnyPermission(auditService, domain.PermDocDraftMaker, domain.PermDocCheckReview)).Post("/api/v1/dossiers/{id}/items/{itemId}/status", dossierHandler.UpdateItemStatus)
				protected.With(api.RequirePermission(auditService, domain.PermDocApprovePDT)).Post("/api/v1/dossiers/{id}/sign-pdt", dossierHandler.SignPDT)
				protected.With(api.RequireAnyPermission(auditService, domain.PermDossierCreate, domain.PermDocCheckReview)).Post("/api/v1/dossiers/{id}/status", dossierHandler.UpdateDossierStatus)

				// Équipe & Habilitations de Sécurité du Dossier (Need-to-Know & Secret Défense Industrielle)
				protected.With(api.RequirePermission(auditService, domain.PermDossierView)).Get("/api/v1/dossiers/{id}/members", dossierHandler.ListMembers)
				protected.With(api.RequireAnyPermission(auditService, domain.PermDossierCreate, domain.PermUserManage)).Post("/api/v1/dossiers/{id}/members", dossierHandler.AssignMember)
				protected.With(api.RequireAnyPermission(auditService, domain.PermDossierCreate, domain.PermUserManage)).Delete("/api/v1/dossiers/{id}/members/{userId}", dossierHandler.RemoveMember)
			}

			// Studio de Compléments ANPP & Traçabilité des Réponses (Jalon 1.7)
			if deficiencyHandler != nil {
				// Liste et création de demandes de compléments au sein d'un dossier
				protected.With(api.RequirePermission(auditService, domain.PermDossierView)).Get("/api/v1/dossiers/{id}/deficiencies", deficiencyHandler.ListDossierDeficiencies)
				protected.With(api.RequireAnyPermission(auditService, domain.PermDossierCreate, domain.PermDocDraftMaker)).Post("/api/v1/dossiers/{id}/deficiencies", deficiencyHandler.CreateDossierDeficiency)

				// Pilotage unitaire, alertes de délai légal et visa PDT
				protected.With(api.RequirePermission(auditService, domain.PermDossierView)).Get("/api/v1/deficiencies/letters/{letterId}", deficiencyHandler.GetDeficiencyLetter)
				protected.With(api.RequirePermission(auditService, domain.PermDossierView)).Get("/api/v1/deficiencies/dashboard", deficiencyHandler.GetDeficiencyDashboard)
				protected.With(api.RequireAnyPermission(auditService, domain.PermDossierCreate, domain.PermDocDraftMaker, domain.PermDocCheckReview)).Put("/api/v1/deficiencies/tasks/{taskId}", deficiencyHandler.UpdateDeficiencyTask)
				protected.With(api.RequirePermission(auditService, domain.PermDocApprovePDT)).Post("/api/v1/deficiencies/tasks/{taskId}/validate-dt", deficiencyHandler.ValidateDeficiencyTaskDT)
			}
		})
	}

	// 4. Lancement du serveur HTTP avec arrêt gracieux
	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("[Serveur] Démarré avec succès sur http://localhost%s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Erreur d'écoute HTTP : %v", err)
		}
	}()

	// Capture des signaux d'arrêt (SIGINT, SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("[Serveur] Arrêt gracieux en cours...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERREUR] Échec de l'arrêt gracieux : %v", err)
	}
	log.Println("[Serveur] Arrêté proprement.")
}
