import os
import re
import sys
import shutil
import subprocess
import pymupdf
import markdown

WORKSPACE_DIR = r"C:\dev\homolog"
DOCS_DIR = os.path.join(WORKSPACE_DIR, "docs")
MD_PATH = os.path.join(DOCS_DIR, "cahier_des_charges.md")
HTML_PATH = os.path.join(DOCS_DIR, "cahier_des_charges.html")
TEMP_PDF_PATH = os.path.join(DOCS_DIR, "temp_cdc.pdf")
FINAL_PDF_PATH = os.path.join(DOCS_DIR, "cahier_des_charges.pdf")

EDGE_PATH = r"C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe"
if not os.path.exists(EDGE_PATH):
    EDGE_PATH = r"C:\Program Files\Microsoft\Edge\Application\msedge.exe"

print(f"Reading markdown from: {MD_PATH}")
with open(MD_PATH, "r", encoding="utf-8") as f:
    raw_content = f.read()

# Safeguard sanitizer for any LaTeX math expressions
def sanitize_latex(text):
    text = text.replace(r'$\to$', '→')
    text = text.replace(r'$\longrightarrow$', '→')
    text = text.replace(r'$\leftrightarrow$', '↔')
    text = text.replace(r'$\le 1$', '≤ 1')
    text = text.replace(r'$\le 4$', '≤ 4')
    text = text.replace(r'$\le$', '≤')
    text = text.replace(r'$\ge$', '≥')
    text = text.replace(r'$\pm U$', '± U')
    text = text.replace(r'$\pm$', '±')
    text = text.replace(r'$\neq$', '≠')
    text = text.replace(r'$C_{max}$', '<em>C</em><sub>max</sub>')
    text = text.replace(r'$AUC$', '<em>AUC</em>')
    text = text.replace(r'$f_1$', '<em>f</em><sub>1</sub>')
    text = text.replace(r'$f_2$', '<em>f</em><sub>2</sub>')
    text = text.replace(r'$C_p$', '<em>C</em><sub>p</sub>')
    text = text.replace(r'$C_{pk}$', '<em>C</em><sub>pk</sub>')
    return text

raw_content = sanitize_latex(raw_content)

# 1. Custom High-Density HTML Diagrams
ctd_diagram_html = """
<div class="diagram-card">
    <div class="diagram-header">
        <span class="diagram-tag">RÉFÉRENTIEL ANPP CTD 2025</span>
        <span class="diagram-title">STRUCTURE COMPARATIVE DU DOSSIER CTD ANPP (GUIDE JUILLET 2025)</span>
    </div>
    <div class="diagram-grid-2">
        <div class="diagram-col princeps-col">
            <div class="col-header">
                <span class="badge badge-princeps">PRODUIT PRINCEPS (INNOVATEUR)</span>
                <span class="col-desc">Exhaustivité Clinique, Toxicologique & Qualité</span>
            </div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 1 :</strong> Renseignements administratifs nationaux & RCP/Notice</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 2 :</strong> Résumés et synthèses d'experts (2.1 à 2.7)</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 3 :</strong> Qualité complète (Substance active 3.2.S & Produit fini 3.2.P)</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 4 :</strong> Rapports d'études non cliniques exhaustifs</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 5 :</strong> Rapports d'études cliniques complètes (Phases I à III)</div>
            <div class="diagram-footer">Modules 1 à 5 obligatoirement exigés et instruits par les commissions ANPP</div>
        </div>
        <div class="diagram-col generic-col">
            <div class="col-header">
                <span class="badge badge-generic">PRODUIT GÉNÉRIQUE (IMPORTÉ OU LOCAL)</span>
                <span class="col-desc">Bioéquivalence & Intégration Modulaire</span>
            </div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 1 :</strong> Exigences administratives algériennes intégrales</div>
            <div class="diagram-step"><span class="badge badge-inc">Incorporé</span> <strong>Module 2 :</strong> Directement intégré aux Modules 3, 4 et 5</div>
            <div class="diagram-step"><span class="badge badge-req">E</span> <strong>Module 3 :</strong> Qualité (Identique princeps + validation industrielle)</div>
            <div class="diagram-step"><span class="badge badge-fac">F</span> <strong>Module 4 :</strong> Facultatif (sur demande motivée ANPP / impuretés)</div>
            <div class="diagram-step"><span class="badge badge-part">P</span> <strong>Module 5 :</strong> Partiellement requis (Démonstration stricte de Bioéquivalence)</div>
            <div class="diagram-footer">Démonstration stricte de Bioéquivalence (5.3.1.2) ou Biowaiver (1.M.P.3)</div>
        </div>
    </div>
</div>
"""

escalation_diagram_html = """
<div class="diagram-card">
    <div class="diagram-header">
        <span class="diagram-tag">MOTEUR RÉGLEMENTAIRE LD-DEPP07/V01</span>
        <span class="diagram-title">ALGORITHME DE QUALIFICATION ET DE RECLASSEMENT AUTOMATIQUE ANPP</span>
    </div>
    <div class="escalation-flow">
        <div class="flow-box main-box">
            <div class="flow-step-num">ÉTAPE 1</div>
            <div class="flow-label">Sélection de la Modification</div>
            <div class="flow-desc">Choix parmi les 34 rubriques réglementaires (Admin, CEP, Substance Active, Produit Fini)</div>
        </div>
        <div class="flow-arrow">➔</div>
        <div class="flow-box eval-box">
            <div class="flow-step-num">ÉTAPE 2</div>
            <div class="flow-label">Contrôle des Conditions</div>
            <div class="flow-desc">Vérification algorithmique de chaque « Condition à remplir » de la rubrique</div>
        </div>
        <div class="flow-arrow">➔</div>
        <div class="flow-split-horizontal">
            <div class="flow-branch branch-success">
                <div class="branch-status">✅ Toutes les conditions remplies</div>
                <div class="branch-card card-nominal">
                    <strong>Classification Nominale</strong><br>
                    Instruction selon le statut initial :<br>
                    <span class="badge badge-mmin">Mmin</span> ou <span class="badge badge-mmod">Mmod</span>
                </div>
            </div>
            <div class="flow-branch branch-warning">
                <div class="branch-status">⚠️ Au moins 1 condition non remplie</div>
                <div class="branch-card card-escalated">
                    <strong>RECLASSEMENT OBLIGATOIRE (Escalade)</strong><br>
                    • <span class="badge badge-mmin">Mmin</span> reclassée en <span class="badge badge-mmod">Mmod</span> ou <span class="badge badge-mmaj">Mmaj</span><br>
                    • <span class="badge badge-mmod">Mmod</span> reclassée en <span class="badge badge-mmaj">Mmaj</span>
                </div>
            </div>
        </div>
    </div>
</div>
"""

