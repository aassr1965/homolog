import os
import re
import subprocess
import markdown

md_path = r"C:\dev\homolog\docs\cahier_des_charges.md"
html_path = r"C:\dev\homolog\docs\cahier_des_charges.html"
pdf_path = r"C:\dev\homolog\docs\cahier_des_charges.pdf"

with open(md_path, "r", encoding="utf-8") as f:
    raw_content = f.read()

# 1. Custom Diagram for CTD Structure
ctd_diagram_html = """
<div class="diagram-card">
    <div class="diagram-title">STRUCTURE COMPARATIVE DU DOSSIER CTD ANPP (GUIDE JUILLET 2025)</div>
    <div class="diagram-grid-2">
        <div class="diagram-col princeps-col">
            <div class="diagram-badge badge-princeps">PRODUIT PRINCEPS (INNOVATEUR)</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 1 :</strong> Renseignements administratifs nationaux</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 2 :</strong> Résumés et synthèses (2.1 à 2.7)</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 3 :</strong> Qualité (Substance active & Produit fini)</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 4 :</strong> Rapports d'études non cliniques</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 5 :</strong> Rapports d'études cliniques complètes</div>
            <div class="diagram-footer">Modules 1 à 5 obligatoires et exhaustifs</div>
        </div>
        <div class="diagram-col generic-col">
            <div class="diagram-badge badge-generic">PRODUIT GÉNÉRIQUE (IMPORTÉ OU LOCAL)</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 1 :</strong> Exigences administratives algériennes</div>
            <div class="diagram-step"><span class="badge badge-inc">Incorporé</span> <strong>Module 2 :</strong> Directement intégré aux Modules 3, 4 et 5</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 3 :</strong> Qualité (Identique princeps + validation locale)</div>
            <div class="diagram-step"><span class="badge badge-fac">F</span> <strong>Module 4 :</strong> Facultatif (sur demande motivée ANPP)</div>
            <div class="diagram-step"><span class="badge badge-part">P</span> <strong>Module 5 :</strong> Partiellement requis (Bioéquivalence & Bioanalyse)</div>
            <div class="diagram-footer">Démonstration stricte de la bioéquivalence ou Biowaiver</div>
        </div>
    </div>
</div>
"""

# 2. Custom Diagram for Escalation / Variations
escalation_diagram_html = """
<div class="diagram-card">
    <div class="diagram-title">ALGORITHME DE QUALIFICATION ET DE RECLASSEMENT ANPP (LD-DEPP07/V01)</div>
    <div class="escalation-flow">
        <div class="flow-box main-box">
            <div class="flow-label">1. SÉLECTION DE LA MODIFICATION</div>
            <div class="flow-desc">Choix parmi les 34 rubriques réglementaires (Admin, CEP, Substance Active, Produit Fini)</div>
        </div>
        <div class="flow-arrow">⬇</div>
        <div class="flow-box eval-box">
            <div class="flow-label">2. CONTRÔLE DES CONDITIONS LÉGALES</div>
            <div class="flow-desc">Vérification algorithmique de chaque « Condition à remplir » de la rubrique</div>
        </div>
        <div class="flow-split">
            <div class="flow-branch branch-success">
                <div class="branch-status">✅ Toutes les conditions sont remplies</div>
                <div class="branch-card card-nominal">
                    <strong>Classification Nominale</strong><br>
                    Instruction selon le statut d'origine :<br>
                    <span class="badge badge-mmin">Mmin</span> ou <span class="badge badge-mmod">Mmod</span>
                </div>
            </div>
            <div class="flow-branch branch-warning">
                <div class="branch-status">⚠️ Au moins une condition NON remplie</div>
                <div class="branch-card card-escalated">
                    <strong>RECLASSEMENT OBLIGATOIRE (Escalade)</strong><br>
                    • <span class="badge badge-mmin">Mmin</span> reclassée en <span class="badge badge-mmod">Mmod</span> ou <span class="badge badge-mmaj">Mmaj</span><br>
                    • <span class="badge badge-mmod">Mmod</span> reclassée en <span class="badge badge-mmaj">Mmaj</span><br>
                    <em>Actualisation immédiate des justificatifs requis</em>
                </div>
            </div>
        </div>
    </div>
</div>
"""

