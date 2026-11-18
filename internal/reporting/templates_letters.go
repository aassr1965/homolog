package reporting

import (
	"fmt"
	"html"
	"time"

	"homolog/internal/domain"
)

// RenderCommitmentLetter1P8 génère la Lettre d'Engagement de Disponibilité Continue sur le Marché National (Module 1.P.8 / ES-ENG-1P8)
func RenderCommitmentLetter1P8(
	d *domain.Dossier,
	estName, estAgreement, estAddress, pdtName, pdtDecision string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, _, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(watermark)

	headerHTML := RenderInstitutionalHeader(estName, estAgreement, estAddress, pdtName, pdtDecision, "Module 1.P.8 • Engagement Solennel d'Approvisionnement Continu")
	footerHTML := RenderLegalFooter("ES-ENG-1P8", watermark, sha256Hash, qrURL, 1, 1, generatedAt)
	signHTML := RenderSignaturesBlock(pdtName, pdtDecision, d.PDTSignedAt, "Direction de la Production & de la Régulation MIPH")

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">Module 1.P.8 — Engagement de Disponibilité Continue sur le Marché National</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-card" style="margin-bottom:14px;">
            <div class="info-card-header">Identification de la Spécialité Pharmaceutique Soumise à l'ANPP</div>
            <div class="field-row"><div class="field-label">Dénomination Commerciale :</div><div class="field-value"><strong>%s</strong></div></div>
            <div class="field-row"><div class="field-label">Dénomination Commune (DCI) :</div><div class="field-value">%s</div></div>
            <div class="field-row"><div class="field-label">Dosage & Forme :</div><div class="field-value">%s — %s</div></div>
            <div class="field-row"><div class="field-label">Procédure CTD Homolog :</div><div class="field-value">%s (Réf. %s)</div></div>
        </div>

        <div class="info-card" style="margin-bottom:14px; background:#ffffff; border-left:4px solid #047857; padding:14px;">
            <div style="font-size:9.5pt; font-weight:800; color:#091e36; margin-bottom:10px; text-transform:uppercase;">
                Déclaration Solennelle sous la Foi du Serment Pharmaceutique
            </div>
            <div style="font-size:8.2pt; color:#1e293b; line-height:1.6; text-align:justify;">
                Je soussigné(e), <strong>%s</strong>, agissant en qualité de <strong>Pharmacien Directeur Technique Titulaire</strong> 
                de l'établissement pharmaceutique <strong>%s</strong> (titulaire de l'Agrément Ministériel N° <strong>%s</strong>), 
                conformément aux dispositions du Décret exécutif n° 20-325 du 22 novembre 2020 et aux exigences de la Loi n° 18-11 relative à la santé :
            </div>
            <div style="margin-top:10px; font-size:8pt; color:#0f172a; line-height:1.55;">
                <ol style="margin:0 0 0 16px; padding:0;">
                    <li style="margin-bottom:6px;">
                        <strong>Garantie d'Approvisionnement Ininterrompu :</strong> M'engage formellement à assurer une disponibilité continue et régulière 
                        de la spécialité pharmaceutique précitée sur l'ensemble du territoire national algérien dès l'octroi de la Décision d'Enregistrement (AMM).
                    </li>
                    <li style="margin-bottom:6px;">
                        <strong>Stock Stratégique de Sécurité :</strong> M'engage à constituer et maintenir en permanence un stock de sécurité inaltérable 
                        équivalant à un minimum de <strong>trois (03) mois de consommation nationale moyenne</strong>, conformément aux normes de gestion des risques sanitaires.
                    </li>
                    <li style="margin-bottom:6px;">
                        <strong>Notification Préalable d'Alerte :</strong> M'engage à notifier sans délai et par écrit à l'Agence Nationale des Produits Pharmaceutiques (ANPP) 
                        et à la Direction de la Production Pharmaceutique (MIPH), au moins <strong>six (06) mois à l'avance</strong>, toute circonstance exceptionnelle ou tout risque prévisible de rupture d'approvisionnement.
                    </li>
                    <li>
                        <strong>Priorité Sanitaire Nationale :</strong> Déclare que les volumes de fabrication locale seront prioritairement alloués aux besoins hospitaliers 
                        (Pharmacie Centrale des Hôpitaux - PCH) et au réseau officinal public et privé.
                    </li>
                </ol>
            </div>
        </div>

        <div style="background:#f8fafc; border:1px solid #cbd5e1; border-radius:4px; padding:10px; font-size:7.5pt; color:#475569; margin-bottom:14px;">
            <strong>Base Réglementaire :</strong> Guide CTD ANPP Juillet 2025 • Module 1.P.8 Spécificités Régionales Algérie • Décret exécutif n° 20-325 • 
            Acte juridiquement opposable scellé dans le coffre-fort BPF Sanad Vault sous empreinte cryptographique SHA-256.
        </div>

        %s
        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(dosage), html.EscapeString(form), html.EscapeString(string(d.ProcedureCode)), html.EscapeString(d.DossierCode), html.EscapeString(pdtName), html.EscapeString(estName), html.EscapeString(estAgreement), signHTML, footerHTML)

	return RenderDocumentWrap("Module 1.P.8 — Engagement de Disponibilité Continue", content)
}