layout_diagram_html = """
<div class="diagram-card layout-mockup-card">
    <div class="diagram-header">
        <span class="diagram-tag">MODÈLE ERGONOMIQUE PHARMA EXPLORER</span>
        <span class="diagram-title">ENVIRONNEMENT DE TRAVAIL RÉGLEMENTAIRE HAUTE DENSITÉ (3 VOLETS)</span>
    </div>
    <div class="mockup-window">
        <div class="mockup-header-bar">
            <span class="mockup-dots">● ● ●</span>
            <span class="mockup-path">Sanad Pharma &gt; Dossiers &gt; ENR-2026-0042 (Amoxicilline 500mg) &gt; CTD Module 1 &gt; 1.P</span>
            <span class="mockup-badge">Mode Enregistrement (NI)</span>
        </div>
        <div class="mockup-body">
            <div class="mockup-pane mockup-tree">
                <div class="pane-title">VOLET 1 : ARBORESCENCE CTD</div>
                <div class="tree-item active">▼ ENR-2026-0042 • Amoxicilline</div>
                <div class="tree-sub">► Mod 1 : Administratif</div>
                <div class="tree-sub sub-open">▼ 1.P Info Produit & Notices</div>
                <div class="tree-leaf">• 1.P.1 RCP Origine / FR <span class="badge-mini ok">100%</span></div>
                <div class="tree-leaf">• 1.P.2 Notice Bilingue FR/AR <span class="badge-mini ok">100%</span></div>
                <div class="tree-leaf">• 1.P.3 Échantillons & Scellés <span class="badge-mini alert">En attente</span></div>
                <div class="tree-leaf">• 1.P.8 Validation Procédé <span class="badge-mini ok">100%</span></div>
                <div class="tree-sub">► Mod 2 : Synthèses CTD</div>
                <div class="tree-sub">► Mod 3 : Qualité Pharmaceutique</div>
                <div class="tree-sub">► Mod 5 : Bioéquivalence Humaine</div>
            </div>
            <div class="mockup-pane mockup-table">
                <div class="pane-title">VOLET 2 : LISTE COMPACTE HAUTE DENSITÉ (32px)</div>
                <table class="inner-mockup-table">
                    <tr><th>Réf</th><th>Statut</th><th>Document</th><th>Ver</th><th>Visa DT</th></tr>
                    <tr><td>1.P.1</td><td><span class="badge badge-req">Validé</span></td><td>RCP_Amoxicilline_v1.0.pdf</td><td>1.0</td><td>✅ Signé DT</td></tr>
                    <tr><td>1.P.2</td><td><span class="badge badge-req">Validé</span></td><td>Notice_Bilingue_FR_AR.pdf</td><td>2.1</td><td>✅ Signé DT</td></tr>
                    <tr><td>1.P.3</td><td><span class="badge badge-mmaj">Attente</span></td><td>Bordereau_Echantillons.pdf</td><td>-</td><td>⏳ En attente</td></tr>
                    <tr><td>1.P.4</td><td><span class="badge badge-req">Validé</span></td><td>CoA_Substance_Active.pdf</td><td>1.0</td><td>✅ Signé DT</td></tr>
                    <tr><td>1.P.5</td><td><span class="badge badge-req">Validé</span></td><td>CoA_Produit_Fini.pdf</td><td>1.0</td><td>✅ Signé DT</td></tr>
                </table>
            </div>
            <div class="mockup-pane mockup-preview">
                <div class="pane-title">VOLET 3 : PRÉVISUALISEUR GED & SIGNATURE</div>
                <div class="preview-box">
                    <div class="preview-thumb">📄 Notice_Bilingue_FR_AR.pdf</div>
                    <div class="preview-meta">
                        <div><strong>Format :</strong> PDF/A-1b scellé (2.4 Mo)</div>
                        <div><strong>Hash SHA-256 :</strong> e3b0c44298fc1c14...</div>
                        <div><strong>Horodatage :</strong> 2026-11-16 18:42:01 UTC</div>
                        <div><strong>Maker :</strong> S. Aitali (Aff. Réglem.)</div>
                        <div><strong>Checker :</strong> M. Benali (Assurance Qualité)</div>
                    </div>
                    <div class="preview-btn">✍️ Apposer Visa Pharmacien DT (21 CFR Part 11)</div>
                </div>
            </div>
        </div>
        <div class="mockup-status-bar">
            <span>Complétude Dossier : 92 % (1 pièce requise en attente)</span>
            <span>Chiffrement : AES-256-GCM Actif</span>
            <span>Intégrité CAS : 100 % Conforme</span>
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
    end = text.find('```', pos)
    if end != -1:
        end = text.find('\n', end)
        if end == -1:
            end = len(text)
    if start != -1 and end != -1:
        return text[:start] + replacement + text[end:]
    return text

content = raw_content
content = replace_block_containing(content, 'STRUCTURE DU DOSSIER CTD ANPP SELON LE GUIDE DU 31 JUILLET 2025', ctd_diagram_html)
content = replace_block_containing(content, 'ALGORITHME DE QUALIFICATION & RECLASSEMENT ANPP', escalation_diagram_html)
content = replace_block_containing(content, 'Amoxicilline 500mg', layout_diagram_html)

# Process Mermaid blocks: replace ```mermaid ... ``` with <div class="mermaid">...</div>
def process_mermaid(text):
    pattern = r'```mermaid\s*\n(.*?)\n```'
    def repl(m):
        code = m.group(1).strip()
        return f'<div class="mermaid-container"><div class="mermaid">\n{code}\n</div></div>'
    return re.sub(pattern, repl, text, flags=re.DOTALL)

content = process_mermaid(content)

# Remove the initial markdown # Cahier des charges ... and ## Plateforme SANAD PHARMA ... up to ## 1. Contexte
# because the Cover Page handles the document presentation majestically
intro_match = re.search(r'^##\s+1\.\s+Contexte', content, re.MULTILINE)
if intro_match:
    content = content[intro_match.start():]

# Process Callouts (> [!IMPORTANT], > [!NOTE], etc.)
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

# Process regular blockquotes that are not GitHub alerts
content_pre = "\n".join(new_lines)

# Transform Chapter Headings (## \d+\. ...) to <h1 class="chapter-heading"> with clean anchor
def format_chapters(text):
    def repl(m):
        chap_text = m.group(1).strip()
        return f'\n\n<h1 class="chapter-heading">{chap_text}</h1>\n\n'
    return re.sub(r'^##\s+(\d+\..*)$', repl, text, flags=re.MULTILINE)

content_formatted = format_chapters(content_pre)

# Convert Markdown to HTML
html_body = markdown.markdown(content_formatted, extensions=['extra', 'tables', 'fenced_code', 'toc'])

# Replace badges in table cells
badge_map = {
    "E": "badge badge-req",
    "NA": "badge badge-na",
    "F": "badge badge-fac",
    "P": "badge badge-part",
    "S": "badge badge-subst",
    "C": "badge badge-cond",
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
    r'<td([^>]*)>\s*(E|NA|F|P|S|C|Mmin|Mmod|Mmaj|NI|RI|MI|NF|RF|MF)\s*</td>',
    badge_sub,
    html_body
)

# Table of Contents
toc_html = """
<div class="toc-container">
    <div class="toc-header">
        <span class="toc-tag">DOCUMENTATION MASTER INDEX</span>
        <h2 class="toc-title">TABLE DES MATIÈRES GÉNÉRALE</h2>
    </div>
    <div class="toc-grid">
        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">01</span> <span class="toc-name">Contexte, Cadre Juridique et Objectifs Stratégiques</span></div>
            <div class="toc-subitems">
                <div>1.1 Contexte Réglementaire Algérien (ANPP / MIPH)</div>
                <div>1.2 Identité de Marque « Sanad Pharma » & Suite Modulaire (9 Modules)</div>
                <div>1.3 Objectifs Stratégiques et Métriques Industrielles Cibles</div>
            </div>
        </div>
        
        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">02</span> <span class="toc-name">Typologie des Procédures et Matrice CTD ANPP (Guide 2025)</span></div>
            <div class="toc-subitems">
                <div>2.1 Les 6 Parcours Officiels de Soumission (NI, RI, MI, NF, RF, MF)</div>
                <div>2.2 Structure des 5 Modules CTD (Princeps vs Générique)</div>
            </div>
        </div>

        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">03</span> <span class="toc-name">Spécification Exhaustive du Module 1 (Spécificités ANPP)</span></div>
            <div class="toc-subitems">
                <div>3.1 Sous-Module 1.A : APPLICANT (Demandeur & Établissement)</div>
                <div>3.2 Sous-Module 1.H : HOLDER (Détenteur de la DE)</div>
                <div>3.3 Sous-Module 1.M : MANUFACTURERS (Fabricants SA & PF)</div>
                <div>3.4 Sous-Module 1.P : PRODUCT INFORMATION (Notices & Contrôle)</div>
            </div>
        </div>

        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">04</span> <span class="toc-name">Modules Techniques CTD 2, 3, 4 et 5 (Exigences Régionales)</span></div>
            <div class="toc-subitems">
                <div>4.1 Module 3 : Qualité Pharmaceutique, Chimique et Biologique</div>
                <div>4.2 Module 5 : Données Cliniques et Bioéquivalence Humaine</div>
            </div>
        </div>

        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">05</span> <span class="toc-name">Moteur Intelligent des Modifications (Ligne Directrice LD-DEPP07/V01)</span></div>
            <div class="toc-subitems">
                <div>5.1 Règle d'Escalade Dynamique & Regroupement (*Grouping*)</div>
                <div>5.2 Documentation Générale Requise pour TOUTE Modification</div>
                <div>5.3 Référentiel des 34 Rubriques de Modifications Réglementaires</div>
            </div>
        </div>

        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">06</span> <span class="toc-name">Architecture Technique et Ergonomie « Pharma Explorer »</span></div>
            <div class="toc-subitems">
                <div>6.1 Stack Technologique Native et Souveraine (Go / PostgreSQL / CAS)</div>
                <div>6.2 Modèle Ergonomique « Pharma Explorer » à 3 Volets & Bilinguisme RTL</div>
                <div>6.3 Console d'Administration & Gestion de la Concurrence (Baux Actifs)</div>
            </div>
        </div>

        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">07</span> <span class="toc-name">Sécurité, Rôles et Conformité 21 CFR Part 11 / BPF</span></div>
            <div class="toc-subitems">
                <div>7.1 Matrice Granulaire des 10 Rôles Pharmaceutiques Stricts</div>
                <div>7.2 Workflow Maker-Checker-Approver & Ségrégation des Fonctions (SoD)</div>
                <div>7.3 Signature Électronique Légale & Intégrité ALCOA+ (Audit Trail Inaltérable)</div>
            </div>
        </div>

        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">08</span> <span class="toc-name">Taxonomie du Coffre-Fort GED Pharmaceutique (Sanad Vault)</span></div>
            <div class="toc-subitems">
                <div>8.1 Architecture de Chiffrement Applicatif Hybride (AES-256-GCM + CAS)</div>
                <div>8.2 Partitionnement des Clés (KMS / HSM) et Sauvegardes GxP</div>
            </div>
        </div>

        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">09</span> <span class="toc-name">Module ANPP Submission Assistant & E-Services Gateway</span></div>
            <div class="toc-subitems">
                <div>9.1 Assistant de Soumission E-TASJIL / E-HOMOLOG & Télé-Services ANPP</div>
                <div>9.2 Chiffrement Hybride de l'Archive Souveraine & Scellement Horodaté</div>
            </div>
        </div>

        <div class="toc-card highlight-card">
            <div class="toc-chapter"><span class="toc-num">10</span> <span class="toc-name">Module Sanad Corporate : Agréments MIPH & Portail Tabadol</span></div>
            <div class="toc-subitems">
                <div>10.1 Cadre Réglementaire (Décret 21-82 & Arrêtés sept/oct 2025)</div>
                <div>10.2 Les 4 Catégories d'Établissements & Parcours Dédiés</div>
                <div>10.3 Formulaires Intelligents & Checklists Officielles Exhaustives</div>
                <div>10.4 Moteur de Rétroplanning MIPH & Export Tabadol Submission Package</div>
            </div>
        </div>

        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">11</span> <span class="toc-name">Spécifications Exhaustives des 28 États de Sortie BPF</span></div>
            <div class="toc-subitems">
                <div>11.1 Catalogue Exhaustif des 28 Rapports et États de Sortie Officiels</div>
                <div>11.2 Standard de Scellement PDF/A-1b et QR-Code de Vérification Décentralisée</div>
            </div>
        </div>

        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">12</span> <span class="toc-name">Spécifications du Copilote IA Réglementaire (Sanad Copilot)</span></div>
            <div class="toc-subitems">
                <div>12.1 Base de Connaissances Vectorielle Locale (PostgreSQL / pgvector)</div>
                <div>12.2 Contrôle de Complétude & Détection des Écarts Réglementaires</div>
            </div>
        </div>

        <div class="toc-card">
            <div class="toc-chapter"><span class="toc-num">13</span> <span class="toc-name">Conclusion, Matrice de Validation GAMP 5 & Signatures</span></div>
            <div class="toc-subitems">
                <div>13.1 Synthèse de Conformité Réglementaire et Avantages Compétitifs</div>
                <div>13.2 Formalisation des Approbations et Signatures Officielles GxP</div>
            </div>
        </div>
    </div>
