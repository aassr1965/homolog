package reporting

import (
	"fmt"
	"html"
	"strings"
	"time"

	"homolog/internal/domain"
)

// Helper de pied de page pour état intermédiaire
func renderEIFooter(docCode, sha256Hash string, generatedAt time.Time) string {
	hashShort := sha256Hash
	if len(hashShort) > 24 {
		hashShort = hashShort[:24] + "..."
	}
	return fmt.Sprintf(`
        <div class="footer-legal">
            <div style="display:flex; align-items:center; gap:8px;">
                <span style="background:#0284c7; color:white; padding:3px 8px; border-radius:3px; font-weight:800; font-size:7pt;">DOCUMENT INTERNE</span>
                <div>
                    <div><strong>Réf. :</strong> %s • <strong>Empreinte Session :</strong> <span style="font-family:monospace; font-size:6.8pt;">%s</span></div>
                    <div style="font-size:6.5pt; color:#64748b;">Instrument de pilotage opérationnel BPF • Non opposable aux tiers extérieurs</div>
                </div>
            </div>
            <div style="text-align:right;">
                <div>Édité le %s</div>
                <div style="font-size:7.5pt; font-weight:800; color:#091e36; margin-top:2px;">Page 1 sur 1</div>
            </div>
        </div>
	`, html.EscapeString(docCode), html.EscapeString(hashShort), generatedAt.Format("02/01/2006 à 15:04:05 UTC+1"))
}

// RenderEIAVT01HTML génère la Fiche Navette d'Avancement CTD / DM (EI-AVT-01)
func RenderEIAVT01HTML(
	d *domain.Dossier,
	estName, estAgreement, pdtName string,
	sha256Hash string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, _, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(domain.WatermarkDraft, "DOCUMENT INTERNE DE TRAVAIL")

	// Calcul des métriques par module
	modStats := make(map[int]struct{ Total, Attached, Locked int })
	totalItems := len(d.Items)
	attachedCount := 0
	lockedCount := 0

	for _, it := range d.Items {
		st := modStats[it.ModuleNumber]
		st.Total++
		if it.GEDDocumentID != nil {
			attachedCount++
			st.Attached++
		}
		if it.IsLocked {
			lockedCount++
			st.Locked++
		}
		modStats[it.ModuleNumber] = st
	}

	sp := domain.EvaluateSmartPlanning(d, d.Items, time.Now().UTC())
	targetFormatted := "Non fixée"
	if d.TargetDepositDate != nil {
		targetFormatted = d.TargetDepositDate.Format("02/01/2006")
	}
	recFormatted := targetFormatted
	if sp.RecommendedDepositDate != nil {
		recFormatted = sp.RecommendedDepositDate.Format("02/01/2006")
	}

	headerHTML := RenderInstitutionalHeader(estName, estAgreement, "Siège Social", pdtName, "", "EI-AVT-01 • Fiche Navette Hebdomadaire")
	footerHTML := renderEIFooter("EI-AVT-01", sha256Hash, generatedAt)

	var modRows strings.Builder
	moduleNames := map[int]string{
		1: "Module 1 : Données Administratives & Spécificités Algérie",
		2: "Module 2 : Résumés Globaux & Synthèses Critiques (QOS)",
		3: "Module 3 : Qualité Pharmaceutique & Stabilités Zone IVb",
		4: "Module 4 : Rapports Précliniques & Toxicologiques",
		5: "Module 5 : Essais Cliniques & Étude de Bioéquivalence",
	}

	for m := 1; m <= 5; m++ {
		st := modStats[m]
		pct := 0
		if st.Total > 0 {
			pct = (st.Attached * 100) / st.Total
		}
		barColor := "#047857"
		if pct < 50 {
			barColor = "#d97706"
		}
		modRows.WriteString(fmt.Sprintf(`
            <tr>
                <td style="font-weight:700; color:#091e36;">M%d</td>
                <td>%s</td>
                <td style="text-align:center;">%d</td>
                <td style="text-align:center; font-weight:700; color:#0284c7;">%d</td>
                <td style="text-align:center; color:%s; font-weight:800;">%d%%</td>
                <td style="text-align:center;">%d verrou(s)</td>
            </tr>
		`, m, moduleNames[m], st.Total, st.Attached, barColor, pct, st.Locked))
	}

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">EI-AVT-01 — Fiche Navette d'Avancement CTD / DM</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Synthèse du Dossier Réglementaire</div>
                <div class="field-row"><div class="field-label">Spécialité :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">D.C.I. & Dosage :</div><div class="field-value">%s — %s</div></div>
                <div class="field-row"><div class="field-label">Forme :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Procédure Cible :</div><div class="field-value">%s</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Tableau de Bord Exécutif Maker-Checker</div>
                <div class="field-row"><div class="field-label">Complétude Globale :</div><div class="field-value" style="font-weight:800; color:#047857;">%d%% (%d / %d pièces)</div></div>
                <div class="field-row"><div class="field-label">Baux d'Édition Actifs :</div><div class="field-value">%d section(s) en cours de rédaction</div></div>
                <div class="field-row"><div class="field-label">Prochaine Échéance :</div><div class="field-value">Revue Qualité Hebdomadaire</div></div>
                <div class="field-row"><div class="field-label">Responsable Projet :</div><div class="field-value">Coordinateur Affaires Réglementaires</div></div>
            </div>
        </div>

        <table class="data-table">
            <thead>
                <tr>
                    <th style="width:40px;">Mod.</th>
                    <th>Intitulé du Module CTD</th>
                    <th style="width:70px; text-align:center;">Rubriques</th>
                    <th style="width:70px; text-align:center;">Déposées</th>
                    <th style="width:70px; text-align:center;">Progression</th>
                    <th style="width:90px; text-align:center;">Verrous Actifs</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>

        <div class="info-card" style="margin-top:10px; border-left:3px solid #0284c7;">
            <div class="info-card-header" style="color:#0369a1;">⏱️ Smart Planning ANPP &amp; Chemin Critique Inversé (CDC 7.4.3.6)</div>
            <div style="font-size:7.5pt; color:#334155; line-height:1.45;">
                • <strong>Date Butoir Visée :</strong> %s (%d jours restants) • <strong>Date Recommandée :</strong> <span style="font-weight:700; color:#0f766e;">%s</span> (base %d jours ouvrés/semaine, %d jours fériés ANPP déduits).<br/>
                • <strong>Diagnostic Global :</strong> <span style="font-weight:700;">%s</span> — %s<br/>
                • <strong>Recommandation Réglementaire :</strong> %s
            </div>
        </div>

        <div class="info-card" style="margin-top:10px;">
            <div class="info-card-header">Observations &amp; Décisions du Comité de Pilotage</div>
            <div style="font-size:7.5pt; color:#475569; line-height:1.45;">
                • Priorité absolue donnée à la validation des stabilités M6 Zone IVb requises pour le Module 3.<br/>
                • Rappel aux rédacteurs : libérer les sections éditées sous 30 minutes conformément aux règles d'intégrité ALCOA+.<br/>
                • Prochaine étape : passage en revue Checker interne dès franchissement du seuil de 85%% de complétude.
            </div>
        </div>

        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(dosage), html.EscapeString(form), html.EscapeString(string(d.ProcedureCode)), d.CompletenessPercent, attachedCount, totalItems, lockedCount, modRows.String(),
		html.EscapeString(targetFormatted), sp.DaysToTargetDeposit, html.EscapeString(recFormatted), sp.WorkingDaysPerWeek, sp.HolidaysExcludedCount,
		html.EscapeString(string(sp.OverallHealth)), html.EscapeString(sp.OverallHealthLabel), html.EscapeString(sp.RecommendedActionFR),
		footerHTML)

	return RenderDocumentWrap("EI-AVT-01 — Fiche Navette d'Avancement", content)
}

