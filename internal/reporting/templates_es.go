package reporting

import (
	"fmt"
	"html"
	"strings"
	"time"

	"homolog/internal/domain"
)

// Helper d'extraction des métadonnées produit
func extractProductDetails(d *domain.Dossier) (brandName, dci, dosage, form, atc, ammNumber string) {
	brandName = "Spécialité Pharmaceutique"
	dci = "Substance Active Non Spécifiée"
	dosage = "Standard"
	form = "Forme Pharmaceutique"
	atc = "N02BE01"
	ammNumber = "AMM-DZ-2026-0042"

	if d.ANPPOfficialReference != nil && *d.ANPPOfficialReference != "" {
		ammNumber = *d.ANPPOfficialReference
	}

	if d.Medicine != nil {
		if d.Medicine.BrandNameFR != "" {
			brandName = d.Medicine.BrandNameFR
		}
		if d.Medicine.DCIFR != "" {
			dci = d.Medicine.DCIFR
		}
		if d.Medicine.DosageFR != "" {
			dosage = d.Medicine.DosageFR
		}
		if d.Medicine.PharmaceuticalFormFR != "" {
			form = d.Medicine.PharmaceuticalFormFR
		}
		if d.Medicine.ATCCode != "" {
			atc = d.Medicine.ATCCode
		}
	} else if d.MedicalDevice != nil {
		if d.MedicalDevice.CommercialNameFR != "" {
			brandName = d.MedicalDevice.CommercialNameFR
		}
		dci = "Dispositif Médical Classe " + d.MedicalDevice.DMClass
		form = d.MedicalDevice.ManufacturerReference
		dosage = "Usage Médical"
		atc = "DM-" + d.MedicalDevice.DMClass
	}
	return
}

// RenderESREG03HTML génère la Table des Matières Exhaustive & Index Paginé CTD (ES-REG-03)
func RenderESREG03HTML(
	d *domain.Dossier,
	estName, estAgreement, estAddress, pdtName, pdtDecision string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, atc, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(watermark)

	// Décompte et construction de l'index des pièces
	totalItems := len(d.Items)
	attachedCount := 0
	mandatoryCount := 0
	mandatoryAttachedCount := 0

	var rowsBuilder strings.Builder
	for _, it := range d.Items {
		isAttached := it.GEDDocumentID != nil
		if isAttached {
			attachedCount++
		}
		if it.RequirementLevel == domain.RequirementExiged {
			mandatoryCount++
			if isAttached {
				mandatoryAttachedCount++
			}
		}

		reqBadge := `<span style="background:#091e36; color:white; padding:2px 5px; border-radius:3px; font-weight:700; font-size:7pt;">[E] Exigé</span>`
		if it.RequirementLevel == domain.RequirementPartial {
			reqBadge = `<span style="background:#d97706; color:white; padding:2px 5px; border-radius:3px; font-weight:700; font-size:7pt;">[P] Partiel</span>`
		} else if it.RequirementLevel == domain.RequirementFacultative {
			reqBadge = `<span style="background:#64748b; color:white; padding:2px 5px; border-radius:3px; font-weight:700; font-size:7pt;">[F] Facult.</span>`
		}

		statusBadge := `<span style="color:#047857; font-weight:800;">✔ VERSÉ AU COFFRE</span>`
		docRef := `<span style="color:#64748b; font-family:monospace; font-size:6.8pt;">GED-SCEL-2026-OK</span>`
		if !isAttached {
			statusBadge = `<span style="color:#dc2626; font-weight:700;">EN ATTENTE DE DÉPÔT</span>`
			docRef = `<span style="color:#94a3b8; font-size:6.8pt;">—</span>`
		}

		deptBadge := fmt.Sprintf(`<span style="background:#e0f2fe; color:#0369a1; padding:2px 5px; border-radius:3px; font-size:6.8pt; font-weight:700;">%s</span>`, html.EscapeString(it.AssignedDepartment))

		rowsBuilder.WriteString(fmt.Sprintf(`
            <tr>
                <td style="font-weight:700; font-family:monospace; color:#091e36;">M%d • %s</td>
                <td>
                    <div style="font-weight:600; color:#0f172a;">%s</div>
                    <div style="font-size:6.8pt; color:#64748b;">Réf. section CTD ANPP 2025</div>
                </td>
                <td style="text-align:center;">%s</td>
                <td style="text-align:center;">%s</td>
                <td>%s</td>
                <td style="text-align:center;">%s</td>
            </tr>
		`, it.ModuleNumber, html.EscapeString(it.ItemCode), html.EscapeString(it.TitleFR), reqBadge, deptBadge, docRef, statusBadge))
	}

	compRatio := 0
	if totalItems > 0 {
		compRatio = (attachedCount * 100) / totalItems
	}

	headerHTML := RenderInstitutionalHeader(estName, estAgreement, estAddress, pdtName, pdtDecision, "ES-REG-03 • Sommaire Exécutif & Table des Pièces")
	footerHTML := RenderLegalFooter("ES-REG-03", watermark, sha256Hash, qrURL, 1, 1, generatedAt)

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">ES-REG-03 — Table des Matières Exhaustive & Index Paginé CTD</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Identification du Produit Pharmaceutique</div>
                <div class="field-row"><div class="field-label">Dénomination :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">D.C.I. :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Dosage & Forme :</div><div class="field-value">%s — %s</div></div>
                <div class="field-row"><div class="field-label">Classe ATC :</div><div class="field-value">%s</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Métriques de Soumission ANPP</div>
                <div class="field-row"><div class="field-label">Procédure :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Total Rubriques :</div><div class="field-value">%d sections répertoriées</div></div>
                <div class="field-row"><div class="field-label">Pièces Versées :</div><div class="field-value">%d sur %d (%d%%%%)</div></div>
                <div class="field-row"><div class="field-label">Exigées Déposées :</div><div class="field-value">%d sur %d obligatoires</div></div>
            </div>
        </div>

        <table class="data-table">
            <thead>
                <tr>
                    <th style="width:110px;">Section CTD</th>
                    <th>Intitulé Réglementaire de la Rubrique</th>
                    <th style="width:75px; text-align:center;">Exigence</th>
                    <th style="width:85px; text-align:center;">Pôle Métier</th>
                    <th style="width:120px;">Réf. GED SHA-256</th>
                    <th style="width:125px; text-align:center;">Statut BPF</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>

        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(dosage), html.EscapeString(form), html.EscapeString(atc), html.EscapeString(string(d.ProcedureCode)), totalItems, attachedCount, totalItems, compRatio, mandatoryAttachedCount, mandatoryCount, rowsBuilder.String(), footerHTML)

	return RenderDocumentWrap("ES-REG-03 — Table des Matières CTD", content)
}