</div>
"""

# Sign-off block at the end of the document
signoff_html = """
<div class="signoff-container">
    <h2 class="signoff-heading">Approbation Réglementaire et Signatures GxP</h2>
    <p>En apposant leur signature électronique ci-dessous, les parties prenantes attestent de l'exactitude des spécifications fonctionnelles, techniques et médico-légales consignées dans le présent document, ainsi que de leur parfaite adéquation avec la réglementation algérienne (ANPP / MIPH) et les standards internationaux BPF / 21 CFR Part 11.</p>
    
    <div class="sign-grid">
        <div class="sign-box">
            <div class="sign-role">RÉDACTEUR (Maker)</div>
            <div class="sign-name">Direction des Affaires Réglementaires</div>
            <div class="sign-title">Pharmacien Chargé d'Affaires Réglementaires</div>
            <div class="sign-space"></div>
            <div class="sign-date">Date : ____ / ____ / 2026</div>
            <div class="sign-status">Visa : [ CONFORME CTD / VARIATIONS ]</div>
        </div>

        <div class="sign-box">
            <div class="sign-role">VÉRIFICATEUR (Checker)</div>
            <div class="sign-name">Direction Assurance Qualité & CQ</div>
            <div class="sign-title">Responsable Assurance Qualité / CSV</div>
            <div class="sign-space"></div>
            <div class="sign-date">Date : ____ / ____ / 2026</div>
            <div class="sign-status">Visa : [ CONFORME BPF / ALCOA+ ]</div>
        </div>

        <div class="sign-box approver-box">
            <div class="sign-role">APPROBATEUR (Approver)</div>
            <div class="sign-name">Pharmacien Directeur Technique (PDT)</div>
            <div class="sign-title">Direction Technique & Industrielle</div>
            <div class="sign-space"></div>
            <div class="sign-date">Date : ____ / ____ / 2026</div>
            <div class="sign-status">Visa : [ BON POUR DÉPLOIEMENT GxP ]</div>
        </div>
    </div>
