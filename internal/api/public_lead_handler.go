package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PublicLeadHandler gère les soumissions Démo, Contact et Pré-Onboarding du portail SANAD Pharma.
type PublicLeadHandler struct {
	pool *pgxpool.Pool
}

// NewPublicLeadHandler instancie le gestionnaire des demandes publiques.
func NewPublicLeadHandler(pool *pgxpool.Pool) *PublicLeadHandler {
	return &PublicLeadHandler{pool: pool}
}

// PublicLeadRequest représente la charge utile d'une demande issue du portail SPA.
type PublicLeadRequest struct {
	LeadType           string `json:"lead_type"`
	FullName           string `json:"full_name"`
	Email              string `json:"email"`
	OrganizationName   string `json:"organization_name"`
	RoleTitle          string `json:"role_title"`
	Phone              string `json:"phone"`
	DeploymentInterest string `json:"deployment_interest"`
	Language           string `json:"language"`
	Message            string `json:"message"`
}

// SubmitLead valide et enregistre une demande de démonstration, contact ou pré-onboarding.
func (h *PublicLeadHandler) SubmitLead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req PublicLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête JSON invalide")
		return
	}

	req.LeadType = strings.ToLower(strings.TrimSpace(req.LeadType))
	if req.LeadType == "" {
		req.LeadType = "demo"
	}
	if req.LeadType != "demo" && req.LeadType != "contact" && req.LeadType != "onboarding" {
		writeJSONError(w, http.StatusBadRequest, "Type de demande invalide (demo, contact ou onboarding requis)")
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.OrganizationName = strings.TrimSpace(req.OrganizationName)
	req.RoleTitle = strings.TrimSpace(req.RoleTitle)
	req.Phone = strings.TrimSpace(req.Phone)
	req.DeploymentInterest = strings.ToLower(strings.TrimSpace(req.DeploymentInterest))
	req.Language = strings.ToLower(strings.TrimSpace(req.Language))
	req.Message = strings.TrimSpace(req.Message)

	if req.FullName == "" || len(req.FullName) < 2 {
		writeJSONError(w, http.StatusBadRequest, "Le nom complet est obligatoire")
		return
	}
	if req.Email == "" || !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		writeJSONError(w, http.StatusBadRequest, "Adresse email professionnelle valide obligatoire")
		return
	}
	if req.DeploymentInterest == "" {
		req.DeploymentInterest = "saas"
	}
	if req.Language != "fr" && req.Language != "ar" && req.Language != "en" {
		req.Language = "fr"
	}

	ip := getClientIP(r)
	ua := r.UserAgent()
	leadID := uuid.New()

	if h.pool != nil {
		_, err := h.pool.Exec(r.Context(), `
			INSERT INTO public_leads (
				id, lead_type, full_name, email, organization_name, role_title,
				phone, deployment_interest, language, message, ip_address, user_agent
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, leadID, req.LeadType, req.FullName, req.Email, req.OrganizationName, req.RoleTitle,
			req.Phone, req.DeploymentInterest, req.Language, req.Message, ip, ua)
		if err != nil {
			log.Printf("[PORTAL LEAD ERROR] %v", err)
			writeJSONError(w, http.StatusInternalServerError, "Erreur lors de l'enregistrement de votre demande")
			return
		}
	} else {
		log.Printf("[PORTAL LEAD] type=%s email=%s org=%s mode=%s", req.LeadType, req.Email, req.OrganizationName, req.DeploymentInterest)
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":    true,
		"lead_id":    leadID.String(),
		"lead_type":  req.LeadType,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"message":    "Demande enregistrée avec succès",
	})
}
