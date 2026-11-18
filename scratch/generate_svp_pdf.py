import os
import re
import subprocess
import markdown

SOURCE_MD = r"C:\Users\soufi\.gemini\antigravity\brain\28d7ef78-a325-4f9c-9822-7255856adaf3\svp_gamp5_sanad.md"
OUTPUT_HTML = r"C:\Users\soufi\.gemini\antigravity\brain\28d7ef78-a325-4f9c-9822-7255856adaf3\scratch\svp_gamp5_sanad.html"
OUTPUT_PDF = r"C:\Users\soufi\.gemini\antigravity\brain\28d7ef78-a325-4f9c-9822-7255856adaf3\SVP_GAMP5_Sanad_Pharma_v1.0.pdf"
CHROME_PATH = r"C:\Program Files\Google\Chrome\Application\chrome.exe"

with open(SOURCE_MD, "r", encoding="utf-8") as f:
    raw_md = f.read()

# Remplacement alertes markdown
def alert_repl(m):
    atype = m.group(1).lower()
    text = m.group(2)
    colors = {
        "important": ("#b45309", "#fef3c7", "#78350f"),
        "warning": ("#b91c1c", "#fee2e2", "#7f1d1d"),
        "note": ("#1d4ed8", "#dbeafe", "#1e3a8a"),
        "tip": ("#047857", "#d1fae5", "#064e3b"),
    }
    border, bg, fg = colors.get(atype, ("#4b5563", "#f3f4f6", "#111827"))
    return f'<div class="callout callout-{atype}" style="border-left: 4px solid {border}; background-color: {bg}; color: {fg}; padding: 12px 16px; margin: 16px 0; border-radius: 4px;"><strong>{atype.upper()} :</strong> {text}</div>'

md_processed = re.sub(r'>\s*\[!(IMPORTANT|WARNING|NOTE|TIP)\]\s*\n((?:>\s*.*\n?)+)', 
                      lambda m: alert_repl(re.match(r'>\s*\[!(IMPORTANT|WARNING|NOTE|TIP)\]\s*\n(.*)', m.group(0), re.DOTALL)),
                      raw_md)

# Traiter les blockquotes standard restants
md_processed = re.sub(r'^>\s*(.*)$', r'<blockquote style="border-left: 4px solid #0284c7; background: #f0f9ff; margin: 14px 0; padding: 10px 14px; font-style: italic; color: #0369a1;">\1</blockquote>', md_processed, flags=re.MULTILINE)

html_content = markdown.markdown(md_processed, extensions=['tables', 'fenced_code', 'nl2br'])