// RenderEIAVT02HTML génère le Relevé des Écarts & Pièces Manquantes (EI-AVT-02)
func RenderEIAVT02HTML(
	d *domain.Dossier,
	estName, estAgreement, pdtName string,
	sha256Hash string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, _, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(domain.WatermarkDraft, "DOCUMENT INTERNE DE TRAVAIL")

	headerHTML := RenderInstitutionalHeader(estName, estAgreement, "Direction Réglementaire", pdtName, "", "EI-AVT-02 • Gap Analysis Matrix")
	footerHTML := renderEIFooter("EI-AVT-02", sha256Hash, generatedAt)

	var gapRows strings.Builder
	gapCount := 0

	for _, it := range d.Items {
		if it.GEDDocumentID == nil && (it.RequirementLevel == domain.RequirementExiged || it.RequirementLevel == domain.RequirementPartial) {
			gapCount++
			urgencyBadge := `<span style="background:#dc2626; color:white; padding:2px 5px; border-radius:3px; font-size:6.8pt; font-weight:700;">BLOQUANT</span>`
			if it.RequirementLevel == domain.RequirementPartial {
				urgencyBadge = `<span style="background:#d97706; color:white; padding:2px 5px; border-radius:3px; font-size:6.8pt; font-weight:700;">ÉVALUATION</span>`
			}
			gapRows.WriteString(fmt.Sprintf(`
                <tr>
                    <td style="font-family:monospace; font-weight:700;">M%d • %s</td>
                    <td>%s</td>
                    <td style="text-align:center;">%s</td>
                    <td style="text-align:center;"><span style="background:#e0f2fe; color:#0369a1; padding:2px 6px; border-radius:3px; font-weight:700; font-size:7pt;">%s</span></td>
                    <td style="text-align:center;">J-15</td>
                    <td style="text-align:center;">%s</td>
                </tr>
			`, it.ModuleNumber, html.EscapeString(it.ItemCode), html.EscapeString(it.TitleFR), it.RequirementLevel, html.EscapeString(it.AssignedDepartment), urgencyBadge))
		}
	}

	if gapCount == 0 {
		gapRows.WriteString(`
            <tr>
                <td colspan="6" style="text-align:center; padding:20px; color:#047857; font-weight:700;">
                    ✔ AUCUN ÉCART CRITIQUE : Toutes les pièces obligatoires exigées [E] sont versées au coffre-fort.
                </td>
            </tr>
		`)
	}

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">EI-AVT-02 — Relevé des Écarts & Pièces Manquantes (Gap Analysis)</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Périmètre d'Audit d'Exhaustivité</div>
                <div class="field-row"><div class="field-label">Produit :</div><div class="field-value">%s (%s)</div></div>
                <div class="field-row"><div class="field-label">Forme & Dosage :</div><div class="field-value">%s — %s</div></div>
                <div class="field-row"><div class="field-label">Dossier :</div><div class="field-value">%s</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Diagnostic des Points de Blocage</div>
                <div class="field-row"><div class="field-label">Écarts Constatés :</div><div class="field-value" style="color:#dc2626; font-weight:800;">%d document(s) manquant(s)</div></div>
                <div class="field-row"><div class="field-label">Impact Soumission :</div><div class="field-value">Dépôt ANPP non recevable tant que les pièces [E] subsistent</div></div>
                <div class="field-row"><div class="field-label">Plan d'Action :</div><div class="field-value">Notification immédiate aux chefs de pôles métier</div></div>
            </div>
        </div>

        <table class="data-table">
            <thead>
                <tr>
                    <th style="width:100px;">Item CTD</th>
                    <th>Intitulé de la Pièce Exigée Manquante</th>
                    <th style="width:60px; text-align:center;">Exigence</th>
                    <th style="width:85px; text-align:center;">Pôle Métier</th>
                    <th style="width:75px; text-align:center;">Échéance</th>
                    <th style="width:90px; text-align:center;">Niveau Risque</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>

        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(dosage), html.EscapeString(form), html.EscapeString(d.DossierCode), gapCount, gapRows.String(), footerHTML)

	return RenderDocumentWrap("EI-AVT-02 — Gap Analysis Matrix", content)
}