# 3. Custom Mockup for Pharma Explorer Layout
layout_diagram_html = """
<div class="diagram-card layout-mockup-card">
    <div class="diagram-title">ERGONOMIE « PHARMA EXPLORER » HAUTE DENSITÉ (MODÈLE 3 VOLETS)</div>
    <div class="mockup-window">
        <div class="mockup-header">
            <span class="mockup-dots">● ● ●</span>
            <span class="mockup-path">📁 Homolog &gt; Dossiers &gt; ENR-2026-0042 (Amoxicilline 500mg) &gt; CTD_Module_1 &gt; 1.P</span>
        </div>
        <div class="mockup-body">
            <div class="mockup-pane mockup-tree">
                <div class="pane-title">NAVIGATION (Tree)</div>
                <div class="tree-item active">▼ ENR-2026-0042</div>
                <div class="tree-sub">► Mod 1 : Administratif</div>
                <div class="tree-sub sub-open">▼ 1.P Info Produit</div>
                <div class="tree-leaf">• 1.P.1 RCP/Notice <span class="badge-mini ok">3/3</span></div>
                <div class="tree-leaf">• 1.P.2 Modèle Algérie <span class="badge-mini ok">4/4</span></div>
                <div class="tree-leaf">• 1.P.3 Échantillons <span class="badge-mini alert">0/1</span></div>
                <div class="tree-leaf">• 1.P.8 Validation <span class="badge-mini ok">1/1</span></div>
                <div class="tree-sub">► Mod 2 : Résumé</div>
                <div class="tree-sub">► Mod 3 : Qualité</div>
                <div class="tree-sub">► Mod 5 : Bioéquivalence</div>
            </div>
            <div class="mockup-pane mockup-table">
                <div class="pane-title">VUE DÉTAILLÉE COMPACTE (32px par ligne)</div>
                <table class="inner-mockup-table">
                    <tr><th>Réf</th><th>Statut</th><th>Document</th><th>Ver</th><th>Signé DT</th></tr>
                    <tr><td>1.P.1</td><td><span class="badge badge-req">Validé</span></td><td>RCP_Origine.pdf</td><td>1.0</td><td>✅ DT</td></tr>
                    <tr><td>1.P.2</td><td><span class="badge badge-req">Validé</span></td><td>Notice_FR_AR.pdf</td><td>2.0</td><td>✅ DT</td></tr>
                    <tr><td>1.P.3</td><td><span class="badge badge-mmaj">Manquant</span></td><td>Bordereau_Échantillons</td><td>-</td><td>❌</td></tr>
                    <tr><td>1.P.4</td><td><span class="badge badge-req">Validé</span></td><td>CoA_Substance_Active.pdf</td><td>1.0</td><td>✅ DT</td></tr>
                    <tr><td>1.P.5</td><td><span class="badge badge-req">Validé</span></td><td>CoA_Produit_Fini.pdf</td><td>1.0</td><td>✅ DT</td></tr>
                    <tr><td>1.P.8</td><td><span class="badge badge-req">Validé</span></td><td>Engagement_NonCommercialisation.pdf</td><td>1.0</td><td>✅ DT</td></tr>
                </table>
            </div>
            <div class="mockup-pane mockup-preview">
                <div class="pane-title">PRÉVISUALISATION (Alt+P)</div>
                <div class="preview-box">
                    <div class="preview-thumb">📄 Notice_FR_AR.pdf</div>
                    <div class="preview-meta">
                        <div><strong>Version :</strong> 2.0 (En vigueur)</div>
                        <div><strong>Hash SHA-256 :</strong> e3b0c44298fc...</div>
                        <div><strong>Horodatage :</strong> 2026-11-10 14:32Z</div>
                    </div>
                    <div class="preview-btn">✍️ Signer Électroniquement (DT)</div>
                </div>
            </div>
        </div>
        <div class="mockup-status">
            <span>Éléments : 6 pièces</span>
            <span>Complétude : 83 % (1 manquante)</span>
            <span>Espace labo : 1.42 Go / 50 Go</span>
            <span>Langue : FR / AR (RTL disponible)</span>
        </div>
    </div>
</div>
"""

# Replace specific diagram blocks safely using exact unique markers
def replace_block_containing(text, marker, replacement):
    pos = text.find(marker)
    if pos == -1:
        return text
    start = text.rfind('```', 0, pos)
    end = text.find('```', pos) + 3
    if start != -1 and end != -1:
        return text[:start] + replacement + text[end:]
    return text

content = raw_content
content = replace_block_containing(content, 'STRUCTURE GLOBALE CTD ANPP', ctd_diagram_html)
content = replace_block_containing(content, 'ALGORITHME DE QUALIFICATION & RECLASSEMENT ANPP', escalation_diagram_html)
content = replace_block_containing(content, 'Amoxicilline 500mg', layout_diagram_html)