// RenderCommitmentLetter1P9 génère la Lettre d'Engagement de Pharmacovigilance & Notification Immédiate (Module 1.P.9 / ES-ENG-1P9)
func RenderCommitmentLetter1P9(
	d *domain.Dossier,
	estName, estAgreement, estAddress, pdtName, pdtDecision string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, _, _ := extractProductDetails(d)
	wmHTML := RenderWatermark(watermark)

	headerHTML := RenderInstitutionalHeader(estName, estAgreement, estAddress, pdtName, pdtDecision, "Module 1.P.9 • Engagement Solennel de Pharmacovigilance")
	footerHTML := RenderLegalFooter("ES-ENG-1P9", watermark, sha256Hash, qrURL, 1, 1, generatedAt)
	signHTML := RenderSignaturesBlock(pdtName, pdtDecision, d.PDTSignedAt, "Centre National de Pharmacovigilance & Matériovigilance (CNPM)")

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">Module 1.P.9 — Engagement de Pharmacovigilance & Notification des Effets Indésirables</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-card" style="margin-bottom:14px;">
            <div class="info-card-header">Spécialité Pharmaceutique Soumise à Surveillance Continue</div>
            <div class="field-row"><div class="field-label">Spécialité :</div><div class="field-value"><strong>%s</strong></div></div>
            <div class="field-row"><div class="field-label">D.C.I. :</div><div class="field-value">%s</div></div>
            <div class="field-row"><div class="field-label">Forme & Dosage :</div><div class="field-value">%s — %s</div></div>
            <div class="field-row"><div class="field-label">Dossier CTD :</div><div class="field-value">%s (Procédure %s)</div></div>
        </div>

        <div class="info-card" style="margin-bottom:14px; background:#ffffff; border-left:4px solid #0284c7; padding:14px;">
            <div style="font-size:9.5pt; font-weight:800; color:#091e36; margin-bottom:10px; text-transform:uppercase;">
                Engagement Formel de Pharmacovigilance & Gestion des Risques
            </div>
            <div style="font-size:8.2pt; color:#1e293b; line-height:1.6; text-align:justify;">
                Je soussigné(e), <strong>%s</strong>, <strong>Pharmacien Directeur Technique</strong> de l'établissement <strong>%s</strong>, 
                atteste sur l'honneur et sous les peines prévues par la législation sanitaire algérienne :
            </div>
            <div style="margin-top:10px; font-size:8pt; color:#0f172a; line-height:1.55;">
                <ol style="margin:0 0 0 16px; padding:0;">
                    <li style="margin-bottom:6px;">
                        <strong>Dispositif Permanent 24/7 :</strong> Qu'un système de pharmacovigilance conforme aux Lignes Directrices de l'ANPP et aux Bonnes Pratiques de Pharmacovigilance (BPPV) 
                        est opérationnel et activé en continu, doté d'une ligne d'urgence joignable sans interruption.
                    </li>
                    <li style="margin-bottom:6px;">
                        <strong>Personne Responsable Qualifiée (QPPV) :</strong> Qu'un Pharmacien Responsable de Pharmacovigilance dûment habilité et qualifié 
                        a été désigné au sein de notre structure, garantissant le recueil, la validation, l'évaluation et l'archivage de toutes les notifications d'effets indésirables.
                    </li>
                    <li style="margin-bottom:6px;">
                        <strong>Délais Impératifs de Télédéclaration :</strong> M'engage formellement à notifier au Centre National de Pharmacovigilance et Matériovigilance (CNPM) 
                        ainsi qu'à la Sous-Direction de la Vigilance ANPP :
                        <ul style="margin:4px 0 0 14px; padding:0; font-size:7.5pt; color:#334155;">
                            <li>Tout effet indésirable grave inattendu dans un délai maximal de <strong>quinze (15) jours calendaires</strong>.</li>
                            <li>Toute nouvelle donnée de sécurité susceptible de modifier le rapport bénéfice/risque dans les <strong>quarante-huit (48) heures</strong>.</li>
                            <li>Toute mesure d'interdiction, de retrait ou de rappel de lot survenue à l'étranger <strong>immédiatement et sans délai</strong>.</li>
                        </ul>
                    </li>
                    <li>
                        <strong>Rapports Périodiques de Sécurité (PSUR / PBRER) :</strong> M'engage à soumettre les rapports périodiques actualisés de sécurité 
                        selon le calendrier réglementaire fixé par l'ANPP.
                    </li>
                </ol>
            </div>
        </div>

        <div style="background:#f8fafc; border:1px solid #cbd5e1; border-radius:4px; padding:10px; font-size:7.5pt; color:#475569; margin-bottom:14px;">
            <strong>Cadre Légal :</strong> Loi sanitaire n° 18-11 • Décret exécutif n° 21-82 • Recommandations ICH E2E / GVP • 
            Document légal visé sous double authentification 2FA Pharmacien DT et scellé à valeur probante.
        </div>

        %s
        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(dosage), html.EscapeString(form), html.EscapeString(d.DossierCode), html.EscapeString(string(d.ProcedureCode)), html.EscapeString(pdtName), html.EscapeString(estName), signHTML, footerHTML)

	return RenderDocumentWrap("Module 1.P.9 — Engagement de Pharmacovigilance", content)
}

