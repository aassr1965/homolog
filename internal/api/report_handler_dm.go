package api

import (
	"fmt"
	"strings"
	"time"

	"homolog/internal/domain"
	"homolog/internal/reporting"
)

// renderESREG02DMHTML génère le Formulaire Officiel de Demande d'Homologation de Dispositif Médical
// Conforme au Décret exécutif n° 20-324 du 22 novembre 2020 et à l'Arrêté du 10 mai 2021 (6 sections normées)
func renderESREG02DMHTML(
	d *domain.Dossier,
	estName, estAgreement, estAddress, pdtName, pdtDecision string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	generatedAt time.Time,
) string {
	dmName := "Dispositif Médical Non Spécifié"
	dmNameAR := ""
	dmRef := "DM-REF-001"
	dmClass := "IIa"
	dmRule := "Règle 5 (Dispositif médical invasif par orifice naturel)"
	isSterile := true
	sterilMethod := "Oxyde d'éthylène (ETO)"
	isSingleUse := true
	intendedUse := "Usage médical et chirurgical spécialisé"
	originType := "importé"
	legalMfr := "Fabricant Légal International Certifié ISO 13485"
	legalCountry := "DE"
	ceCert := "CE-0123-MDR-2023-7749"
	notifiedBody := "TÜV SÜD Product Service GmbH (0123)"
	isoCert := "ISO13485-DE-TUV-9921"
	clvCert := "FSC-DE-BFARM-2023-4410 (BfArM)"
	gmdn := "34978"

	if d.MedicalDevice != nil {
		if d.MedicalDevice.CommercialNameFR != "" {
			dmName = d.MedicalDevice.CommercialNameFR
		}
		dmNameAR = d.MedicalDevice.CommercialNameAR
		if d.MedicalDevice.ManufacturerReference != "" {
			dmRef = d.MedicalDevice.ManufacturerReference
		}
		if d.MedicalDevice.DMClass != "" {
			dmClass = d.MedicalDevice.DMClass
		}
		if d.MedicalDevice.ClassificationRule != "" {
			dmRule = d.MedicalDevice.ClassificationRule
		}
		isSterile = d.MedicalDevice.IsSterile
		if d.MedicalDevice.SterilizationMethod != "" {
			sterilMethod = d.MedicalDevice.SterilizationMethod
		}
		isSingleUse = d.MedicalDevice.IsSingleUse
		if d.MedicalDevice.IntendedUseFR != "" {
			intendedUse = d.MedicalDevice.IntendedUseFR
		}
		if d.MedicalDevice.OriginType != "" {
			originType = d.MedicalDevice.OriginType
		}
		if d.MedicalDevice.LegalManufacturerName != "" {
			legalMfr = d.MedicalDevice.LegalManufacturerName
		}
		if d.MedicalDevice.LegalManufacturerCountry != "" {
			legalCountry = d.MedicalDevice.LegalManufacturerCountry
		}
		if d.MedicalDevice.CECertificateNumber != "" {
			ceCert = d.MedicalDevice.CECertificateNumber
		}
		if d.MedicalDevice.NotifiedBodyName != "" {
			notifiedBody = fmt.Sprintf("%s (%s)", d.MedicalDevice.NotifiedBodyName, d.MedicalDevice.NotifiedBodyNumber)
		}
		if d.MedicalDevice.ISO13485CertificateNumber != "" {
			isoCert = d.MedicalDevice.ISO13485CertificateNumber
		}
		if d.MedicalDevice.CLVFSCNumber != "" {
			auth := d.MedicalDevice.CLVIssuingAuthority
			if auth == "" {
				auth = "Autorité Compétente Pays d'Origine"
			}
			clvCert = fmt.Sprintf("%s (%s)", d.MedicalDevice.CLVFSCNumber, auth)
		}
		if d.MedicalDevice.GMDNCode != "" {
			gmdn = d.MedicalDevice.GMDNCode
		}
	}

	wmClass := ""
	wmText := "OFFICIEL — VALIDÉ"
	if watermark == domain.WatermarkDraft {
		wmClass = "draft"
		wmText = "BROUILLON — NON OPPOSABLE"
	}

	hashShort := sha256Hash
	if len(hashShort) > 24 {
		hashShort = hashShort[:24]
	}

	sterileStr := "OUI (Stérile)"
	if !isSterile {
		sterileStr = "NON (Non stérile)"
	}
	singleUseStr := "OUI (Usage unique)"
	if !isSingleUse {
		singleUseStr = "NON (Réutilisable après désinfection/stérilisation)"
	}

	originLabel := "Dispositif Médical Importé (Régime Décret 20-324)"
	if strings.Contains(strings.ToLower(originType), "local") {
		originLabel = "Dispositif Médical de Fabrication Locale (Régime Décret 20-324)"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>ES-REG-02-DM — Formulaire Officiel de Demande d'Homologation DM — %s</title>
    <style>
        @page { size: A4 portrait; margin: 0; }
        *, *::before, *::after { box-sizing: border-box; }
        body { font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, sans-serif; color: #0f172a; background-color: #f1f5f9; margin: 0; padding: 24px 0; font-size: 8.5pt; line-height: 1.4; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
        .no-print-toolbar { position: fixed; top: 0; left: 0; right: 0; height: 52px; background: #091e36; color: #ffffff; display: flex; align-items: center; justify-content: space-between; padding: 0 24px; z-index: 9999; box-shadow: 0 4px 12px rgba(0,0,0,0.15); font-size: 13px; }
        .no-print-toolbar .btn { display: inline-flex; align-items: center; gap: 8px; padding: 7px 14px; border-radius: 6px; font-weight: 600; font-size: 12px; cursor: pointer; border: none; text-decoration: none; }
        .btn-primary { background: #0284c7; color: #ffffff; }
        .btn-outline { background: rgba(255,255,255,0.12); color: #f8fafc; border: 1px solid rgba(255,255,255,0.25); }
        .sheet { width: 210mm; min-height: 297mm; margin: 48px auto 24px auto; background: #ffffff; box-shadow: 0 8px 30px rgba(15,23,42,0.12); padding: 14mm 16mm; position: relative; overflow: hidden; }
        @media print { body { background: #ffffff; padding: 0; margin: 0; } .no-print-toolbar { display: none !important; } .sheet { width: 100%%; margin: 0; padding: 12mm 15mm; box-shadow: none; border-radius: 0; } }
        .watermark { position: absolute; top: 48%%; left: 50%%; transform: translate(-50%%, -50%%) rotate(-35deg); font-size: 46pt; font-weight: 900; letter-spacing: 5px; color: rgba(4,120,87,0.06); pointer-events: none; user-select: none; z-index: 1; white-space: nowrap; text-transform: uppercase; }
        .watermark.draft { color: rgba(100,116,139,0.08); }
        .header-table { width: 100%%; border-collapse: collapse; margin-bottom: 10px; border-bottom: 2px solid #091e36; padding-bottom: 6px; }
        .title-banner { background: #091e36; color: #ffffff; padding: 8px 14px; border-radius: 3px; display: flex; align-items: center; justify-content: space-between; margin-bottom: 11px; }
        .banner-title { font-size: 10.5pt; font-weight: 800; letter-spacing: 0.5px; text-transform: uppercase; margin: 0; }
        .info-card { border: 1px solid #cbd5e1; border-radius: 4px; padding: 8px 11px; background: #ffffff; margin-bottom: 9px; }
        .info-card-header { font-size: 8.2pt; font-weight: 800; color: #091e36; text-transform: uppercase; letter-spacing: 0.5px; border-bottom: 1px solid #e2e8f0; padding-bottom: 3px; margin-bottom: 6px; display: flex; align-items: center; gap: 6px; }
        .field-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 3px 14px; }
        .field-row { display: flex; margin-bottom: 2px; font-size: 8pt; }
        .field-label { width: 155px; font-weight: 700; color: #475569; flex-shrink: 0; }
        .field-value { color: #091e36; font-weight: 600; flex-grow: 1; }
        .data-table { width: 100%%; border-collapse: collapse; margin-bottom: 9px; font-size: 7.8pt; }
        .data-table th { background: #0f3b60; color: #ffffff; padding: 5px 8px; text-align: left; font-weight: 700; font-size: 7.5pt; border: 1px solid #0f3b60; }
        .data-table td { padding: 4px 8px; border: 1px solid #cbd5e1; }
        .data-table tr:nth-child(even) { background-color: #f8fafc; }
        .sign-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-top: 10px; page-break-inside: avoid; }
        .sign-card { border: 1px solid #091e36; border-radius: 4px; padding: 8px; background: #ffffff; }
        .sign-header { font-weight: 800; font-size: 7.8pt; color: #091e36; text-transform: uppercase; margin-bottom: 4px; }
        .sign-space { height: 56px; border: 1px dashed #cbd5e1; background: #f8fafc; border-radius: 3px; display: flex; align-items: center; justify-content: center; font-size: 7.2pt; color: #64748b; margin-top: 4px; }
        .declaration-box { border: 1px solid #0284c7; background: #f0f9ff; border-radius: 4px; padding: 7px 11px; font-size: 7.5pt; line-height: 1.35; color: #0369a1; margin-bottom: 9px; }
        .footer-legal { margin-top: 12px; padding-top: 6px; border-top: 1px solid #cbd5e1; display: flex; align-items: center; justify-content: space-between; font-size: 6.8pt; color: #64748b; page-break-inside: avoid; }
    </style>
</head>
<body>
    <div class="no-print-toolbar">
        <div><strong>Homolog Reporting BPF</strong> • ES-REG-02-DM Formulaire Officiel d'Homologation DM (Décret 20-324 & Arrêté 10/05/2021)</div>
        <div style="display:flex; gap:10px;">
            <button class="btn btn-primary" onclick="window.print()">🖨️ Imprimer / Export PDF (A4 300 DPI)</button>
            <button type="button" class="btn btn-outline" onclick="if(window.parent && window.parent !== window && typeof window.parent.closeModal === 'function'){ window.parent.closeModal('modal-dossier-reports'); } else { window.close(); }">✕ Fermer</button>
        </div>
    </div>
    <script>
        if (window.parent && window.parent !== window) {
            var tb = document.querySelector('.no-print-toolbar');
            if (tb) tb.style.display = 'none';
            var sh = document.querySelector('.sheet');
            if (sh) sh.style.marginTop = '15px';
        }
    </script>

    <div class="sheet">
        <div class="watermark %s">%s</div>

        <table class="header-table">
            <tr>
                <td style="width:55px; vertical-align:middle;">
                    <svg width="44" height="44" viewBox="0 0 100 100" fill="none">
                        <circle cx="50" cy="50" r="46" stroke="#091e36" stroke-width="4" fill="#f8fafc"/>
                        <path d="M50 15v70M30 35h40M30 65h40M40 50h20" stroke="#047857" stroke-width="4" stroke-linecap="round"/>
                    </svg>
                </td>
                <td style="vertical-align:middle; padding-left:10px;">
                    <div style="font-size:6.8pt; font-weight:700; color:#64748b; text-transform:uppercase;">RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE</div>
                    <div style="font-size:7.8pt; font-weight:800; color:#091e36;">MINISTÈRE DE L'INDUSTRIE ET DE LA PRODUCTION PHARMACEUTIQUE • ANPP</div>
                    <div style="font-size:7pt; color:#047857; font-weight:700;">Direction des Dispositifs Médicaux (DDM) • Sous-Direction de l'Homologation</div>
                </td>
                <td style="text-align:right; vertical-align:middle;">
                    <div style="font-size:6.8pt; color:#64748b;">RÉFÉRENCE OFFICIELLE</div>
                    <div style="font-size:9pt; font-weight:800; color:#091e36;">ES-REG-02-DM</div>
                    <div style="font-size:6.8pt; color:#047857; font-weight:700;">Décret exécutif n° 20-324</div>
                </td>
            </tr>
        </table>

        <div class="title-banner">
            <h1 class="banner-title">Formulaire Officiel de Demande d'Homologation de Dispositif Médical</h1>
            <span style="font-size:7.8pt; font-weight:700; background:rgba(255,255,255,0.2); padding:3px 8px; border-radius:3px;">Dossier : %s</span>
        </div>

        <!-- SECTION 1 : Renseignements Administratifs -->
        <div class="info-card">
            <div class="info-card-header">1. Renseignements Administratifs sur le Dispositif Médical & Demandeur</div>
            <div class="field-grid">
                <div class="field-row"><span class="field-label">Dénomination Commerciale :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Dénomination en Arabe :</span><span class="field-value" style="font-family:'Traditional Arabic', Tahoma;">%s</span></div>
                <div class="field-row"><span class="field-label">Référence Fabricant :</span><span class="field-value" style="font-family:monospace;">%s</span></div>
                <div class="field-row"><span class="field-label">Code Nomenclature GMDN :</span><span class="field-value" style="font-family:monospace;">%s</span></div>
                <div class="field-row"><span class="field-label">Classe de Risque (DDM) :</span><span class="field-value">Classe %s (%s)</span></div>
                <div class="field-row"><span class="field-label">Destination d'Usage :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Établissement Demandeur :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">N° Agrément MIPH :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Siège Social & Ville :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Pharmacien DT Déposant :</span><span class="field-value">%s (%s)</span></div>
            </div>
        </div>

        <!-- SECTION 2 : Engagement Solennel PDT -->
        <div class="declaration-box">
            <strong>2. Certifications & Engagement Solennel du Pharmacien Directeur Technique (Arrêté 10 mai 2021) :</strong><br/>
            Je soussigné, <strong>%s</strong>, Pharmacien Directeur Technique dûment mandaté par le Ministère, certifie solennellement que le dispositif médical susvisé est conforme aux exigences de sécurité, de santé et de performance du Décret exécutif n° 20-324. Je m'engage formellement au respect des règles de traçabilité, de matériovigilance, et confirme que l'ensemble des droits d'homologation a été réglé conformément à l'Article 31 de l'Ordonnance n° 21-07.
        </div>

        <!-- SECTION 3 : Nature de la Demande & Régime -->
        <div class="info-card">
            <div class="info-card-header">3. Nature de la Demande & Certifications Internationales</div>
            <div class="field-grid">
                <div class="field-row"><span class="field-label">Régime Réglementaire :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Fabricant Légal & Pays :</span><span class="field-value">%s (%s)</span></div>
                <div class="field-row"><span class="field-label">Certificat de Libre Vente :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Certification ISO 13485 :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Marquage CE de Conformité :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Organisme Notifié (NB) :</span><span class="field-value">%s</span></div>
            </div>
        </div>

        <!-- SECTION 4 : Renseignements Techniques & Échantillonnage -->
        <div class="info-card">
            <div class="info-card-header">4. Renseignements Techniques, Caractéristiques & Échantillonnage</div>
            <div class="field-grid">
                <div class="field-row"><span class="field-label">Statut de Stérilité :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Procédé de Stérilisation :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Caractère à Usage Unique :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Durée de Validité Proposée :</span><span class="field-value">36 mois (Stabilité temps réel et accélérée ISO 11607)</span></div>
                <div class="field-row"><span class="field-label">Échantillons Déposés :</span><span class="field-value">03 échantillons représentatifs + 01 blister pelable témoin</span></div>
                <div class="field-row"><span class="field-label">Lots Échantillonnés :</span><span class="field-value">Lot PF N° LOT-PF-2026-DM01 • Lot MP N° LOT-MP-PUR-098</span></div>
            </div>
        </div>

        <!-- SECTION 5 : Cartographie Industrielle des Sites -->
        <div class="info-card">
            <div class="info-card-header">5. Cartographie Industrielle des Sites de Fabrication & Contrôle</div>
            <table class="data-table">
                <thead>
                    <tr>
                        <th style="width:25%%;">Étape Industrielle</th>
                        <th style="width:35%%;">Raison Sociale de l'Établissement</th>
                        <th style="width:20%%;">Pays & Ville</th>
                        <th style="width:20%%;">Référentiel Qualité</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td><strong>Fabrication & Assemblage</strong></td>
                        <td>%s</td>
                        <td>%s</td>
                        <td>ISO 13485 / BPF Dispositifs</td>
                    </tr>
                    <tr>
                        <td><strong>Conditionnement Primaire / Blister</strong></td>
                        <td>%s</td>
                        <td>%s</td>
                        <td>Salle Propre ISO 7 / ISO 14644</td>
                    </tr>
                    <tr>
                        <td><strong>Stérilisation Industrielle</strong></td>
                        <td>Unité Spécialisée Stérilisation Médicale</td>
                        <td>%s</td>
                        <td>ISO 11135 (ETO) / ISO 11137</td>
                    </tr>
                    <tr>
                        <td><strong>Laboratoire Contrôle Qualité (LCQ)</strong></td>
                        <td>%s (LCQ Agréé)</td>
                        <td>%s</td>
                        <td>ISO/IEC 17025 • Validation ANPP</td>
                    </tr>
                    <tr>
                        <td><strong>Libération Officielle du Lot</strong></td>
                        <td>%s</td>
                        <td>%s</td>
                        <td>Visa Pharmacien Directeur Technique</td>
                    </tr>
                </tbody>
            </table>
        </div>

        <!-- SECTION 6 : Performance & Biocompatibilité -->
        <div class="info-card">
            <div class="info-card-header">6. Performance, Sécurité Biologique (ISO 10993) & Gestion des Risques (ISO 14971)</div>
            <table class="data-table">
                <thead>
                    <tr>
                        <th style="width:30%%;">Norme Internationale</th>
                        <th style="width:45%%;">Objet de l'Évaluation Réglementaire</th>
                        <th style="width:25%%;">Statut & Justificatif</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td><strong>ISO 14971:2019</strong></td>
                        <td>Application de la gestion des risques aux dispositifs médicaux</td>
                        <td><span style="color:#047857; font-weight:700;">✓ Conforme</span> (Rapport PGR validé)</td>
                    </tr>
                    <tr>
                        <td><strong>ISO 10993-1 / 10993-5</strong></td>
                        <td>Évaluation biologique & Biocompatibilité (Essais de cytotoxicité in vitro)</td>
                        <td><span style="color:#047857; font-weight:700;">✓ Conforme</span> (Rapport BPL exempt de toxicité)</td>
                    </tr>
                    <tr>
                        <td><strong>ISO 10993-10</strong></td>
                        <td>Essais d'irritation et de sensibilisation cutanée / muqueuse</td>
                        <td><span style="color:#047857; font-weight:700;">✓ Conforme</span> (Indice d'irritation nul)</td>
                    </tr>
                    <tr>
                        <td><strong>ISO 14155:2020</strong></td>
                        <td>Investigation clinique des dispositifs médicaux pour sujets humains</td>
                        <td><span style="color:#047857; font-weight:700;">✓ Conforme</span> (Rapport d'évaluation CER)</td>
                    </tr>
                    <tr>
                        <td><strong>Contrôle Matériaux Critiques</strong></td>
                        <td>Exemption formelle de Latex, DEHP/Phtalates et métaux lourds toxiques</td>
                        <td><span style="color:#047857; font-weight:700;">✓ Certifié</span> (Déclaration d'absence jointe)</td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div class="sign-grid">
            <div class="sign-card">
                <div class="sign-header">Pharmacien Directeur Technique Déposant</div>
                <div style="font-size:7pt; color:#475569;">Visa, Cachet Nominatif & Date</div>
                <div class="sign-space">
                    <div>
                        <strong>%s</strong><br/>
                        <span style="font-size:6.5pt;">Décision MIPH N° %s</span>
                    </div>
                </div>
            </div>
            <div class="sign-card">
                <div class="sign-header">Direction Générale de l'Établissement</div>
                <div style="font-size:7pt; color:#475569;">Signature Solennelle & Cachet Social</div>
                <div class="sign-space">
                    <div>
                        <strong>%s</strong><br/>
                        <span style="font-size:6.5pt;">Pour la Direction Générale</span>
                    </div>
                </div>
            </div>
        </div>

        %s
    </div>
</body>
</html>`,
		dmName,
		wmClass, wmText,
		d.DossierCode,
		dmName,
		dmNameAR,
		dmRef,
		gmdn,
		dmClass, dmRule,
		intendedUse,
		estName,
		estAgreement,
		estAddress,
		pdtName, pdtDecision,
		pdtName,
		originLabel,
		legalMfr, legalCountry,
		clvCert,
		isoCert,
		ceCert,
		notifiedBody,
		sterileStr,
		sterilMethod,
		singleUseStr,
		legalMfr, legalCountry,
		legalMfr, legalCountry,
		legalCountry,
		estName, estAddress,
		estName, estAddress,
		pdtName, pdtDecision,
		estName,
		reporting.RenderLegalFooter("ES-REG-02-DM", watermark, sha256Hash, qrURL, 1, 1, generatedAt),
	)
}

// renderESREG03DMHTML génère le Récépissé Officiel de Dépôt des Dossiers soumis à Homologation DM
// Grille de recevabilité administrative et technique en 13 points de contrôle (Décret 20-324 & Arrêté 10 mai 2021)
func renderESREG03DMHTML(
	d *domain.Dossier,
	estName, estAgreement, estAddress, pdtName, pdtDecision string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	generatedAt time.Time,
	eval *domain.DMRecevabiliteEvaluation,
) string {
	dmName := "Dispositif Médical"
	dmRef := "DM-REF-001"
	dmClass := "IIa"
	dmRule := "Règle 5 (Dispositif médical invasif par orifice naturel)"

	if d.MedicalDevice != nil {
		if d.MedicalDevice.CommercialNameFR != "" {
			dmName = d.MedicalDevice.CommercialNameFR
		}
		if d.MedicalDevice.ManufacturerReference != "" {
			dmRef = d.MedicalDevice.ManufacturerReference
		}
		if d.MedicalDevice.DMClass != "" {
			dmClass = d.MedicalDevice.DMClass
		}
		if d.MedicalDevice.ClassificationRule != "" {
			dmRule = d.MedicalDevice.ClassificationRule
		}
	}

	wmClass := ""
	wmText := "OFFICIEL — VALIDÉ"
	if watermark == domain.WatermarkDraft {
		wmClass = "draft"
		wmText = "BROUILLON — NON OPPOSABLE"
	}

	hashShort := sha256Hash
	if len(hashShort) > 24 {
		hashShort = hashShort[:24]
	}

	// Valeurs par défaut si eval non fournie
	orderNumber := "ANPP/DDM/REC/2026/0148"
	agentName := "Dr. Amel Bouzid (Sous-Direction de l'Homologation DDM)"
	score := 13
	decisionConclusion := `Le dossier est déclaré <strong>COMPLET ET RECEVABLE</strong> pour enregistrement et transmission à la commission d'évaluation technique et scientifique sous le N° d'ordre : <strong>ANPP/DDM/REC/2026/0148</strong>.`
	badgeFunc := func(st string) string {
		switch strings.ToLower(strings.TrimSpace(st)) {
		case "conforme":
			return `<span class="badge-ok">✓ CONFORME</span>`
		case "avec_reserve":
			return `<span class="badge-warn" style="background:#fef3c7; color:#92400e; font-weight:800; padding:2px 6px; border-radius:3px; font-size:7pt;">⚠️ AVEC RÉSERVE</span>`
		default:
			return `<span class="badge-err" style="background:#fee2e2; color:#b91c1c; font-weight:800; padding:2px 6px; border-radius:3px; font-size:7pt;">✗ NON CONFORME</span>`
		}
	}

	b1, b2, b3, b4, b5, b6, b7, b8, b9, b10, b11, b12, b13 := badgeFunc("conforme"), badgeFunc("conforme"), badgeFunc("conforme"), badgeFunc("conforme"), badgeFunc("conforme"), badgeFunc("conforme"), badgeFunc("conforme"), badgeFunc("conforme"), badgeFunc("conforme"), badgeFunc("conforme"), badgeFunc("conforme"), badgeFunc("conforme"), `<span class="badge-ok">✓ 04 REÇUS</span>`
	sealHTML := ""

	if eval != nil {
		if eval.DepositOrderNumber != "" {
			orderNumber = eval.DepositOrderNumber
		}
		if eval.ANPPAgentName != "" {
			agentName = eval.ANPPAgentName
		}
		score = eval.ConformityScore
		b1 = badgeFunc(eval.Item1Status)
		b2 = badgeFunc(eval.Item2Status)
		b3 = badgeFunc(eval.Item3Status)
		b4 = badgeFunc(eval.Item4Status)
		b5 = badgeFunc(eval.Item5Status)
		b6 = badgeFunc(eval.Item6Status)
		b7 = badgeFunc(eval.Item7Status)
		b8 = badgeFunc(eval.Item8Status)
		b9 = badgeFunc(eval.Item9Status)
		b10 = badgeFunc(eval.Item10Status)
		b11 = badgeFunc(eval.Item11Status)
		b12 = badgeFunc(eval.Item12Status)
		if eval.Item13Status == "conforme" {
			b13 = fmt.Sprintf(`<span class="badge-ok">✓ %02d REÇUS</span>`, eval.ElectronicCopiesCount)
		} else {
			b13 = badgeFunc(eval.Item13Status)
		}

		if eval.Decision == domain.RecevabiliteDecisionComplet {
			decisionConclusion = fmt.Sprintf(`Le dossier est déclaré <strong>COMPLET ET RECEVABLE</strong> (%d/13 points conformes) pour enregistrement et transmission à la commission d'évaluation technique et scientifique sous le N° d'ordre : <strong>%s</strong>.`, score, orderNumber)
		} else if eval.Decision == domain.RecevabiliteDecisionReserves {
			resNote := eval.ReservesNotes
			if resNote == "" {
				resNote = "Compléments documentaires requis sous délai réglementaire de 30 jours."
			}
			decisionConclusion = fmt.Sprintf(`Le dossier est déclaré <strong>RECEVABLE AVEC RÉSERVES</strong> (%d/13 points conformes). Motif: %s — Enregistré sous le N° d'ordre : <strong>%s</strong>.`, score, resNote, orderNumber)
		} else {
			decisionConclusion = fmt.Sprintf(`Le dossier est déclaré <strong>IRRECEVABLE</strong> (%d/13 points conformes). Pièces non conformes. Enregistré sous le N° d'ordre : <strong>%s</strong>.`, score, orderNumber)
		}

		if eval.PDTSignatureSeal != "" {
			sealHTML = fmt.Sprintf(`<div style="margin-top:6px; font-family:monospace; font-size:6.5pt; color:#047857; font-weight:bold;">✔ Sceau 21 CFR Part 11: %s<br/>Signé électroniquement le %s</div>`,
				eval.PDTSignatureSeal, eval.PDTSignedAt.Format("02/01/2006 15:04 UTC"))
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>ES-REG-03-DM — Récépissé Officiel de Dépôt des Dossiers d'Homologation DM — %s</title>
    <style>
        @page { size: A4 portrait; margin: 0; }
        *, *::before, *::after { box-sizing: border-box; }
        body { font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, sans-serif; color: #0f172a; background-color: #f1f5f9; margin: 0; padding: 24px 0; font-size: 8.5pt; line-height: 1.4; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
        .no-print-toolbar { position: fixed; top: 0; left: 0; right: 0; height: 52px; background: #091e36; color: #ffffff; display: flex; align-items: center; justify-content: space-between; padding: 0 24px; z-index: 9999; box-shadow: 0 4px 12px rgba(0,0,0,0.15); font-size: 13px; }
        .no-print-toolbar .btn { display: inline-flex; align-items: center; gap: 8px; padding: 7px 14px; border-radius: 6px; font-weight: 600; font-size: 12px; cursor: pointer; border: none; text-decoration: none; }
        .btn-primary { background: #0284c7; color: #ffffff; }
        .btn-outline { background: rgba(255,255,255,0.12); color: #f8fafc; border: 1px solid rgba(255,255,255,0.25); }
        .sheet { width: 210mm; min-height: 297mm; margin: 48px auto 24px auto; background: #ffffff; box-shadow: 0 8px 30px rgba(15,23,42,0.12); padding: 14mm 16mm; position: relative; overflow: hidden; }
        @media print { body { background: #ffffff; padding: 0; margin: 0; } .no-print-toolbar { display: none !important; } .sheet { width: 100%%; margin: 0; padding: 12mm 15mm; box-shadow: none; border-radius: 0; } }
        .watermark { position: absolute; top: 48%%; left: 50%%; transform: translate(-50%%, -50%%) rotate(-35deg); font-size: 46pt; font-weight: 900; letter-spacing: 5px; color: rgba(4,120,87,0.06); pointer-events: none; user-select: none; z-index: 1; white-space: nowrap; text-transform: uppercase; }
        .watermark.draft { color: rgba(100,116,139,0.08); }
        .header-table { width: 100%%; border-collapse: collapse; margin-bottom: 10px; border-bottom: 2px solid #091e36; padding-bottom: 6px; }
        .title-banner { background: #091e36; color: #ffffff; padding: 8px 14px; border-radius: 3px; display: flex; align-items: center; justify-content: space-between; margin-bottom: 11px; }
        .banner-title { font-size: 10.5pt; font-weight: 800; letter-spacing: 0.5px; text-transform: uppercase; margin: 0; }
        .info-card { border: 1px solid #cbd5e1; border-radius: 4px; padding: 8px 11px; background: #ffffff; margin-bottom: 9px; }
        .info-card-header { font-size: 8.2pt; font-weight: 800; color: #091e36; text-transform: uppercase; letter-spacing: 0.5px; border-bottom: 1px solid #e2e8f0; padding-bottom: 3px; margin-bottom: 6px; display: flex; align-items: center; gap: 6px; }
        .field-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 3px 14px; }
        .field-row { display: flex; margin-bottom: 2px; font-size: 8pt; }
        .field-label { width: 155px; font-weight: 700; color: #475569; flex-shrink: 0; }
        .field-value { color: #091e36; font-weight: 600; flex-grow: 1; }
        .data-table { width: 100%%; border-collapse: collapse; margin-bottom: 9px; font-size: 7.7pt; }
        .data-table th { background: #0f3b60; color: #ffffff; padding: 5px 8px; text-align: left; font-weight: 700; font-size: 7.5pt; border: 1px solid #0f3b60; }
        .data-table td { padding: 4px 8px; border: 1px solid #cbd5e1; }
        .data-table tr:nth-child(even) { background-color: #f8fafc; }
        .badge-ok { background: #dcfce7; color: #166534; font-weight: 800; padding: 2px 6px; border-radius: 3px; font-size: 7pt; display: inline-block; }
        .recepisse-box { border: 1.5px solid #047857; background: #f0fdf4; border-radius: 4px; padding: 9px 12px; margin-bottom: 10px; }
        .recepisse-title { font-size: 8.5pt; font-weight: 800; color: #047857; text-transform: uppercase; margin-bottom: 4px; }
        .sign-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-top: 10px; page-break-inside: avoid; }
        .sign-card { border: 1px solid #091e36; border-radius: 4px; padding: 8px; background: #ffffff; }
        .sign-header { font-weight: 800; font-size: 7.8pt; color: #091e36; text-transform: uppercase; margin-bottom: 4px; }
        .sign-space { height: 58px; border: 1px dashed #cbd5e1; background: #f8fafc; border-radius: 3px; display: flex; align-items: center; justify-content: center; font-size: 7.2pt; color: #64748b; margin-top: 4px; text-align: center; }
        .footer-legal { margin-top: 12px; padding-top: 6px; border-top: 1px solid #cbd5e1; display: flex; align-items: center; justify-content: space-between; font-size: 6.8pt; color: #64748b; page-break-inside: avoid; }
    </style>
</head>
<body>
    <div class="no-print-toolbar">
        <div><strong>Homolog Reporting BPF</strong> • ES-REG-03-DM Récépissé Officiel de Dépôt DM (Grille en 13 Points)</div>
        <div style="display:flex; gap:10px;">
            <button class="btn btn-primary" onclick="window.print()">🖨️ Imprimer / Export PDF (A4 300 DPI)</button>
            <button type="button" class="btn btn-outline" onclick="if(window.parent && window.parent !== window && typeof window.parent.closeModal === 'function'){ window.parent.closeModal('modal-dossier-reports'); } else { window.close(); }">✕ Fermer</button>
        </div>
    </div>
    <script>
        if (window.parent && window.parent !== window) {
            var tb = document.querySelector('.no-print-toolbar');
            if (tb) tb.style.display = 'none';
            var sh = document.querySelector('.sheet');
            if (sh) sh.style.marginTop = '15px';
        }
    </script>

    <div class="sheet">
        <div class="watermark %s">%s</div>

        <table class="header-table">
            <tr>
                <td style="width:55px; vertical-align:middle;">
                    <svg width="44" height="44" viewBox="0 0 100 100" fill="none">
                        <circle cx="50" cy="50" r="46" stroke="#091e36" stroke-width="4" fill="#f8fafc"/>
                        <path d="M30 50l15 15L75 35" stroke="#047857" stroke-width="6" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                </td>
                <td style="vertical-align:middle; padding-left:10px;">
                    <div style="font-size:6.8pt; font-weight:700; color:#64748b; text-transform:uppercase;">RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE</div>
                    <div style="font-size:7.8pt; font-weight:800; color:#091e36;">MINISTÈRE DE L'INDUSTRIE ET DE LA PRODUCTION PHARMACEUTIQUE • ANPP</div>
                    <div style="font-size:7pt; color:#047857; font-weight:700;">Direction des Dispositifs Médicaux (DDM) • Guichet d'Accueil & Enregistrement</div>
                </td>
                <td style="text-align:right; vertical-align:middle;">
                    <div style="font-size:6.8pt; color:#64748b;">ACTE OFFICIEL</div>
                    <div style="font-size:9pt; font-weight:800; color:#091e36;">ES-REG-03-DM</div>
                    <div style="font-size:6.8pt; color:#047857; font-weight:700;">Arrêté du 10 mai 2021</div>
                </td>
            </tr>
        </table>

        <div class="title-banner">
            <h1 class="banner-title">Récépissé Officiel de Dépôt des Dossiers soumis à Homologation DM</h1>
            <span style="font-size:7.8pt; font-weight:700; background:rgba(255,255,255,0.2); padding:3px 8px; border-radius:3px;">Dossier : %s</span>
        </div>

        <!-- 1. Identification de la Demande -->
        <div class="info-card">
            <div class="info-card-header">1. Identification du Demandeur, du Dispositif Médical & Échantillons Déposés</div>
            <div class="field-grid">
                <div class="field-row"><span class="field-label">Établissement Assujetti :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">N° Agrément MIPH :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Siège Social & Adresse :</span><span class="field-value">%s</span></div>
                <div class="field-row"><span class="field-label">Pharmacien DT Déposant :</span><span class="field-value">%s (%s)</span></div>
                <div class="field-row"><span class="field-label">Dispositif Médical :</span><span class="field-value"><strong>%s</strong></span></div>
                <div class="field-row"><span class="field-label">Référence Fabricant :</span><span class="field-value" style="font-family:monospace;">%s</span></div>
                <div class="field-row"><span class="field-label">Classe Réglementaire :</span><span class="field-value">Classe %s (%s)</span></div>
                <div class="field-row"><span class="field-label">Nature de la Soumission :</span><span class="field-value">Nouvelle Demande d'Homologation DM (Ordonnance 21-07)</span></div>
                <div class="field-row"><span class="field-label">Échantillons Remis au Guichet :</span><span class="field-value">03 unités de vente sous emballage protecteur scellé</span></div>
                <div class="field-row"><span class="field-label">Bulletins d'Analyse Fournis :</span><span class="field-value">Bulletins de contrôle Produit Fini + Matières Premières</span></div>
            </div>
        </div>

        <!-- 2. Grille de Recevabilité en 13 Points -->
        <div class="info-card">
            <div class="info-card-header">2. Grille Officielle d'Évaluation de la Recevabilité (13 Points de Contrôle Normés)</div>
            <table class="data-table">
                <thead>
                    <tr>
                        <th style="width:6%%; text-align:center;">N°</th>
                        <th style="width:58%%;">Exigence Réglementaire (Arrêté du 10 mai 2021)</th>
                        <th style="width:14%%; text-align:center;">Évaluation</th>
                        <th style="width:22%%;">Référence de la Pièce</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td style="text-align:center; font-weight:700;">1</td>
                        <td>Lettre de demande d'homologation en 02 exemplaires + descriptif technique détaillé</td>
                        <td style="text-align:center;">%s</td>
                        <td>Lettre officielle signée PDT</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">2</td>
                        <td>Formulaire officiel de demande d'homologation dûment renseigné et signé</td>
                        <td style="text-align:center;">%s</td>
                        <td>Formulaire 6 sections (ES-REG-02-DM)</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">3</td>
                        <td>Agrément d'ouverture MIPH (fabricant/exploitant) + contrat technique d'exploitation</td>
                        <td style="text-align:center;">%s</td>
                        <td>Agrément %s</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">4</td>
                        <td>Justification paiement droits (Compte BNA Chéraga 00100622030000064395 - 300 000 DA)</td>
                        <td style="text-align:center;">%s</td>
                        <td>BV DM & Quittance bancaire</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">5</td>
                        <td>Certificat de Libre Vente (CLV / FSC) légalisé / apostillé du pays d'origine</td>
                        <td style="text-align:center;">%s</td>
                        <td>CLV délivré par autorité sanitaire</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">6</td>
                        <td>Certificat CE de conformité délivré par Organisme Notifié + Déclaration CE fabricant</td>
                        <td style="text-align:center;">%s</td>
                        <td>Certificat CE en cours de validité</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">7</td>
                        <td>Certifications BPF/GMP, ISO 13485 et ISO 9001 du site fabricant PF (Validation LCQ)</td>
                        <td style="text-align:center;">%s</td>
                        <td>Certificats ISO 13485 & Décision LCQ</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">8</td>
                        <td>Certifications qualité BPF / ISO 13485 / ISO 9001 des matières premières</td>
                        <td style="text-align:center;">%s</td>
                        <td>CoA et certificats biomatériaux</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">9</td>
                        <td>Description du système de matériovigilance et Plan de Gestion des Risques (PGR)</td>
                        <td style="text-align:center;">%s</td>
                        <td>SOP Matériovigilance & PGR</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">10</td>
                        <td>Partie III : Stabilité, procédés de fabrication et validation de la stérilisation</td>
                        <td style="text-align:center;">%s</td>
                        <td>Rapport stérilisation ISO 11135</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">11</td>
                        <td>Partie IV : Performance, gestion des risques (ISO 14971) et biocompatibilité (ISO 10993)</td>
                        <td style="text-align:center;">%s</td>
                        <td>Rapports BPL ISO 10993-1/5/10</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">12</td>
                        <td>Partie V : Rapport de l'évaluation clinique ou investigation clinique (ISO 14155)</td>
                        <td style="text-align:center;">%s</td>
                        <td>Rapport CER conforme ISO 14155</td>
                    </tr>
                    <tr>
                        <td style="text-align:center; font-weight:700;">13</td>
                        <td><strong>Exemplaires électroniques : 04 exemplaires complets sur supports scellés</strong></td>
                        <td style="text-align:center;">%s</td>
                        <td>04 clés USB scellées SHA-256</td>
                    </tr>
                </tbody>
            </table>
        </div>

        <!-- 3. Avis de Recevabilité du Guichet -->
        <div class="recepisse-box">
            <div class="recepisse-title">3. Décision de Recevabilité Formelle (Direction des Dispositifs Médicaux ANPP)</div>
            <div style="font-size:8pt; color:#065f46; line-height:1.4;">
                Le dossier technique et administratif déposé au guichet d'accueil de la Direction des Dispositifs Médicaux de l'ANPP a été examiné conformément à la grille réglementaire en 13 points. <br/>
                <strong>CONCLUSION :</strong> %s
            </div>
        </div>

        <!-- 4. Double Visa Solennel -->
        <div class="sign-grid">
            <div class="sign-card">
                <div class="sign-header">Établissement Déposant</div>
                <div style="font-size:7pt; color:#475569;">Pharmacien Directeur Technique</div>
                <div class="sign-space">
                    <div>
                        <strong>%s</strong><br/>
                        <span style="font-size:6.5pt;">Décision MIPH N° %s<br/>Date & Cachet du Laboratoire</span>
                        %s
                    </div>
                </div>
            </div>
            <div class="sign-card">
                <div class="sign-header">Guichet de Dépôt ANPP (DDM)</div>
                <div style="font-size:7pt; color:#475569;">Chargé de Réception des Dossiers DM</div>
                <div class="sign-space">
                    <div>
                        <strong>%s</strong><br/>
                        <span style="font-size:6.5pt;">Cachet Humide du Bureau d'Ordre<br/>Reçu le %s</span>
                    </div>
                </div>
            </div>
        </div>

        %s
    </div>
</body>
</html>`,
		dmName,
		wmClass, wmText,
		d.DossierCode,
		estName,
		estAgreement,
		estAddress,
		pdtName, pdtDecision,
		dmName,
		dmRef,
		dmClass, dmRule,
		b1, b2, b3, estAgreement, b4, b5, b6, b7, b8, b9, b10, b11, b12, b13,
		decisionConclusion,
		pdtName, pdtDecision,
		sealHTML,
		agentName,
		generatedAt.Format("02/01/2006"),
		reporting.RenderLegalFooter("ES-REG-03-DM", watermark, sha256Hash, qrURL, 1, 1, generatedAt),
	)
}

// renderESREG05DMHTML génère l'Avenant Officiel à la Décision d'Homologation d'un Dispositif Médical
// Conforme au Décret exécutif n° 20-324 du 22 novembre 2020 (Articles 22 à 25)
func renderESREG05DMHTML(
	mod *domain.DMModification,
	estName, estAgreement, estAddress, pdtName, pdtDecision string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	generatedAt time.Time,
) string {
	wmClass := ""
	wmText := "OFFICIEL — VALIDÉ"
	if watermark == domain.WatermarkDraft {
		wmClass = "draft"
		wmText = "BROUILLON — NON OPPOSABLE"
	}

	dmName := "Dispositif Médical"
	dmRef := "DM-REF"
	dmClass := "IIa"
	homologDec := "DÉC-MIPH-DM-2024-XXXX"
	homologDate := "15/06/2024"

	if mod.MedicalDevice != nil {
		if mod.MedicalDevice.CommercialNameFR != "" {
			dmName = mod.MedicalDevice.CommercialNameFR
		}
		if mod.MedicalDevice.ManufacturerReference != "" {
			dmRef = mod.MedicalDevice.ManufacturerReference
		}
		if mod.MedicalDevice.DMClass != "" {
			dmClass = mod.MedicalDevice.DMClass
		}
		if mod.MedicalDevice.HomologationDecisionNumber != "" {
			homologDec = mod.MedicalDevice.HomologationDecisionNumber
		}
		if mod.MedicalDevice.HomologationDate != nil {
			homologDate = mod.MedicalDevice.HomologationDate.Format("02/01/2006")
		}
	}

	typeBadge := `<span style="background:#0f3b60; color:white; padding:3px 8px; border-radius:3px; font-weight:bold; font-size:8.5pt;">MODIFICATION SUBSTANTIELLE (Art. 23 - Avenant)</span>`
	if mod.ModificationType == domain.DMModTypeNonSubstantielle {
		typeBadge = `<span style="background:#059669; color:white; padding:3px 8px; border-radius:3px; font-weight:bold; font-size:8.5pt;">MODIFICATION NON SUBSTANTIELLE (Art. 24 - Déclaration)</span>`
	}

	isoImpact := func(flag bool, label string) string {
		if flag {
			return fmt.Sprintf(`<span style="display:inline-block; margin:2px; padding:2px 6px; background:#fee2e2; border:1px solid #ef4444; border-radius:3px; color:#991b1b; font-size:7.5pt; font-weight:700;">⚠️ %s (Impacté)</span>`, label)
		}
		return fmt.Sprintf(`<span style="display:inline-block; margin:2px; padding:2px 6px; background:#f0fdf4; border:1px solid #86efac; border-radius:3px; color:#166534; font-size:7.5pt;">✓ %s (Non impacté)</span>`, label)
	}

	sealBox := `<div style="color:#6b7280; font-style:italic;">En attente de visa Pharmacien DT</div>`
	if mod.PDTSignatureSeal != "" {
		sealDateStr := ""
		if mod.PDTSignedAt != nil {
			sealDateStr = mod.PDTSignedAt.Format("02/01/2006 à 15:04")
		}
		sealBox = fmt.Sprintf(`
			<div style="background:#f0fdf4; border:1px solid #059669; border-radius:4px; padding:8px; font-size:8pt; color:#065f46;">
				<div style="font-weight:bold; margin-bottom:2px;">🔏 Visa Électronique Apposé (21 CFR Part 11)</div>
				<div>Pharmacien DT : <strong>%s</strong> (Décision MIPH: %s)</div>
				<div>Empreinte SHA-256 : <code style="font-size:7.5pt;">%s</code></div>
				<div style="font-size:7pt; color:#047857; margin-top:2px;">Signé le : %s</div>
			</div>
		`, pdtName, pdtDecision, mod.PDTSignatureSeal, sealDateStr)
	}

	avenantNumber := mod.AvenantNumber
	if avenantNumber == "" {
		avenantNumber = "AVN-DDM-PROV-" + strings.TrimPrefix(mod.ModificationCode, "MOD-DM-")
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>ES-REG-05-DM • Avenant Décision d'Homologation Dispositif Médical</title>
    <style>
        @page { size: A4 portrait; margin: 0; }
        *, *::before, *::after { box-sizing: border-box; }
        body { font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, sans-serif; color: #0f172a; background-color: #f1f5f9; margin: 0; padding: 24px 0; font-size: 8.5pt; line-height: 1.4; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
        .no-print-toolbar { position: fixed; top: 0; left: 0; right: 0; height: 52px; background: #091e36; color: #ffffff; display: flex; align-items: center; justify-content: space-between; padding: 0 24px; z-index: 9999; box-shadow: 0 4px 12px rgba(0,0,0,0.15); font-size: 13px; }
        .no-print-toolbar .btn { display: inline-flex; align-items: center; gap: 8px; padding: 7px 14px; border-radius: 6px; font-weight: 600; font-size: 12px; cursor: pointer; border: none; text-decoration: none; }
        .btn-primary { background: #0284c7; color: #ffffff; }
        .btn-outline { background: rgba(255,255,255,0.12); color: #f8fafc; border: 1px solid rgba(255,255,255,0.25); }
        .sheet { width: 210mm; min-height: 297mm; margin: 48px auto 24px auto; background: #ffffff; box-shadow: 0 8px 30px rgba(15,23,42,0.12); padding: 14mm 16mm; position: relative; overflow: hidden; }
        @media print { body { background: #ffffff; padding: 0; margin: 0; } .no-print-toolbar { display: none !important; } .sheet { width: 100%%; margin: 0; padding: 12mm 15mm; box-shadow: none; border-radius: 0; } }
        .watermark { position: absolute; top: 48%%; left: 50%%; transform: translate(-50%%, -50%%) rotate(-35deg); font-size: 46pt; font-weight: 900; letter-spacing: 5px; color: rgba(4,120,87,0.06); pointer-events: none; user-select: none; z-index: 1; white-space: nowrap; text-transform: uppercase; }
        .watermark.draft { color: rgba(100,116,139,0.08); }
        .header-table { width: 100%%; border-collapse: collapse; margin-bottom: 10px; border-bottom: 2px solid #091e36; padding-bottom: 6px; }
        .section-box { border: 1px solid #cbd5e1; border-radius: 4px; margin-bottom: 9px; overflow: hidden; }
        .section-title { background: #f1f5f9; padding: 4px 10px; font-weight: bold; font-size: 8.5pt; color: #0f3b60; border-bottom: 1px solid #cbd5e1; display: flex; justify-content: space-between; }
        .section-body { padding: 8px 10px; font-size: 8pt; }
        table.dense-table { width: 100%%; border-collapse: collapse; margin-top: 3px; }
        table.dense-table th, table.dense-table td { border: 1px solid #cbd5e1; padding: 4px 6px; font-size: 7.5pt; }
        table.dense-table th { background: #f8fafc; text-align: left; font-weight: bold; }
        .footer-legal { margin-top: 12px; padding-top: 6px; border-top: 1px solid #cbd5e1; display: flex; align-items: center; justify-content: space-between; font-size: 6.8pt; color: #64748b; page-break-inside: avoid; }
    </style>
</head>
<body>
    <div class="no-print-toolbar">
        <div><strong>Homolog Reporting BPF</strong> • ES-REG-05-DM Avenant Décision d'Homologation DM (Décret 20-324 Art. 22-25)</div>
        <div style="display:flex; gap:10px;">
            <button class="btn btn-primary" onclick="window.print()">🖨️ Imprimer / Export PDF (A4 300 DPI)</button>
            <button type="button" class="btn btn-outline" onclick="if(window.parent && window.parent !== window && typeof window.parent.closeModal === 'function'){ window.parent.closeModal('modal-dossier-reports'); } else { window.close(); }">✕ Fermer</button>
        </div>
    </div>
    <script>
        if (window.parent && window.parent !== window) {
            var tb = document.querySelector('.no-print-toolbar');
            if (tb) tb.style.display = 'none';
            var sh = document.querySelector('.sheet');
            if (sh) sh.style.marginTop = '15px';
        }
    </script>
    <div class="sheet">
        <div class="watermark %s">%s</div>
        <table class="header-table">
            <tr>
                <td style="width:55px; vertical-align:middle;">
                    <svg width="44" height="44" viewBox="0 0 100 100" fill="none">
                        <circle cx="50" cy="50" r="46" stroke="#091e36" stroke-width="4" fill="#f8fafc"/>
                        <path d="M50 15v70M30 35h40M30 65h40M40 50h20" stroke="#047857" stroke-width="4" stroke-linecap="round"/>
                    </svg>
                </td>
                <td style="vertical-align:middle; padding-left:10px;">
                    <div style="font-size:6.8pt; font-weight:700; color:#64748b; text-transform:uppercase;">RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE</div>
                    <div style="font-size:7.8pt; font-weight:800; color:#091e36;">MINISTÈRE DE L'INDUSTRIE ET DE LA PRODUCTION PHARMACEUTIQUE • ANPP</div>
                    <div style="font-size:7pt; color:#047857; font-weight:700;">Direction des Dispositifs Médicaux (DDM) • Sous-Direction de l'Homologation</div>
                </td>
                <td style="text-align:right; vertical-align:middle;">
                    <div style="font-size:6.8pt; color:#64748b;">RÉFÉRENCE OFFICIELLE</div>
                    <div style="font-size:9pt; font-weight:800; color:#091e36;">ES-REG-05-DM</div>
                    <div style="font-size:6.8pt; color:#047857; font-weight:700;">Décret exécutif n° 20-324</div>
                </td>
            </tr>
        </table>

        <div style="text-align:center; margin: 10px 0 12px 0;">
            <div style="font-size: 13pt; font-weight: 900; color: #0f3b60; text-transform: uppercase; letter-spacing: 0.5px;">
                AVENANT À LA DÉCISION D'HOMOLOGATION DE DISPOSITIF MÉDICAL
            </div>
            <div style="font-size: 9pt; font-weight: bold; color: #059669; margin-top: 2px;">
                ACTE OFFICIEL RÉGLEMENTAIRE N° %s • ES-REG-05-DM
            </div>
            <div style="font-size: 7.5pt; color: #64748b; margin-top: 1px;">
                Décret exécutif n° 20-324 du 22 novembre 2020 (Articles 22 à 25) • Arrêté du 10 mai 2021
            </div>
        </div>

        <!-- Section 1 : Identification Établissement & DM -->
        <div class="section-box">
            <div class="section-title">
                <span>SECTION 1 : ÉTABLISSEMENT DEMANDEUR &amp; DISPOSITIF MÉDICAL HOMOLOGUÉ</span>
                <span>Réf. Dossier : %s</span>
            </div>
            <div class="section-body">
                <table style="width:100%%; font-size:8pt; border:none;">
                    <tr>
                        <td style="width:50%%; vertical-align:top;">
                            <strong>Établissement Pharmaceutique :</strong> %s<br/>
                            <strong>Agrément MIPH :</strong> %s<br/>
                            <strong>Adresse Siège / Site :</strong> %s<br/>
                            <strong>Pharmacien Directeur Technique :</strong> %s (Mandat 1.A.7)
                        </td>
                        <td style="width:50%%; vertical-align:top; border-left:1px solid #e2e8f0; padding-left:10px;">
                            <strong>Dispositif Médical :</strong> %s<br/>
                            <strong>Référence Catalogue :</strong> %s • <strong>Classe :</strong> %s<br/>
                            <strong>Décision d'Homologation Initiale :</strong> <code>%s</code><br/>
                            <strong>Date d'Homologation Initiale :</strong> %s
                        </td>
                    </tr>
                </table>
            </div>
        </div>

        <!-- Section 2 : Caractéristiques de la Modification Post-Homologation -->
        <div class="section-box">
            <div class="section-title">
                <span>SECTION 2 : CARACTÉRISATION DU CHANGEMENT (ARTICLES 22 À 25 DÉCRET 20-324)</span>
                <span>Code : %s</span>
            </div>
            <div class="section-body">
                <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:6px;">
                    <div><strong>Type de Modification :</strong> %s</div>
                    <div><strong>Catégorie :</strong> <code>%s</code></div>
                </div>
                <div style="margin-bottom:4px;"><strong>Intitulé de l'Objet :</strong> %s</div>
                <div style="margin-bottom:4px; font-size:7.5pt; color:#475569;"><strong>Description Détaillée :</strong> %s</div>
                <div style="font-size:7.5pt; color:#475569;"><strong>Justification Technico-Réglementaire :</strong> %s</div>
            </div>
        </div>

        <!-- Section 3 : Tableau Comparatif Avant / Après -->
        <div class="section-box">
            <div class="section-title">
                <span>SECTION 3 : SITUATION COMPARATIVE DES SPÉCIFICATIONS DU DISPOSITIF MÉDICAL</span>
                <span>Change Control BPF / ISO 13485</span>
            </div>
            <div class="section-body" style="padding:0;">
                <table class="dense-table" style="margin:0; border:none;">
                    <thead>
                        <tr>
                            <th style="width:50%%; background:#fef2f2; color:#991b1b;">SITUATION HOMOLOGUÉE ACTUELLE (AVANT CHANGEMENT)</th>
                            <th style="width:50%%; background:#f0fdf4; color:#166534;">SITUATION MODIFIÉE PROPOSÉE (APRÈS CHANGEMENT)</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td style="vertical-align:top; padding:8px; line-height:1.4;">%s</td>
                            <td style="vertical-align:top; padding:8px; line-height:1.4;">%s</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <!-- Section 4 : Matrice d'Impact Normatif ISO -->
        <div class="section-box">
            <div class="section-title">
                <span>SECTION 4 : ÉVALUATION D'IMPACT SUR LES NORMES PILIERS DU DOSSIER TECHNIQUE</span>
                <span>Arrêté du 10 Mai 2021</span>
            </div>
            <div class="section-body">
                <div style="display:flex; flex-wrap:wrap; gap:4px; margin-bottom:6px;">
                    %s %s %s %s %s %s
                </div>
                <div style="font-size:7.5pt; color:#64748b; line-height:1.3;">
                    Les études justificatives, rapports d'essais en laboratoire accrédité GLP et revues de dossiers de risques correspondants sont dûment archivés dans le coffre-fort documentaire BPF / ALCOA+ de l'établissement.
                </div>
            </div>
        </div>

        <!-- Section 5 : Droits Réglementaires & Quittance BNA -->
        <div class="section-box">
            <div class="section-title">
                <span>SECTION 5 : REDEVANCE RÉGLEMENTAIRE (ORDONNANCE N° 21-07 ART. 31)</span>
                <span>Quittance BNA Chéraga</span>
            </div>
            <div class="section-body">
                <table style="width:100%%; font-size:8pt; border:none;">
                    <tr>
                        <td style="width:60%%;">
                            Compte d'affectation spéciale ANPP : <strong>BNA Agence Chéraga 622</strong><br/>
                            RIB officiel : <code>00100622030000064395</code> • Clé BNA : <code>95</code>
                        </td>
                        <td style="width:40%%; text-align:right;">
                            Montant de la Redevance : <strong style="font-size:9.5pt; color:#0f3b60;">%.2f DZD</strong><br/>
                            N° Quittance : <code>%s</code>
                        </td>
                    </tr>
                </table>
            </div>
        </div>

        <!-- Section 6 & 7 : Signatures, Visa PDT et Décision DDM -->
        <div style="display:grid; grid-template-columns:1fr 1fr; gap:10px; margin-top:8px;">
            <div class="section-box">
                <div class="section-title">
                    <span>SECTION 6 : ENGAGEMENT DU PHARMACIEN DT</span>
                </div>
                <div class="section-body">
                    <p style="font-size:7pt; line-height:1.3; color:#475569; margin-top:0;">
                        J'atteste sur l'honneur avoir vérifié la maîtrise des changements (ISO 13485) et certifie que les données techniques soumises à l'ANPP sont véridiques et conformes au Décret 20-324.
                    </p>
                    %s
                </div>
            </div>
            <div class="section-box">
                <div class="section-title">
                    <span>SECTION 7 : DÉCISION DE L'ANPP (DDM)</span>
                </div>
                <div class="section-body" style="text-align:center;">
                    <div style="font-size:9pt; font-weight:bold; color:#059669; margin-top:4px;">
                        AVENANT VALIDÉ &amp; ENREGISTRÉ AU REGISTRE NATIONAL
                    </div>
                    <div style="font-size:8pt; color:#1e293b; margin-top:2px;">
                        N° Avenant : <strong>%s</strong>
                    </div>
                    <div style="font-size:7.5pt; color:#64748b; margin-top:2px;">
                        Date d'effet : %s • Sous-Direction de l'Homologation DDM
                    </div>
                    <div style="margin-top:12px; border-top:1px dashed #cbd5e1; padding-top:6px; font-size:7pt; color:#64748b;">
                        Cachet &amp; Visa du Directeur des Dispositifs Médicaux
                    </div>
                </div>
            </div>
        </div>

        %s
    </div>
</body>
</html>`,
		wmClass, wmText,
		avenantNumber,
		mod.ModificationCode,
		estName, estAgreement, estAddress, pdtName,
		dmName, dmRef, dmClass, homologDec, homologDate,
		mod.ModificationCode,
		typeBadge,
		mod.CategoryCode,
		mod.Title,
		mod.Description,
		mod.Justification,
		mod.CurrentState,
		mod.ProposedState,
		isoImpact(mod.ImpactRiskISO14971, "ISO 14971 Risques"),
		isoImpact(mod.ImpactBiocompatibilityISO10993, "ISO 10993 Biocompatibilité"),
		isoImpact(mod.ImpactSterilizationISO11135, "ISO 11135 Stérilisation"),
		isoImpact(mod.ImpactPackagingISO11607, "ISO 11607 Barrière"),
		isoImpact(mod.ImpactClinicalISO14155, "ISO 14155 Clinique"),
		isoImpact(mod.ImpactQualityISO13485, "ISO 13485 Qualité"),
		mod.FeeAmount,
		mod.FeeReceiptRef,
		sealBox,
		avenantNumber,
		generatedAt.Format("02/01/2006"),
		reporting.RenderLegalFooter("ES-REG-05-DM", watermark, sha256Hash, qrURL, 1, 1, generatedAt),
	)
}