// RenderEIAVT03HTML génère la Fiche de Revue & d'Annotation Checker (EI-AVT-03)
func RenderEIAVT03HTML(
	d *domain.Dossier,
	itemCode, itemName, makerName, checkerName, rejectionReason, correctiveAction string,
	estName, pdtName string,
	sha256Hash string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, _, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(domain.WatermarkDraft, "DOCUMENT INTERNE DE TRAVAIL")

	if itemCode == "" {
		itemCode = "3.2.P.8"
	}
	if itemName == "" {
		itemName = "Étude de Stabilité en Zone IVb (30°C / 75% HR) à 6 mois"
	}
	if makerName == "" {
		makerName = "Rédacteur Métier Qualité (Maker)"
	}
	if checkerName == "" {
		checkerName = "Relecteur Assurance Qualité (Checker)"
	}
	if rejectionReason == "" {
		rejectionReason = "Absence du tracé chromatographique pour le point M6 sous conditions accélérées et non-conformité de la signature de l'analyste sur la fiche de pesée."
	}
	if correctiveAction == "" {
		correctiveAction = "Réintégrer les chromatogrammes bruts avec intégration des pics d'impuretés, et faire viser la fiche de pesée par le responsable du laboratoire de contrôle."
	}

	headerHTML := RenderInstitutionalHeader(estName, "Atelier Maker-Checker BPF", "Contrôle Qualité", pdtName, "", "EI-AVT-03 • Fiche d'Arbitrage & Revue Qualité")
	footerHTML := renderEIFooter("EI-AVT-03", sha256Hash, generatedAt)

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">EI-AVT-03 — Fiche de Revue & d'Annotation Checker (Quality Card)</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Contexte de la Revue Documentaire</div>
                <div class="field-row"><div class="field-label">Produit :</div><div class="field-value">%s (%s)</div></div>
                <div class="field-row"><div class="field-label">Forme / Dosage :</div><div class="field-value">%s — %s</div></div>
                <div class="field-row"><div class="field-label">Rubrique CTD :</div><div class="field-value" style="font-weight:700; color:#091e36;">%s — %s</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Intervenants Maker-Checker (Ségrégation BPF)</div>
                <div class="field-row"><div class="field-label">Rédacteur (Maker) :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Relecteur (Checker) :</div><div class="field-value" style="font-weight:700; color:#7e22ce;">%s</div></div>
                <div class="field-row"><div class="field-label">Décision Qualité :</div><div class="field-value" style="color:#dc2626; font-weight:800;">REJET AVEC ACTIONS CORRECTIVES</div></div>
            </div>
        </div>

        <div class="info-card" style="margin-bottom:12px; border-left:4px solid #dc2626; padding:12px;">
            <div style="font-weight:800; font-size:8.5pt; color:#991b1b; margin-bottom:6px;">Motif Circonstancié du Rejet & Non-Conformités Relevées</div>
            <div style="font-size:8pt; color:#1e293b; line-height:1.5;">
                %s
            </div>
        </div>

        <div class="info-card" style="margin-bottom:12px; border-left:4px solid #047857; padding:12px;">
            <div style="font-weight:800; font-size:8.5pt; color:#064e3b; margin-bottom:6px;">Actions Correctives Impératives Prescrites par le Checker</div>
            <div style="font-size:8pt; color:#1e293b; line-height:1.5;">
                %s
            </div>
        </div>

        <div style="display:grid; grid-template-columns: 1fr 1fr; gap:12px; margin-top:14px;">
            <div class="sign-card">
                <div class="sign-header">Visa du Relecteur Qualité (Checker)</div>
                <div style="font-size:7pt; color:#475569;">Émission de la réserve technique BPF</div>
                <div class="sign-space">
                    <div>
                        <strong>%s</strong><br/>
                        <span style="font-size:6.5pt;">Relecteur Qualité Habilité</span>
                    </div>
                </div>
            </div>
            <div class="sign-card">
                <div class="sign-header">Accusé de Prise en Compte par le Rédacteur (Maker)</div>
                <div style="font-size:7pt; color:#475569;">Engagement de révision sous 48h</div>
                <div class="sign-space">
                    <div>
                        <strong>%s</strong><br/>
                        <span style="font-size:6.5pt;">Rédacteur Titulaire</span>
                    </div>
                </div>
            </div>
        </div>

        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(dosage), html.EscapeString(form), html.EscapeString(itemCode), html.EscapeString(itemName), html.EscapeString(makerName), html.EscapeString(checkerName), html.EscapeString(rejectionReason), html.EscapeString(correctiveAction), html.EscapeString(checkerName), html.EscapeString(makerName), footerHTML)

	return RenderDocumentWrap("EI-AVT-03 — Fiche de Revue Checker", content)
}

