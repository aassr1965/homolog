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

	"homolog/internal/config"
	"homolog/internal/database"
	"homolog/internal/storage"
	"homolog/web"
)

func main() {
	log.Println("==================================================================")
	log.Println("  HOMOLOG - Plateforme Réglementaire ANPP & GED Pharmaceutique   ")
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
	db, err = database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("[AVERTISSEMENT] Connexion PostgreSQL non disponible au démarrage : %v", err)
		log.Println("[INFO] Le serveur démarre en attente de la base de données (Setup / Diagnostics).")
	} else {
		defer db.Close()
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
			"app": "Homolog",
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
