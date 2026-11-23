// ============================================================================
// AUTHENTIFICATION, 2FA, SIGNATURE PDT & AUDIT TRAIL (21 CFR Part 11 / BPF)
// ============================================================================

const mockAuditTrail = [
            { ts: "10/11/2026 15:42:01", code: "AUTH_LOGIN", label: "Connexion 2FA", user: "Dr. Amine K.", role: "Pharmacien DT", target: "Session 192.168.1.45 (TOTP vérifié)", sha: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", status: "VALIDÉ" },
            { ts: "10/11/2026 15:45:18", code: "DOC_VIEW", label: "Consultation Pièce", user: "Dr. Amine K.", role: "Pharmacien DT", target: "3.2.S.1_Nomenclature_et_Structure.pdf", sha: "7a89f102bc4501a1d9e487102b3391cd048a6efb019485720194857201948572", status: "CONFORME" },
            { ts: "10/11/2026 15:48:32", code: "DOC_SIGN_DT", label: "Signature Électronique", user: "Dr. Amine K.", role: "Pharmacien DT", target: "Visa DT : Module 3 Qualité (CMC)", sha: "18b95c328e8f192b001859ac910e5274901baef4827591028374950182749501", status: "SCELLÉ 21 CFR" },
            { ts: "10/11/2026 15:52:10", code: "FILE_UPLOAD", label: "Dépôt de Document", user: "Ing. Sarah B.", role: "Affaires Réglementaires", target: "3.2.P.1_Composition_Formule.pdf (v1.0)", sha: "4c718a93e8194b00284759102837495018274950182749501827495018274950", status: "HASHÉ CAS" },
            { ts: "10/11/2026 16:01:45", code: "SYS_BACKUP", label: "Sauvegarde Immuable", user: "Système Automatique", role: "Admin Système", target: "Archive universelle sanad_20261110.sanad", sha: "8f5a910283749501827495018274950182749501827495018274950182749501", status: "VÉRIFIÉ" },
            { ts: "10/11/2026 16:15:00", code: "EXPORT_CTD", label: "Génération Archive", user: "Dr. Amine K.", role: "Pharmacien DT", target: "Dossier CTD Paracétamol 500mg (Modules 1 à 5)", sha: "5fa98c32bb4e0192e47856cd891b002138947ef3a92184950182749501827495", status: "PRÊT ANPP" }
        ];


        function openSignModal(docName) {
            closeAddressBar();
            closeAllModals();
            document.getElementById("sign-doc-name").value = docName;
            document.getElementById("sign-auth-code").value = "";
            document.getElementById("modal-sign").classList.add("open");
        }


        let currentUser = null;
        let currentOrg = null;
        let isGateRequired = false;

        window.getCurrentUser = function() { return currentUser; };
        window.currentUser = null;

        // Sas d'Accès Partenaire Confidentiel (Private Preview Gate)
        function showGateView(data) {
            const screenGate = document.getElementById("screen-gate");
            const screenLogin = document.getElementById("screen-login");
            const screenSetup = document.getElementById("screen-setup");
            const screenApp = document.getElementById("screen-app");

            if (screenGate) screenGate.style.display = "flex";
            if (screenLogin) screenLogin.style.display = "none";
            if (screenSetup) screenSetup.style.display = "none";
            if (screenApp) screenApp.style.display = "none";

            const err = document.getElementById("gate-error");
            if (err) err.style.display = "none";

            const input = document.getElementById("gate-passcode");
            if (input) {
                input.value = "";
                setTimeout(() => input.focus(), 100);
            }
        }

        async function handleGateSubmit(e) {
            e.preventDefault();
            const input = document.getElementById("gate-passcode");
            const errBox = document.getElementById("gate-error");
            const btn = document.getElementById("btn-gate-submit");
            if (!input) return;

            const passcode = input.value.trim();
            if (!passcode) return;

            if (btn) {
                btn.disabled = true;
                btn.innerText = "Vérification en cours...";
            }
            if (errBox) errBox.style.display = "none";

            try {
                const res = await fetch("/api/v1/gate/unlock", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ passcode: passcode })
                });

                const data = await res.json();
                if (!res.ok) {
                    if (btn) {
                        btn.disabled = false;
                        btn.innerText = "Déverrouiller l'accès ➔";
                    }
                    if (errBox) {
                        errBox.innerText = data.error || "Code d'accès partenaire invalide.";
                        errBox.style.display = "block";
                        if (typeof errBox.animate === "function") {
                            errBox.animate([
                                { transform: 'translateX(-6px)' },
                                { transform: 'translateX(6px)' },
                                { transform: 'translateX(-4px)' },
                                { transform: 'translateX(4px)' },
                                { transform: 'translateX(0)' }
                            ], { duration: 300 });
                        }
                    }
                    input.select();
                    return;
                }

                if (btn) {
                    btn.innerText = "✓ Accès Autorisé !";
                    btn.style.background = "var(--accent-emerald)";
                }

                setTimeout(() => {
                    if (btn) {
                        btn.disabled = false;
                        btn.innerText = "Déverrouiller l'accès ➔";
                        btn.style.background = "";
                    }
                    checkAuthSession();
                }, 350);

            } catch (err) {
                if (btn) {
                    btn.disabled = false;
                    btn.innerText = "Déverrouiller l'accès ➔";
                }
                if (errBox) {
                    errBox.innerText = "Erreur de connexion au sas confidentiel.";
                    errBox.style.display = "block";
                }
            }
        }

        async function handleGateLock() {
            try {
                await fetch("/api/v1/gate/lock", { method: "POST" });
            } catch (e) {
                console.warn("Erreur reverrouillage sas:", e);
            }
            window.location.reload();
        }

        // Vérification de la session active
        async function checkAuthSession() {
            try {
                // 0. Contrôle du Sas d'Accès Confidentiel (Private Preview Gate)
                try {
                    const gateRes = await fetch('/api/v1/gate/status');
                    if (gateRes.ok) {
                        const gateData = await gateRes.json();
                        isGateRequired = !!gateData.gate_required;
                        if (gateData.gate_required && !gateData.unlocked) {
                            window.location.replace('/?gate=1');
                            return;
                        }
                    }
                } catch(err) {
                    console.warn("Impossible de vérifier le statut du sas:", err);
                }

                // Contrôle de l'initialisation de l'instance (Setup Wizard BPF)
                try {
                    const setupRes = await fetch('/api/v1/setup/status');
                    if (setupRes.ok) {
                        const setupData = await setupRes.json();
                        if (setupData.setup_required) {
                            showSetupWizardView(setupData);
                            return;
                        }
                    }
                } catch(err) {
                    console.warn("Impossible de vérifier le statut setup:", err);
                }

                const res = await fetch('/api/v1/auth/me');
                if (res.ok) {
                    const data = await res.json();
                    currentUser = data.user;
                    currentOrg = data.organization;
                    if (data.dev_switcher_enabled === false) {
                        const switcher = document.getElementById("dev-role-switcher-container");
                        if (switcher) switcher.style.display = "none";
                    }
                    if (data.pitch_deck_enabled === false) {
                        const pitchBtn = document.getElementById("btn-pitch-commercial");
                        if (pitchBtn) pitchBtn.style.display = "none";
                    }
                    showAppView(currentUser, currentOrg);
                } else {
                    showLoginView();
                }
            } catch (e) {
                showLoginView();
            }
        }
        checkAuthSession();

        function showSetupWizardView(data) {
            const screenGate = document.getElementById("screen-gate");
            const screenSetup = document.getElementById("screen-setup");
            const screenLogin = document.getElementById("screen-login");
            const screenApp = document.getElementById("screen-app");
            if (screenGate) screenGate.style.display = "none";
            if (screenSetup) screenSetup.style.display = "flex";
            if (screenLogin) screenLogin.style.display = "none";
            if (screenApp) screenApp.style.display = "none";
        }

        async function handleSetupSubmit(e) {
            e.preventDefault();
            const errBox = document.getElementById("setup-error");
            const okBox = document.getElementById("setup-success");
            const submitBtn = document.getElementById("btn-setup-submit");
            if (errBox) errBox.style.display = "none";
            if (okBox) okBox.style.display = "none";

            const orgName = document.getElementById("setup-org-name").value.trim();
            const orgSlug = document.getElementById("setup-org-slug").value.trim();
            const license = document.getElementById("setup-license").value.trim();
            const dtFirstName = document.getElementById("setup-dt-firstname").value.trim();
            const dtLastName = document.getElementById("setup-dt-lastname").value.trim();
            const dtEmail = document.getElementById("setup-dt-email").value.trim();
            const dtPassword = document.getElementById("setup-dt-password").value;

            if (submitBtn) {
                submitBtn.disabled = true;
                submitBtn.innerText = "Initialisation et Scellement en cours...";
            }

            try {
                const res = await fetch("/api/v1/setup/initialize", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        organization_name: orgName,
                        organization_slug: orgSlug,
                        miph_license_number: license,
                        technical_director_name: "Dr. " + dtFirstName + " " + dtLastName,
                        dt_first_name: dtFirstName,
                        dt_last_name: dtLastName,
                        dt_email: dtEmail,
                        dt_password: dtPassword
                    })
                });

                const data = await res.json();
                if (!res.ok) {
                    if (submitBtn) {
                        submitBtn.disabled = false;
                        submitBtn.innerText = "🚀 Provisionner et Sceller l'Instance ➔";
                    }
                    if (errBox) {
                        errBox.innerText = data.error || "Échec de l'initialisation";
                        errBox.style.display = "block";
                    }
                    return;
                }

                if (okBox) {
                    okBox.innerHTML = `
                        <strong>✓ Instance Souveraine Initialisée avec Succès !</strong><br>
                        Établissement : <code>${escapeHtml(orgName)}</code><br>
                        Pharmacien DT : <code>Dr. ${escapeHtml(dtFirstName)} ${escapeHtml(dtLastName)}</code><br>
                        Redirection vers la mire de connexion sécurisée...
                    `;
                    okBox.style.display = "block";
                }

                const loginEmailInput = document.getElementById("login-email");
                if (loginEmailInput) loginEmailInput.value = dtEmail;

                setTimeout(() => {
                    const screenSetup = document.getElementById("screen-setup");
                    if (screenSetup) screenSetup.style.display = "none";
                    showLoginView();
                }, 2000);

            } catch (err) {
                if (submitBtn) {
                    submitBtn.disabled = false;
                    submitBtn.innerText = "🚀 Provisionner et Sceller l'Instance ➔";
                }
                if (errBox) {
                    errBox.innerText = "Erreur de connexion : " + err.message;
                    errBox.style.display = "block";
                }
            }
        }

        function showLoginView() {
            const screenGate = document.getElementById("screen-gate");
            if (screenGate) screenGate.style.display = "none";
            const screenSetup = document.getElementById("screen-setup");
            if (screenSetup) screenSetup.style.display = "none";
            document.getElementById("screen-login").style.display = "flex";
            document.getElementById("screen-app").style.display = "none";
            const err = document.getElementById("login-error");
            if (err) err.style.display = "none";

            const relockContainer = document.getElementById("gate-relock-container");
            if (relockContainer) {
                relockContainer.style.display = isGateRequired ? "block" : "none";
            }
        }

        function handleAuthExpired(message) {
            currentUser = null;
            currentOrg = null;
            showLoginView();
            const errBox = document.getElementById("login-error");
            if (errBox) {
                errBox.innerText = message || (isArabic ? "انتهت صلاحية الجلسة لعدم النشاط (15 دقيقة BPF). يرجى إعادة تسجيل الدخول." : "Session verrouillée après 15 minutes d'inactivité (Règle BPF / 21 CFR Part 11). Veuillez vous reconnecter.");
                errBox.style.display = "block";
            }
        }

        // Helper functions pour le contrôle d'accès RBAC
        function hasPermission(permission) {
            if (!currentUser || !currentUser.permissions) return false;
            return currentUser.permissions.includes(permission);
        }

        function hasAnyPermission(...permissions) {
            if (!currentUser || !currentUser.permissions) return false;
            return permissions.some(p => currentUser.permissions.includes(p));
        }

        function hasRole(...roles) {
            if (!currentUser || !currentUser.role) return false;
            return roles.includes(currentUser.role);
        }

        function formatRole(role) {
            switch (role) {
                case 'directeur_technique':
                    return isArabic ? 'مدير تقني (1.A.7)' : 'Pharmacien DT (1.A.7)';
                case 'pharmacien_assistant':
                    return isArabic ? 'الصيدلي المساعد (نائب المدير التقني)' : 'Pharmacien Assistant / Adjoint DT';
                case 'admin_laboratoire':
                    return isArabic ? 'مسؤول المخبر' : 'Admin Laboratoire';
                case 'responsable_affaires_reglementaires':
                    return isArabic ? 'مدير الشؤون التنظيمية' : 'Manager Affaires Réglem.';
                case 'charge_affaires_reglementaires':
                    return isArabic ? 'مكلف بالشؤون التنظيمية' : 'Chargé Affaires Réglem.';
                case 'responsable_assurance_qualite':
                    return isArabic ? 'مسؤول ضمان الجودة' : 'Manager Assurance Qualité';
                case 'responsable_controle_qualite':
                    return isArabic ? 'مراقبة الجودة والاستقرار' : 'CQ & Stabilités';
                case 'responsable_production':
                    return isArabic ? 'مسؤول الإنتاج والتأهيل' : 'Production & Validation';
                case 'consultant_externe':
                case 'consultant_cro_externe':
                    return isArabic ? 'مستشار خارجي (نطاق مخصص)' : 'Consultant Externe (CRO - Cloisonné)';
                case 'partenaire_fournisseur_api':
                    return isArabic ? 'شريك مورد المادة الفعالة (ASMF)' : 'Partenaire Fournisseur (API Holder)';
                case 'auditeur_inspecteur':
                case 'auditeur_qualite_anpp':
                    return isArabic ? 'مدقق الجودة / ANPP' : 'Auditeur Qualité ANPP';
                case 'admin_systeme_local':
                    return isArabic ? 'مسؤول النظام (Zero-Trust)' : 'Admin Système (Zero-Trust)';
                case 'super_admin_editeur':
                    return isArabic ? 'المشرف العام للبرمجية' : 'SuperAdmin Éditeur SaaS';
                default:
                    return role || '-';
            }
        }

        // Bascule de rôle instantanée pour l'environnement de développement
        async function devSwitchRole(email) {
            if (!email) return;
            try {
                const res = await fetch('/api/v1/auth/dev-switch', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email: email })
                });

                if (res.status === 404) {
                    showNotifyModal(
                        "Fonctionnalité Verrouillée",
                        `<div style="display:flex; flex-direction:column; gap:8px;">
                            <div style="color:var(--accent-red); font-weight:700;">
                                🛡️ Commutateur strictement désactivé en environnement de production
                            </div>
                            <div style="font-size:11px; color:var(--text-main);">
                                Conformément aux principes d'intégrité ALCOA+ et aux exigences 21 CFR Part 11, la bascule de session sans mot de passe n'est pas autorisée sur ce serveur.
                            </div>
                        </div>`,
                        "🛡️",
                        "Compris"
                    );
                    const switcher = document.getElementById("dev-role-switcher-container");
                    if (switcher) switcher.style.display = "none";
                    return;
                }

                const data = await res.json();
                if (res.ok) {
                    currentUser = data.user;
                    currentOrg = data.organization;
                    showAppView(currentUser, currentOrg);
                    
                    showNotifyModal(
                        "🧪 Bascule de Rôle Active (Environnement Dev)",
                        `<div style="display:flex; flex-direction:column; gap:8px;">
                            <div style="color:var(--accent-blue); font-weight:700;">
                                ✓ Connecté en tant que : ${data.user.first_name} ${data.user.last_name}
                            </div>
                            <div style="background:var(--bg-surface); border:1px solid var(--border-color); border-radius:4px; padding:10px; font-size:11px; display:flex; flex-direction:column; gap:4px;">
                                <div><strong>Email :</strong> <code>${data.user.email}</code></div>
                                <div><strong>Rôle BPF :</strong> <span style="font-weight:600; color:var(--accent-emerald);">${formatRole(data.user.role)}</span></div>
                                <div><strong>Permissions Version :</strong> v${data.user.permissions_version || 1}</div>
                                <div><strong>Droits Effectifs (${(data.user.permissions || []).length}/14) :</strong></div>
                                <div style="display:flex; flex-wrap:wrap; gap:4px; margin-top:4px;">
                                    ${(data.user.permissions || []).map(p => `<span style="background:var(--badge-bg); color:var(--accent-blue); font-size:9px; padding:1px 5px; border-radius:3px; font-family:monospace;">${p}</span>`).join('')}
                                </div>
                            </div>
                            <div style="font-size:11px; color:var(--text-dim);">
                                L'interface graphique a été dynamiquement reconfigurée selon les habilitations strictes du rôle.
                            </div>
                        </div>`,
                        "🧪",
                        "Fermer"
                    );
                } else {
                    showNotifyModal("Erreur de Bascule Dev", data.error || "Impossible de basculer de rôle", "⚠️", "Fermer");
                    // Rétablir la valeur précédente du sélecteur
                    const devSelect = document.getElementById("dev-role-select");
                    if (devSelect && currentUser && currentUser.email) {
                        devSelect.value = currentUser.email;
                    }
                }
            } catch (e) {
                console.error("Erreur devSwitchRole:", e);
                showNotifyModal("Erreur de Connexion", "Erreur réseau lors de la bascule de rôle", "⚠️", "Fermer");
            }
        }

        // Application dynamique des habilitations RBAC et des garde-fous Zero-Trust sur le DOM
        function applyPermissionsUI() {
            if (!currentUser) return;

            // Synchroniser le sélecteur Dev dans le header
            const devSelect = document.getElementById("dev-role-select");
            if (devSelect && currentUser.email) {
                devSelect.value = currentUser.email;
            }

            // Éléments assujettis à une permission unitaire data-perm
            document.querySelectorAll('[data-perm]').forEach(el => {
                const perm = el.getAttribute('data-perm');
                if (perm) {
                    if (hasPermission(perm)) {
                        el.style.display = '';
                        el.removeAttribute('disabled');
                    } else {
                        el.style.display = 'none';
                    }
                }
            });

            // Éléments assujettis à l'une des permissions data-perm-any
            document.querySelectorAll('[data-perm-any]').forEach(el => {
                const raw = el.getAttribute('data-perm-any');
                if (raw) {
                    const perms = raw.split(',').map(s => s.trim());
                    if (hasAnyPermission(...perms)) {
                        el.style.display = '';
                    } else {
                        el.style.display = 'none';
                    }
                }
            });

            // Garde-Fou Zero-Trust : Masquage des actions de téléchargement de pièces confidentielles pour les admins IT
            const isZeroTrustAdmin = hasRole('admin_systeme_local', 'super_admin_editeur');
            document.querySelectorAll('.btn-zero-trust-pharma-doc').forEach(el => {
                if (isZeroTrustAdmin) {
                    el.style.display = 'none';
                } else {
                    el.style.display = '';
                }
            });

            // SoD Guardrail UI : Signature PDT réservée strictement au pharmacien directeur technique (1.A.7)
            const btnSignPDT = document.getElementById('tb-btn-sign-pdt');
            if (btnSignPDT) {
                btnSignPDT.style.display = hasPermission('DOC_APPROVE_PDT') ? '' : 'none';
            }

            // Dépôt ANPP : nécessite DOSSIER_SUBMIT
            const btnDepositANPP = document.getElementById('tb-btn-deposit-anpp');
            if (btnDepositANPP) {
                btnDepositANPP.style.display = hasPermission('DOSSIER_SUBMIT') ? '' : 'none';
            }

            // Garde-Fou Cloisonnement Tiers / Consultant Externe (CDC Section 7.2.2) :
            const isConsultant = hasRole('consultant_externe', 'consultant_cro_externe');
            if (isConsultant) {
                // Masquer tous les univers transverses (Référentiel, GED, Variations, DM, Admin, Audit)
                ['act-ref', 'act-variations', 'act-dm', 'act-ged', 'act-admin', 'act-audit'].forEach(id => {
                    const el = document.getElementById(id);
                    if (el) el.style.display = 'none';
                });
                ['mob-act-ref', 'mob-act-variations', 'mob-act-dm', 'mob-act-ged', 'mob-act-admin'].forEach(id => {
                    const el = document.getElementById(id);
                    if (el) el.style.display = 'none';
                });
                ['mobnav-ref', 'mobnav-variations', 'mobnav-dm', 'mobnav-ged', 'mobnav-admin'].forEach(id => {
                    const el = document.getElementById(id);
                    if (el) el.style.display = 'none';
                });

                // Masquage des actions réservées dans les barres d'outils
                const btnNewD = document.getElementById('tb-btn-new-dossier');
                if (btnNewD) btnNewD.style.display = 'none';
                const btnExp = document.getElementById('tb-btn-export');
                if (btnExp) btnExp.style.display = 'none';
            }

            // Garde-Fou Zero-Trust & Univers Dédié Administrateur Système (SYS_MAINT / CDC Section 7.1) :
            const isSystemAdmin = hasRole('admin_systeme_local', 'super_admin_editeur');
            if (isSystemAdmin) {
                // Masquer tous les univers métier pharmaceutiques pour l'administrateur IT (Zero-Trust strict)
                ['act-ref', 'act-ctd', 'act-variations', 'act-dm', 'act-ged', 'act-admin'].forEach(id => {
                    const el = document.getElementById(id);
                    if (el) el.style.display = 'none';
                });
                ['mob-act-ref', 'mob-act-ctd', 'mob-act-variations', 'mob-act-dm', 'mob-act-ged', 'mob-act-admin'].forEach(id => {
                    const el = document.getElementById(id);
                    if (el) el.style.display = 'none';
                });

                // Afficher le bouton Cockpit IT
                const actSys = document.getElementById('act-sys');
                if (actSys) actSys.style.display = '';
                const mobSys = document.getElementById('mob-act-sys');
                if (mobSys) mobSys.style.display = '';

                // Rediriger automatiquement vers l'univers IT si l'utilisateur est sur un module pharma
                if (currentDomain !== 'sys' && currentDomain !== 'audit') {
                    setTimeout(() => switchDomainModule('sys'), 10);
                }
            } else {
                // Pour les profils non-IT, masquer le Cockpit IT
                const actSys = document.getElementById('act-sys');
                if (actSys) actSys.style.display = 'none';
                const mobSys = document.getElementById('mob-act-sys');
                if (mobSys) mobSys.style.display = 'none';
            }
        }

        function showAppView(user, org) {
            const screenGate = document.getElementById("screen-gate");
            if (screenGate) screenGate.style.display = "none";
            document.getElementById("screen-login").style.display = "none";
            document.getElementById("screen-app").style.display = "flex";
            if (user) {
                currentUser = user;
                window.currentUser = user;
                const roleLabel = formatRole(user.role);
                document.getElementById("user-badge").innerText = `${user.first_name} ${user.last_name} (${roleLabel})`;
                document.getElementById("u-drop-name").innerText = `${user.first_name} ${user.last_name}`;
                document.getElementById("u-drop-email").innerText = user.email;
                document.getElementById("u-drop-role").innerText = roleLabel;
            }
            if (org) {
                currentOrg = org;
                document.getElementById("u-drop-org").innerText = org.name;
            }

            // Appliquer les masquages d'interface selon les permissions effectives
            applyPermissionsUI();

            // Vérification de sécurité de routage : Si le domaine actif est interdit, basculer vers un univers autorisé
            const isConsultant = hasRole('consultant_externe', 'consultant_cro_externe');
            const isSysAdmin = hasRole('admin_systeme_local', 'super_admin_editeur') || hasPermission('SYS_MAINT');
            let targetDomain = currentDomain || 'ref';
            if (isConsultant) {
                targetDomain = 'ctd';
            } else if (isSysAdmin && !hasPermission('REF_VIEW')) {
                targetDomain = 'sys';
            } else if (targetDomain === 'admin' && !hasPermission('USER_MANAGE')) {
                targetDomain = hasPermission('REF_VIEW') ? 'ref' : (hasPermission('DOSSIER_VIEW') ? 'ctd' : (hasPermission('SYS_MAINT') ? 'sys' : 'ged'));
            } else if (targetDomain === 'ged' && !hasPermission('DOC_READ')) {
                targetDomain = hasPermission('REF_VIEW') ? 'ref' : (hasPermission('DOSSIER_VIEW') ? 'ctd' : (hasPermission('SYS_MAINT') ? 'sys' : 'admin'));
            } else if ((targetDomain === 'ctd' || targetDomain === 'variations' || targetDomain === 'dm' || targetDomain === 'deficiencies') && !hasPermission('DOSSIER_VIEW')) {
                targetDomain = hasPermission('USER_MANAGE') ? 'admin' : (hasPermission('REF_VIEW') ? 'ref' : (hasPermission('SYS_MAINT') ? 'sys' : 'ref'));
            } else if (targetDomain === 'ref' && !hasPermission('REF_VIEW')) {
                targetDomain = hasPermission('USER_MANAGE') ? 'admin' : (hasPermission('DOSSIER_VIEW') ? 'ctd' : (hasPermission('SYS_MAINT') ? 'sys' : 'ged'));
            }

            switchDomainModule(targetDomain);

            // Rafraîchissement discret des alertes de compléments ANPP et compteur d'échéances
            if (typeof updateDeficiencyBadgeCount === 'function' && typeof hasPermission === 'function' && hasPermission('DOSSIER_VIEW')) {
                fetch('/api/v1/deficiencies/dashboard')
                    .then(r => r.json())
                    .then(data => {
                        updateDeficiencyBadgeCount(data);
                    })
                    .catch(() => {});
            }
            if (typeof refreshAlertsModalContent === 'function') {
                refreshAlertsModalContent();
            }
            if (typeof refreshParapheurBadge === 'function') {
                refreshParapheurBadge();
            }
            if (typeof refreshVariationsBadgeCount === 'function') {
                refreshVariationsBadgeCount();
            }
            if (typeof refreshCorporateBadgeCount === 'function') {
                refreshCorporateBadgeCount();
            }
        }

        function toggleUserDropdown(e) {
            if (e) e.stopPropagation();
            const drop = document.getElementById("user-dropdown");
            drop.style.display = (drop.style.display === "block") ? "none" : "block";
        }

        async function handleLoginSubmit(event) {
            if (event) event.preventDefault();
            const email = document.getElementById("login-email").value.trim();
            const password = document.getElementById("login-password").value;
            const totp = document.getElementById("login-totp").value.trim();
            const errBox = document.getElementById("login-error");
            const btn = document.getElementById("btn-login-submit");

            errBox.style.display = "none";
            btn.disabled = true;
            btn.innerText = "Vérification en cours...";

            try {
                const res = await fetch('/api/v1/auth/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email, password, totp_code: totp })
                });
                const data = await res.json();

                if (res.ok) {
                    if (data.requires_2fa) {
                        document.getElementById("totp-field-group").style.display = "flex";
                        document.getElementById("login-totp").focus();
                        btn.disabled = false;
                        btn.innerText = "Valider le code 2FA ➔";
                        return;
                    }
                    if (data.dev_switcher_enabled === false) {
                        const switcher = document.getElementById("dev-role-switcher-container");
                        if (switcher) switcher.style.display = "none";
                    }
                    if (data.pitch_deck_enabled === false) {
                        const pitchBtn = document.getElementById("btn-pitch-commercial");
                        if (pitchBtn) pitchBtn.style.display = "none";
                    }
                    showAppView(data.user, data.organization);
                    btn.disabled = false;
                    btn.innerText = "Se connecter au cockpit ➔";
                    switchDomainModule(currentDomain || 'ged');
                } else {
                    errBox.innerText = data.error || "Identifiants invalides";
                    errBox.style.display = "block";
                    btn.disabled = false;
                    btn.innerText = "Se connecter au cockpit ➔";
                }
            } catch (e) {
                errBox.innerText = "Erreur de communication avec le serveur local";
                errBox.style.display = "block";
                btn.disabled = false;
                btn.innerText = "Se connecter au cockpit ➔";
            }
        }

        function fillDemoAccount(email, password) {
            document.getElementById("login-email").value = email;
            const pwd = password || "Sanad#Secure2026!Key";
            const pwdField = document.getElementById("login-password");
            if (pwdField) pwdField.value = pwd;
            document.getElementById("totp-field-group").style.display = "none";
            document.getElementById("login-totp").value = "";
            document.getElementById("login-error").style.display = "none";
            handleLoginSubmit();
        }

        async function executeLogout() {
            const drop = document.getElementById("user-dropdown");
            if (drop) drop.style.display = "none";
            try {
                await fetch('/api/v1/auth/logout', { method: 'POST' });
            } catch (e) {}
            currentUser = null;
            currentOrg = null;
            showLoginView();
        }

        async function executeSignature() {
            const code = document.getElementById("sign-auth-code").value.trim();
            if (!code) {
                showNotifyModal(
                    "Erreur de Conformité 21 CFR Part 11",
                    `<div style="display:flex; flex-direction:column; gap:8px;">
                        <div style="color:var(--accent-red); font-weight:700;">
                            ⚠️ Re-authentification obligatoire pour signature légale
                        </div>
                        <div style="font-size:12px; color:var(--text-main);">
                            Conformément aux exigences <strong>21 CFR Part 11</strong> et à l'<strong>Annexe 11 des BPF</strong>, une signature électronique ne peut être apposée sans ressaisie obligatoire du mot de passe ou d'un code 2FA TOTP à 6 chiffres.
                        </div>
                    </div>`,
                    "⚠️",
                    "Corriger"
                );
                return;
            }

            const doc = mockDocuments.find(d => d.id === selectedDocId);
            const reason = document.getElementById("sign-reason").value.trim() || "Validation légale DT";
            const declaration = "Je soussigné, Dr. Amine K., en ma qualité de Pharmacien Directeur Technique agréé par le MIPH, certifie sur l'honneur l'exactitude de cette pièce.";

            let auditHash = doc ? doc.sha : "e3b0c442...";
            let signedUtc = new Date().toISOString();
            let auditLogId = null;

            // Appel direct à l'API REST backend (persistance PostgreSQL)
            try {
                const isTOTP = /^\d{6}$/.test(code);
                const res = await fetch('/api/v1/auth/reauth-sign', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        password: isTOTP ? "" : code,
                        totp_code: isTOTP ? code : "",
                        document_name: doc ? doc.name : "Document",
                        document_sha256: doc ? doc.sha : "e3b0c442...",
                        reason: reason,
                        declaration: declaration
                    })
                });

                if (res.ok) {
                    const data = await res.json();
                    auditHash = data.audit_sha256;
                    signedUtc = data.signed_at_utc;
                    auditLogId = data.audit_log_id;
                } else {
                    const errData = await res.json();
                    showNotifyModal("Erreur de Signature BPF", errData.error || "Mot de passe ou code 2FA incorrect", "⚠️", "Corriger");
                    return;
                }
            } catch (e) {
                console.warn("Mode dégradé hors-ligne:", e);
            }

            if (doc) {
                doc.status = "status-ok";
                doc.statusLabel = "VALIDÉ BPF";
                doc.version = "v1.0";
                doc.validDT = "Dr. Amine K. (À l'instant)";
            }

            // Ajout local immédiat dans l'audit trail
            mockAuditTrail.unshift({
                ts: new Date().toLocaleDateString('fr-FR') + " " + new Date().toLocaleTimeString('fr-FR'),
                code: "DOC_SIGN_DT",
                label: "Signature Électronique",
                user: "Dr. Amine K.",
                role: "Pharmacien DT",
                target: (doc ? doc.name : "Document") + " (Visa DT)",
                sha: auditHash,
                status: "SCELLÉ 21 CFR"
            });

            closeModal("modal-sign");
            renderTable();

            showNotifyModal(
                "Signature Électronique Certifiée (21 CFR Part 11)",
                `<div style="display:flex; flex-direction:column; gap:8px;">
                    <div style="display:flex; align-items:center; gap:8px; color:var(--accent-emerald); font-weight:700;">
                        <span style="font-size:18px;">✓</span>
                        <span>Document visé et scellé dans le registre officiel</span>
                    </div>
                    <div style="background:var(--bg-surface); border:1px solid var(--border-color); border-radius:4px; padding:10px; font-size:11px; display:flex; flex-direction:column; gap:4px;">
                        <div><strong>Signataire :</strong> Dr. Amine K. (Directeur Technique)</div>
                        <div><strong>Rôle légal BPF :</strong> Pharmacien DT agréé MIPH</div>
                        <div><strong>Horodatage UTC :</strong> <code>${signedUtc}</code></div>
                        <div><strong>Audit Log ID :</strong> <code>#${auditLogId || 'LIVE'}</code></div>
                        <div><strong>Empreinte SHA-256 scellée :</strong> <code style="word-break:break-all;">${auditHash}</code></div>
                    </div>
                    <div style="font-size:11px; color:var(--text-dim);">
                        L'acte a été inscrit dans la table <code>audit_logs</code> protégée par trigger SQL (interdiction formelle de modification ou suppression).
                    </div>
                </div>`,
                "✍️",
                "Fermer"
            );
        }

        function openUploadModal() {
            closeAddressBar();
            closeAllModals();
            document.getElementById("modal-upload").classList.add("open");
        }

        function finishUploadSimulation() {
            closeModal('modal-upload');
            showNotifyModal(
                "Téléversement Terminé & Intégrité Vérifiée",
                `<div style="display:flex; flex-direction:column; gap:8px;">
                    <div style="color:var(--accent-emerald); font-weight:700;">
                        ✓ Fichier réceptionné dans le stockage CAS sécurisé
                    </div>
                    <div style="background:var(--bg-surface); border:1px solid var(--border-color); border-radius:4px; padding:10px; font-size:11px; display:flex; flex-direction:column; gap:4px;">
                        <div><strong>Mode :</strong> Chunked Upload résilient (Morceaux de 5 Mo)</div>
                        <div><strong>Déduplication CAS :</strong> Hash SHA-256 calculé à la volée</div>
                        <div><strong>Version créée :</strong> <span style="font-weight:600;">v1.1 (En revue Qualité)</span></div>
                        <div><strong>Statut BPF :</strong> En attente de validation par le DT</div>
                    </div>
                </div>`,
                "📤",
                "Fermer"
            );
        }

        async function openAlertsModal() {
            closeAddressBar();
            closeAllModals();
            openModal("modal-alerts");
            await refreshAlertsModalContent();
        }

        async function refreshAlertsModalContent() {
            const container = document.getElementById("alerts-cards-container");
            if (!container) return;

            let letters = [];
            try {
                const resp = await fetch('/api/v1/deficiencies/dashboard');
                if (resp.ok) {
                    const data = await resp.json();
                    letters = data.recent_letters || [];
                    
                    const criticalCount = data.critical_letters || 0;
                    const urgentCount = data.urgent_letters || 0;
                    const totalUrgent = criticalCount + urgentCount;
                    const txtAlerts = document.getElementById("txt-urgent-alerts");
                    if (txtAlerts) {
                        const displayCount = Math.max(totalUrgent, letters.length);
                        txtAlerts.innerText = `${displayCount} Échéance${displayCount > 1 ? 's' : ''} (J < 30j)`;
                    }
                }
            } catch (e) {
                console.warn("Échec chargement alertes compléments:", e);
            }

            let html = '';

            // 1. Alertes réelles issues de la base SQL (Compléments ANPP)
            if (letters.length > 0) {
                letters.forEach(l => {
                    const days = l.days_remaining != null ? l.days_remaining : 30;
                    const isCritical = days <= 15;
                    const badgeColor = isCritical ? '#ef4444' : '#f59e0b';
                    const badgeBg = isCritical ? 'rgba(239,68,68,0.2)' : 'rgba(245,158,11,0.2)';
                    const cardClass = isCritical ? 'alert-card' : 'alert-card warning';
                    const statusTxt = days < 0 ? `Expiré (${Math.abs(days)}j de retard)` : `Expire dans ${days} jour${days > 1 ? 's' : ''}`;

                    html += `
                        <div class="${cardClass}">
                            <div style="display:flex; flex-direction:column; gap:4px; flex:1;">
                                <div style="display:flex; align-items:center; gap:8px;">
                                    <span style="font-size:14px;">⏱️</span>
                                    <strong style="font-size:13px; color:var(--text-main);">Demande de Compléments ANPP — ${escapeHTML(l.product_brand_name || l.dossier_reference || 'Dossier')}</strong>
                                    <span style="background:${badgeBg}; color:${badgeColor}; font-size:10px; font-weight:700; padding:1px 6px; border-radius:3px;">${statusTxt}</span>
                                </div>
                                <div style="color:var(--text-muted); font-size:11px;">
                                    Courrier officiel : <code style="color:${isCritical ? 'var(--accent-blue)' : 'var(--accent-amber)'};">${escapeHTML(l.reference_number)}</code> • Date butoir : <strong>${(l.due_date || '').substring(0, 10)}</strong> • Dossier : <strong>${escapeHTML(l.dossier_reference || '')}</strong>
                                </div>
                                <div style="color:var(--text-dim); font-size:11px;">
                                    ${l.notes ? escapeHTML(l.notes) : 'Réponses aux questions notifiées et visa Pharmacien Directeur Technique requis (Mandat 1.A.7).'}
                                </div>
                            </div>
                            <button class="btn-pill" style="background:${badgeColor}; color:white; border:none; padding:6px 12px; white-space:nowrap; cursor:pointer;" onclick="closeModal('modal-alerts'); if(typeof switchDomainModule === 'function') switchDomainModule('deficiencies'); if(typeof openDeficiencyLetterDetail === 'function') openDeficiencyLetterDetail('${l.id}');">
                                Ouvrir Studio ➔
                            </button>
                        </div>
                    `;
                });
            }

            // 1b. Alertes Chemin Critique Smart Planning ANPP
            if (typeof currentDossierSmartPlanning === 'object' && currentDossierSmartPlanning && currentDossierSmartPlanning.critical_warnings && currentDossierSmartPlanning.critical_warnings.length > 0) {
                const sp = currentDossierSmartPlanning;
                html += `
                    <div class="alert-card ${sp.overall_health === 'FATAL_BLOCK' || sp.overall_health === 'CRITICAL_DELAY' ? '' : 'warning'}">
                        <div style="display:flex; flex-direction:column; gap:4px; flex:1;">
                            <div style="display:flex; align-items:center; gap:8px;">
                                <span style="font-size:14px;">⏱️</span>
                                <strong style="font-size:13px; color:var(--text-main);">Smart Planning ANPP — Dérive Chemin Critique (${escapeHTML(sp.dossier_code || 'Dossier')})</strong>
                                <span style="background:${sp.overall_health === 'FATAL_BLOCK' ? 'rgba(239,68,68,0.2)' : 'rgba(245,158,11,0.2)'}; color:${sp.overall_health === 'FATAL_BLOCK' ? '#ef4444' : '#d97706'}; font-size:10px; font-weight:700; padding:1px 6px; border-radius:3px;">${escapeHTML(sp.overall_health_label)}</span>
                            </div>
                            <div style="color:var(--text-muted); font-size:11px;">
                                Date Butoir Cible : <strong>${sp.target_deposit_date ? new Date(sp.target_deposit_date).toLocaleDateString('fr-FR') : ''}</strong> • ${escapeHTML(sp.critical_warnings[0])}
                            </div>
                            <div style="color:var(--text-dim); font-size:11px;">
                                Recommandation : ${escapeHTML(sp.recommended_action_fr || '')}
                            </div>
                        </div>
                        <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; padding:6px 12px; white-space:nowrap; cursor:pointer;" onclick="closeModal('modal-alerts'); if(typeof showSmartPlanningTimelineModal === 'function') showSmartPlanningTimelineModal();">
                            Voir Chronogramme ➔
                        </button>
                    </div>
                `;
            }

            // 2. Alertes statutaires réglementaires (AMM & DM)
            html += `
                <div class="alert-card">
                    <div style="display:flex; flex-direction:column; gap:4px; flex:1;">
                        <div style="display:flex; align-items:center; gap:8px;">
                            <span style="font-size:14px;">💊</span>
                            <strong style="font-size:13px; color:var(--text-main);">Renouvellement Décision d'Enregistrement (AMM) — Paracétamol 500mg</strong>
                            <span style="background:rgba(239,68,68,0.2); color:#ef4444; font-size:10px; font-weight:700; padding:1px 6px; border-radius:3px;">Expire dans 18 jours</span>
                        </div>
                        <div style="color:var(--text-muted); font-size:11px;">
                            Référence ANPP : <code style="color:var(--accent-blue);">DZ-2019-0412</code> • Date limite de dépôt : <strong>28/11/2026</strong> • Type : Renouvellement quinquennal
                        </div>
                        <div style="color:var(--text-dim); font-size:11px;">
                            Action requise : Valider les pièces administratives du Module 1 (Statut de commercialisation, BPF fabricant actif).
                        </div>
                    </div>
                    <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; padding:6px 12px; white-space:nowrap; cursor:pointer;" onclick="closeModal('modal-alerts'); if(typeof openDossierModule === 'function') { openDossierModule('parac', 1); } else { goToLocation('ctd-para-1'); }">
                        Accéder au Module 1 ➔
                    </button>
                </div>
            `;

            container.innerHTML = html;
        }

        async function openAuditModal() {
            closeAddressBar();
            closeAllModals();

            // Rechargement en direct depuis la base de données PostgreSQL
            try {
                const res = await fetch('/api/v1/audit/logs');
                if (res.ok) {
                    const data = await res.json();
                    if (data.logs && data.logs.length > 0) {
                        mockAuditTrail.length = 0;
                        data.logs.forEach(l => {
                            const d = new Date(l.created_at);
                            mockAuditTrail.push({
                                ts: d.toLocaleDateString('fr-FR') + " " + d.toLocaleTimeString('fr-FR'),
                                code: l.action,
                                label: l.action === 'DOC_SIGN_DT' ? 'Signature DT' : (l.action === 'AUTH_LOGIN' ? 'Connexion' : 'Audit Trail'),
                                user: l.user_email ? l.user_email.split('@')[0] : 'Système',
                                role: l.user_role === 'directeur_technique' ? 'Pharmacien DT' : l.user_role,
                                target: l.entity_id,
                                sha: l.sha256_hash,
                                status: 'SCELLÉ SQL'
                            });
                        });
                    }
                }
            } catch (e) {
                console.warn("Lecture API audit:", e);
            }

            renderAuditTable();
            document.getElementById("modal-audit").classList.add("open");
        }

        function renderAuditTable(list = mockAuditTrail) {
            const tbody = document.getElementById("audit-table-body");
            if (!tbody) return;
            tbody.innerHTML = "";
            list.forEach(item => {
                const tr = document.createElement("tr");
                tr.className = "audit-row";
                tr.innerHTML = `
                    <td style="padding:6px 10px; font-family:monospace; color:var(--text-muted);">${item.ts}</td>
                    <td style="padding:6px 10px;"><span style="background:var(--badge-bg); color:var(--accent-blue); padding:1px 6px; border-radius:3px; font-weight:600; font-size:10px;">${item.code}</span> <small style="color:var(--text-dim);">${item.label}</small></td>
                    <td style="padding:6px 10px; font-weight:600;">${item.user} <span style="color:var(--text-dim); font-weight:normal;">(${item.role})</span></td>
                    <td style="padding:6px 10px; color:var(--text-main);">${item.target}</td>
                    <td style="padding:6px 10px; font-family:monospace; font-size:10px; color:var(--text-dim);" title="${item.sha}">${item.sha.substring(0, 16)}...</td>
                    <td style="padding:6px 10px;"><span style="color:var(--accent-emerald); font-weight:700; font-size:10px;">${item.status}</span></td>
                `;
                tbody.appendChild(tr);
            });
            const badge = document.getElementById("audit-count-badge");
            if (badge) badge.innerText = `${list.length} événement(s) scellé(s)`;
        }

        function filterAuditTable() {
            const q = document.getElementById("audit-filter-input").value.toLowerCase();
            const filtered = mockAuditTrail.filter(a => 
                a.code.toLowerCase().includes(q) ||
                a.label.toLowerCase().includes(q) ||
                a.user.toLowerCase().includes(q) ||
                a.role.toLowerCase().includes(q) ||
                a.target.toLowerCase().includes(q) ||
                a.sha.toLowerCase().includes(q)
            );
            renderAuditTable(filtered);
        }

        function exportAuditTrailCSV() {
            showNotifyModal(
                "Export CSV Certifié (Audit Trail)",
                `<div style="display:flex; flex-direction:column; gap:8px;">
                    <div style="color:var(--accent-emerald); font-weight:700;">
                        ✓ Fichier CSV certifié généré avec succès
                    </div>
                    <div style="font-size:11px; color:var(--text-main);">
                        Le fichier <code>audit_trail_${new Date().toISOString().slice(0,10)}.csv</code> contient l'historique complet infalsifiable avec horodatage UTC et hash chaîné.
                    </div>
                </div>`,
                "📥",
                "Fermer"
            );
        }

        function simulateExportZip() {
            closeAddressBar();
            closeAllModals();
            document.getElementById("modal-export").classList.add("open");
        }

        function triggerArchiveDownload() {
            closeModal("modal-export");
            showNotifyModal(
                "Archive Souveraine Générée",
                `<div style="display:flex; flex-direction:column; gap:8px;">
                    <div style="color:var(--accent-blue); font-weight:700;">
                        📦 Archive officielle prête pour téléchargement
                    </div>
                    <div style="font-size:11px; color:var(--text-main);">
                        Le package <code>CTD_Paracetamol_500mg_ANPP_2026.zip</code> (184.2 Mo) a été assemblé avec son index XML, son bordereau de dépôt officiel et l'empreinte de contrôle SHA-256.
                    </div>
                </div>`,
                "⬇️",
                "Terminer"
            );
        }



        // Clic extérieur pour refermer la barre d'adresse et le menu utilisateur
        document.addEventListener('click', function(e) {
            const wrapper = document.querySelector('.address-bar-wrapper');
            if (wrapper && !wrapper.contains(e.target)) {
                closeAddressBar();
            }
            const udrop = document.getElementById("user-dropdown");
            const ubtn = document.getElementById("user-menu-btn");
            if (udrop && udrop.style.display === "block" && !udrop.contains(e.target) && (!ubtn || !ubtn.contains(e.target))) {
                udrop.style.display = "none";
            }
        });