# Process callouts
lines = content.split('\n')
new_lines = []
in_callout = False
callout_type = ""
callout_lines = []

for line in lines:
    m = re.match(r"^>\s*\[!(IMPORTANT|NOTE|TIP|WARNING|CAUTION)\]\s*$", line.strip())
    if m:
        if in_callout:
            new_lines.append(f'<div class="callout callout-{callout_type.lower()}"><div class="callout-header"><strong class="callout-title">{callout_type}</strong></div><div class="callout-body">' + "\n".join(callout_lines) + '</div></div>')
            callout_lines = []
        in_callout = True
        callout_type = m.group(1)
        continue
    if in_callout:
        if line.startswith(">"):
            callout_lines.append(line.lstrip("> ").rstrip())
        else:
            new_lines.append(f'<div class="callout callout-{callout_type.lower()}"><div class="callout-header"><strong class="callout-title">{callout_type}</strong></div><div class="callout-body">\n' + "\n".join(callout_lines) + '\n</div></div>\n')
            in_callout = False
            callout_type = ""
            callout_lines = []
            new_lines.append(line)
    else:
        new_lines.append(line)

if in_callout:
    new_lines.append(f'<div class="callout callout-{callout_type.lower()}"><div class="callout-header"><strong class="callout-title">{callout_type}</strong></div><div class="callout-body">\n' + "\n".join(callout_lines) + '\n</div></div>\n')

md_to_convert = "\n".join(new_lines)

# Convert Markdown to HTML
html_body = markdown.markdown(md_to_convert, extensions=['extra', 'tables', 'fenced_code', 'toc'])

# Replace badges in table cells
badge_map = {
    "E": "badge badge-req",
    "NA": "badge badge-na",
    "F": "badge badge-fac",
    "P": "badge badge-part",
    "Mmin": "badge badge-mmin",
    "Mmod": "badge badge-mmod",
    "Mmaj": "badge badge-mmaj",
    "NI": "badge badge-proc",
    "RI": "badge badge-proc",
    "MI": "badge badge-proc",
    "NF": "badge badge-proc",
    "RF": "badge badge-proc",
    "MF": "badge badge-proc",
}

def badge_sub(m):
    pre_tag = m.group(1)
    cell_val = m.group(2).strip()
    cls = badge_map.get(cell_val, "")
    if cls:
        return f'<td{pre_tag} class="text-center"><span class="{cls}">{cell_val}</span></td>'
    return m.group(0)

html_body = re.sub(
    r'<td([^>]*)>\s*(E|NA|F|P|Mmin|Mmod|Mmaj|NI|RI|MI|NF|RF|MF)\s*</td>',
    badge_sub,
    html_body
)