html_full = f"""<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>Plan de Validation Système (SVP) — Sanad Pharma</title>
    <style>
        @page {{
            size: A4 portrait;
            margin: 20mm 15mm 22mm 15mm;
            @top-right {{
                content: "VAL-SVP-SANAD-001 | v1.0";
                font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
                font-size: 8pt;
                color: #64748b;
            }}
            @bottom-left {{
                content: "STRICTEMENT CONFIDENTIEL — USAGE RÉGLEMENTAIRE PHARMACEUTIQUE (BPF / GAMP 5)";
                font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
                font-size: 7.5pt;
                color: #94a3b8;
            }}
            @bottom-right {{
                content: "Page " counter(page);
                font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
                font-size: 8pt;
                color: #64748b;
                font-weight: bold;
            }}
        }}

        body {{
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            font-size: 9.5pt;
            line-height: 1.5;
            color: #1e293b;
            background: #ffffff;
            margin: 0;
            padding: 0;
        }}

        /* Cartouche En-tête Officiel */
        .header-box {{
            border: 2px solid #0f766e;
            border-radius: 6px;
            padding: 16px 20px;
            margin-bottom: 24px;
            background: linear-gradient(135deg, #f0fdfa 0%, #ffffff 100%);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }}

        .header-title-box h1 {{
            font-size: 18pt;
            font-weight: 800;
            color: #0f766e;
            margin: 0 0 4px 0;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }}

        .header-title-box .sub {{
            font-size: 10pt;
            font-weight: 600;
            color: #334155;
            margin: 0;
        }}

        .header-badge {{
            text-align: right;
            border-left: 2px solid #cbd5e1;
            padding-left: 16px;
        }}

        .status-pill {{
            display: inline-block;
            background: #0d9488;
            color: #ffffff;
            font-size: 8.5pt;
            font-weight: 700;
            padding: 4px 10px;
            border-radius: 12px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 4px;
        }}

        .meta-ref {{
            font-size: 8pt;
            color: #64748b;
            font-family: Consolas, monospace;
        }}

        h2 {{
            color: #0f766e;
            font-size: 13pt;
            font-weight: 700;
            border-bottom: 1.5px solid #0f766e;
            padding-bottom: 4px;
            margin-top: 24px;
            margin-bottom: 12px;
            page-break-after: avoid;
        }}

        h3 {{
            color: #1e293b;
            font-size: 11pt;
            font-weight: 700;
            margin-top: 16px;
            margin-bottom: 8px;
            page-break-after: avoid;
        }}

        h4 {{
            color: #334155;
            font-size: 10pt;
            font-weight: 700;
            margin-top: 12px;
            margin-bottom: 6px;
            page-break-after: avoid;
        }}

        p {{
            margin: 0 0 8px 0;
            text-align: justify;
        }}

        /* Tableaux Haute Densité Réglementaires */
        table {{
            width: 100%;
            border-collapse: collapse;
            margin: 12px 0 18px 0;
            font-size: 8.5pt;
            page-break-inside: auto;
        }}

        tr {{
            page-break-inside: avoid;
            page-break-after: auto;
        }}

        thead {{
            display: table-header-group;
        }}

        th {{
            background-color: #0f766e;
            color: #ffffff;
            font-weight: 700;
            text-align: left;
            padding: 7px 10px;
            border: 1px solid #0f766e;
            font-size: 8.5pt;
            letter-spacing: 0.3px;
        }}

        td {{
            padding: 6px 10px;
            border: 1px solid #cbd5e1;
            vertical-align: top;
        }}

        tr:nth-child(even) td {{
            background-color: #f8fafc;
        }}

        /* Encadrés Cas de Tests OQ */
        h4:has(+ p) {{
            background-color: #f1f5f9;
            padding: 6px 10px;
            border-left: 3px solid #0f766e;
            border-radius: 0 4px 4px 0;
        }}

        ul, ol {{
            margin: 4px 0 10px 0;
            padding-left: 20px;
        }}

        li {{
            margin-bottom: 3px;
        }}

        code {{
            font-family: Consolas, "Courier New", monospace;
            background: #f1f5f9;
            color: #0f766e;
            padding: 1px 4px;
            border-radius: 3px;
            font-size: 8.5pt;
            border: 1px solid #e2e8f0;
        }}

        pre {{
            background: #0f172a;
            color: #f8fafc;
            padding: 12px;
            border-radius: 5px;
            font-family: Consolas, monospace;
            font-size: 8pt;
            line-height: 1.4;
            overflow-x: auto;
            margin: 10px 0 16px 0;
            page-break-inside: avoid;
        }}

        pre code {{
            background: transparent;
            color: inherit;
            padding: 0;
            border: none;
        }}

        hr {{
            border: none;
            border-top: 1px solid #e2e8f0;
            margin: 20px 0;
        }}

        .page-break {{
            page-break-before: always;
        }}
    </style>
</head>
<body>

    <div class="header-box">
        <div class="header-title-box">
            <h1>Sanad Pharma</h1>
            <div class="sub">Plateforme Réglementaire ANPP &amp; GED Pharmaceutique</div>
            <div style="font-size: 8.5pt; color: #64748b; margin-top: 4px;">Système Validé : <strong>Homolog</strong> | Conforme 21 CFR Part 11 &amp; GAMP 5</div>
        </div>
        <div class="header-badge">
            <div class="status-pill">Document Approuvé</div>
            <div class="meta-ref">RÉF : VAL-SVP-SANAD-001</div>
            <div class="meta-ref">VERSION : 1.0</div>
            <div class="meta-ref">DATE : 18 NOVEMBRE 2026</div>
        </div>
    </div>

    {html_content}

</body>
</html>
"""

os.makedirs(os.path.dirname(OUTPUT_HTML), exist_ok=True)
with open(OUTPUT_HTML, "w", encoding="utf-8") as f:
    f.write(html_full)

print(f"HTML généré avec succès : {OUTPUT_HTML}")

# Conversion en PDF via Chrome headless
cmd = [
    CHROME_PATH,
    "--headless=new",
    "--disable-gpu",
    "--no-pdf-header-footer",
    f"--print-to-pdf={OUTPUT_PDF}",
    OUTPUT_HTML
]

res = subprocess.run(cmd, capture_output=True, text=True)
if os.path.exists(OUTPUT_PDF) and os.path.getsize(OUTPUT_PDF) > 1000:
    print(f"SUCCES : PDF généré avec succès ({os.path.getsize(OUTPUT_PDF)} octets) : {OUTPUT_PDF}")
else:
    print(f"ERREUR lors de la génération PDF : {res.stderr}")