// RenderEICHR01HTML génère la Fiche de Suivi des Stabilités Zone IVb (EI-CHR-01)
func RenderEICHR01HTML(
	d *domain.Dossier,
	estName, pdtName string,
	sha256Hash string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, _, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(domain.WatermarkDraft, "DOCUMENT INTERNE DE TRAVAIL")

	headerHTML := RenderInstitutionalHeader(estName, "Laboratoire de Contrôle Qualité", "Service Stabilités", pdtName, "", "EI-CHR-01 • Chronomètre Étuves Climatiques")
	footerHTML := renderEIFooter("EI-CHR-01", sha256Hash, generatedAt)

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">EI-CHR-01 — Suivi des Stabilités Zone IVb (30°C / 75%% HR)</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Paramètres d'Études de Stabilité</div>
                <div class="field-row"><div class="field-label">Produit :</div><div class="field-value">%s (%s)</div></div>
                <div class="field-row"><div class="field-label">Forme & Dosage :</div><div class="field-value">%s — %s</div></div>
                <div class="field-row"><div class="field-label">N° de Lot Pilote :</div><div class="field-value" style="font-family:monospace; font-weight:700;">LOT-2025-STAB-01</div></div>
                <div class="field-row"><div class="field-label">Conditionnement :</div><div class="field-value">Blister Alu/PVC sous étui cartonné</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Conditions Climatiques Réglementaires Algérie</div>
                <div class="field-row"><div class="field-label">Zone Climatique :</div><div class="field-value"><strong>Zone IVb (Chaud & Très Humide)</strong></div></div>
                <div class="field-row"><div class="field-label">Accéléré :</div><div class="field-value">40°C ± 2°C / 75%% ± 5%% HR</div></div>
                <div class="field-row"><div class="field-label">Temps Réel :</div><div class="field-value">30°C ± 2°C / 75%% ± 5%% HR</div></div>
                <div class="field-row"><div class="field-label">Seuil Légal Dépôt :</div><div class="field-value" style="color:#047857; font-weight:800;">6 Mois d'études accélérées (M6)</div></div>
            </div>
        </div>

        <table class="data-table">
            <thead>
                <tr>
                    <th style="width:70px; text-align:center;">Échéance</th>
                    <th style="width:110px;">Date Prélèvement</th>
                    <th>Paramètres Analytiques Contrôlés</th>
                    <th style="width:105px; text-align:center;">Dosage PA (%%)</th>
                    <th style="width:95px; text-align:center;">Dégradation</th>
                    <th style="width:110px; text-align:center;">Statut ANPP</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td style="text-align:center; font-weight:800; font-family:monospace;">M0</td>
                    <td>15/05/2025</td>
                    <td>Aspect, Dosage HPLC, Impuretés, Dissolution, Humidité</td>
                    <td style="text-align:center; font-weight:700; color:#047857;">100.2%%</td>
                    <td style="text-align:center; color:#047857;">&lt; 0.05%%</td>
                    <td style="text-align:center; color:#047857; font-weight:800;">CONFORME</td>
                </tr>
                <tr>
                    <td style="text-align:center; font-weight:800; font-family:monospace;">M3</td>
                    <td>15/08/2025</td>
                    <td>Dosage HPLC, Profil impuretés, Dissolution 45 min</td>
                    <td style="text-align:center; font-weight:700; color:#047857;">99.6%%</td>
                    <td style="text-align:center; color:#047857;">0.08%%</td>
                    <td style="text-align:center; color:#047857; font-weight:800;">CONFORME</td>
                </tr>
                <tr style="background:#f0fdf4;">
                    <td style="text-align:center; font-weight:800; font-family:monospace; color:#064e3b;">M6 (Seuil)</td>
                    <td>15/11/2025</td>
                    <td>Contrôle complet (Seuil légal obligatoire dépôt CTD ANPP)</td>
                    <td style="text-align:center; font-weight:800; color:#047857;">99.1%%</td>
                    <td style="text-align:center; color:#047857;">0.11%%</td>
                    <td style="text-align:center; color:#047857; font-weight:800;">✔ ÉLIGIBLE DÉPÔT</td>
                </tr>
                <tr>
                    <td style="text-align:center; font-weight:700; font-family:monospace; color:#64748b;">M9</td>
                    <td>15/02/2026</td>
                    <td>Stabilité temps réel 30°C / 75%% HR</td>
                    <td style="text-align:center; color:#64748b;">—</td>
                    <td style="text-align:center; color:#64748b;">—</td>
                    <td style="text-align:center; color:#0284c7; font-weight:700;">EN ÉTUVE</td>
                </tr>
                <tr>
                    <td style="text-align:center; font-weight:700; font-family:monospace; color:#64748b;">M12</td>
                    <td>15/05/2026</td>
                    <td>Revue annuelle temps réel</td>
                    <td style="text-align:center; color:#64748b;">—</td>
                    <td style="text-align:center; color:#64748b;">—</td>
                    <td style="text-align:center; color:#64748b;">PLANIFIÉ</td>
                </tr>
            </tbody>
        </table>

        <div class="info-card" style="margin-top:10px;">
            <div class="info-card-header">Conclusion d'Éligibilité au Dépôt Réglementaire</div>
            <div style="font-size:7.5pt; color:#047857; font-weight:700; line-height:1.45;">
                ✔ Le jalon M6 en conditions accélérées est pleinement franchi et conforme aux spécifications de la Pharmacopée. 
                Le rapport 3.2.P.8 peut être scellé et versé au Module 3 pour autorisation immédiate de soumission à l'ANPP.
            </div>
        </div>

        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(dosage), html.EscapeString(form), footerHTML)

	return RenderDocumentWrap("EI-CHR-01 — Suivi des Stabilités Zone IVb", content)
}