# Table of Contents
toc_html = """
<div class="toc-container">
    <div class="toc-title">TABLE DES MATIÈRES</div>
    <div class="toc-grid">
        <div class="toc-item"><span class="toc-num">1</span> <span class="toc-text">Contexte, Cadre Juridique et Objectifs Stratégiques</span></div>
        <div class="toc-sub"><span class="toc-subnum">1.1</span> Cadre Réglementaire Algérien (ANPP / MIPH)</div>
        <div class="toc-sub"><span class="toc-subnum">1.2</span> Objectifs Stratégiques de la Plateforme « Homolog »</div>
        
        <div class="toc-item"><span class="toc-num">2</span> <span class="toc-text">Typologie des Procédures et Matrice CTD ANPP (Guide Juillet 2025)</span></div>
        <div class="toc-sub"><span class="toc-subnum">2.1</span> Les 6 Parcours Officiels de Soumission (NI, RI, MI, NF, RF, MF)</div>
        <div class="toc-sub"><span class="toc-subnum">2.2</span> Structure des 5 Modules CTD selon le Type d'Origine (Princeps vs Générique)</div>

        <div class="toc-item"><span class="toc-num">3</span> <span class="toc-text">Spécification Exhaustive du Module 1 (Spécificités ANPP Algérie)</span></div>
        <div class="toc-sub"><span class="toc-subnum">3.1</span> Sous-Module 1.A : APPLICANT (Demandeur & Établissement)</div>
        <div class="toc-sub"><span class="toc-subnum">3.2</span> Sous-Module 1.H : HOLDER (Détenteur de la DE)</div>
        <div class="toc-sub"><span class="toc-subnum">3.3</span> Sous-Module 1.M : MANUFACTURERS (Fabricants SA & Produit Fini)</div>
        <div class="toc-sub"><span class="toc-subnum">3.4</span> Sous-Module 1.P : PRODUCT INFORMATION (Informations Produit & Contrôle)</div>

        <div class="toc-item"><span class="toc-num">4</span> <span class="toc-text">Modules Techniques CTD 2, 3, 4 et 5 (Exigences et Spécificités Régionales)</span></div>
        <div class="toc-sub"><span class="toc-subnum">4.1</span> Module 3 : Qualité Pharmaceutique, Chimique et Biologique</div>
        <div class="toc-sub"><span class="toc-subnum">4.2</span> Module 5 : Données Cliniques et Bioéquivalence</div>

        <div class="toc-item"><span class="toc-num">5</span> <span class="toc-text">Moteur Intelligent des Modifications (Ligne Directrice LD-DEPP07/V01)</span></div>
        <div class="toc-sub"><span class="toc-subnum">5.1</span> Cadre Conceptuel et Règle d'Escalade Dynamique (Mmin, Mmod, Mmaj)</div>
        <div class="toc-sub"><span class="toc-subnum">5.2</span> Regroupement de Modifications (Grouping - Arrêté du 07/11/2021)</div>
        <div class="toc-sub"><span class="toc-subnum">5.3</span> Documentation Générale Requise pour TOUTE Demande de Modification</div>
        <div class="toc-sub"><span class="toc-subnum">5.4</span> Référentiel des 34 Rubriques de Modifications Réglementaires</div>

        <div class="toc-item"><span class="toc-num">6</span> <span class="toc-text">Architecture Technique et Ergonomie « Pharma Explorer »</span></div>
        <div class="toc-sub"><span class="toc-subnum">6.1</span> Stack Technologique Native et Souveraine (Go / PostgreSQL / CAS)</div>
        <div class="toc-sub"><span class="toc-subnum">6.2</span> Modèle Ergonomique « Pharma Explorer » à 3 Volets & Bilinguisme RTL</div>

        <div class="toc-item"><span class="toc-num">7</span> <span class="toc-text">Sécurité, Rôles et Conformité 21 CFR Part 11 / BPF</span></div>
        <div class="toc-sub"><span class="toc-subnum">7.1</span> Matrice des Rôles Stricts (9 Profils)</div>
        <div class="toc-sub"><span class="toc-subnum">7.2</span> Signature Électronique Légale et Intégrité ALCOA+</div>

        <div class="toc-item"><span class="toc-num">8</span> <span class="toc-text">Taxonomie du Coffre-Fort GED Pharmaceutique (Standard BPF)</span></div>

        <div class="toc-item"><span class="toc-num">9</span> <span class="toc-text">Conclusion et Validation</span></div>
    </div>
</div>
<div class="page-break"></div>
"""

