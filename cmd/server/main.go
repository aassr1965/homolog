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
	var pharmaHandler *api.PharmaHandler
	var gedHandler *api.GEDHandler
	var dossierHandler *api.DossierHandler

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

		sessionManager = auth.NewSessionManager(db.Pool, cfg.SessionExpiryMinutes, cfg.IdleTimeoutMinutes)
		auditService = audit.NewService(db.Pool)
		authHandler = api.NewAuthHandler(db.Pool, sessionManager, auditService)

		pharmaRepo := repository.NewPharmaRepository(db.Pool)
		pharmaHandler = api.NewPharmaHandler(pharmaRepo, auditService)

		gedRepo := repository.NewGEDRepository(db.Pool)
		gedHandler = api.NewGEDHandler(gedRepo, storageProvider, auditService, db.Pool)

		dossierRepo := repository.NewDossierRepository(db.Pool)
		dossierHandler = api.NewDossierHandler(dossierRepo, pharmaRepo, gedRepo, auditService, db.Pool)
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

	// Routes de diagnostic et d'état de santé
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		dbStatus := "connected"
		if db == nil {
			dbStatus = "disconnected"
		}
		jsonResp := fmt.Sprintf(`{"status":"ok","mode":"%s","database":"%s","time_utc":"%s","version":"0.7.0"}`,
			cfg.DeploymentMode, dbStatus, time.Now().UTC().Format(time.RFC3339))
		w.Write([]byte(jsonResp))
	})

	r.Get("/api/v1/system/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := fmt.Sprintf(`{
			"app": "Sanad Pharma",
			"version": "0.7.0",
			"deployment_mode": "%s",
			"org_slug": "%s",
			"org_name": "%s",
			"idle_timeout_min": %d,
			"max_upload_bytes": %d,
			"timestamp_utc": "%s"
		}`, cfg.DeploymentMode, cfg.OnPremOrgSlug, cfg.OnPremOrgName, cfg.IdleTimeoutMinutes, cfg.MaxUploadSizeBytes, time.Now().UTC().Format(time.RFC3339))
		w.Write([]byte(resp))
	})

	// Routes API Authentification, 2FA, Signature DT et Audit Trail
	if authHandler != nil {
		r.Post("/api/v1/auth/login", authHandler.Login)
		r.Post("/api/v1/auth/logout", authHandler.Logout)

		// Routes protégées par session HttpOnly
		r.Group(func(protected chi.Router) {
			protected.Use(api.AuthMiddleware(sessionManager))
			protected.Get("/api/v1/auth/me", authHandler.Me)
			protected.Post("/api/v1/auth/reauth-sign", authHandler.ReauthSign)
			protected.Get("/api/v1/audit/logs", authHandler.GetAuditLogs)

			// Référentiel Établissements, Sites & Produits (Jalon 1.3 - Guide ANPP 2025)
			if pharmaHandler != nil {
				// Établissements
				protected.Get("/api/v1/establishments", pharmaHandler.ListEstablishments)
				protected.Post("/api/v1/establishments", pharmaHandler.CreateEstablishment)
				protected.Get("/api/v1/establishments/{id}", pharmaHandler.GetEstablishment)
				protected.Put("/api/v1/establishments/{id}", pharmaHandler.UpdateEstablishment)

				// Sites intervenants (1.H.8)
				protected.Get("/api/v1/manufacturing-sites", pharmaHandler.ListManufacturingSites)
				protected.Post("/api/v1/manufacturing-sites", pharmaHandler.CreateManufacturingSite)
				protected.Get("/api/v1/manufacturing-sites/{id}", pharmaHandler.GetManufacturingSite)

				// Médicaments (Princeps, Générique importé, Générique local)
				protected.Get("/api/v1/medicines", pharmaHandler.ListMedicines)
				protected.Post("/api/v1/medicines", pharmaHandler.CreateMedicine)
				protected.Get("/api/v1/medicines/{id}", pharmaHandler.GetMedicine)

				// Dispositifs Médicaux (Décret 20-324 & Arrêté 10/05/2021)
				protected.Get("/api/v1/medical-devices", pharmaHandler.ListMedicalDevices)
				protected.Post("/api/v1/medical-devices", pharmaHandler.CreateMedicalDevice)
				protected.Get("/api/v1/medical-devices/{id}", pharmaHandler.GetMedicalDevice)
			}

			// Coffre-Fort GED Pharmaceutique (Jalon 1.4 - BPF & ALCOA+)
			if gedHandler != nil {
				// Taxonomie en 5 branches maîtresses
				protected.Get("/api/v1/ged/taxonomy", gedHandler.GetTaxonomy)

				// Documents & Versions
				protected.Get("/api/v1/ged/documents", gedHandler.ListDocuments)
				protected.Post("/api/v1/ged/documents", gedHandler.CreateDocument)
				protected.Get("/api/v1/ged/documents/{id}", gedHandler.GetDocument)
				protected.Get("/api/v1/ged/documents/{id}/versions", gedHandler.GetDocumentVersions)
				protected.Post("/api/v1/ged/documents/{id}/versions", gedHandler.UploadNewVersion)
				protected.Get("/api/v1/ged/documents/{id}/download", gedHandler.DownloadVersion)

				// Cycle de vie & Verrouillage
				protected.Post("/api/v1/ged/documents/{id}/checkout", gedHandler.CheckOut)
				protected.Post("/api/v1/ged/documents/{id}/cancel-checkout", gedHandler.CancelCheckOut)
				protected.Post("/api/v1/ged/documents/{id}/status", gedHandler.UpdateStatus)

				// Signature électronique officielle du Pharmacien Directeur Technique (PDT)
				protected.Post("/api/v1/ged/documents/{id}/sign-dt", gedHandler.SignDocumentDT)
			}

			// Moteur de Dossiers ANPP & Checklists Dynamiques (Jalon 1.5 - Guide CTD Juillet 2025)
			if dossierHandler != nil {
				protected.Get("/api/v1/dossiers", dossierHandler.ListDossiers)
				protected.Post("/api/v1/dossiers", dossierHandler.CreateDossier)
				protected.Get("/api/v1/dossiers/{id}", dossierHandler.GetDossier)
				protected.Post("/api/v1/dossiers/{id}/items/{itemId}/attach", dossierHandler.AttachDocument)
				protected.Post("/api/v1/dossiers/{id}/items/{itemId}/detach", dossierHandler.DetachDocument)
				protected.Post("/api/v1/dossiers/{id}/items/{itemId}/status", dossierHandler.UpdateItemStatus)
				protected.Post("/api/v1/dossiers/{id}/sign-pdt", dossierHandler.SignPDT)
				protected.Post("/api/v1/dossiers/{id}/status", dossierHandler.UpdateDossierStatus)
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