// RenderCommitmentLetter1H10 génère la Déclaration Solennelle de Non-Modification (Module 1.H.10 / ES-ENG-1H10)
func RenderCommitmentLetter1H10(
	d *domain.Dossier,
	estName, estAgreement, estAddress, pdtName, pdtDecision string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	generatedAt time.Time,
) string {
	prodName, dci, dosage, form, _, ammNumber := extractProductDetails(d)
	wmHTML := RenderWatermark(watermark)

	headerHTML := RenderInstitutionalHeader(estName, estAgreement, estAddress, pdtName, pdtDecision, "Module 1.H.10 • Déclaration Solennelle de Non-Modification (Renouvellement)")
	footerHTML := RenderLegalFooter("ES-ENG-1H10", watermark, sha256Hash, qrURL, 1, 1, generatedAt)
	signHTML := RenderSignaturesBlock(pdtName, pdtDecision, d.PDTSignedAt, "Direction de l'Enregistrement des Produits Pharmaceutiques (DEPP)")

	content := fmt.Sprintf(`
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">Module 1.H.10 — Déclaration Solennelle de Non-Modification du Dossier d'AMM</div>
            <div class="banner-code">%s</div>
        </div>

        <div class="info-card" style="margin-bottom:14px;">
            <div class="info-card-header">Dossier de Demande de Renouvellement Quinquennal (Procédure %s)</div>
            <div class="field-row"><div class="field-label">Spécialité :</div><div class="field-value"><strong>%s</strong></div></div>
            <div class="field-row"><div class="field-label">D.C.I. :</div><div class="field-value">%s</div></div>
            <div class="field-row"><div class="field-label">Forme & Dosage :</div><div class="field-value">%s — %s</div></div>
            <div class="field-row"><div class="field-label">N° Décision d'AMM Initiale :</div><div class="field-value" style="font-family:monospace; font-weight:700;">%s</div></div>
        </div>

        <div class="info-card" style="margin-bottom:14px; background:#ffffff; border-left:4px solid #7e22ce; padding:14px;">
            <div style="font-size:9.5pt; font-weight:800; color:#091e36; margin-bottom:10px; text-transform:uppercase;">
                Attestation d'Inaltérabilité & de Conformité Bit-à-Bit
            </div>
            <div style="font-size:8.2pt; color:#1e293b; line-height:1.6; text-align:justify;">
                Je soussigné(e), <strong>%s</strong>, <strong>Pharmacien Directeur Technique</strong> de l'établissement demandeur <strong>%s</strong>, 
                atteste et certifie solennellement sur l'honneur devant l'Agence Nationale des Produits Pharmaceutiques :
            </div>
            <div style="margin-top:10px; font-size:8pt; color:#0f172a; line-height:1.55;">
                <ol style="margin:0 0 0 16px; padding:0;">
                    <li style="margin-bottom:6px;">
                        <strong>Intégrité Formule & Procédé :</strong> Qu'aucune modification d'aucune nature (qualitative, quantitative, excipients, procédé de fabrication, 
                        spécifications de contrôle analytique, ou sites de production) n'a été introduite dans le dossier de la spécialité depuis la date de délivrance 
                        de la décision d'enregistrement précédente, à l'exception exclusive des variations formellement autorisées et visées par l'ANPP.
                    </li>
                    <li style="margin-bottom:6px;">
                        <strong>Historique Complet des Variations :</strong> Que l'intégralité des modifications survenues au cours du quinquennat a fait l'objet de notifications 
                        ou demandes d'approbation préalables conformes aux Lignes Directrices LD-DEPP07/V01, et que leurs accusés de réception ou décisions d'approbation 
                        figurent sans omission dans le présent dossier de renouvellement.
                    </li>
                    <li>
                        <strong>Qualité & Stabilité Maintenues :</strong> Que les données de stabilité en temps réel (Zone Climatique IVb : 30°C / 75%%%% HR) confirment la durée 
                        de conservation approuvée sans aucune dérive de qualité ni non-conformité analytique.
                    </li>
                </ol>
            </div>
        </div>

        <div style="background:#f8fafc; border:1px solid #cbd5e1; border-radius:4px; padding:10px; font-size:7.5pt; color:#475569; margin-bottom:14px;">
            <strong>Portée Juridique :</strong> Arrêté ministériel du 16 décembre 2020 fixant les modalités de renouvellement de l'enregistrement • 
            Toute fausse déclaration expose le signataire et l'établissement aux sanctions pénales et administratives prévues par la Loi 18-11 
            (retrait d'AMM, suspension d'agrément, poursuites judiciaires).
        </div>

        %s
        %s
    </div>
	`, wmHTML, headerHTML, html.EscapeString(d.DossierCode), html.EscapeString(string(d.ProcedureCode)), html.EscapeString(prodName), html.EscapeString(dci), html.EscapeString(dosage), html.EscapeString(form), html.EscapeString(ammNumber), html.EscapeString(pdtName), html.EscapeString(estName), signHTML, footerHTML)

	return RenderDocumentWrap("Module 1.H.10 — Déclaration Solennelle de Non-Modification", content)
}
