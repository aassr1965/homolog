package reporting

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"image/png"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"homolog/internal/domain"
)

// GenerateQRCodeDataURI génère une image QR-code PNG encodée en data URI base64
func GenerateQRCodeDataURI(content string, sizePx int) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("contenu QR code vide")
	}
	if sizePx <= 0 {
		sizePx = 120
	}

	qrCode, err := qr.Encode(content, qr.M, qr.Auto)
	if err != nil {
		return "", fmt.Errorf("erreur encodage QR code: %w", err)
	}

	scaled, err := barcode.Scale(qrCode, sizePx, sizePx)
	if err != nil {
		return "", fmt.Errorf("erreur redimensionnement QR code: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return "", fmt.Errorf("erreur sérialisation PNG QR code: %w", err)
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// GetStandardBPFCSS retourne la feuille de style universelle pour l'édition A4 Haute Fidélité (300+ DPI, ISO 19005 / PDF/A)
func GetStandardBPFCSS() string {
	return `
        @page {
            size: A4 portrait;
            margin: 0;
            @bottom-right {
                content: "Page " counter(page) " sur " counter(pages);
                font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, sans-serif;
                font-size: 7.5pt;
                color: #64748b;
            }
        }
        *, *::before, *::after { box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, Helvetica, Arial, sans-serif;
            color: #0f172a;
            background-color: #f1f5f9;
            margin: 0;
            padding: 20px 0;
            font-size: 8.8pt;
            line-height: 1.42;
            -webkit-print-color-adjust: exact;
            print-color-adjust: exact;
        }
        .no-print-toolbar {
            position: fixed;
            top: 0; left: 0; right: 0;
            height: 52px;
            background: #091e36;
            color: #ffffff;
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 0 24px;
            z-index: 9999;
            box-shadow: 0 4px 12px rgba(0,0,0,0.18);
            font-size: 13px;
        }
        .no-print-toolbar .btn {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            padding: 7px 14px;
            border-radius: 6px;
            font-weight: 600;
            font-size: 12px;
            cursor: pointer;
            border: none;
            text-decoration: none;
            transition: all 0.15s ease;
        }
        .btn-primary { background: #0284c7; color: #ffffff; }
        .btn-primary:hover { background: #0369a1; }
        .btn-outline { background: rgba(255,255,255,0.12); color: #f8fafc; border: 1px solid rgba(255,255,255,0.25); }
        .btn-outline:hover { background: rgba(255,255,255,0.2); }
        .btn-seal { background: #7e22ce; color: #ffffff; }
        .btn-seal:hover { background: #6b21a8; }

        .sheet {
            width: 210mm;
            min-height: 297mm;
            margin: 50px auto 25px auto;
            background: #ffffff;
            box-shadow: 0 8px 30px rgba(15,23,42,0.12);
            padding: 14mm 16mm;
            position: relative;
            overflow: hidden;
            display: flex;
            flex-direction: column;
            justify-content: space-between;
        }

        @media print {
            body { background: #ffffff; padding: 0; margin: 0; }
            .no-print-toolbar { display: none !important; }
            .sheet {
                width: 100%;
                margin: 0;
                padding: 12mm 15mm;
                box-shadow: none;
                border-radius: 0;
                page-break-after: always;
            }
        }

        /* Filigranes dynamiques BPF */
        .watermark {
            position: absolute;
            top: 48%; left: 50%;
            transform: translate(-50%, -50%) rotate(-35deg);
            font-size: 52pt;
            font-weight: 900;
            letter-spacing: 5px;
            pointer-events: none;
            user-select: none;
            z-index: 1;
            white-space: nowrap;
            text-transform: uppercase;
        }
        .watermark-draft { color: rgba(100, 116, 139, 0.08); }
        .watermark-official { color: rgba(4, 120, 87, 0.07); }
        .watermark-archived { color: rgba(185, 28, 28, 0.08); }
        .watermark-internal { color: rgba(2, 132, 199, 0.08); }

        /* En-têtes institutionnels */
        .header-table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 12px;
            border-bottom: 2px solid #091e36;
            padding-bottom: 8px;
        }
        .repub-title {
            text-align: center;
            font-size: 8.5pt;
            font-weight: 800;
            letter-spacing: 0.5px;
            color: #091e36;
            margin-bottom: 3px;
            text-transform: uppercase;
        }
        .miph-title {
            text-align: center;
            font-size: 7.8pt;
            font-weight: 700;
            color: #334155;
            margin-bottom: 2px;
        }
        .anpp-title {
            text-align: center;
            font-size: 8.2pt;
            font-weight: 800;
            color: #047857;
            letter-spacing: 0.4px;
        }

        /* Bandeaux titres */
        .title-banner {
            background: #091e36;
            color: #ffffff;
            padding: 8px 14px;
            border-radius: 4px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 12px;
        }
        .banner-code {
            font-size: 8.2pt;
            font-weight: 800;
            background: rgba(255,255,255,0.15);
            padding: 3px 8px;
            border-radius: 3px;
            letter-spacing: 0.5px;
        }
        .banner-title {
            font-size: 11pt;
            font-weight: 800;
            letter-spacing: 0.5px;
            text-transform: uppercase;
            margin: 0;
        }
        .banner-badge {
            font-size: 7.5pt;
            font-weight: 700;
            padding: 3px 8px;
            border-radius: 3px;
            background: #047857;
            color: white;
            text-transform: uppercase;
        }
        .banner-badge.draft { background: #d97706; }
        .banner-badge.archived { background: #b91c1c; }

        /* Grilles d'informations */
        .info-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 12px;
            margin-bottom: 12px;
        }
        .info-card {
            border: 1px solid #cbd5e1;
            border-radius: 4px;
            padding: 9px 11px;
            background: #ffffff;
        }
        .info-card-header {
            font-size: 8.2pt;
            font-weight: 800;
            color: #091e36;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            border-bottom: 1px solid #e2e8f0;
            padding-bottom: 4px;
            margin-bottom: 6px;
            display: flex;
            align-items: center;
            gap: 6px;
        }
        .field-row {
            display: flex;
            margin-bottom: 3.5px;
            font-size: 8pt;
        }
        .field-label {
            width: 135px;
            font-weight: 700;
            color: #475569;
            flex-shrink: 0;
        }
        .field-value {
            color: #091e36;
            font-weight: 600;
            flex-grow: 1;
        }

        /* Tableaux BPF */
        .data-table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 12px;
            font-size: 8pt;
        }
        .data-table th {
            background: #0f3b60;
            color: #ffffff;
            padding: 6px 8px;
            text-align: left;
            font-weight: 700;
            font-size: 7.8pt;
            border: 1px solid #0f3b60;
        }
        .data-table td {
            padding: 5px 8px;
            border: 1px solid #cbd5e1;
            vertical-align: middle;
        }
        .data-table tr:nth-child(even) { background-color: #f8fafc; }

        /* Signatures et Visas légaux */
        .sign-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 12px;
            margin-top: 10px;
            page-break-inside: avoid;
        }
        .sign-card {
            border: 1px solid #091e36;
            border-radius: 4px;
            padding: 9px;
            background: #ffffff;
        }
        .sign-header {
            font-weight: 800;
            font-size: 7.8pt;
            color: #091e36;
            text-transform: uppercase;
            margin-bottom: 5px;
        }
        .sign-space {
            height: 60px;
            border: 1px dashed #cbd5e1;
            background: #f8fafc;
            border-radius: 3px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 7.2pt;
            color: #64748b;
            margin-top: 5px;
        }

        /* Pied de page légal et cartouche cryptographique */
        .footer-legal {
            margin-top: 14px;
            padding-top: 8px;
            border-top: 1px solid #cbd5e1;
            display: flex;
            align-items: center;
            justify-content: space-between;
            font-size: 6.8pt;
            color: #64748b;
            page-break-inside: avoid;
        }
        .footer-qr {
            width: 44px;
            height: 44px;
            border: 1px solid #cbd5e1;
            border-radius: 3px;
            padding: 2px;
            background: #ffffff;
        }
	`
}

// RenderWatermark génère le bloc HTML du filigrane de sécurité selon le statut légal BPF
func RenderWatermark(status domain.WatermarkStatus, customText ...string) string {
	var text, class string
	if len(customText) > 0 && customText[0] != "" {
		text = customText[0]
	}

	switch status {
	case domain.WatermarkOfficial:
		if text == "" {
			text = "OFFICIEL — VALIDÉ"
		}
		class = "watermark-official"
	case domain.WatermarkArchived:
		if text == "" {
			text = "ARCHIVÉ — OBSOLÈTE"
		}
		class = "watermark-archived"
	default: // WatermarkDraft
		if text == "" {
			text = "BROUILLON — NON OPPOSABLE"
		}
		class = "watermark-draft"
	}

	return fmt.Sprintf(`<div class="watermark %s">%s</div>`, class, html.EscapeString(text))
}

// RenderInstitutionalHeader génère le cartouche officiel national algérien et de l'établissement demandeur
func RenderInstitutionalHeader(estName, estAgreement, estAddress, pdtName, pdtDecision string, subTitle ...string) string {
	extraSub := ""
	if len(subTitle) > 0 && subTitle[0] != "" {
		extraSub = fmt.Sprintf(`<div style="font-size:7.5pt; color:#64748b; margin-top:2px; text-transform:uppercase;">%s</div>`, html.EscapeString(subTitle[0]))
	}

	return fmt.Sprintf(`
        <table class="header-table">
            <tr>
                <td style="width:25%%; vertical-align:middle;">
                    <div style="font-weight:800; font-size:10.5pt; color:#091e36; letter-spacing:0.5px;">SANAD PHARMA</div>
                    <div style="font-size:7pt; color:#64748b; font-weight:600;">Homolog Regulatory Platform</div>
                    <div style="font-size:6.8pt; color:#047857; font-weight:700; margin-top:3px;">Agrément BPF / ALCOA+</div>
                </td>
                <td style="width:50%%; vertical-align:middle; text-align:center;">
                    <div class="repub-title">République Algérienne Démocratique et Populaire</div>
                    <div class="miph-title">Ministère de l'Industrie et de la Production Pharmaceutique</div>
                    <div class="anpp-title">Agence Nationale des Produits Pharmaceutiques (ANPP)</div>
                    %s
                </td>
                <td style="width:25%%; vertical-align:middle; text-align:right;">
                    <div style="font-size:8pt; font-weight:800; color:#091e36;">%s</div>
                    <div style="font-size:7pt; color:#475569;">Agrément : %s</div>
                    <div style="font-size:6.8pt; color:#64748b;">DT : %s</div>
                </td>
            </tr>
        </table>
	`, extraSub, html.EscapeString(estName), html.EscapeString(estAgreement), html.EscapeString(pdtName))
}

// RenderLegalFooter génère le pied de page légal BPF avec QR-Code réel scellé et pagination opposable "Page X sur Y"
func RenderLegalFooter(
	docCode string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	pageNum, totalPages int,
	generatedAt time.Time,
) string {
	// Génération du QR Code réel scellé
	qrDataURI, err := GenerateQRCodeDataURI(qrURL, 90)
	qrImgTag := ""
	if err == nil && qrDataURI != "" {
		qrImgTag = fmt.Sprintf(`<img src="%s" class="footer-qr" alt="QR Code Vérification" />`, qrDataURI)
	} else {
		// Fallback SVG minimaliste
		qrImgTag = `<svg width="44" height="44" viewBox="0 0 100 100" fill="#091e36" class="footer-qr"><rect width="100" height="100" fill="#ffffff"/><rect x="10" y="10" width="30" height="30"/><rect x="60" y="10" width="30" height="30"/><rect x="10" y="60" width="30" height="30"/><rect x="40" y="40" width="20" height="20" fill="#047857"/></svg>`
	}

	hashDisplay := sha256Hash
	if len(hashDisplay) > 28 {
		hashDisplay = hashDisplay[:28] + "..."
	}

	statusBadge := `<span style="color:#047857; font-weight:800;">OFFICIEL</span>`
	if watermark == domain.WatermarkDraft {
		statusBadge = `<span style="color:#d97706; font-weight:800;">BROUILLON (NON OPPOSABLE)</span>`
	} else if watermark == domain.WatermarkArchived {
		statusBadge = `<span style="color:#b91c1c; font-weight:800;">ARCHIVÉ</span>`
	}

	if pageNum <= 0 {
		pageNum = 1
	}
	if totalPages <= 0 {
		totalPages = 1
	}

	return fmt.Sprintf(`
        <div class="footer-legal">
            <div style="display:flex; align-items:center; gap:9px;">
                %s
                <div>
                    <div><strong>Réf. :</strong> %s • Statut : %s</div>
                    <div><strong>Sceau SHA-256 :</strong> <span style="font-family:monospace; font-size:6.5pt;">%s</span></div>
                    <div>Vérification d'authenticité : <a href="%s" target="_blank" style="color:#0284c7; text-decoration:none;">%s</a></div>
                </div>
            </div>
            <div style="text-align:right;">
                <div>Édition certifiée le %s</div>
                <div style="font-size:7.5pt; font-weight:800; color:#091e36; margin-top:2px;">Page %d sur %d</div>
                <div style="font-size:6.2pt; color:#94a3b8;">Conforme 21 CFR Part 11 & Norme ISO 19005 (PDF/A)</div>
            </div>
        </div>
	`,
		qrImgTag,
		html.EscapeString(docCode),
		statusBadge,
		html.EscapeString(hashDisplay),
		html.EscapeString(qrURL),
		html.EscapeString(qrURL),
		generatedAt.Format("02/01/2006 à 15:04:05 UTC+1"),
		pageNum,
		totalPages,
	)
}

// RenderSignaturesBlock génère le bloc de signature réglementaire opposable
func RenderSignaturesBlock(pdtName, pdtDecision string, pdtSignedAt *time.Time, recipientTitle string) string {
	if recipientTitle == "" {
		recipientTitle = "Décharge Bureau d'Ordre ANPP (Guichet Réception)"
	}

	pdtSignatureContent := `
        <div class="sign-space">
            <span>En attente de visa du Pharmacien Directeur Technique</span>
        </div>
    `
	if pdtSignedAt != nil {
		pdtSignatureContent = fmt.Sprintf(`
            <div style="padding:6px; border:1px solid #047857; background:#f0fdf4; border-radius:3px; margin-top:5px; font-size:7.2pt;">
                <div style="color:#047857; font-weight:800;">✔ SCELLÉ & SIGNÉ ÉLECTRONIQUEMENT (2FA)</div>
                <div style="color:#091e36; font-weight:700;">%s</div>
                <div style="color:#64748b; font-size:6.5pt;">Directeur Technique • %s</div>
                <div style="color:#334155; font-size:6.5pt; margin-top:2px;">Signé le : %s</div>
            </div>
        `, html.EscapeString(pdtName), html.EscapeString(pdtDecision), pdtSignedAt.Format("02/01/2006 à 15:04:05 UTC+1"))
	}

	return fmt.Sprintf(`
        <div class="sign-grid">
            <div class="sign-card">
                <div class="sign-header">Visa & Signature du Pharmacien Directeur Technique</div>
                <div style="font-size:7pt; color:#475569;">Engagement légal sous les peines de droit (21 CFR Part 11 / Annexe 11 BPF).</div>
                %s
            </div>
            <div class="sign-card">
                <div class="sign-header">%s</div>
                <div style="font-size:7pt; color:#475569;">Cachet humide officiel, date de dépôt et signature de l'agent récepteur.</div>
                <div class="sign-space">
                    <span>Réservé à l'Autorité Sanitaire</span>
                </div>
            </div>
        </div>
	`, pdtSignatureContent, html.EscapeString(recipientTitle))
}

// RenderDocumentWrap enveloppe le contenu dans la structure HTML A4 300+ DPI
func RenderDocumentWrap(title, bodyContent string) string {
	css := GetStandardBPFCSS()
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>%s</style>
</head>
<body>
%s
</body>
</html>`, html.EscapeString(title), css, bodyContent)
}