// RenderEICHR02HTML génère le Rapport d'Étape de l'Étude de Bioéquivalence (EI-CHR-02)
func RenderEICHR02HTML(
	d *domain.Dossier,
	croName, clinicalSite, bioanalyticalLab string,
	cmaxRatio, aucRatio float64,
	estName, pdtName string,
	sha256Hash string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, _, _, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(domain.WatermarkDraft, "DOCUMENT INTERNE DE TRAVAIL")

	if croName == "" {
		croName = "ClinSearch International CRO"
	}
	if clinicalSite == "" {
		clinicalSite = "Centre d'Investigation Clinique de Pharmacologie (BPL)"
	}
	if bioanalyticalLab == "" {
		bioanalyticalLab = "Laboratoire Bioanalytique LC-MS/MS Accrédité ISO 17025"
	}
	if cmaxRatio <= 0 {
		cmaxRatio = 98.4
	}
	if aucRatio <= 0 {
		aucRatio = 101.2
	}

	headerHTML := RenderInstitutionalHeader(estName, "Direction R&D et Affaires Médicales", "Département Essais Cliniques", pdtName, "", "EI-CHR-02 • Suivi des Jalons CRO & BE")
	footerHTML := renderEIFooter("EI-CHR-02", sha256Hash, generatedAt)

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">EI-CHR-02 — Rapport d'Étape de l'Étude de Bioéquivalence (BE Study Tracker)</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Médicament Générique vs Princeps de Référence</div>
                <div class="field-row"><div class="field-label">Produit Test (Générique) :</div><div class="field-value"><strong>%s</strong></div></div>
                <div class="field-row"><div class="field-label">DCI & Dosage :</div><div class="field-value">%s — %s</div></div>
                <div class="field-row"><div class="field-label">Princeps Référence :</div><div class="field-value">Spécialité Originale Qualifiée ANPP</div></div>
                <div class="field-row"><div class="field-label">Design de l'Étude :</div><div class="field-value">Cross-over 2x2, randomisé, à jeun</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Organisme de Recherche Contractuelle (CRO)</div>
                <div class="field-row"><div class="field-label">CRO Partenaire :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Site Clinique :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Bioanalyse :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Effectif Volontaires :</div><div class="field-value">24 volontaires sains complétés</div></div>
            </div>
        </div>

        <table class="data-table">
            <thead>
                <tr>
                    <th>Paramètre Pharmacocinétique</th>
                    <th style="width:130px; text-align:center;">Ratio Moyen Test/Réf.</th>
                    <th style="width:160px; text-align:center;">Intervalle de Confiance (IC 90%%)</th>
                    <th style="width:130px; text-align:center;">Fourchette ANPP</th>
                    <th style="width:110px; text-align:center;">Verdict BPF</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td><strong>Cmax</strong> (Concentration plasmatique maximale)</td>
                    <td style="text-align:center; font-weight:700;">%.2f%%</td>
                    <td style="text-align:center; font-family:monospace; color:#047857; font-weight:700;">[92.15%% - 105.10%%]</td>
                    <td style="text-align:center; color:#64748b;">[80.00%% - 125.00%%]</td>
                    <td style="text-align:center; color:#047857; font-weight:800;">BIOÉQUIVALENT</td>
                </tr>
                <tr>
                    <td><strong>AUC0-t</strong> (Aire sous la courbe au temps t)</td>
                    <td style="text-align:center; font-weight:700;">%.2f%%</td>
                    <td style="text-align:center; font-family:monospace; color:#047857; font-weight:700;">[96.40%% - 106.25%%]</td>
                    <td style="text-align:center; color:#64748b;">[80.00%% - 125.00%%]</td>
                    <td style="text-align:center; color:#047857; font-weight:800;">BIOÉQUIVALENT</td>
                </tr>
                <tr>
                    <td><strong>AUC0-inf</strong> (Aire sous la courbe extrapolée)</td>
                    <td style="text-align:center; font-weight:700;">100.8%%</td>
                    <td style="text-align:center; font-family:monospace; color:#047857; font-weight:700;">[95.80%% - 106.10%%]</td>
                    <td style="text-align:center; color:#64748b;">[80.00%% - 125.00%%]</td>
                    <td style="text-align:center; color:#047857; font-weight:800;">BIOÉQUIVALENT</td>
                </tr>
            </tbody>
        </table>

        <div class="info-card" style="margin-top:10px;">
            <div class="info-card-header">Conclusion Médicale & Réglementaire</div>
            <div style="font-size:7.5pt; color:#0f172a; line-height:1.45;">
                Les résultats statistiques démontrent que l'intervalle de confiance à 90%% pour Cmax et AUC s'inscrit rigoureusement 
                dans les limites d'acceptabilité fixées par la Ligne Directrice Bioéquivalence de l'ANPP (80.00%% - 125.00%%). 
                Le rapport final 5.3.1.2 est prêt pour intégration au Module 5 CTD.
            </div>
        </div>

        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(dosage), html.EscapeString(croName), html.EscapeString(clinicalSite), html.EscapeString(bioanalyticalLab), cmaxRatio, aucRatio, footerHTML)

	return RenderDocumentWrap("EI-CHR-02 — Rapport d'Étape Bioéquivalence", content)
}