</div>
"""

full_html = f"""<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>Cahier des Charges Fonctionnel & Technique — Sanad Pharma v1.0</title>
    <style>
        @page {{
            size: A4 portrait;
            margin: 18mm 14mm 18mm 14mm;
        }}

        @page:first {{
            margin: 0;
        }}

        * {{
            box-sizing: border-box;
        }}

        body {{
            font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, "Helvetica Neue", Arial, sans-serif;
            font-size: 8.6pt;
            line-height: 1.45;
            color: #1e293b;
            background-color: #ffffff;
            margin: 0;
            padding: 0;
        }}

        /* COVER PAGE */
        .cover-page {{
            page-break-after: always;
            break-after: page;
            height: 100vh;
            display: flex;
            flex-direction: column;
            justify-content: space-between;
            background: linear-gradient(140deg, #071626 0%, #0a2540 45%, #053b58 100%);
            color: #ffffff;
            padding: 55px 45px 45px 45px;
            position: relative;
            overflow: hidden;
        }}

        .cover-page::after {{
            content: "";
            position: absolute;
            top: -120px;
            right: -120px;
            width: 480px;
            height: 480px;
            background: radial-gradient(circle, rgba(20, 184, 166, 0.22) 0%, rgba(20, 184, 166, 0) 70%);
            border-radius: 50%;
            pointer-events: none;
        }}

        .cover-header {{
            display: flex;
            justify-content: space-between;
            align-items: center;
            border-bottom: 1px solid rgba(255, 255, 255, 0.15);
            padding-bottom: 20px;
        }}

        .cover-brand {{
            display: flex;
            align-items: center;
            gap: 12px;
        }}

        .cover-brand-logo {{
            width: 42px;
            height: 42px;
            background: linear-gradient(135deg, #0d9488, #14b8a6);
            border-radius: 9px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 22pt;
            font-weight: 800;
            color: white;
            box-shadow: 0 4px 14px rgba(13, 148, 136, 0.4);
        }}

        .cover-brand-name {{
            font-size: 20pt;
            font-weight: 800;
            letter-spacing: -0.5px;
            color: #ffffff;
        }}

        .cover-brand-sub {{
            font-size: 7.5pt;
            color: #94a3b8;
            letter-spacing: 0.5px;
            text-transform: uppercase;
        }}

        .cover-header-tag {{
            background: rgba(13, 148, 136, 0.25);
            border: 1px solid rgba(20, 184, 166, 0.5);
            color: #5eead4;
            font-size: 7.5pt;
            font-weight: 700;
            padding: 5px 14px;
            border-radius: 20px;
            text-transform: uppercase;
            letter-spacing: 0.8px;
        }}

        .cover-main {{
            margin-top: auto;
            margin-bottom: auto;
            padding: 15px 0;
        }}

        .cover-badge-doc {{
            display: inline-block;
            background: rgba(56, 189, 248, 0.2);
            border: 1px solid #38bdf8;
            color: #bae6fd;
            padding: 4px 12px;
            border-radius: 16px;
            font-size: 7.8pt;
            font-weight: 700;
            letter-spacing: 1px;
            text-transform: uppercase;
            margin-bottom: 18px;
        }}

        .cover-title {{
            font-size: 26pt;
            font-weight: 800;
            line-height: 1.15;
            margin: 0 0 12px 0;
            letter-spacing: -0.5px;
            color: #ffffff;
        }}

        .cover-subtitle {{
            font-size: 11.5pt;
            font-weight: 600;
            line-height: 1.4;
            color: #38bdf8;
            margin: 0 0 16px 0;
            max-width: 95%;
        }}

        .cover-summary {{
            font-size: 8.4pt;
            line-height: 1.5;
            color: #cbd5e1;
            max-width: 96%;
            margin-bottom: 25px;
            background: rgba(15, 23, 42, 0.5);
            border-left: 3px solid #14b8a6;
            padding: 10px 14px;
            border-radius: 0 6px 6px 0;
        }}

        .cover-meta-grid {{
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 10px;
            background: rgba(255, 255, 255, 0.05);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.12);
            border-radius: 8px;
            padding: 14px 16px;
        }}

        .meta-item {{
            display: flex;
            flex-direction: column;
            gap: 2px;
        }}

        .meta-label {{
            font-size: 6.8pt;
            text-transform: uppercase;
            color: #94a3b8;
            font-weight: 700;
            letter-spacing: 0.5px;
        }}

        .meta-val {{
            font-size: 8pt;
            color: #ffffff;
            font-weight: 600;
        }}

        .cover-footer {{
            border-top: 1px solid rgba(255, 255, 255, 0.15);
            padding-top: 14px;
            display: flex;
            justify-content: space-between;
            font-size: 7.2pt;
            color: #94a3b8;
        }}

        /* TABLE OF CONTENTS */
        .toc-container {{
            padding: 20px 0;
            page-break-after: always;
            break-after: page;
        }}

        .toc-header {{
            margin-bottom: 20px;
            border-bottom: 2.5px solid #0f2942;
            padding-bottom: 10px;
        }}

        .toc-tag {{
            font-size: 7pt;
            font-weight: 800;
            letter-spacing: 1.2px;
            color: #0d9488;
            text-transform: uppercase;
        }}

        .toc-title {{
            font-size: 17pt;
            font-weight: 800;
            color: #0f172a;
            margin: 4px 0 0 0;
        }}

        .toc-grid {{
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 11px;
        }}

        .toc-card {{
            border: 1px solid #e2e8f0;
            border-radius: 6px;
            padding: 9px 12px;
            background: #ffffff;
            box-shadow: 0 1px 3px rgba(0,0,0,0.03);
            page-break-inside: avoid;
            break-inside: avoid;
        }}

        .toc-card.highlight-card {{
            border-color: #0d9488;
            background: #f0fdfa;
        }}

        .toc-chapter {{
            display: flex;
            align-items: center;
            gap: 8px;
            font-weight: 700;
            font-size: 8.4pt;
            color: #0f2942;
            margin-bottom: 5px;
            border-bottom: 1px solid #f1f5f9;
            padding-bottom: 4px;
        }}

        .toc-num {{
            background: #0f2942;
            color: #ffffff;
            font-size: 7pt;
            font-weight: 800;
            padding: 2px 5px;
            border-radius: 4px;
        }}

        .highlight-card .toc-num {{
            background: #0d9488;
        }}

        .toc-name {{
            flex: 1;
        }}

        .toc-subitems {{
            font-size: 7.1pt;
            color: #475569;
            display: flex;
            flex-direction: column;
            gap: 3px;
            padding-left: 6px;
        }}

        /* HEADINGS */
        .chapter-heading {{
            color: #091e36;
            font-size: 13.5pt;
            font-weight: 800;
            border-bottom: 2.5px solid #0284c7;
            padding-bottom: 5px;
            margin: 26px 0 12px 0;
            page-break-after: avoid;
            break-after: avoid;
            page-break-inside: avoid;
            break-inside: avoid;
            letter-spacing: -0.3px;
        }}

        h2 {{
            color: #0f2942;
            font-size: 11pt;
            font-weight: 700;
            border-left: 4px solid #0d9488;
            padding-left: 10px;
            margin: 18px 0 10px 0;
            page-break-after: avoid;
            break-after: avoid;
        }}

        h3 {{
            color: #1e293b;
            font-size: 9.5pt;
            font-weight: 700;
            margin: 14px 0 7px 0;
            page-break-after: avoid;
            break-after: avoid;
        }}

        h4 {{
            color: #334155;
            font-size: 8.6pt;
            font-weight: 700;
            margin: 10px 0 5px 0;
            page-break-after: avoid;
            break-after: avoid;
        }}

        p {{
            margin: 0 0 8px 0;
            line-height: 1.48;
        }}

        ul, ol {{
            margin: 4px 0 10px 18px;
            padding: 0;
        }}

        li {{
            margin-bottom: 3.5px;
            line-height: 1.45;
        }}

        blockquote {{
            margin: 10px 0;
            padding: 8px 12px;
            background: #f8fafc;
            border-left: 4px solid #0284c7;
            border-radius: 0 5px 5px 0;
            color: #334155;
            font-size: 8.2pt;
            page-break-inside: avoid;
            break-inside: avoid;
        }}

        /* TABLES */
        table {{
            width: 100%;
            border-collapse: collapse;
            font-size: 7.2pt;
            margin: 8px 0 12px 0;
            page-break-inside: auto;
        }}

        th {{
            background: #0f2942;
            color: #ffffff;
            font-weight: 700;
            text-align: left;
            padding: 5px 6px;
            border: 1px solid #0f2942;
            font-size: 6.8pt;
            text-transform: uppercase;
            letter-spacing: 0.3px;
        }}

        td {{
            padding: 4px 5.5px;
            border: 1px solid #cbd5e1;
            color: #334155;
            vertical-align: top;
            line-height: 1.32;
        }}

        tr:nth-child(even) td {{
            background-color: #f8fafc;
        }}

        tr {{
            page-break-inside: avoid;
            break-inside: avoid;
        }}

        .text-center {{
            text-align: center;
        }}

        /* BADGES */
        .badge {{
            display: inline-block;
            padding: 1.5px 5px;
            border-radius: 3px;
            font-size: 6.4pt;
            font-weight: 700;
            letter-spacing: 0.3px;
            text-transform: uppercase;
            white-space: nowrap;
        }}

        .badge-req {{ background: #dcfce7; color: #166534; border: 1px solid #86efac; }}
        .badge-na {{ background: #f1f5f9; color: #64748b; border: 1px solid #cbd5e1; }}
        .badge-fac {{ background: #fef3c7; color: #92400e; border: 1px solid #fcd34d; }}
        .badge-part {{ background: #e0e7ff; color: #3730a3; border: 1px solid #a5b4fc; }}
        .badge-subst {{ background: #fee2e2; color: #991b1b; border: 1px solid #fca5a5; }}
        .badge-cond {{ background: #ffedd5; color: #9a3412; border: 1px solid #fed7aa; }}
        .badge-mmin {{ background: #ccfbf1; color: #115e59; border: 1px solid #5eead4; }}
        .badge-mmod {{ background: #e0f2fe; color: #075985; border: 1px solid #7dd3fc; }}
        .badge-mmaj {{ background: #fce7f3; color: #9d174d; border: 1px solid #f472b6; }}
        .badge-proc {{ background: #ede9fe; color: #5b21b6; border: 1px solid #c4b5fd; font-weight: 800; }}
        .badge-princeps {{ background: #0284c7; color: white; padding: 2.5px 7px; font-size: 7pt; border-radius: 10px; }}
        .badge-generic {{ background: #0d9488; color: white; padding: 2.5px 7px; font-size: 7pt; border-radius: 10px; }}
        .badge-inc {{ background: #f1f5f9; color: #334155; border: 1px solid #cbd5e1; }}

        /* CALLOUTS */
        .callout {{
            border-radius: 5px;
            padding: 8px 12px;
            margin: 12px 0;
            page-break-inside: avoid;
            break-inside: avoid;
            font-size: 8.2pt;
        }}

        .callout-important {{
            background-color: #fef2f2;
            border: 1px solid #fecaca;
            border-left: 4px solid #dc2626;
        }}

        .callout-note {{
            background-color: #f0f9ff;
            border: 1px solid #bae6fd;
            border-left: 4px solid #0284c7;
        }}

        .callout-tip {{
            background-color: #f0fdf4;
            border: 1px solid #bbf7d0;
            border-left: 4px solid #16a34a;
        }}

        .callout-warning, .callout-caution {{
            background-color: #fffbeb;
            border: 1px solid #fde68a;
            border-left: 4px solid #d97706;
        }}

        .callout-header {{
            margin-bottom: 3px;
        }}

        .callout-title {{
            font-weight: 800;
            font-size: 8.2pt;
            letter-spacing: 0.5px;
            text-transform: uppercase;
        }}

        .callout-important .callout-title {{ color: #991b1b; }}
        .callout-note .callout-title {{ color: #0369a1; }}
        .callout-tip .callout-title {{ color: #15803d; }}
        .callout-warning .callout-title {{ color: #b45309; }}

        .callout-body {{
            line-height: 1.45;
            color: #1e293b;
        }}

        /* DIAGRAM CARDS */
        .diagram-card {{
            background: #ffffff;
            border: 1px solid #cbd5e1;
            border-radius: 6px;
            padding: 10px 12px;
            margin: 12px 0;
            page-break-inside: avoid;
            break-inside: avoid;
            box-shadow: 0 1px 3px rgba(0,0,0,0.03);
        }}

        .diagram-header {{
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 8px;
            border-bottom: 1.5px solid #e2e8f0;
            padding-bottom: 5px;
        }}

        .diagram-tag {{
            background: #0f2942;
            color: #ffffff;
            font-size: 6.4pt;
            font-weight: 800;
            padding: 2px 5px;
            border-radius: 3px;
            letter-spacing: 0.8px;
        }}

        .diagram-title {{
            font-size: 8.2pt;
            font-weight: 800;
            color: #0f172a;
        }}

        .diagram-grid-2 {{
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 12px;
        }}

        .diagram-col {{
            background: #f8fafc;
            border: 1px solid #e2e8f0;
            border-radius: 5px;
            padding: 9px;
            display: flex;
            flex-direction: column;
            gap: 5px;
        }}

        .col-header {{
            display: flex;
            flex-direction: column;
            gap: 2px;
            margin-bottom: 3px;
        }}

        .col-desc {{
            font-size: 6.6pt;
            color: #64748b;
            font-weight: 600;
            text-transform: uppercase;
        }}

        .diagram-step {{
            font-size: 7pt;
            color: #334155;
            display: flex;
            align-items: center;
            gap: 5px;
            background: #ffffff;
            padding: 3.5px 5px;
            border-radius: 3px;
            border: 1px solid #f1f5f9;
        }}

        .diagram-footer {{
            margin-top: auto;
            font-size: 6.6pt;
            font-style: italic;
            color: #64748b;
            border-top: 1px dashed #cbd5e1;
            padding-top: 4px;
        }}

        /* ESCALATION FLOW */
        .escalation-flow {{
            display: flex;
            align-items: center;
            gap: 8px;
        }}

        .flow-box {{
            flex: 1;
            background: #f8fafc;
            border: 1px solid #cbd5e1;
            border-radius: 5px;
            padding: 7px;
            font-size: 7pt;
        }}

        .flow-step-num {{
            font-size: 6pt;
            font-weight: 800;
            color: #0284c7;
            letter-spacing: 0.5px;
        }}

        .flow-label {{
            font-weight: 700;
            color: #0f172a;
            font-size: 7.4pt;
            margin: 2px 0;
        }}

        .flow-desc {{
            color: #64748b;
            font-size: 6.6pt;
            line-height: 1.3;
        }}

        .flow-arrow {{
            color: #94a3b8;
            font-size: 13pt;
            font-weight: bold;
        }}

        .flow-split-horizontal {{
            flex: 1.8;
            display: flex;
            gap: 6px;
        }}

        .flow-branch {{
            flex: 1;
            border-radius: 5px;
            padding: 6px;
            font-size: 6.6pt;
        }}

        .branch-success {{
            background: #f0fdf4;
            border: 1px solid #86efac;
        }}

        .branch-warning {{
            background: #fffbeb;
            border: 1px solid #fcd34d;
        }}

        .branch-status {{
            font-weight: 700;
            margin-bottom: 3px;
            font-size: 6.6pt;
        }}

        /* MOCKUP WINDOW */
        .mockup-window {{
            border: 1px solid #cbd5e1;
            border-radius: 5px;
            overflow: hidden;
        }}

        .mockup-header-bar {{
            background: #0f2942;
            color: #ffffff;
            padding: 5px 8px;
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 6.8pt;
        }}

        .mockup-dots {{
            color: #38bdf8;
            font-size: 7.5pt;
        }}

        .mockup-path {{
            flex: 1;
            font-family: monospace;
            color: #cbd5e1;
        }}

        .mockup-badge {{
            background: #0d9488;
            padding: 2px 5px;
            border-radius: 8px;
            font-size: 6pt;
            font-weight: bold;
        }}

        .mockup-body {{
            display: grid;
            grid-template-columns: 28% 44% 28%;
            background: #ffffff;
            border-bottom: 1px solid #e2e8f0;
            min-height: 175px;
        }}

        .mockup-pane {{
            padding: 5px;
            font-size: 6.6pt;
        }}

        .pane-title {{
            font-weight: 800;
            color: #64748b;
            font-size: 6.2pt;
            letter-spacing: 0.5px;
            margin-bottom: 4px;
            border-bottom: 1px solid #e2e8f0;
            padding-bottom: 2px;
        }}

        .mockup-tree {{
            background: #f8fafc;
            border-right: 1px solid #cbd5e1;
            display: flex;
            flex-direction: column;
            gap: 2px;
        }}

        .tree-item {{ font-weight: bold; color: #0f3a5d; }}
        .tree-sub {{ padding-left: 6px; color: #334155; font-weight: 600; }}
        .sub-open {{ color: #0d9488; }}
        .tree-leaf {{ padding-left: 12px; color: #475569; display: flex; justify-content: space-between; }}

        .badge-mini {{ font-size: 5.8pt; font-weight: 700; padding: 1px 3px; border-radius: 3px; }}
        .badge-mini.ok {{ background: #dcfce7; color: #166534; }}
        .badge-mini.alert {{ background: #fee2e2; color: #991b1b; }}

        .mockup-table {{
            border-right: 1px solid #cbd5e1;
            padding: 4px;
        }}

        .inner-mockup-table {{
            width: 100%;
            font-size: 6.4pt;
            margin: 0;
        }}

        .inner-mockup-table th {{ padding: 2.5px; font-size: 6pt; background: #334155; }}
        .inner-mockup-table td {{ padding: 2px 3.5px; font-size: 6.2pt; }}

        .mockup-preview {{
            background: #f8fafc;
            display: flex;
            flex-direction: column;
            gap: 4px;
        }}

        .preview-box {{
            background: white;
            border: 1px solid #cbd5e1;
            border-radius: 4px;
            padding: 5px;
            display: flex;
            flex-direction: column;
            gap: 4px;
        }}

        .preview-thumb {{
            background: #e0f2fe;
            border: 1px dashed #0284c7;
            border-radius: 3px;
            padding: 6px 3px;
            text-align: center;
            font-weight: bold;
            color: #0369a1;
            font-size: 6.6pt;
        }}

        .preview-meta {{ font-size: 6pt; color: #475569; display: flex; flex-direction: column; gap: 1.5px; }}
        .preview-btn {{
            background: #0d9488;
            color: white;
            padding: 2.5px;
            border-radius: 3px;
            text-align: center;
            font-size: 6pt;
            font-weight: bold;
        }}

        .mockup-status-bar {{
            background: #f1f5f9;
            padding: 3px 8px;
            font-size: 6.2pt;
            color: #475569;
            display: flex;
            justify-content: space-between;
            font-weight: 600;
        }}

        /* MERMAID */
        .mermaid-container {{
            background: #ffffff;
            border: 1px solid #e2e8f0;
            border-radius: 6px;
            padding: 10px;
            margin: 12px 0;
            text-align: center;
            page-break-inside: avoid;
            break-inside: avoid;
        }}

        /* CODE & PRE */
        pre {{
            background-color: #0f172a;
            color: #f1f5f9;
            padding: 8px 10px;
            border-radius: 4px;
            font-family: 'Cascadia Code', Consolas, 'Courier New', monospace;
            font-size: 7pt;
            overflow-x: auto;
            margin: 8px 0;
            page-break-inside: avoid;
            break-inside: avoid;
            line-height: 1.35;
            border: 1px solid #1e293b;
        }}

        code {{
            font-family: 'Cascadia Code', Consolas, monospace;
            font-size: 7.4pt;
            background-color: #f1f5f9;
            color: #0f172a;
            padding: 1px 3px;
            border-radius: 3px;
            border: 1px solid #e2e8f0;
        }}

        pre code {{
            background-color: transparent;
            color: inherit;
            padding: 0;
            border: none;
        }}

        hr {{
            display: none !important;
        }}

        .page-break {{
            page-break-after: always;
            break-after: page;
        }}

        /* SIGNOFF GRID */
        .signoff-container {{
            margin-top: 25px;
            page-break-inside: avoid;
            break-inside: avoid;
            border-top: 2px solid #0f2942;
            padding-top: 15px;
        }}

        .signoff-heading {{
            font-size: 11pt;
            font-weight: 800;
            color: #0f2942;
            margin-bottom: 8px;
        }}

        .sign-grid {{
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 10px;
            margin-top: 12px;
        }}

        .sign-box {{
            border: 1px solid #cbd5e1;
            border-radius: 5px;
            padding: 9px;
            background: #f8fafc;
            font-size: 7.2pt;
        }}

        .approver-box {{
            border-color: #0d9488;
            background: #f0fdfa;
        }}

        .sign-role {{
            font-weight: 800;
            color: #091e36;
            margin-bottom: 3px;
            text-transform: uppercase;
            font-size: 6.8pt;
        }}

        .sign-name {{
            font-weight: 700;
            color: #0f172a;
            font-size: 7.4pt;
        }}

        .sign-title {{
            color: #64748b;
            font-size: 6.6pt;
            margin-bottom: 6px;
        }}

        .sign-space {{
            height: 42px;
            border-bottom: 1px dashed #94a3b8;
            margin-bottom: 6px;
        }}

        .sign-date {{
            color: #64748b;
            font-size: 6.4pt;
            margin-bottom: 3px;
        }}

        .sign-status {{
            font-weight: 700;
            color: #0d9488;
            font-size: 6.6pt;
        }}
    </style>
    <script src="mermaid.min.js"></script>
    <script>
        document.addEventListener("DOMContentLoaded", function() {{
            if (window.mermaid) {{
                mermaid.initialize({{
                    startOnLoad: true,
                    theme: 'neutral',
                    flowchart: {{ useMaxWidth: true, htmlLabels: true, curve: 'basis' }}
                }});
            }}
        }});
    </script>
</head>
<body>

<!-- COVER PAGE -->
<div class="cover-page">
    <div class="cover-header">
        <div class="cover-brand">
            <div class="cover-brand-logo">S</div>
            <div>
                <div class="cover-brand-name">SANAD PHARMA</div>
                <div class="cover-brand-sub">سَنَد فارما • Système Réglementaire & GED BPF</div>
            </div>
        </div>
        <div class="cover-header-tag">Standard Réglementaire ANPP & MIPH</div>
    </div>

    <div class="cover-main">
        <div class="cover-badge-doc">Cahier des Charges Fonctionnel & Technique • Version 1.0</div>
        <div class="cover-title">SANAD PHARMA</div>
        <div class="cover-subtitle">Plateforme Collaborative de Gestion de l'Enregistrement des Médicaments, de l'Homologation des Dispositifs Médicaux et GED Pharmaceutique en Algérie</div>
        <div class="cover-summary">
            Spécifications fonctionnelles, techniques et médico-légales exhaustives harmonisées avec le <strong>Guide CTD ANPP (Juillet 2025)</strong>, la <strong>Ligne Directrice des Modifications LD-DEPP07/V01 (Avril 2026)</strong>, les décrets exécutifs 20-324 / 20-325, l'arrêté du 10 mai 2021 (DM), l'arrêté du 7 novembre 2021 (Grouping), ainsi que le <strong>Décret 21-82 et les Arrêtés de sept/oct 2025</strong> régissant les Agréments d'Établissements (Plateforme Tabadol). Standard de conformité <strong>21 CFR Part 11</strong> et intégrité <strong>ALCOA+</strong>.
        </div>

        <div class="cover-meta-grid">
            <div class="meta-item">
                <span class="meta-label">Autorités Réglementaires de Référence</span>
                <span class="meta-val">ANPP (Enregistrement/Homologation) & MIPH (Agréments Tabadol)</span>
            </div>
            <div class="meta-item">
                <span class="meta-label">Cadre Juridique Algérien</span>
                <span class="meta-val">Décrets 20-324, 20-325, 21-82 & Arrêtés 2021/2025</span>
            </div>
            <div class="meta-item">
                <span class="meta-label">Architecture Technologique</span>
                <span class="meta-val">Go 1.26+ | PostgreSQL 16+ (pgvector) | CAS SHA-256 | AES-256-GCM</span>
            </div>
            <div class="meta-item">
                <span class="meta-label">Intégrité & Sécurité BPF</span>
                <span class="meta-val">21 CFR Part 11, TOTP Hors-Ligne, RBAC 10 Profils, SoD Stricte</span>
            </div>
            <div class="meta-item">
                <span class="meta-label">Statut & Date d'Effet</span>
                <span class="meta-val">Version 1.0 Validée & Scellée • Novembre 2026</span>
            </div>
            <div class="meta-item">
                <span class="meta-label">Gouvernance Réglementaire</span>
                <span class="meta-val">Pharmacien Directeur Technique (PDT) & Direction Qualité</span>
            </div>
        </div>
    </div>

    <div class="cover-footer">
        <span>République Algérienne Démocratique et Populaire • Ministère de l'Industrie et de la Production Pharmaceutique</span>
        <span>Usage Professionnel Restreint — Propriété Intellectuelle & Réglementaire</span>
    </div>
</div>

<!-- TABLE OF CONTENTS -->
{toc_html}

<!-- CONTENT -->
<div class="content-container">
{html_body}
</div>

<!-- SIGNOFF -->
{signoff_html}

</body>
</html>
"""

print(f"Writing complete HTML to: {HTML_PATH}")
with open(HTML_PATH, "w", encoding="utf-8") as f:
    f.write(full_html)
print(f"HTML successfully written ({os.path.getsize(HTML_PATH):,} bytes)")

# 2. Compile to PDF with Edge Headless
print("Compiling HTML to PDF using headless Edge...")
if os.path.exists(TEMP_PDF_PATH):
    os.remove(TEMP_PDF_PATH)

cmd = [
    EDGE_PATH,
    "--headless=new",
    "--disable-gpu",
    "--run-all-compositor-stages-before-draw",
    "--virtual-time-budget=8000",
    "--no-pdf-header-footer",
    f"--print-to-pdf={TEMP_PDF_PATH}",
    f"file:///{HTML_PATH.replace(os.sep, '/')}"
]

process = subprocess.run(cmd, capture_output=True, text=True, timeout=120)
if not os.path.exists(TEMP_PDF_PATH) or os.path.getsize(TEMP_PDF_PATH) == 0:
    print("Edge generation failed! Return code:", process.returncode)
    print("Stderr:", process.stderr[:1000])
    sys.exit(1)

print(f"Initial PDF generated by Edge ({os.path.getsize(TEMP_PDF_PATH):,} bytes)")

# 3. Post-Process with PyMuPDF: Stamp Professional Running Header & Footer
print("Post-processing PDF with PyMuPDF (vector headers, footers, metadata)...")
doc = pymupdf.open(TEMP_PDF_PATH)
total_pages = len(doc)
print(f"Total pages in document: {total_pages}")

# Header & Footer styling parameters
font_name = "helv"
font_bold = "hebo"
font_size = 7.0
header_color = (0.25, 0.35, 0.45)
line_color = (0.82, 0.86, 0.90)
footer_color = (0.45, 0.50, 0.55)
page_num_color = (0.10, 0.20, 0.30)

# Stamp pages from index 1 (Page 2) to end (omit cover page 0)
for p in range(1, total_pages):
    page = doc[p]
    page_rect = page.rect
    width = page_rect.width
    height = page_rect.height
    
    # Left and right margins
    margin_left = 38.0
    margin_right = width - 38.0
    
    # Header: Y = 28
    header_y = 26.0
    line_header_y = 32.0
    
    # Draw header rule
    page.draw_line(
        pymupdf.Point(margin_left, line_header_y),
        pymupdf.Point(margin_right, line_header_y),
        color=line_color,
        width=0.65
    )
    
    # Header Left
    page.insert_text(
        pymupdf.Point(margin_left, header_y),
        "SANAD PHARMA • CAHIER DES CHARGES FONCTIONNEL & TECHNIQUE (v1.0)",
        fontname=font_bold,
        fontsize=font_size,
        color=header_color
    )
    
    # Header Right
    hdr_right_text = "RÉFÉRENTIEL ANPP / MIPH • BPF / 21 CFR PART 11"
    hdr_right_len = pymupdf.get_text_length(hdr_right_text, fontname=font_name, fontsize=font_size)
    page.insert_text(
        pymupdf.Point(margin_right - hdr_right_len, header_y),
        hdr_right_text,
        fontname=font_name,
        fontsize=font_size,
        color=header_color
    )
    
    # Footer: Y = height - 26
    line_footer_y = height - 32.0
    footer_y = height - 22.0
    
    # Draw footer rule
    page.draw_line(
        pymupdf.Point(margin_left, line_footer_y),
        pymupdf.Point(margin_right, line_footer_y),
        color=line_color,
        width=0.65
    )
    
    # Footer Left
    page.insert_text(
        pymupdf.Point(margin_left, footer_y),
        "CONFIDENTIEL • DIRECTION DES AFFAIRES RÉGLEMENTAIRES & ASSURANCE QUALITÉ",
        fontname=font_name,
        fontsize=font_size,
        color=footer_color
    )
    
    # Footer Right (Page X sur Y)
    page_text = f"Page {p + 1} sur {total_pages}"
    page_len = pymupdf.get_text_length(page_text, fontname=font_bold, fontsize=font_size)
    page.insert_text(
        pymupdf.Point(margin_right - page_len, footer_y),
        page_text,
        fontname=font_bold,
        fontsize=font_size,
        color=page_num_color
    )

# Set PDF Metadata
doc.set_metadata({
    "title": "Cahier des Charges Fonctionnel & Technique — Plateforme Sanad Pharma",
    "author": "Direction des Affaires Réglementaires & Qualité",
    "subject": "Spécifications ANPP CTD 2025, LD-DEPP07/V01, Décret 21-82 & Portail Tabadol MIPH",
    "keywords": "Sanad Pharma, ANPP, CTD, MIPH, Tabadol, BPF, 21 CFR Part 11, ALCOA+, Médicaments, Dispositifs Médicaux",
    "creator": "Sanad Pharma GxP Publishing Engine",
    "producer": "PyMuPDF & Microsoft Edge Blink Engine"
})

# Save finalized PDF
print(f"Saving finalized PDF to: {FINAL_PDF_PATH}")
doc.save(FINAL_PDF_PATH, garbage=4, deflate=True)
doc.close()

# Clean up temp PDF
if os.path.exists(TEMP_PDF_PATH):
    os.remove(TEMP_PDF_PATH)

print(f"SUCCESS: Finalized PDF generated ({os.path.getsize(FINAL_PDF_PATH):,} bytes, {total_pages} pages)")