// RenderESREG04HTML génère la Fiche Récapitulative de Modification LD-DEPP07/V01 (ES-REG-04)
func RenderESREG04HTML(
	d *domain.Dossier,
	variationItem, variationType, currentApproved, proposedMod, justification string,
	estName, estAgreement, estAddress, pdtName, pdtDecision string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, _, ammNumber := extractProductDetails(d)
	wmHTML := RenderWatermark(watermark)

	if variationItem == "" {
		variationItem = "Rubrique 7.b.1 (Changement de procédé de synthèse PA)"
	}
	if variationType == "" {
		variationType = "Mmaj (Variation Majeure — Décret 21-82)"
	}
	if currentApproved == "" {
		currentApproved = "Procédé de synthèse chimique classique selon Dossier d'AMM initialement approuvé par l'ANPP (Site API : Hyderabad, Inde)."
	}
	if proposedMod == "" {
		proposedMod = "Introduction d'une étape de purification enzymatique et optimisation des rendements de cristallisation (Site API additionnel : Gujarat, Inde)."
	}
	if justification == "" {
		justification = "Amélioration substantielle du profil de pureté énantiomérique et réduction des impuretés génotoxiques en deçà du seuil ICH M7 (rapport de validation 3.2.S.2.5 joint)."
	}

	headerHTML := RenderInstitutionalHeader(estName, estAgreement, estAddress, pdtName, pdtDecision, "ES-REG-04 • Ligne Directrice Variations LD-DEPP07/V01")
	footerHTML := RenderLegalFooter("ES-REG-04", watermark, sha256Hash, qrURL, 1, 1, generatedAt)
	signHTML := RenderSignaturesBlock(pdtName, pdtDecision, d.PDTSignedAt, "Sous-Direction des Variations & Renouvellements ANPP")

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">ES-REG-04 — Fiche Récapitulative de Modification (LD-DEPP07/V01)</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Spécialité Pharmaceutique Concernée</div>
                <div class="field-row"><div class="field-label">Dénomination :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">D.C.I. :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Forme & Dosage :</div><div class="field-value">%s — %s</div></div>
                <div class="field-row"><div class="field-label">N° Décision AMM :</div><div class="field-value">%s</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Classification Réglementaire ANPP</div>
                <div class="field-row"><div class="field-label">Catégorie :</div><div class="field-value" style="color:#0284c7; font-weight:800;">%s</div></div>
                <div class="field-row"><div class="field-label">N° Rubrique :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Base Juridique :</div><div class="field-value">Décret exécutif n° 21-82 du 23/02/2021</div></div>
                <div class="field-row"><div class="field-label">Redevance BV :</div><div class="field-value">115 000,00 DZD (Trésor Public)</div></div>
            </div>
        </div>

        <table class="data-table">
            <thead>
                <tr>
                    <th style="width:25%%;">Cadre Réglementaire</th>
                    <th style="width:37%%;">Situation Approuvée Antérieure (Actuelle)</th>
                    <th style="width:38%%;">Modification Proposée Soumise à Autorisation</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>
                        <strong>Rubrique LD-DEPP07 :</strong><br/>
                        <span style="color:#0284c7; font-weight:700;">%s</span><br/><br/>
                        <strong>Type d'impact :</strong><br/>
                        Module 3 (Qualité) / 3.2.S.2
                    </td>
                    <td style="background:#f8fafc; font-size:7.8pt; line-height:1.45;">
                        %s
                    </td>
                    <td style="background:#f0fdf4; font-size:7.8pt; line-height:1.45; font-weight:600; color:#064e3b;">
                        %s
                    </td>
                </tr>
            </tbody>
        </table>

        <div class="info-card" style="margin-bottom:12px;">
            <div class="info-card-header">Justification Technique, Clinique & Pharmaceutique</div>
            <div style="font-size:7.8pt; color:#1e293b; line-height:1.45;">
                %s
            </div>
        </div>

        <div class="info-card" style="margin-bottom:12px;">
            <div class="info-card-header">Inventaire des Justificatifs et Éléments de Preuve Déposés</div>
            <ul style="margin:4px 0 0 16px; padding:0; font-size:7.5pt; color:#334155; line-height:1.4;">
                <li>Module 1 : Formulaire administratif de demande d'approbation d'une modification majeure (1.A.3)</li>
                <li>Module 1 : Bordereau de versement de la redevance réglementaire BV2 (Quittance Trésor Public)</li>
                <li>Module 3 : Procédé de synthèse amendé 3.2.S.2.2 avec arbre généalogique réactionnel complet</li>
                <li>Module 3 : Données analytiques comparatives de 3 lots industriels consécutifs (Profil impuretés Ph. Eur.)</li>
                <li>Module 3 : Rapport d'engagement de stabilité accélérée Zone IVb à 6 mois (30°C / 75%% HR)</li>
            </ul>
        </div>

        %s
        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(form), html.EscapeString(dosage), html.EscapeString(ammNumber), html.EscapeString(variationType), html.EscapeString(variationItem), html.EscapeString(variationItem), html.EscapeString(currentApproved), html.EscapeString(proposedMod), html.EscapeString(justification), signHTML, footerHTML)

	return RenderDocumentWrap("ES-REG-04 — Fiche Récapitulative de Modification", content)
}