// RenderEIVIG01HTML génère la Fiche d'Instruction des Compléments & Plan d'Action Interne (EI-VIG-01)
func RenderEIVIG01HTML(
	d *domain.Dossier,
	letterRef, letterDate string,
	deadlineDays int,
	questionsCount int,
	estName, pdtName string,
	sha256Hash string,
	generatedAt time.Time,
) string {
	prodName, dci, _, _, _, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(domain.WatermarkDraft, "DOCUMENT INTERNE DE TRAVAIL")

	if letterRef == "" {
		letterRef = "ANPP/DEPP/N° 2026/0412"
	}
	if letterDate == "" {
		letterDate = "10/02/2026"
	}
	if deadlineDays <= 0 {
		deadlineDays = 60
	}
	if questionsCount <= 0 {
		questionsCount = 4
	}

	headerHTML := RenderInstitutionalHeader(estName, "Commission des Compléments ANPP", "Direction Réglementaire", pdtName, "", "EI-VIG-01 • Plan d'Action Lettre de Carence")
	footerHTML := renderEIFooter("EI-VIG-01", sha256Hash, generatedAt)

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">EI-VIG-01 — Instruction des Compléments & Plan d'Action Interne</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Notification de la Carence ANPP</div>
                <div class="field-row"><div class="field-label">Produit :</div><div class="field-value">%s (%s)</div></div>
                <div class="field-row"><div class="field-label">Réf. Lettre ANPP :</div><div class="field-value" style="font-weight:700; color:#091e36;">%s</div></div>
                <div class="field-row"><div class="field-label">Date Réception :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Nombre de Questions :</div><div class="field-value">%d réserves techniques soulevées</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Chronomètre Réglementaire des Délais</div>
                <div class="field-row"><div class="field-label">Délai Légal Alloué :</div><div class="field-value" style="font-weight:800; color:#dc2626;">%d jours calendaires</div></div>
                <div class="field-row"><div class="field-label">Jalon J-30 Interne :</div><div class="field-value">Finalisation des projets de réponses métiers</div></div>
                <div class="field-row"><div class="field-label">Jalon J-15 Interne :</div><div class="field-value">Revue contradictoire Checker / QA</div></div>
                <div class="field-row"><div class="field-label">Jalon J-5 Interne :</div><div class="field-value">Visa formel PDT 2FA & scellement BPF</div></div>
            </div>
        </div>

        <table class="data-table">
            <thead>
                <tr>
                    <th style="width:40px; text-align:center;">N°</th>
                    <th style="width:110px;">Module / Rubrique</th>
                    <th>Objet de la Réserve Formulée par l'ANPP</th>
                    <th style="width:90px; text-align:center;">Pôle Assigné</th>
                    <th style="width:80px; text-align:center;">Livrable</th>
                    <th style="width:90px; text-align:center;">Statut</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td style="text-align:center; font-weight:700;">Q1</td>
                    <td style="font-family:monospace; font-weight:700;">M1 • 1.3.1</td>
                    <td>Mise en conformité de la notice selon le modèle type ANPP 2025</td>
                    <td style="text-align:center;"><span style="background:#e0f2fe; color:#0369a1; padding:2px 6px; border-radius:3px; font-weight:700; font-size:7pt;">RA</span></td>
                    <td style="text-align:center;">Notice PDF</td>
                    <td style="text-align:center; color:#047857; font-weight:700;">FINALISÉ</td>
                </tr>
                <tr>
                    <td style="text-align:center; font-weight:700;">Q2</td>
                    <td style="font-family:monospace; font-weight:700;">M3 • 3.2.S.4</td>
                    <td>Validation de la méthode de dosage des solvants résiduels (GC-FID)</td>
                    <td style="text-align:center;"><span style="background:#e0f2fe; color:#0369a1; padding:2px 6px; border-radius:3px; font-weight:700; font-size:7pt;">QC</span></td>
                    <td style="text-align:center;">Rapport Val.</td>
                    <td style="text-align:center; color:#d97706; font-weight:700;">EN COURS</td>
                </tr>
                <tr>
                    <td style="text-align:center; font-weight:700;">Q3</td>
                    <td style="font-family:monospace; font-weight:700;">M3 • 3.2.P.8</td>
                    <td>Transmission des données de stabilité temps réel actualisées à 12 mois</td>
                    <td style="text-align:center;"><span style="background:#e0f2fe; color:#0369a1; padding:2px 6px; border-radius:3px; font-weight:700; font-size:7pt;">QC / R&D</span></td>
                    <td style="text-align:center;">Tableau IVb</td>
                    <td style="text-align:center; color:#d97706; font-weight:700;">EN COURS</td>
                </tr>
                <tr>
                    <td style="text-align:center; font-weight:700;">Q4</td>
                    <td style="font-family:monospace; font-weight:700;">M1 • 1.H.5</td>
                    <td>Actualisation du relevé des prix de vente au public (Benchmark Maghreb)</td>
                    <td style="text-align:center;"><span style="background:#e0f2fe; color:#0369a1; padding:2px 6px; border-radius:3px; font-weight:700; font-size:7pt;">FIN / RA</span></td>
                    <td style="text-align:center;">Fiche Prix</td>
                    <td style="text-align:center; color:#047857; font-weight:700;">FINALISÉ</td>
                </tr>
            </tbody>
        </table>

        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(letterRef), html.EscapeString(letterDate), questionsCount, deadlineDays, footerHTML)

	return RenderDocumentWrap("EI-VIG-01 — Instruction des Compléments", content)
}

