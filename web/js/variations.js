// ============================================================================
// MOTEUR DE VARIATIONS ANPP (Jalon 2.2 - Ligne Directrice LD-DEPP07/V01)
// ============================================================================

let selectedVarIndex = 0;

        const mockVariations = [
            {
                id: "VAR-2026-001",
                item: "7.b.1",
                label: "Remplacement ou ajout d'un matériau de départ / réactif intermédiaire",
                product: "Paracétamol LPA 500 mg",
                dci: "Paracétamol",
                category: "Mmin",
                categoryClass: "status-rev",
                status: "EN INSTRUCTION ANPP",
                statusClass: "status-rev",
                submissionDate: "12/01/2026",
                docsCount: "4 pièces (3.2.S.2 / BPF)",
                desc: "Changement de fournisseur du p-aminophénol (matériau de départ). Spécifications et profil d'impuretés conformes Ph. Eur."
            },
            {
                id: "VAR-2026-002",
                item: "7.b.2",
                label: "Modification substantielle du procédé de synthèse de la substance active",
                product: "Amoxicilline LPA 1 g",
                dci: "Amoxicilline trihydratée",
                category: "Mmaj",
                categoryClass: "status-req",
                status: "DÉPÔT EFFECTUÉ",
                statusClass: "status-ok",
                submissionDate: "03/02/2026",
                docsCount: "7 pièces (3.2.S.2 / Validation / Stabilité)",
                desc: "Nouveau site de synthèse chimique avec étape de purification enzymatique. Dépôt majeur nécessitant avis de la Commission d'Enregistrement."
            },
            {
                id: "VAR-2026-003",
                item: "12.a",
                label: "Prolongation de la durée de conservation du produit fini (36 à 48 mois)",
                product: "Paracétamol LPA 500 mg",
                dci: "Paracétamol",
                category: "Mmod",
                categoryClass: "status-rev",
                status: "APPROUVÉ ANPP",
                statusClass: "status-ok",
                submissionDate: "15/10/2025",
                docsCount: "3 pièces (3.2.P.8 Stabilité temps réel 48m zone IVb)",
                desc: "Données de stabilité en temps réel à 30°C / 75% HR conformément aux exigences algériennes."
            }
        ];


        function renderVariationsTable() {
            const thead = document.getElementById("table-head");
            const tbody = document.getElementById("doc-tbody");
            tbody.innerHTML = "";

            thead.innerHTML = `
                <tr>
                    <th style="width: 80px;">Classification</th>
                    <th>N° Item &amp; Intitulé de la Variation (LD-DEPP07)</th>
                    <th>Spécialité / Produit</th>
                    <th>DCI</th>
                    <th>Date Dépôt</th>
                    <th>Pièces Régionales (3.2.R)</th>
                    <th>Statut ANPP</th>
                </tr>
            `;

            mockVariations.forEach((v, idx) => {
                const tr = document.createElement("tr");
                if (idx === selectedVarIndex) tr.classList.add("selected");
                tr.onclick = () => { selectedVarIndex = idx; renderVariationsTable(); };
                tr.innerHTML = `
                    <td><span class="status-tag ${v.categoryClass}">${v.category}</span></td>
                    <td style="font-weight:600; color:var(--text-main);"><code style="color:var(--accent-blue);">${v.item}</code> ${v.label}</td>
                    <td style="font-weight:600;">${v.product}</td>
                    <td style="color:var(--text-muted);">${v.dci}</td>
                    <td style="font-family:monospace; color:var(--text-dim);">${v.submissionDate}</td>
                    <td style="font-size:11px;">${v.docsCount}</td>
                    <td><span class="status-tag ${v.statusClass}">${v.status}</span></td>
                `;
                tbody.appendChild(tr);
            });
            renderVariationPreview(mockVariations[selectedVarIndex]);
        }

        function renderVariationPreview(v) {
            const preview = document.getElementById("preview-body");
            if (!v) return;
            preview.innerHTML = `
                <div class="pdf-sim-page">
                    <div class="pdf-sim-header">
                        AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)<br>
                        NOTIFICATION DE VARIATION POST-AMM (LD-DEPP07/V01)
                    </div>
                    <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                        ITEM ${v.item} : CLASSE ${v.category}
                    </div>
                    <div style="font-size:11px; line-height:1.6;">
                        <strong>Produit concerné :</strong> ${v.product} (${v.dci})<br>
                        <strong>Intitulé légal :</strong> ${v.label}<br><br>
                        <strong>Justification scientifique &amp; réglementaire :</strong><br>
                        ${v.desc}<br><br>
                        <strong>Conformité Ligne Directrice ANPP :</strong><br>
                        Dossier technique complet soumis avec visa du Pharmacien Directeur Technique conformément au Décret Exécutif 22-247.
                    </div>
                </div>

                <div class="meta-box">
                    <div class="meta-row"><span class="meta-label">Réf. Dossier :</span><span class="meta-val">${v.id}</span></div>
                    <div class="meta-row"><span class="meta-label">Classification :</span><span class="meta-val">${v.category} (${v.category === 'Mmin' ? 'Mineure' : (v.category === 'Mmod' ? 'Modérée' : 'Majeure')})</span></div>
                    <div class="meta-row"><span class="meta-label">Date soumission :</span><span class="meta-val">${v.submissionDate}</span></div>
                    <div class="meta-row"><span class="meta-label">Statut commission :</span><span class="meta-val">${v.status}</span></div>
                </div>
            `;
        }