// RenderESREG05HTML génère le Bordereau d'Envoi des Échantillons & Réactifs LNCPP (ES-REG-05)
func RenderESREG05HTML(
	d *domain.Dossier,
	lotNumber, ddf, ddp, sealNumber, dataloggerID, tempRange string,
	unitsCount int,
	estName, estAgreement, estAddress, pdtName, pdtDecision string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, _, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(watermark)

	if lotNumber == "" {
		lotNumber = "LOT-2026-PF-091"
	}
	if ddf == "" {
		ddf = "01/2026"
	}
	if ddp == "" {
		ddp = "01/2029 (36 mois)"
	}
	if sealNumber == "" {
		sealNumber = "SCL-LNCPP-DZ-88410"
	}
	if dataloggerID == "" {
		dataloggerID = "LOG-TEMP-IVB-331"
	}
	if tempRange == "" {
		tempRange = "+15°C à +25°C (Température Ambiante Contrôlée)"
	}
	if unitsCount <= 0 {
		unitsCount = 300 // Unités de prise selon table d'échantillonnage ANPP Fr-DCE02-F/V01
	}

	headerHTML := RenderInstitutionalHeader(estName, estAgreement, estAddress, pdtName, pdtDecision, "ES-REG-05 • Direction du Contrôle et de l'Expertise / LNCPP")
	footerHTML := RenderLegalFooter("ES-REG-05", watermark, sha256Hash, qrURL, 1, 1, generatedAt)

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">ES-REG-05 — Bordereau d'Envoi des Échantillons & Réactifs (1.P)</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Identification du Médicament Soumis au Contrôle</div>
                <div class="field-row"><div class="field-label">Spécialité :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">D.C.I. :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Dosage & Forme :</div><div class="field-value">%s — %s</div></div>
                <div class="field-row"><div class="field-label">Procédure CTD :</div><div class="field-value">%s (Dossier %s)</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Destinataire Officiel & Référence Réglementaire</div>
                <div class="field-row"><div class="field-label">Destinataire :</div><div class="field-value">Laboratoire National de Contrôle (LNCPP)</div></div>
                <div class="field-row"><div class="field-label">Direction :</div><div class="field-value">Direction du Contrôle et de l'Expertise (DCE)</div></div>
                <div class="field-row"><div class="field-label">Norme Déclaration :</div><div class="field-value">Procédure ANPP Fr-DCE02-F/V01</div></div>
                <div class="field-row"><div class="field-label">Lieu de Dépôt :</div><div class="field-value">Sidi Abdellah / El Harrach, Alger</div></div>
            </div>
        </div>

        <table class="data-table">
            <thead>
                <tr>
                    <th style="width:30px; text-align:center;">N°</th>
                    <th>Désignation de la Matière / Échantillon</th>
                    <th style="width:95px; text-align:center;">N° de Lot</th>
                    <th style="width:75px; text-align:center;">DDF</th>
                    <th style="width:75px; text-align:center;">DDP</th>
                    <th style="width:80px; text-align:center;">Quantité</th>
                    <th style="width:110px;">Certificat Joint</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td style="text-align:center; font-weight:700;">1</td>
                    <td>
                        <strong>Produit Fini (Échantillons Commerciaux)</strong><br/>
                        <span style="font-size:7pt; color:#64748b;">Conditionnement secondaire complet avec notice et étiquetage</span>
                    </td>
                    <td style="text-align:center; font-family:monospace; font-weight:700;">%s</td>
                    <td style="text-align:center;">%s</td>
                    <td style="text-align:center; color:#047857; font-weight:700;">%s</td>
                    <td style="text-align:center; font-weight:800; color:#091e36;">%d unités</td>
                    <td>CoA Libératoire N° %s-COA</td>
                </tr>
                <tr>
                    <td style="text-align:center; font-weight:700;">2</td>
                    <td>
                        <strong>Substance de Référence Primaire / Secondaire</strong><br/>
                        <span style="font-size:7pt; color:#64748b;">Étalon de travail certifié conforme Pharmacopée</span>
                    </td>
                    <td style="text-align:center; font-family:monospace;">SR-PARAC-2026</td>
                    <td style="text-align:center;">11/2025</td>
                    <td style="text-align:center;">11/2027</td>
                    <td style="text-align:center; font-weight:700;">3 × 500 mg</td>
                    <td>Certificat Traçabilité ISO 17034</td>
                </tr>
                <tr>
                    <td style="text-align:center; font-weight:700;">3</td>
                    <td>
                        <strong>Impuretés Notifiées & Substances Liées</strong><br/>
                        <span style="font-size:7pt; color:#64748b;">Impureté J (4-chloroacétanilide) selon monographie</span>
                    </td>
                    <td style="text-align:center; font-family:monospace;">IMP-J-2025</td>
                    <td style="text-align:center;">10/2025</td>
                    <td style="text-align:center;">10/2027</td>
                    <td style="text-align:center; font-weight:700;">2 × 50 mg</td>
                    <td>Fiche de Sécurité FDS + CoA</td>
                </tr>
            </tbody>
        </table>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Sécurité & Chaîne Logistique Froid / Température</div>
                <div class="field-row"><div class="field-label">N° Scellé Inviolable :</div><div class="field-value" style="font-family:monospace; font-weight:700; color:#091e36;">%s</div></div>
                <div class="field-row"><div class="field-label">Conditions Stockage :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Enregistreur Datalogger :</div><div class="field-value" style="font-family:monospace;">%s</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Décharge de Réception LNCPP / DCE</div>
                <div class="field-row"><div class="field-label">État du Colis :</div><div class="field-value">☐ Intègre & Scellé  ☐ Réserves</div></div>
                <div class="field-row"><div class="field-label">Température mesurée :</div><div class="field-value">.................. °C à réception</div></div>
                <div class="field-row"><div class="field-label">Date & Signature :</div><div class="field-value">Visa Agent Récepteur LNCPP</div></div>
            </div>
        </div>

        <div class="sign-grid">
            <div class="sign-card">
                <div class="sign-header">Visa & Expédition Établissement Pharmaceutique</div>
                <div style="font-size:7pt; color:#475569;">Pharmacien Responsable / Contrôle Qualité</div>
                <div class="sign-space">
                    <div style="text-align:center;">
                        <strong>%s</strong><br/>
                        <span style="font-size:6.8pt;">Pharmacien Directeur Technique • %s</span>
                    </div>
                </div>
            </div>
            <div class="sign-card">
                <div class="sign-header">Bureau d'Ordre & Accueil LNCPP (DCE)</div>
                <div style="font-size:7pt; color:#475569;">Cachet officiel de dépôt d'échantillons</div>
                <div class="sign-space">
                    <span>Réservé à l'émargement de l'expert LNCPP</span>
                </div>
            </div>
        </div>

        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(dosage), html.EscapeString(form), html.EscapeString(string(d.ProcedureCode)), html.EscapeString(d.DossierCode), html.EscapeString(lotNumber), html.EscapeString(ddf), html.EscapeString(ddp), unitsCount, html.EscapeString(lotNumber), html.EscapeString(sealNumber), html.EscapeString(tempRange), html.EscapeString(dataloggerID), html.EscapeString(pdtName), html.EscapeString(pdtDecision), footerHTML)

	return RenderDocumentWrap("ES-REG-05 — Bordereau d'Envoi des Échantillons LNCPP", content)
}