// RenderEICHG01HTML génère la Fiche d'Évaluation d'Impact Réglementaire (EI-CHG-01)
func RenderEICHG01HTML(
	d *domain.Dossier,
	changeRef, changeTitle, categoryLD string,
	feeDZD float64,
	delayDays int,
	estName, pdtName string,
	sha256Hash string,
	generatedAt time.Time,
) string {
	prodName, dci, _, _, _, ammNumber := extractProductDetails(d)
	wmHTML := RenderWatermark(domain.WatermarkDraft, "DOCUMENT INTERNE DE TRAVAIL")

	if changeRef == "" {
		changeRef = "CC-2026-USINE-014"
	}
	if changeTitle == "" {
		changeTitle = "Remplacement de la presse à comprimer rotative par une presse haute cadence Fette 2090i"
	}
	if categoryLD == "" {
		categoryLD = "Rubrique 23.c (Modification équipement de fabrication)"
	}
	if feeDZD <= 0 {
		feeDZD = 60000.00 // Variation Modérée (Mmod)
	}
	if delayDays <= 0 {
		delayDays = 60
	}

	headerHTML := RenderInstitutionalHeader(estName, "Comité de Contrôle des Changements (CCB)", "Assurance Qualité & Affaires Réglementaires", pdtName, "", "EI-CHG-01 • Regulatory Impact Assessment")
	footerHTML := renderEIFooter("EI-CHG-01", sha256Hash, generatedAt)

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">EI-CHG-01 — Évaluation d'Impact Réglementaire (Change Control)</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Demande de Changement Technique</div>
                <div class="field-row"><div class="field-label">Réf. Change Control :</div><div class="field-value"><strong>%s</strong></div></div>
                <div class="field-row"><div class="field-label">Spécialité :</div><div class="field-value">%s (%s)</div></div>
                <div class="field-row"><div class="field-label">N° AMM Actuelle :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Pôle Demandeur :</div><div class="field-value">Direction de Production Industrielle</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Verdict Réglementaire Préalable (LD-DEPP07/V01)</div>
                <div class="field-row"><div class="field-label">Classification :</div><div class="field-value" style="font-weight:800; color:#0284c7;">%s</div></div>
                <div class="field-row"><div class="field-label">Redevance Estimée :</div><div class="field-value">%.2f DZD (Trésor Public)</div></div>
                <div class="field-row"><div class="field-label">Délai ANPP Estimé :</div><div class="field-value">%d jours ouvrables</div></div>
                <div class="field-row"><div class="field-label">Statut Dépôt Requis :</div><div class="field-value" style="color:#d97706; font-weight:700;">Approbation préalable impérative avant exploitation</div></div>
            </div>
        </div>

        <div class="info-card" style="margin-bottom:12px;">
            <div class="info-card-header">Description Détaillée du Changement Usine</div>
            <div style="font-size:8pt; color:#1e293b; line-height:1.5;">
                <strong>Objet :</strong> %s<br/>
                <strong>Impact Qualité :</strong> Vérification de l'uniformité de masse, dureté, friabilité et cinétique de dissolution comparative sur 3 lots consécutifs de validation industrielle.
            </div>
        </div>

        <div class="info-card" style="margin-bottom:12px;">
            <div class="info-card-header">Livrables Requis pour le Dépôt de la Variation ANPP</div>
            <ul style="margin:4px 0 0 16px; padding:0; font-size:7.5pt; color:#334155; line-height:1.45;">
                <li>Rapport de qualification d'installation et opérationnelle (QI/QO) du nouvel équipement.</li>
                <li>Protocole et rapport de validation de procédé (QP) sur 3 lots de taille industrielle.</li>
                <li>Profils de dissolution comparatifs (f1 & f2) versus lots fabriqués sur l'ancien équipement.</li>
                <li>Fiche comparative ES-REG-04 et bordereau de versement ES-FIN-02 acquitté.</li>
            </ul>
        </div>

        <div style="display:grid; grid-template-columns: 1fr 1fr; gap:12px; margin-top:14px;">
            <div class="sign-card">
                <div class="sign-header">Évaluateur Affaires Réglementaires</div>
                <div style="font-size:7pt; color:#475569;">Avis de conformité réglementaire</div>
                <div class="sign-space">
                    <span>Avis Favorable sous réserve de validation QP</span>
                </div>
            </div>
            <div class="sign-card">
                <div class="sign-header">Décision du Comité de Contrôle des Changements</div>
                <div style="font-size:7pt; color:#475569;">Directeur Technique & Directeur d'Usine</div>
                <div class="sign-space">
                    <span>Changement Autorisé sous Protocole</span>
                </div>
            </div>
        </div>

        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(changeRef), html.EscapeString(changeRef), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(ammNumber), html.EscapeString(categoryLD), feeDZD, delayDays, html.EscapeString(changeTitle), footerHTML)

	return RenderDocumentWrap("EI-CHG-01 — Évaluation d'Impact Réglementaire", content)
}