full_html = f"""<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>Cahier des charges fonctionnel & technique - Homolog</title>
    <style>
        @page {{
            size: A4 portrait;
            margin: 18mm 16mm 20mm 16mm;
            @top-right {{
                content: "Homolog • Cahier des Charges v1.0";
                font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
                font-size: 8pt;
                color: #64748b;
                border-bottom: 1px solid #cbd5e1;
                padding-bottom: 4px;
            }}
            @bottom-left {{
                content: "Confidentiel — Direction Affaires Réglementaires & Qualité";
                font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
                font-size: 8pt;
                color: #94a3b8;
            }}
            @bottom-right {{
                content: "Page " counter(page);
                font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
                font-size: 8pt;
                color: #334155;
                font-weight: 700;
            }}
        }}

        @page:first {{
            margin: 0;
            @top-right {{ content: normal; }}
            @bottom-left {{ content: normal; }}
            @bottom-right {{ content: normal; }}
        }}

        * {{
            box-sizing: border-box;
        }}

        body {{
            font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, "Helvetica Neue", Arial, sans-serif;
            font-size: 9.2pt;
            line-height: 1.5;
            color: #1e293b;
            background-color: #ffffff;
            margin: 0;
            padding: 0;
        }}

        /* COVER PAGE */
        .cover-page {{
            page-break-after: always;
            height: 100vh;
            display: flex;
            flex-direction: column;
            justify-content: space-between;
            background: linear-gradient(135deg, #091e3a 0%, #0d3b66 50%, #082846 100%);
            color: #ffffff;
            padding: 60px 50px 50px 50px;
            position: relative;
            overflow: hidden;
        }}

        .cover-page::after {{
            content: "";
            position: absolute;
            top: -100px;
            right: -100px;
            width: 450px;
            height: 450px;
            background: radial-gradient(circle, rgba(13, 148, 136, 0.28) 0%, rgba(13, 148, 136, 0) 70%);
            border-radius: 50%;
        }}

        .cover-header {{
            display: flex;
            justify-content: space-between;
            align-items: center;
            border-bottom: 1px solid rgba(255, 255, 255, 0.15);
            padding-bottom: 25px;
        }}

        .cover-logo {{
            font-size: 26pt;
            font-weight: 800;
            letter-spacing: -1px;
            display: flex;
            align-items: center;
            gap: 12px;
        }}

        .cover-logo-icon {{
            width: 44px;
            height: 44px;
            background: #0d9488;
            border-radius: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 20pt;
            font-weight: 800;
            color: white;
            box-shadow: 0 4px 12px rgba(13, 148, 136, 0.4);
        }}

        .cover-tag {{
            background: rgba(13, 148, 136, 0.25);
            border: 1px solid rgba(13, 148, 136, 0.5);
            color: #2dd4bf;
            font-size: 9pt;
            font-weight: 700;
            padding: 6px 16px;
            border-radius: 20px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }}

        .cover-main {{
            margin-top: auto;
            margin-bottom: auto;
            padding: 20px 0;
        }}

        .cover-title {{
            font-size: 28pt;
            font-weight: 800;
            line-height: 1.15;
            margin: 0 0 16px 0;
            letter-spacing: -0.5px;
            color: #ffffff;
        }}

        .cover-subtitle {{
            font-size: 13pt;
            font-weight: 400;
            line-height: 1.45;
            color: #cbd5e1;
            margin: 0 0 28px 0;
            max-width: 92%;
        }}

        .cover-meta-grid {{
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 14px;
            background: rgba(255, 255, 255, 0.06);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.15);
            border-radius: 12px;
            padding: 20px;
            margin-top: 10px;
        }}

        .meta-item {{
            display: flex;
            flex-direction: column;
            gap: 3px;
        }}

        .meta-label {{
            font-size: 7.5pt;
            text-transform: uppercase;
            letter-spacing: 0.8px;
            color: #94a3b8;
            font-weight: 600;
        }}

        .meta-val {{
            font-size: 9.5pt;
            font-weight: 600;
            color: #f8fafc;
        }}

        .cover-footer {{
            border-top: 1px solid rgba(255, 255, 255, 0.15);
            padding-top: 16px;
            font-size: 8.5pt;
            color: #94a3b8;
            display: flex;
            justify-content: space-between;
        }}

        /* TOC */
        .toc-container {{
            background: #f8fafc;
            border: 1px solid #e2e8f0;
            border-radius: 10px;
            padding: 24px;
            margin: 20px 0 30px 0;
            page-break-after: always;
        }}

        .toc-title {{
            font-size: 15pt;
            font-weight: 800;
            color: #0f3a5d;
            border-bottom: 2px solid #0d9488;
            padding-bottom: 8px;
            margin-bottom: 18px;
            letter-spacing: 0.5px;
        }}

        .toc-grid {{
            display: flex;
            flex-direction: column;
            gap: 8px;
        }}

        .toc-item {{
            font-weight: 700;
            color: #0f3a5d;
            font-size: 10pt;
            display: flex;
            align-items: center;
            gap: 10px;
            margin-top: 6px;
        }}

        .toc-num {{
            background: #0d9488;
            color: white;
            width: 22px;
            height: 22px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            border-radius: 5px;
            font-size: 8.5pt;
            font-weight: 800;
        }}

        .toc-sub {{
            font-size: 8.5pt;
            color: #475569;
            padding-left: 32px;
            margin-top: -2px;
        }}

        .toc-subnum {{
            font-weight: 600;
            color: #0d9488;
            margin-right: 6px;
        }}

        /* HEADINGS */
        h1, h2, h3, h4 {{
            color: #0f2d4a;
            page-break-after: avoid;
            break-after: avoid;
        }}

        h1 {{
            font-size: 17pt;
            font-weight: 800;
            border-bottom: 2.5px solid #0d9488;
            padding-bottom: 8px;
            margin-top: 32px;
            margin-bottom: 14px;
            color: #0d3b66;
        }}

        h2 {{
            font-size: 12pt;
            font-weight: 700;
            border-left: 4.5px solid #0d9488;
            padding-left: 10px;
            margin-top: 22px;
            margin-bottom: 12px;
            background: linear-gradient(90deg, #f0fdfa 0%, #ffffff 100%);
            padding-top: 4px;
            padding-bottom: 4px;
        }}

        h3 {{
            font-size: 10.5pt;
            font-weight: 700;
            color: #1e3a8a;
            margin-top: 16px;
            margin-bottom: 8px;
        }}

        h4 {{
            font-size: 9.2pt;
            font-weight: 700;
            color: #334155;
            margin-top: 12px;
            margin-bottom: 6px;
        }}

        p {{
            margin-top: 0;
            margin-bottom: 10px;
            text-align: justify;
        }}

        ul, ol {{
            margin-top: 0;
            margin-bottom: 12px;
            padding-left: 20px;
        }}

        li {{
            margin-bottom: 4px;
        }}

        /* TABLES */
        table {{
            width: 100%;
            border-collapse: collapse;
            margin-top: 10px;
            margin-bottom: 16px;
            font-size: 8pt;
            page-break-inside: avoid;
            border-radius: 6px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
        }}

        th {{
            background: #0f3a5d;
            color: #ffffff;
            font-weight: 700;
            padding: 6px 7px;
            text-align: left;
            border: 1px solid #0f3a5d;
            font-size: 8pt;
            letter-spacing: 0.2px;
        }}

        td {{
            padding: 5px 6px;
            border: 1px solid #cbd5e1;
            vertical-align: middle;
        }}

        tr:nth-child(even) {{
            background-color: #f8fafc;
        }}

        .text-center {{
            text-align: center;
        }}

        /* BADGES */
        .badge {{
            display: inline-block;
            padding: 2.5px 6.5px;
            font-size: 7.2pt;
            font-weight: 800;
            border-radius: 4px;
            text-align: center;
            letter-spacing: 0.3px;
        }}

        .badge-req {{
            background-color: #dcfce7;
            color: #15803d;
            border: 1px solid #86efac;
        }}

        .badge-na {{
            background-color: #f1f5f9;
            color: #64748b;
            border: 1px solid #cbd5e1;
        }}

        .badge-fac {{
            background-color: #fef9c3;
            color: #854d0e;
            border: 1px solid #fde047;
        }}

        .badge-part {{
            background-color: #e0f2fe;
            color: #0369a1;
            border: 1px solid #7dd3fc;
        }}

        .badge-inc {{
            background-color: #f3e8ff;
            color: #7e22ce;
            border: 1px solid #d8b4fe;
        }}

        .badge-mmin {{
            background-color: #dbeafe;
            color: #1e40af;
            border: 1px solid #93c5fd;
        }}

        .badge-mmod {{
            background-color: #ffedd5;
            color: #c2410c;
            border: 1px solid #fdba74;
        }}

        .badge-mmaj {{
            background-color: #fee2e2;
            color: #b91c1c;
            border: 1px solid #fca5a5;
        }}

        .badge-proc {{
            background-color: #ede9fe;
            color: #6d28d9;
            border: 1px solid #c4b5fd;
        }}

        /* CALLOUTS */
        .callout {{
            margin: 14px 0;
            border-radius: 8px;
            border-left: 4.5px solid;
            padding: 12px 14px;
            page-break-inside: avoid;
        }}

        .callout-important {{
            background-color: #fef2f2;
            border-color: #ef4444;
            color: #991b1b;
        }}

        .callout-note {{
            background-color: #f0fdfa;
            border-color: #0d9488;
            color: #115e59;
        }}

        .callout-header {{
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 5px;
            font-size: 8.8pt;
        }}

        .callout-title {{
            text-transform: uppercase;
            letter-spacing: 0.5px;
            font-weight: 800;
        }}

        .callout-body p {{
            margin-bottom: 0;
            color: inherit;
        }}

        /* DIAGRAM CARDS */
        .diagram-card {{
            background: #ffffff;
            border: 1px solid #cbd5e1;
            border-radius: 10px;
            padding: 14px;
            margin: 14px 0;
            box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);
            page-break-inside: avoid;
        }}

        .diagram-title {{
            font-size: 9pt;
            font-weight: 800;
            color: #0f3a5d;
            text-align: center;
            letter-spacing: 0.5px;
            border-bottom: 1.5px solid #0d9488;
            padding-bottom: 6px;
            margin-bottom: 12px;
        }}

        .diagram-grid-2 {{
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 12px;
        }}

        .diagram-col {{
            border-radius: 8px;
            padding: 10px;
            display: flex;
            flex-direction: column;
            gap: 6px;
        }}

        .princeps-col {{
            background: #f0f9ff;
            border: 1px solid #bae6fd;
        }}

        .generic-col {{
            background: #f0fdfa;
            border: 1px solid #99f6e4;
        }}

        .diagram-badge {{
            font-size: 7.8pt;
            font-weight: 800;
            padding: 3px 6px;
            border-radius: 5px;
            text-align: center;
            margin-bottom: 2px;
            letter-spacing: 0.4px;
        }}

        .badge-princeps {{
            background: #0284c7;
            color: white;
        }}

        .badge-generic {{
            background: #0d9488;
            color: white;
        }}

        .diagram-step {{
            font-size: 7.8pt;
            display: flex;
            align-items: center;
            gap: 6px;
            background: white;
            padding: 4px 6px;
            border-radius: 4px;
            border: 1px solid rgba(0,0,0,0.06);
        }}

        .diagram-footer {{
            font-size: 7.2pt;
            font-style: italic;
            text-align: center;
            color: #475569;
            margin-top: 4px;
        }}

        /* ESCALATION FLOW */
        .escalation-flow {{
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 8px;
        }}

        .flow-box {{
            width: 85%;
            border-radius: 8px;
            padding: 8px 12px;
            text-align: center;
        }}

        .main-box {{
            background: #eff6ff;
            border: 1.5px solid #3b82f6;
        }}

        .eval-box {{
            background: #fefce8;
            border: 1.5px solid #eab308;
        }}

        .flow-label {{
            font-size: 8.2pt;
            font-weight: 800;
            color: #1e3a8a;
            letter-spacing: 0.4px;
        }}

        .flow-desc {{
            font-size: 7.8pt;
            color: #475569;
            margin-top: 1px;
        }}

        .flow-arrow {{
            font-size: 12pt;
            color: #0d9488;
            font-weight: bold;
        }}

        .flow-split {{
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 12px;
            width: 100%;
            margin-top: 2px;
        }}

        .flow-branch {{
            border-radius: 8px;
            padding: 8px;
            display: flex;
            flex-direction: column;
            gap: 5px;
        }}

        .branch-success {{
            background: #f0fdf4;
            border: 1px solid #86efac;
        }}

        .branch-warning {{
            background: #fff1f2;
            border: 1.5px dashed #f43f5e;
        }}

        .branch-status {{
            font-size: 7.8pt;
            font-weight: 800;
            text-align: center;
        }}

        .branch-card {{
            background: white;
            padding: 6px 8px;
            border-radius: 6px;
            font-size: 7.6pt;
            text-align: center;
            border: 1px solid rgba(0,0,0,0.08);
        }}

        /* MOCKUP CARD */
        .mockup-window {{
            background: #ffffff;
            border: 1px solid #94a3b8;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 4px 12px rgba(0,0,0,0.08);
        }}

        .mockup-header {{
            background: #1e293b;
            color: #f1f5f9;
            padding: 5px 10px;
            font-size: 7.2pt;
            display: flex;
            align-items: center;
            gap: 10px;
            font-family: monospace;
        }}

        .mockup-dots {{
            color: #64748b;
            font-size: 8.5pt;
        }}

        .mockup-body {{
            display: grid;
            grid-template-columns: 28% 44% 28%;
            min-height: 190px;
            background: #ffffff;
            border-bottom: 1px solid #e2e8f0;
        }}

        .mockup-pane {{
            padding: 7px;
            font-size: 7.2pt;
        }}

        .mockup-tree {{
            background: #f8fafc;
            border-right: 1px solid #cbd5e1;
            display: flex;
            flex-direction: column;
            gap: 2px;
        }}

        .pane-title {{
            font-size: 6.8pt;
            font-weight: 800;
            color: #64748b;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 5px;
            border-bottom: 1px solid #e2e8f0;
            padding-bottom: 2px;
        }}

        .tree-item {{
            font-weight: bold;
            color: #0f3a5d;
        }}

        .tree-sub {{
            padding-left: 8px;
            color: #334155;
            font-weight: 600;
        }}

        .sub-open {{
            color: #0d9488;
        }}

        .tree-leaf {{
            padding-left: 16px;
            color: #475569;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }}

        .badge-mini {{
            font-size: 6.2pt;
            font-weight: 700;
            padding: 1px 3px;
            border-radius: 3px;
        }}

        .badge-mini.ok {{
            background: #dcfce7;
            color: #166534;
        }}

        .badge-mini.alert {{
            background: #fee2e2;
            color: #991b1b;
        }}

        .mockup-table {{
            border-right: 1px solid #cbd5e1;
            padding: 5px;
        }}

        .inner-mockup-table {{
            width: 100%;
            font-size: 6.8pt;
            margin: 0;
            box-shadow: none;
        }}

        .inner-mockup-table th {{
            padding: 3px;
            font-size: 6.8pt;
            background: #334155;
        }}

        .inner-mockup-table td {{
            padding: 2.5px 4px;
            font-size: 6.8pt;
        }}

        .mockup-preview {{
            background: #f8fafc;
            display: flex;
            flex-direction: column;
            gap: 5px;
        }}

        .preview-box {{
            background: white;
            border: 1px solid #cbd5e1;
            border-radius: 6px;
            padding: 7px;
            display: flex;
            flex-direction: column;
            gap: 5px;
        }}

        .preview-thumb {{
            background: #e0f2fe;
            border: 1px dashed #0284c7;
            border-radius: 4px;
            padding: 10px 4px;
            text-align: center;
            font-weight: bold;
            color: #0369a1;
            font-size: 7.2pt;
        }}

        .preview-meta {{
            font-size: 6.5pt;
            color: #475569;
            display: flex;
            flex-direction: column;
            gap: 2px;
        }}

        .preview-btn {{
            background: #0d9488;
            color: white;
            padding: 3.5px;
            border-radius: 4px;
            text-align: center;
            font-size: 6.8pt;
            font-weight: bold;
        }}

        .mockup-status {{
            background: #f1f5f9;
            padding: 3.5px 10px;
            font-size: 6.8pt;
            color: #475569;
            display: flex;
            justify-content: space-between;
            font-weight: 600;
        }}

        /* CODE & PRE */
        pre {{
            background-color: #0f172a;
            color: #e2e8f0;
            padding: 10px 12px;
            border-radius: 6px;
            font-family: 'Cascadia Code', Consolas, 'Courier New', monospace;
            font-size: 7.8pt;
            overflow-x: auto;
            margin: 10px 0;
            page-break-inside: avoid;
            border: 1px solid #1e293b;
        }}

        code {{
            font-family: 'Cascadia Code', Consolas, 'Courier New', monospace;
            background-color: #f1f5f9;
            color: #0f172a;
            padding: 1px 4px;
            border-radius: 4px;
            font-size: 8pt;
        }}

        pre code {{
            background-color: transparent;
            color: inherit;
            padding: 0;
        }}

        hr {{
            border: none;
            border-top: 1px solid #e2e8f0;
            margin: 20px 0;
        }}

        .page-break {{
            page-break-before: always;
            break-before: page;
        }}
    </style>
</head>
<body>

<!-- COVER PAGE -->
<div class="cover-page">
    <div class="cover-header">
        <div class="cover-logo">
            <div class="cover-logo-icon">H</div>
            <span>HOMOLOG</span>
        </div>
        <div class="cover-tag">Conformité Réglementaire ANPP</div>
    </div>

    <div class="cover-main">
        <h1 class="cover-title">Cahier des Charges Fonctionnel & Technique</h1>
        <p class="cover-subtitle">Plateforme collaborative de gestion de l'enregistrement des médicaments, de l'homologation des dispositifs médicaux et GED pharmaceutique en Algérie</p>

        <div class="cover-meta-grid">
            <div class="meta-item">
                <span class="meta-label">Version du document</span>
                <span class="meta-val">Version 1.0 (Harmonisation Officielle ANPP)</span>
            </div>
            <div class="meta-item">
                <span class="meta-label">Date d'édition</span>
                <span class="meta-val">Novembre 2026</span>
            </div>
            <div class="meta-item">
                <span class="meta-label">Référentiel CTD</span>
                <span class="meta-val">Guide ANPP CTD (DEPP v01 - Juillet 2025)</span>
            </div>
            <div class="meta-item">
                <span class="meta-label">Référentiel Variations</span>
                <span class="meta-val">Ligne Directrice LD-DEPP07/V01 (Avril 2026)</span>
            </div>
            <div class="meta-item">
                <span class="meta-label">Conformité Qualité & Audit</span>
                <span class="meta-val">BPF / 21 CFR Part 11 / ALCOA+</span>
            </div>
            <div class="meta-item">
                <span class="meta-label">Architecture Cible</span>
                <span class="meta-val">Go 1.26+ • PostgreSQL 16+ • CAS SHA-256</span>
            </div>
        </div>
    </div>

    <div class="cover-footer">
        <span>République Algérienne Démocratique et Populaire • Ministère de l'Industrie Pharmaceutique</span>
        <span>Diffusion Restreinte — Usage Professionnel</span>
    </div>
</div>

<!-- TABLE OF CONTENTS -->
{toc_html}

<!-- CONTENT -->
<div class="content-container">
{html_body}
</div>

</body>
</html>
"""

with open(html_path, "w", encoding="utf-8") as f:
    f.write(full_html)

print("Safely updated HTML written to:", html_path)