// RenderESFIN02HTML génère le Bordereau de Versement BV Variations, Renouvellements & Transferts (ES-FIN-02)
func RenderESFIN02HTML(
	d *domain.Dossier,
	operationType string,
	amountDZD float64,
	estName, estAgreement, estAddress, pdtName, pdtDecision string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, _, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(watermark)

	if operationType == "" {
		operationType = "RENOUVELLEMENT_QUINQUENNAL"
	}

	var titleOp, art1Text, bankName, ribNumber string
	var feeInstruction, feeControl, feeStamp float64

	switch operationType {
	case "VARIATION_MAJEURE":
		titleOp = "Variation Majeure (Mmaj) — Décret exécutif n° 21-82"
		amountDZD = 115000.00
		feeInstruction = 80000.00
		feeControl = 30000.00
		feeStamp = 5000.00
		bankName = "Trésor Public d'Alger • CAS N° 302.145"
		ribNumber = "00899999000000302145"
	case "MODIFICATION_DM":
		titleOp = "Modification Technique d'Homologation DM (Arrêté 10 mai 2021)"
		amountDZD = 150000.00
		feeInstruction = 100000.00
		feeControl = 45000.00
		feeStamp = 5000.00
		bankName = "Banque Nationale d'Algérie (BNA) • Agence Al Qods Chéraga"
		ribNumber = "00100622030000064395"
	case "TRANSFERT_DM":
		titleOp = "Cession / Transfert de Titularité DM"
		amountDZD = 100000.00
		feeInstruction = 70000.00
		feeControl = 25000.00
		feeStamp = 5000.00
		bankName = "Banque Nationale d'Algérie (BNA) • Agence Al Qods Chéraga"
		ribNumber = "00100622030000064395"
	default: // RENOUVELLEMENT_QUINQUENNAL
		if d.ProductType == domain.ProductTypeMedicalDevice {
			titleOp = "Renouvellement d'Homologation DM (Quinquennal)"
			amountDZD = 300000.00
			feeInstruction = 200000.00
			feeControl = 95000.00
			feeStamp = 5000.00
			bankName = "Banque Nationale d'Algérie (BNA) • Agence Al Qods Chéraga"
			ribNumber = "00100622030000064395"
		} else {
			titleOp = "Renouvellement Quinquennal AMM Médicament (RI / RF)"
			amountDZD = 205000.00
			feeInstruction = 140000.00
			feeControl = 60000.00
			feeStamp = 5000.00
			bankName = "Trésor Public d'Alger • CAS N° 302.145"
			ribNumber = "00899999000000302145"
		}
	}
	art1Text = fmt.Sprintf("Redevance réglementaire au titre de : %s", titleOp)
	wordsFR, wordsAR := domain.ConvertAmountDZDToWords(amountDZD)
	wordsFull := fmt.Sprintf("%s (%s)", wordsFR, wordsAR)

	headerHTML := RenderInstitutionalHeader(estName, estAgreement, estAddress, pdtName, pdtDecision, "ES-FIN-02 • Bordereau de Liquidation Décret 21-82")
	footerHTML := RenderLegalFooter("ES-FIN-02", watermark, sha256Hash, qrURL, 1, 1, generatedAt)

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">ES-FIN-02 — Bordereau de Versement Variations, Renouvellements & Transferts</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">Identification de la Demande & Produit</div>
                <div class="field-row"><div class="field-label">Spécialité :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">D.C.I. :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Forme & Dosage :</div><div class="field-value">%s — %s</div></div>
                <div class="field-row"><div class="field-label">Dossier Homolog :</div><div class="field-value">%s</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">Domiciliation Bancaire & Trésor</div>
                <div class="field-row"><div class="field-label">Établissement :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Régie Recettes :</div><div class="field-value">Agence Nationale des Produits Pharmaceutiques</div></div>
                <div class="field-row"><div class="field-label">Compte / RIB :</div><div class="field-value" style="font-family:monospace; font-weight:700;">%s</div></div>
                <div class="field-row"><div class="field-label">Motif Exigé :</div><div class="field-value">Redevance ANPP • %s</div></div>
            </div>
        </div>

        <table class="data-table">
            <thead>
                <tr>
                    <th style="width:70px; text-align:center;">Article</th>
                    <th>Rubrique Budgétaire Réglementaire (Décret 21-82)</th>
                    <th style="width:120px; text-align:right;">Montant DZD</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td style="text-align:center; font-family:monospace;">Art. 01</td>
                    <td>Frais d'instruction administrative et d'évaluation du dossier (%s)</td>
                    <td style="text-align:right; font-weight:700;">%.2f DA</td>
                </tr>
                <tr>
                    <td style="text-align:center; font-family:monospace;">Art. 02</td>
                    <td>Frais d'expertise technique et d'évaluation analytique spécialisée</td>
                    <td style="text-align:right; font-weight:700;">%.2f DA</td>
                </tr>
                <tr>
                    <td style="text-align:center; font-family:monospace;">Art. 03</td>
                    <td>Droits d'enregistrement et timbre fiscal réglementaire</td>
                    <td style="text-align:right; font-weight:700;">%.2f DA</td>
                </tr>
                <tr style="background:#f0fdf4; font-weight:800; font-size:9pt;">
                    <td colspan="2" style="color:#064e3b;">TOTAL GÉNÉRAL À LIQUIDER (DZD)</td>
                    <td style="text-align:right; color:#047857;">%.2f DA</td>
                </tr>
            </tbody>
        </table>

        <div class="info-card" style="margin-bottom:12px; background:#f8fafc;">
            <div class="info-card-header">Arrêté du Présent Bordereau en Toutes Lettres</div>
            <div style="font-size:8.2pt; font-weight:700; color:#091e36; text-transform:uppercase;">
                Arrêté le présent bordereau à la somme de : %s.
            </div>
        </div>

        <div style="display:grid; grid-template-columns: 1fr 1fr 1fr; gap:10px; margin-top:10px;">
            <div class="sign-card">
                <div class="sign-header">Volet 1 : Demandeur</div>
                <div style="font-size:6.8pt; color:#475569;">Directeur Financier & DT</div>
                <div class="sign-space">
                    <span>Visa & Sceau Émetteur</span>
                </div>
            </div>
            <div class="sign-card">
                <div class="sign-header">Volet 2 : Organisme Financier</div>
                <div style="font-size:6.8pt; color:#475569;">Avis d'opéré bancaire</div>
                <div class="sign-space">
                    <span>Cachet Banque / Trésor</span>
                </div>
            </div>
            <div class="sign-card">
                <div class="sign-header">Volet 3 : Régie ANPP</div>
                <div style="font-size:6.8pt; color:#475569;">Quittance délivrée</div>
                <div class="sign-space">
                    <span>Visa Régisseur ANPP</span>
                </div>
            </div>
        </div>

        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(form), html.EscapeString(dosage), html.EscapeString(d.DossierCode), html.EscapeString(bankName), html.EscapeString(ribNumber), html.EscapeString(d.DossierCode), html.EscapeString(art1Text), feeInstruction, feeControl, feeStamp, amountDZD, html.EscapeString(wordsFull), footerHTML)

	return RenderDocumentWrap("ES-FIN-02 — Bordereau de Versement Variations & Renouvellements", content)
}