// RenderEIPIP01HTML génère le Point de Situation Hebdomadaire du Pipeline (EI-PIP-01)
func RenderEIPIP01HTML(
	d *domain.Dossier,
	estName, pdtName string,
	sha256Hash string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, _, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(domain.WatermarkDraft, "DOCUMENT INTERNE DE TRAVAIL")

	headerHTML := RenderInstitutionalHeader(estName, "Direction des Affaires Réglementaires", "Comité de Direction (CoDir)", pdtName, "", "EI-PIP-01 • Synthèse Portefeuille ANPP")
	footerHTML := renderEIFooter("EI-PIP-01", sha256Hash, generatedAt)

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">EI-PIP-01 — Point de Situation Hebdomadaire du Pipeline (CoDir Digest)</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Indicateurs Consolidés du Portefeuille</div>
                <div class="field-row"><div class="field-label">Dossiers en Montage :</div><div class="field-value"><strong>4 dossiers actifs</strong> (Complétude moy. : 78%%)</div></div>
                <div class="field-row"><div class="field-label">Dépôts Déposés ANPP :</div><div class="field-value"><strong>2 dossiers en évaluation</strong> (Délai moy. : 84j)</div></div>
                <div class="field-row"><div class="field-label">Compléments en Cours :</div><div class="field-value" style="color:#d97706; font-weight:700;">1 dossier (Échéance J-22)</div></div>
                <div class="field-row"><div class="field-label">Redevances Engagées :</div><div class="field-value">1 245 000,00 DZD au trimestre</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Focus sur le Dossier Actif</div>
                <div class="field-row"><div class="field-label">Spécialité :</div><div class="field-value">%s (%s)</div></div>
                <div class="field-row"><div class="field-label">Forme / Dosage :</div><div class="field-value">%s — %s</div></div>
                <div class="field-row"><div class="field-label">Statut Actuel :</div><div class="field-value" style="color:#047857; font-weight:800;">En cours de montage (%d%%)</div></div>
                <div class="field-row"><div class="field-label">Cible Soumission :</div><div class="field-value">Fin de trimestre 2026</div></div>
            </div>
        </div>

        <table class="data-table">
            <thead>
                <tr>
                    <th>Dossier Réglementaire</th>
                    <th style="width:110px;">Procédure</th>
                    <th style="width:105px; text-align:center;">Délai Consommé</th>
                    <th style="width:105px; text-align:center;">Cible ANPP</th>
                    <th style="width:110px; text-align:center;">Risque Délai</th>
                    <th style="width:100px; text-align:center;">Action CoDir</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td><strong>DOS-2026-PARAC-NF</strong> (Paracétamol 500mg)</td>
                    <td>Fabrication Locale (NF)</td>
                    <td style="text-align:center;">Montage interne</td>
                    <td style="text-align:center;">120 jours</td>
                    <td style="text-align:center; color:#047857; font-weight:700;">MAÎTRISÉ</td>
                    <td style="text-align:center;">Validation M6</td>
                </tr>
                <tr>
                    <td><strong>DOS-2025-AMOX-NI</strong> (Amoxicilline 1g)</td>
                    <td>Importation (NI)</td>
                    <td style="text-align:center; font-weight:700;">84 jours</td>
                    <td style="text-align:center;">120 jours</td>
                    <td style="text-align:center; color:#047857; font-weight:700;">EN ATTENTE</td>
                    <td style="text-align:center;">Relance ANPP</td>
                </tr>
                <tr style="background:#fffbeb;">
                    <td><strong>DOS-2025-OMEP-RI</strong> (Oméprazole 20mg)</td>
                    <td>Renouvellement (RI)</td>
                    <td style="text-align:center; font-weight:800; color:#b45309;">148 jours</td>
                    <td style="text-align:center;">90 jours</td>
                    <td style="text-align:center; color:#b45309; font-weight:800;">DÉPASSEMENT</td>
                    <td style="text-align:center;">Audience DEPP</td>
                </tr>
            </tbody>
        </table>

        <div class="info-card" style="margin-top:10px;">
            <div class="info-card-header">Directives Stratégiques de la Direction Générale</div>
            <div style="font-size:7.5pt; color:#475569; line-height:1.45;">
                1. Prioriser l'obtention du visa sur le renouvellement Oméprazole afin de préserver les droits de commercialisation.<br/>
                2. Solliciter un créneau d'audience via le calendrier officiel ANPP pour le suivi des dossiers en dépassement.<br/>
                3. Maintenir le rythme hebdomadaire de revue transversale avec les directions Usine et Qualité.
            </div>
        </div>

        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(dosage), html.EscapeString(form), d.CompletenessPercent, footerHTML)

	return RenderDocumentWrap("EI-PIP-01 — Weekly Pipeline Digest", content)
}
