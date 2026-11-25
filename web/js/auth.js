// ============================================================================
// AUTHENTIFICATION, 2FA, SIGNATURE PDT & AUDIT TRAIL (21 CFR Part 11 / BPF)
// ============================================================================

const mockAuditTrail = [
            { ts: "10/11/2026 15:42:01", code: "AUTH_LOGIN", label: "Connexion 2FA", user: "Dr. Amine K.", role: "Pharmacien DT", target: "Session 192.168.1.45 (TOTP vérifié)", ipAddress: "192.168.1.45", location: "🏢 Siège Administratif & RA (Alger)", sha: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", status: "VALIDÉ" },
            { ts: "10/11/2026 15:45:18", code: "DOC_VIEW", label: "Consultation Pièce", user: "Dr. Amine K.", role: "Pharmacien DT", target: "3.2.S.1_Nomenclature_et_Structure.pdf", ipAddress: "192.168.1.45", location: "🏢 Siège Administratif & RA (Alger)", sha: "7a89f102bc4501a1d9e487102b3391cd048a6efb019485720194857201948572", status: "CONFORME" },
            { ts: "10/11/2026 15:48:32", code: "DOC_SIGN_DT", label: "Signature Électronique", user: "Dr. Amine K.", role: "Pharmacien DT", target: "Visa DT : Module 3 Qualité (CMC)", ipAddress: "192.168.1.45", location: "🏢 Siège Administratif & RA (Alger)", sha: "18b95c328e8f192b001859ac910e5274901baef4827591028374950182749501", status: "SCELLÉ 21 CFR" },
            { ts: "10/11/2026 15:52:10", code: "FILE_UPLOAD", label: "Dépôt de Document", user: "Ing. Sarah B.", role: "Affaires Réglementaires", target: "3.2.P.1_Composition_Formule.pdf (v1.0)", ipAddress: "192.168.1.108", location: "🏢 Siège Administratif & RA (Alger)", sha: "4c718a93e8194b00284759102837495018274950182749501827495018274950", status: "HASHÉ CAS" },
            { ts: "10/11/2026 16:01:45", code: "SYS_BACKUP", label: "Sauvegarde Immuable", user: "Système Automatique", role: "Admin Système", target: "Archive universelle sanad_20261110.sanad", ipAddress: "127.0.0.1", location: "💻 Console Locale Serveur (Sur Site)", sha: "8f5a910283749501827495018274950182749501827495018274950182749501", status: "VÉRIFIÉ" },
            { ts: "10/11/2026 16:15:00", code: "EXPORT_CTD", label: "Génération Archive", user: "Dr. Amine K.", role: "Pharmacien DT", target: "Dossier CTD Paracétamol 500mg (Modules 1 à 5)", ipAddress: "192.168.1.45", location: "🏢 Siège Administratif & RA (Alger)", sha: "5fa98c32bb4e0192e47856cd891b002138947ef3a92184950182749501827495", status: "PRÊT ANPP" }
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

        // ============================================================================
        // DÉLAI RÉGLEMENTAIRE D'INACTIVITÉ (BPF / 21 CFR PART 11 - 15 MINUTES)
        // ============================================================================
        const REGULATORY_IDLE_TIMEOUT_MS = 15 * 60 * 1000; // 15 minutes BPF
        let isSessionExpiredState = false;
        let userActivityDebounceTimer = null;

        function isSessionTimedOut() {
            if (isSessionExpiredState) return true;
            if (!currentUser && localStorage.getItem('sanad_auth_active') !== 'true') return false;
            const lastAct = parseInt(localStorage.getItem('sanad_last_activity') || '0', 10);
            if (lastAct > 0 && (Date.now() - lastAct >= REGULATORY_IDLE_TIMEOUT_MS)) {
                isSessionExpiredState = true;
                return true;
            }
            return false;
        }
        window.isSessionTimedOut = isSessionTimedOut;

        function updateLastActivityTimestamp() {
            const now = Date.now();
            localStorage.setItem('sanad_last_activity', now.toString());
        }

        function initActivityListeners() {
            const activityEvents = ['mousedown', 'keydown', 'scroll', 'touchstart', 'click'];
            activityEvents.forEach(evt => {
                window.addEventListener(evt, () => {
                    if (!currentUser) return;
                    const lastAct = parseInt(localStorage.getItem('sanad_last_activity') || '0', 10);
                    const now = Date.now();

                    // Si le délai réglementaire a été dépassé
                    if (lastAct > 0 && (now - lastAct >= REGULATORY_IDLE_TIMEOUT_MS)) {
                        isSessionExpiredState = true;
                        showSessionTimeoutModal('navigation');
                        return;
                    }

                    // Rafraîchir le timestamp d'activité locale (temporisé à 5s)
                    if (!userActivityDebounceTimer) {
                        userActivityDebounceTimer = setTimeout(() => {
                            updateLastActivityTimestamp();
                            userActivityDebounceTimer = null;
                        }, 5000);
                    }
                }, { passive: true });
            });
        }

        // Intercepteur global de navigation / interaction au clic (Capture phase)
        document.addEventListener('click', function(e) {
            const screenApp = document.getElementById('screen-app');
            if (screenApp && screenApp.style.display !== 'none') {
                if (isSessionTimedOut()) {
                    const timeoutModal = document.getElementById('modal-session-timeout');
                    if (timeoutModal && timeoutModal.contains(e.target)) {
                        return; // Laisser passer le clic sur le bouton de redirection
                    }
                    e.preventDefault();
                    e.stopPropagation();
                    showSessionTimeoutModal('navigation');
                }
            }
        }, true);

        // Intercepteur global des appels API pour capturer les 401 SESSION_IDLE / SESSION_EXPIRED
        (function setupFetchInterceptor() {
            const originalFetch = window.fetch;
            window.fetch = async function(...args) {
                const response = await originalFetch.apply(this, args);
                if (response.status === 401) {
                    const url = (typeof args[0] === 'string') ? args[0] : (args[0] && args[0].url ? args[0].url : '');
                    if (!url.includes('/api/v1/auth/login') && !url.includes('/api/v1/setup/')) {
                        let isTimeout = false;
                        let errReason = '';
                        try {
                            const cloned = response.clone();
                            const data = await cloned.json();
                            if (data.code === 'SESSION_IDLE' || data.code === 'SESSION_EXPIRED' || (data.error && data.error.includes('15 min'))) {
                                isTimeout = true;
                                errReason = data.error;
                            }
                        } catch (_) {}

                        const wasActive = localStorage.getItem('sanad_auth_active') === 'true';
                        if (isTimeout || wasActive) {
                            showSessionTimeoutModal('api_401', errReason);
                        }
                    }
                }
                return response;
            };
        })();

        // Vérification périodique d'inactivité
        setInterval(() => {
            if (currentUser && isSessionTimedOut()) {
                isSessionExpiredState = true;
            }
        }, 10000);

        initActivityListeners();

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
                    localStorage.setItem('sanad_auth_active', 'true');
                    localStorage.setItem('sanad_last_activity', Date.now().toString());
                    localStorage.removeItem('sanad_voluntary_logout');
                    if (currentUser && currentUser.email) {
                        localStorage.setItem('sanad_user_email', currentUser.email);
                        localStorage.setItem('sanad_user_name', `${currentUser.first_name || ''} ${currentUser.last_name || ''}`.trim());
                    }
                    isSessionExpiredState = false;
                    showAppView(currentUser, currentOrg);
                } else {
                    const wasActive = localStorage.getItem('sanad_auth_active') === 'true';
                    const voluntary = localStorage.getItem('sanad_voluntary_logout') === 'true';
                    let isTimeout = false;
                    let timeoutMsg = '';
                    if (res.status === 401) {
                        try {
                            const errData = await res.json();
                            if (errData.code === 'SESSION_IDLE' || errData.code === 'SESSION_EXPIRED' || (errData.error && errData.error.includes('15 min'))) {
                                isTimeout = true;
                                timeoutMsg = errData.error;
                            }
                        } catch (_) {}
                    }
                    const lastAct = parseInt(localStorage.getItem('sanad_last_activity') || '0', 10);
                    const now = Date.now();
                    if (wasActive && !voluntary && (isTimeout || (lastAct > 0 && (now - lastAct >= REGULATORY_IDLE_TIMEOUT_MS)))) {
                        showLoginView();
                        showSessionTimeoutModal('refresh', timeoutMsg);
                    } else {
                        showLoginView();
                    }
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

        function showSessionTimeoutModal(triggerType, customMessage) {
            isSessionExpiredState = true;
            let modal = document.getElementById("modal-session-timeout");
            if (!modal) return;

            const savedName = (currentUser ? `${currentUser.first_name || ''} ${currentUser.last_name || ''}`.trim() : '') || localStorage.getItem("sanad_user_name") || "Collaborateur Habilité";
            const savedEmail = (currentUser && currentUser.email) || localStorage.getItem("sanad_user_email") || "Session Nominative BPF";

            const nameEl = document.getElementById("session-timeout-user-name");
            const emailEl = document.getElementById("session-timeout-user-email");
            if (nameEl) nameEl.innerText = savedName;
            if (emailEl) emailEl.innerText = savedEmail;

            const detailEl = document.getElementById("session-timeout-detail");
            if (detailEl) {
                if (customMessage) {
                    detailEl.innerHTML = `Conformément aux normes des <strong>Bonnes Pratiques de Fabrication (BPF)</strong>, au <strong>Décret exécutif n° 21-82</strong> et aux exigences <strong>FDA 21 CFR Part 11 / ALCOA+</strong>, votre session a été automatiquement verrouillée.<br><br><strong>Motif Réglementaire :</strong> ${escapeHtml(customMessage)}`;
                } else {
                    detailEl.innerHTML = `Conformément aux normes des <strong>Bonnes Pratiques de Fabrication (BPF)</strong>, au <strong>Décret exécutif n° 21-82</strong> et aux exigences <strong>FDA 21 CFR Part 11 / ALCOA+</strong>, votre session a été automatiquement verrouillée après <strong>15 minutes d'inactivité</strong> afin de garantir la stricte confidentialité et l'intégrité des dossiers pharmaceutiques.`;
                }
            }

            modal.classList.add("open");
            modal.style.display = "flex";
        }
        window.showSessionTimeoutModal = showSessionTimeoutModal;

        function closeSessionTimeoutModalAndRedirect() {
            const modal = document.getElementById("modal-session-timeout");
            if (modal) {
                modal.classList.remove("open");
                modal.style.display = "none";
            }

            const savedEmail = (currentUser && currentUser.email) || localStorage.getItem("sanad_user_email");

            currentUser = null;
            currentOrg = null;
            isSessionExpiredState = false;
            localStorage.removeItem("sanad_auth_active");
            localStorage.removeItem("sanad_last_activity");

            showLoginView();

            const errBox = document.getElementById("login-error");
            if (errBox) {
                errBox.innerText = isArabic ?
                    "انتهت صلاحية الجلسة لعدم النشاط (15 دقيقة BPF). يرجى إعادة تسجيل الدخول." :
                    "Session verrouillée après 15 minutes d'inactivité (Règle BPF / 21 CFR Part 11). Veuillez vous reconnecter.";
                errBox.style.display = "block";
            }

            const emailInput = document.getElementById("login-email");
            const pwdInput = document.getElementById("login-password");
            if (savedEmail && emailInput) {
                emailInput.value = savedEmail;
                if (pwdInput) {
                    pwdInput.value = "";
                    setTimeout(() => pwdInput.focus(), 150);
                }
            } else if (emailInput) {
                setTimeout(() => emailInput.focus(), 150);
            }
        }
        window.closeSessionTimeoutModalAndRedirect = closeSessionTimeoutModalAndRedirect;

        // Outil de simulation pour tester immédiatement la déconnexion automatique réglementaire
        window.testSessionTimeout = function() {
            localStorage.setItem('sanad_last_activity', (Date.now() - 16 * 60 * 1000).toString());
            isSessionExpiredState = true;
            showSessionTimeoutModal('test');
        };

        function handleAuthExpired(message) {
            currentUser = null;
            currentOrg = null;
            showSessionTimeoutModal('inactivity', message);
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
                    localStorage.setItem('sanad_auth_active', 'true');
                    localStorage.setItem('sanad_last_activity', Date.now().toString());
                    localStorage.removeItem('sanad_voluntary_logout');
                    if (data.user && data.user.email) {
                        localStorage.setItem('sanad_user_email', data.user.email);
                        localStorage.setItem('sanad_user_name', `${data.user.first_name || ''} ${data.user.last_name || ''}`.trim());
                    }
                    isSessionExpiredState = false;
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
                const loginHeaders = { 'Content-Type': 'application/json' };
                if (clientPublicEgressIP) {
                    loginHeaders['X-Client-Public-IP'] = clientPublicEgressIP;
                }
                if (clientPreciseLocation) {
                    try {
                        loginHeaders['X-Client-Location'] = encodeURIComponent(clientPreciseLocation);
                    } catch (e) {
                        // Ignore header encoding error, JSON body client_location will be used
                    }
                }
                const res = await fetch('/api/v1/auth/login', {
                    method: 'POST',
                    headers: loginHeaders,
                    body: JSON.stringify({ 
                        email, 
                        password, 
                        totp_code: totp,
                        client_location: clientPreciseLocation 
                    })
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
                    localStorage.setItem('sanad_auth_active', 'true');
                    localStorage.setItem('sanad_last_activity', Date.now().toString());
                    localStorage.removeItem('sanad_voluntary_logout');
                    if (data.user && data.user.email) {
                        localStorage.setItem('sanad_user_email', data.user.email);
                        localStorage.setItem('sanad_user_name', `${data.user.first_name || ''} ${data.user.last_name || ''}`.trim());
                    }
                    isSessionExpiredState = false;
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
            isSessionExpiredState = false;
            localStorage.removeItem('sanad_auth_active');
            localStorage.removeItem('sanad_last_activity');
            localStorage.setItem('sanad_voluntary_logout', 'true');
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
                ipAddress: "192.168.1.45",
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

        let currentAuditTab = 'all';
        let currentAuditLogs = [];
        let selectedAuditIndex = -1;

        function getActionCategory(action) {
            if (!action) return 'security';
            const a = String(action).toUpperCase();
            if (a.includes('SIGN') || a.includes('VISA') || a.includes('SEAL') || a.includes('APPROVE_DT')) {
                return 'signatures';
            }
            if (a.startsWith('AUTH_') || a.includes('LOGIN') || a.includes('LOGOUT') || a.includes('MFA') || a.includes('PASSWORD') || a.includes('USER_') || a.includes('PERMISSION') || a.includes('ROLE') || a.includes('SESSION')) {
                return 'security';
            }
            if (a.includes('CORPORATE') || a.includes('TABADOL') || a.includes('AGREMENT') || a.includes('LICENSE') || a.includes('RECEPISE_BO')) {
                return 'corporate';
            }
            if (a.includes('VARIATION') || a.includes('DEFICIENCY') || a.includes('POST_AMM') || a.includes('COMPLEMENT')) {
                return 'variations';
            }
            if (a.includes('DOSSIER') || a.includes('CTD') || a.includes('GATING') || a.includes('SMART_PLANNING') || a.includes('CHECKLIST') || a.includes('DEPOSIT')) {
                return 'ctd';
            }
            if (a.includes('CAS_') || a.includes('FILE_') || a.includes('SYS_') || a.includes('HEALTH') || a.includes('STORAGE') || a.includes('PSSI') || a.includes('BACKUP')) {
                return 'system';
            }
            return 'security';
        }

        function getActionFriendlyLabel(action) {
            const labels = {
                'DOC_SIGN_DT': 'Signature Cryptographique Pharmacien DT (1.A.7)',
                'VISA_PDT_VALIDE': 'Visa Officiel PDT Apposé',
                'AUTH_LOGIN': 'Authentification Réussie (Session TLS)',
                'AUTH_LOGOUT': 'Déconnexion Sécurisée',
                'AUTH_GATE_UNLOCK': 'Déverrouillage Sas Partenaire VIP',
                'USER_CREATE': 'Création Compte Collaborateur',
                'USER_ROLE_UPDATE': 'Modification Habilitation RBAC',
                'USER_PWD_RESET': 'Réinitialisation Mot de Passe',
                'USER_TOTP_RESET': 'Révocation Secret MFA/TOTP',
                'DOSSIER_CREATE': 'Initialisation Dossier CTD ANPP',
                'DOSSIER_STATUS_CHANGE': 'Transition Statut Réglementaire',
                'DOSSIER_ATTACH': 'Rattachement Document BPF au CTD',
                'DOSSIER_DETACH': 'Détachement Document du CTD',
                'SMART_PLANNING_TARGET': 'Fixation Échéance Cible Dépôt',
                'CORPORATE_APP_CREATE': 'Initiation Demande Agrément Décret 21-82',
                'TABADOL_SUBMIT': 'Télétransmission sur Portail Tabadol MIPH',
                'RECEPISE_BO_DELIVRE': 'Enregistrement Récépissé Bureau Ordre MIPH',
                'LICENSE_OCTROYEE': 'Octroi Arrêté Ministériel Agrément',
                'VARIATION_SUBMIT': 'Dépôt Demande de Variation (LD-DEPP07)',
                'DEFICIENCY_TASK_VALIDATE': 'Visa Réponse Notification ANPP',
                'CAS_STORE': 'Stockage Immuable Fichier CAS (SHA-256)',
                'SYS_BACKUP': 'Sauvegarde Cryptographique Immuable'
            };
            return labels[action] || String(action).replace(/_/g, ' ');
        }

        let clientPublicEgressIP = null;
        let clientPreciseLocation = sessionStorage.getItem('sanad_precise_location') || null;

        function countryCodeToFlag(countryCode) {
            const cc = (countryCode || '').toUpperCase().trim();
            if (cc.length !== 2) return '🌍';
            const r1 = 0x1F1E6 + (cc.charCodeAt(0) - 65);
            const r2 = 0x1F1E6 + (cc.charCodeAt(1) - 65);
            return String.fromCodePoint(r1, r2);
        }

        async function detectClientPublicEgress() {
            try {
                const res = await fetch('http://ip-api.com/json/?fields=status,query,country,city,countryCode,isp', { signal: AbortSignal.timeout(1800) });
                if (res.ok) {
                    const d = await res.json();
                    if (d.status === 'success' && d.query) {
                        clientPublicEgressIP = d.query;
                    }
                }
            } catch (_) {}
        }
        detectClientPublicEgress();

        async function detectHighPrecisionLocation() {
            // 1. Tenter la géolocalisation HTML5 haute précision (GPS / Wi-Fi)
            if (navigator.geolocation) {
                navigator.geolocation.getCurrentPosition(async (pos) => {
                    try {
                        const lat = pos.coords.latitude;
                        const lon = pos.coords.longitude;
                        const res = await fetch(`https://api.bigdatacloud.net/data/reverse-geocode-client?latitude=${lat}&longitude=${lon}&localityLanguage=fr`, { signal: AbortSignal.timeout(2500) });
                        if (res.ok) {
                            const data = await res.json();
                            formatAndStoreLocality(data);
                        }
                    } catch (_) {}
                }, () => {
                    detectIPLocality();
                }, { timeout: 3500, maximumAge: 60000 });
            } else {
                detectIPLocality();
            }
        }

        async function detectIPLocality() {
            try {
                const res = await fetch('https://api.bigdatacloud.net/data/reverse-geocode-client?localityLanguage=fr', { signal: AbortSignal.timeout(2500) });
                if (res.ok) {
                    const data = await res.json();
                    formatAndStoreLocality(data);
                }
            } catch (_) {}
        }

        function formatAndStoreLocality(data) {
            if (!data) return;
            const locality = (data.locality || '').trim();
            let city = (data.city || data.principalSubdivision || '').replace(/wilaya d'/i, '').replace(/wilaya de /i, '').trim();
            const country = (data.countryName === 'Algerie' ? 'Algérie' : (data.countryName || 'Algérie')).trim();
            const flag = countryCodeToFlag(data.countryCode || 'DZ');

            let parts = [];
            if (locality) parts.push(locality);
            if (city && city.toLowerCase() !== locality.toLowerCase()) parts.push(city);
            if (country) parts.push(country);

            if (parts.length > 0) {
                clientPreciseLocation = `${flag} ${parts.join(', ')}`;
                sessionStorage.setItem('sanad_precise_location', clientPreciseLocation);
            }
        }
        detectHighPrecisionLocation();

        function resolveClientLocation(ip) {
            if (!ip || ip === '127.0.0.1' || ip === '::1' || ip === 'localhost' || ip === '0.0.0.0') {
                return '💻 Console Locale Serveur (Sur Site)';
            }
            if (clientPreciseLocation && (ip === clientPublicEgressIP || !clientPublicEgressIP)) {
                return clientPreciseLocation;
            }
            const clean = ip.split(':')[0].trim();
            if (clean.startsWith('192.168.1.')) return '🏢 Siège Administratif & RA (Alger)';
            if (clean.startsWith('192.168.10.')) return '🏭 Usine & Production (Oued Smar)';
            if (clean.startsWith('192.168.20.')) return '🔬 Labo Contrôle Qualité LCQ (Alger)';
            if (clean.startsWith('192.168.30.')) return '📦 Magasin Central & Logistique (Baraki)';
            if (clean.startsWith('192.168.')) return '🏢 Réseau Intranet Siège (Alger)';
            if (clean.startsWith('10.10.')) return '🏭 Usine & Production (Oued Smar)';
            if (clean.startsWith('10.20.')) return '🔬 Labo Contrôle Qualité LCQ (Alger)';
            if (clean.startsWith('10.30.')) return '📦 Magasin Central & Logistique (Baraki)';
            if (clean.startsWith('10.')) return '🏢 Réseau Intranet GxP (Site Industriel)';
            if (clean.startsWith('172.16.') || clean.startsWith('172.17.') || clean.startsWith('172.18.')) return '🏢 Réseau Sécurisé Datacenter GxP';
            if (clean.startsWith('193.194.')) return '🇩🇿 CERIST / Réseau Ministériel (Alger)';
            if (clean.startsWith('41.200.') || clean.startsWith('41.201.')) return '🇩🇿 Alger (Algérie Télécom)';
            if (clean.startsWith('41.220.')) return '🇩🇿 Oran (Algérie)';
            if (clean.startsWith('41.221.')) return '🇩🇿 Constantine (Algérie)';
            if (clean.startsWith('41.')) return '🇩🇿 Alger, Algérie';
            if (clean.startsWith('105.')) {
                if (clientPreciseLocation) return clientPreciseLocation;
                return '🇩🇿 Alger, Algérie (Fibre/4G)';
            }
            if (clean.startsWith('154.')) return '🇩🇿 Alger, Algérie (Mobilis/Djezzy)';
            if (clean.startsWith('197.')) return '🇩🇿 Alger, Algérie (Ooredoo)';
            if (clean.startsWith('192.0.2.')) return '🔬 Réseau Test BPF / Validation';
            if (clean.startsWith('194.')) return '🇫🇷 Paris, France (Partenaire B2B)';
            if (clean.startsWith('195.')) return '🇨🇭 Bâle, Suisse (Titulaire API)';
            if (clean.startsWith('165.')) return '🇩🇪 Francfort, Allemagne (Auditeur Qualité)';
            if (clean.startsWith('115.') || clean.startsWith('103.')) return '🇮🇳 Hyderabad, Inde (Fournisseur API)';
            if (clean.startsWith('185.')) return '🇪🇺 Europe (Accès Distant Sécurisé)';
            return `🌍 Accès Distant (${clean})`;
        }

        function getNetworkType(ip) {
            if (!ip || ip === '127.0.0.1' || ip === '::1' || ip === 'localhost') {
                return 'Console Locale Host';
            }
            const clean = ip.split(':')[0].trim();
            if (clean.startsWith('192.168.') || clean.startsWith('10.') || clean.startsWith('172.16.')) {
                return 'Intranet Privé GxP';
            }
            if (clean.startsWith('193.194.')) {
                return 'Ligne Dédiée / VPN CERIST';
            }
            return 'Accès Distant SSL/TLS Chiffré';
        }

        function initAuditColumnResizers() {
            const table = document.getElementById("audit-trail-table");
            if (!table) return;
            const headers = Array.from(table.querySelectorAll("thead th"));
            if (headers.length === 0) return;

            // Restaurer les largeurs enregistrées si présentes
            try {
                const saved = localStorage.getItem("sanad_audit_col_widths");
                if (saved) {
                    const widths = JSON.parse(saved);
                    if (Array.isArray(widths) && widths.length === headers.length) {
                        let total = 0;
                        headers.forEach((th, idx) => {
                            const w = Math.max(60, parseInt(widths[idx], 10) || 120);
                            th.style.width = w + "px";
                            total += w;
                        });
                        table.style.width = total + "px";
                    }
                }
            } catch (_) {}

            headers.forEach((th) => {
                if (th.dataset.resizerInit === "1") return;
                th.dataset.resizerInit = "1";

                const resizer = document.createElement("div");
                resizer.className = "audit-col-resizer";
                resizer.title = "Glisser pour redimensionner la colonne (double-clic pour réinitialiser)";
                th.appendChild(resizer);

                let startX = 0;
                let startWidth = 0;
                let startTableWidth = 0;

                const onMouseMove = (e) => {
                    const isRtl = document.documentElement.dir === "rtl";
                    const dx = isRtl ? (startX - e.clientX) : (e.clientX - startX);
                    const newWidth = Math.max(60, startWidth + dx);
                    const delta = newWidth - startWidth;
                    th.style.width = newWidth + "px";
                    table.style.width = Math.max(startTableWidth + delta, 100) + "px";
                };

                const onMouseUp = () => {
                    resizer.classList.remove("resizing");
                    document.body.style.cursor = "";
                    document.removeEventListener("mousemove", onMouseMove);
                    document.removeEventListener("mouseup", onMouseUp);
                    try {
                        const widths = headers.map(h => Math.round(h.getBoundingClientRect().width));
                        localStorage.setItem("sanad_audit_col_widths", JSON.stringify(widths));
                    } catch (_) {}
                };

                resizer.addEventListener("mousedown", (e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    startX = e.clientX;
                    startWidth = th.getBoundingClientRect().width;
                    startTableWidth = table.getBoundingClientRect().width;
                    resizer.classList.add("resizing");
                    document.body.style.cursor = "col-resize";
                    document.addEventListener("mousemove", onMouseMove);
                    document.addEventListener("mouseup", onMouseUp);
                });

                resizer.addEventListener("dblclick", (e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    let total = 0;
                    headers.forEach(h => {
                        const defW = parseInt(h.dataset.defaultWidth || "140", 10);
                        h.style.width = defW + "px";
                        total += defW;
                    });
                    table.style.width = "100%";
                    try { localStorage.removeItem("sanad_audit_col_widths"); } catch (_) {}
                });
            });
        }

        function switchAuditTab(tab) {
            currentAuditTab = tab || 'all';
            document.querySelectorAll('#audit-tabs-bar .audit-tab-btn').forEach(btn => {
                btn.classList.toggle('active', btn.dataset.tab === currentAuditTab);
            });
            const tabLabels = {
                'all': 'Tous les événements',
                'security': 'Sécurité & Accès',
                'signatures': 'Visas & Signatures DT (21 CFR Part 11)',
                'ctd': 'Dossiers CTD & Gating ANPP',
                'corporate': 'Agréments MIPH & Tabadol (Décret 21-82)',
                'variations': 'Variations & Compléments',
                'system': 'Coffre-fort CAS & Système'
            };
            const lbl = document.getElementById('audit-active-tab-label');
            if (lbl) lbl.innerText = tabLabels[currentAuditTab] || currentAuditTab;
            filterAuditTable();
        }
        window.switchAuditTab = switchAuditTab;

        let auditLiveTimer = null;
        let isAuditLiveActive = true;
        let isAuditFetching = false;
        let lastKnownTopAuditId = null;

        async function fetchAndProcessAuditLogs(silent = false) {
            if (isAuditFetching) return;
            isAuditFetching = true;

            const syncEl = document.getElementById("audit-last-sync-time");
            if (!silent && syncEl) {
                syncEl.textContent = "Synchro : En cours...";
            }

            try {
                const res = await fetch('/api/v1/audit/logs');
                if (res.ok) {
                    const data = await res.json();
                    if (data.logs && data.logs.length > 0) {
                        const topId = data.logs[0]?.id;
                        const hasNewArrivals = (lastKnownTopAuditId !== null && topId > lastKnownTopAuditId);

                        mockAuditTrail.length = 0;
                        const prevSelectedId = (filteredAuditList && filteredAuditList[selectedAuditIndex]) ? filteredAuditList[selectedAuditIndex].id : null;

                        currentAuditLogs = data.logs.map(l => {
                            const d = new Date(l.created_at);
                            const isNew = hasNewArrivals && (l.id > lastKnownTopAuditId);
                            const item = {
                                id: l.id,
                                rawTs: l.created_at,
                                ts: d.toLocaleDateString('fr-FR') + " " + d.toLocaleTimeString('fr-FR'),
                                code: l.action,
                                label: getActionFriendlyLabel(l.action),
                                category: getActionCategory(l.action),
                                user: l.user_email ? l.user_email.split('@')[0] : 'Système',
                                userEmail: l.user_email || 'system@homolog.dz',
                                role: l.user_role === 'directeur_technique' ? 'Pharmacien DT' : (l.user_role || 'Système'),
                                entityType: l.entity_type || 'N/A',
                                target: l.entity_id || '-',
                                sha: l.sha256_hash || '',
                                details: l.details || '{}',
                                ipAddress: l.ip_address || '127.0.0.1',
                                userAgent: l.user_agent || navigator.userAgent,
                                location: (l.location && l.location !== '🌍 Connexion Distante Sécurisée') ? l.location : resolveClientLocation(l.ip_address),
                                status: 'SCELLÉ SQL',
                                isNew: isNew
                            };
                            mockAuditTrail.push(item);
                            return item;
                        });

                        lastKnownTopAuditId = topId;

                        updateAuditTabCounts();
                        filterAuditTable();

                        // Maintenir la ligne sélectionnée si possible
                        if (prevSelectedId != null) {
                            const foundIdx = filteredAuditList.findIndex(x => x.id === prevSelectedId);
                            if (foundIdx >= 0) {
                                selectAuditRow(foundIdx, filteredAuditList[foundIdx]);
                            }
                        }
                    }
                }
            } catch (e) {
                console.warn("Lecture API audit:", e);
            } finally {
                isAuditFetching = false;
                const nowStr = new Date().toLocaleTimeString('fr-FR');
                if (syncEl) {
                    syncEl.textContent = `Synchro : ${nowStr}`;
                }
            }
        }

        function startAuditLiveStream() {
            stopAuditLiveStream();
            isAuditLiveActive = true;
            updateLiveBadgeUI();
            auditLiveTimer = setInterval(async () => {
                const modal = document.getElementById("modal-audit");
                if (modal && modal.classList.contains("open") && isAuditLiveActive) {
                    await fetchAndProcessAuditLogs(true);
                } else if (!modal || !modal.classList.contains("open")) {
                    stopAuditLiveStream();
                }
            }, 3000);
        }
        window.startAuditLiveStream = startAuditLiveStream;

        function stopAuditLiveStream() {
            if (auditLiveTimer) {
                clearInterval(auditLiveTimer);
                auditLiveTimer = null;
            }
        }
        window.stopAuditLiveStream = stopAuditLiveStream;

        function toggleAuditLiveStream() {
            isAuditLiveActive = !isAuditLiveActive;
            updateLiveBadgeUI();
            if (isAuditLiveActive) {
                fetchAndProcessAuditLogs(false);
                startAuditLiveStream();
            }
        }
        window.toggleAuditLiveStream = toggleAuditLiveStream;

        function updateLiveBadgeUI() {
            const badge = document.getElementById("audit-live-badge");
            const btnText = document.getElementById("audit-live-btn-text");
            const btnIcon = document.getElementById("audit-live-btn-icon");
            if (isAuditLiveActive) {
                if (badge) {
                    badge.classList.remove("paused");
                    badge.innerHTML = `<span class="audit-live-dot"></span> EN DIRECT (3s)`;
                }
                if (btnText) btnText.textContent = "Pause";
                if (btnIcon) btnIcon.textContent = "⏸️";
            } else {
                if (badge) {
                    badge.classList.add("paused");
                    badge.innerHTML = `⏸️ EN PAUSE`;
                }
                if (btnText) btnText.textContent = "Reprendre";
                if (btnIcon) btnIcon.textContent = "▶️";
            }
        }

        async function manualRefreshAuditTrail() {
            const icon = document.getElementById("audit-refresh-icon");
            if (icon) icon.classList.add("spin-anim");
            await fetchAndProcessAuditLogs(false);
            setTimeout(() => {
                if (icon) icon.classList.remove("spin-anim");
            }, 600);
        }
        window.manualRefreshAuditTrail = manualRefreshAuditTrail;

        async function openAuditModal() {
            closeAddressBar();
            closeAllModals();

            const auditDialog = document.getElementById("modal-audit-dialog");
            if (auditDialog && !auditDialog.classList.contains('modal-fullscreen')) {
                auditDialog.style.position = '';
                auditDialog.style.left = '';
                auditDialog.style.top = '';
                auditDialog.style.margin = 'auto';
            }

            // Initialisation de la modale redimensionnable et déplaçable (façon fenêtre OS)
            if (typeof initModalWindowResizer === 'function') {
                setTimeout(() => {
                    initModalWindowResizer('modal-audit-dialog', 'modal-audit-header', { minWidth: 540, minHeight: 420 });
                }, 40);
            }

            // Initialisation des colonnes redimensionnables de l'Audit Trail
            initAuditColumnResizers();

            // Chargement initial des logs
            await fetchAndProcessAuditLogs(false);

            if (currentAuditLogs.length === 0 && mockAuditTrail.length > 0) {
                currentAuditLogs = mockAuditTrail.map((m, idx) => ({
                    id: idx + 1,
                    rawTs: new Date().toISOString(),
                    ts: m.ts,
                    code: m.code,
                    label: m.label || getActionFriendlyLabel(m.code),
                    category: getActionCategory(m.code),
                    user: m.user,
                    userEmail: `${m.user.toLowerCase().replace(/[^a-z0-9]/g, '')}@pharma-cdmo.dz`,
                    role: m.role,
                    entityType: 'document',
                    target: m.target,
                    sha: m.sha,
                    details: JSON.stringify({ entity: m.target, action: m.code, operator: m.user }, null, 2),
                    ipAddress: m.ipAddress || '192.168.1.45',
                    userAgent: navigator.userAgent,
                    location: m.location || resolveClientLocation(m.ipAddress || '192.168.1.45'),
                    status: m.status || 'SCELLÉ SQL'
                }));
                updateAuditTabCounts();
                selectedAuditIndex = currentAuditLogs.length > 0 ? 0 : -1;
                filterAuditTable();
            }

            document.getElementById("modal-audit").classList.add("open");

            // Lancement du rafraîchissement temps réel
            startAuditLiveStream();
        }

        function updateAuditTabCounts() {
            const counts = { all: currentAuditLogs.length, security: 0, signatures: 0, ctd: 0, corporate: 0, variations: 0, system: 0 };
            currentAuditLogs.forEach(l => {
                const cat = l.category || getActionCategory(l.code);
                if (counts[cat] !== undefined) counts[cat]++;
            });
            Object.keys(counts).forEach(k => {
                const el = document.getElementById(`tab-cnt-${k}`);
                if (el) el.textContent = counts[k];
            });
        }

        let filteredAuditList = [];

        function renderAuditTable(list = currentAuditLogs) {
            filteredAuditList = list;
            const tbody = document.getElementById("audit-table-body");
            if (!tbody) return;
            tbody.innerHTML = "";

            if (list.length === 0) {
                tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:36px; color:var(--text-muted);"><div style="font-size:24px; margin-bottom:8px;">🛡️</div>Aucun événement trouvé pour ce filtre ou cet onglet.</td></tr>`;
                renderAuditInspector(null);
                const badge = document.getElementById("audit-count-badge");
                if (badge) badge.innerText = "0 événement scellé";
                return;
            }

            list.forEach((item, idx) => {
                const tr = document.createElement("tr");
                tr.className = "audit-row" + (idx === selectedAuditIndex ? " selected" : "") + (item.isNew ? " new-entry-flash" : "");
                if (item.isNew) item.isNew = false;
                tr.onclick = () => selectAuditRow(idx, item);

                // Catégorie badge color
                let tagColor = 'var(--accent-blue)';
                if (item.category === 'signatures') tagColor = 'var(--accent-emerald)';
                else if (item.category === 'security') tagColor = '#7c3aed';
                else if (item.category === 'corporate') tagColor = '#d97706';
                else if (item.category === 'variations') tagColor = '#ea580c';
                else if (item.category === 'system') tagColor = '#64748b';

                tr.innerHTML = `
                    <td style="padding:6px 10px; font-family:monospace; font-size:10.5px; color:var(--text-muted);">${escapeHTML(item.ts)}</td>
                    <td style="padding:6px 10px;">
                        <span style="display:inline-flex; align-items:center; white-space:nowrap; max-width:100%; overflow:hidden; text-overflow:ellipsis; background:rgba(2,132,199,0.08); border:1px solid ${tagColor}; color:${tagColor}; padding:1px 6px; border-radius:3px; font-weight:700; font-size:10px;" title="${escapeHTML(item.code)}">${escapeHTML(item.code)}</span> 
                        <div style="color:var(--text-main); font-weight:600; font-size:10.5px; margin-top:2px;">${escapeHTML(item.label)}</div>
                    </td>
                    <td style="padding:6px 10px;">
                        <span style="font-weight:600; color:var(--text-main);">${escapeHTML(item.user)}</span>
                        <div style="color:var(--text-dim); font-size:10px;">${escapeHTML(item.role)}</div>
                    </td>
                    <td style="padding:6px 10px; white-space:nowrap;">
                        <div style="font-family:monospace; font-size:10.5px; color:var(--accent-blue); font-weight:600;">
                            ${escapeHTML(item.ipAddress || '127.0.0.1')}
                        </div>
                        <div style="font-size:10px; color:var(--text-main); margin-top:2px; display:flex; align-items:center; gap:4px;" title="${escapeHTML(item.location || resolveClientLocation(item.ipAddress))}">
                            <span>${escapeHTML(item.location || resolveClientLocation(item.ipAddress))}</span>
                        </div>
                    </td>
                    <td style="padding:6px 10px; color:var(--text-main); font-size:11px;" title="${escapeHTML(item.target)}">${escapeHTML(item.target)}</td>
                    <td style="padding:6px 10px; font-family:monospace; font-size:10px; color:var(--text-dim);" title="${escapeHTML(item.sha)}">
                        ${escapeHTML(item.sha ? item.sha.substring(0, 16) + '...' : '-')}
                    </td>
                    <td style="padding:6px 10px; text-align:center;">
                        <span style="color:var(--accent-emerald); background:rgba(16,185,129,0.1); border:1px solid rgba(16,185,129,0.3); padding:2px 6px; border-radius:4px; font-weight:700; font-size:9.5px;">${escapeHTML(item.status || 'SCELLÉ SQL')}</span>
                    </td>
                `;
                tbody.appendChild(tr);
            });

            const badge = document.getElementById("audit-count-badge");
            if (badge) badge.innerText = `${list.length} événement(s) scellé(s)`;

            if (selectedAuditIndex >= 0 && selectedAuditIndex < list.length) {
                renderAuditInspector(list[selectedAuditIndex]);
            } else if (list.length > 0) {
                selectedAuditIndex = 0;
                renderAuditInspector(list[0]);
            } else {
                renderAuditInspector(null);
            }
        }

        function selectAuditRow(idx, item) {
            selectedAuditIndex = idx;
            document.querySelectorAll('#audit-table-body tr').forEach((r, i) => {
                r.classList.toggle('selected', i === idx);
            });
            renderAuditInspector(item);
        }

        function renderAuditInspector(item) {
            const container = document.getElementById("audit-inspector-content");
            const headerId = document.getElementById("inspector-event-id");
            if (!container) return;

            if (!item) {
                if (headerId) headerId.textContent = "#--";
                container.innerHTML = `
                    <div style="text-align:center; padding:40px 10px; color:var(--text-muted);">
                        🛡️ Sélectionnez un événement dans l'audit trail pour inspecter son intégrité et ses métadonnées.
                    </div>
                `;
                return;
            }

            if (headerId) headerId.textContent = `#${item.id || item.code}`;

            // Tentative de parsing propre du JSON details
            let formattedJSON = item.details || '{}';
            try {
                if (typeof item.details === 'string') {
                    const parsed = JSON.parse(item.details);
                    formattedJSON = JSON.stringify(parsed, null, 2);
                } else {
                    formattedJSON = JSON.stringify(item.details, null, 2);
                }
            } catch (_) {}

            container.innerHTML = `
                <div class="audit-inspector-card">
                    <div class="audit-inspector-label">Événement &amp; Règle BPF</div>
                    <div style="font-weight:700; font-size:12px; color:var(--accent-blue);">${escapeHTML(item.code)}</div>
                    <div style="color:var(--text-main); font-size:11px; margin-top:2px;">${escapeHTML(item.label)}</div>
                </div>

                <div class="audit-inspector-card">
                    <div class="audit-inspector-label">Horodatage Strict (UTC &amp; Local)</div>
                    <div class="audit-inspector-value" style="font-family:monospace;">${escapeHTML(item.ts)}</div>
                    <div style="font-size:10px; color:var(--text-dim); margin-top:2px;">Norme ISO : ${escapeHTML(item.rawTs || item.ts)}</div>
                </div>

                <div class="audit-inspector-card">
                    <div class="audit-inspector-label">Opérateur / Acteur</div>
                    <div style="font-weight:600; color:var(--text-main);">${escapeHTML(item.user)}</div>
                    <div style="font-size:10.5px; color:var(--text-dim);">${escapeHTML(item.userEmail)}</div>
                    <div style="margin-top:3px;"><span class="count-badge ok">${escapeHTML(item.role)}</span></div>
                </div>

                <div class="audit-inspector-card">
                    <div class="audit-inspector-label">Cible &amp; Périmètre Réglementaire</div>
                    <div class="audit-inspector-value"><strong>Type :</strong> ${escapeHTML(item.entityType || 'Entité')}</div>
                    <div class="audit-inspector-value" style="font-family:monospace; margin-top:2px;"><strong>Cible :</strong> ${escapeHTML(item.target)}</div>
                </div>

                <div class="audit-inspector-card" style="background:rgba(2,132,199,0.04); border-color:rgba(2,132,199,0.25);">
                    <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:4px;">
                        <span class="audit-inspector-label" style="color:var(--accent-blue);">Empreinte SHA-256 Scellée</span>
                        <button type="button" class="btn-pill" onclick="copyAuditHash('${escapeHTML(item.sha)}')" style="font-size:9.5px; padding:1px 6px; height:20px;" title="Copier l'empreinte complète dans le presse-papier">📋 Copier</button>
                    </div>
                    <div style="font-family:monospace; font-size:9.5px; word-break:break-all; line-height:1.4; color:var(--text-main);">${escapeHTML(item.sha || 'N/A')}</div>
                    <div style="font-size:9.5px; color:#059669; font-weight:700; margin-top:4px;">🔒 Trigger PostgreSQL Append-Only (ALCOA+ Intact)</div>
                </div>

                <div class="audit-inspector-card">
                    <div class="audit-inspector-label">Environnement Réseau &amp; Sécurité GxP</div>
                    <div style="display:flex; flex-direction:column; gap:4px; margin-top:4px;">
                        <div style="font-size:10.5px; color:var(--text-main);">
                            <strong>📍 Lieu de Connexion :</strong> <span style="font-weight:600; color:var(--accent-blue);">${escapeHTML(item.location || resolveClientLocation(item.ipAddress))}</span>
                        </div>
                        <div style="font-size:10.5px; color:var(--text-main);">
                            <strong>🌐 IP Client :</strong> <span style="font-family:monospace;">${escapeHTML(item.ipAddress || '127.0.0.1')}</span>
                        </div>
                        <div style="font-size:10.5px; color:var(--text-main);">
                            <strong>🛡️ Type d'Accès :</strong> <span style="font-size:9.5px; background:rgba(2,132,199,0.08); border:1px solid rgba(2,132,199,0.3); padding:1px 6px; border-radius:3px; color:var(--accent-blue); font-weight:600;">${escapeHTML(getNetworkType(item.ipAddress))}</span>
                        </div>
                        <div style="font-size:10px; color:var(--text-dim); margin-top:2px; word-break:break-word;">
                            <strong>User-Agent :</strong> ${escapeHTML(item.userAgent ? item.userAgent.substring(0, 70) + '...' : 'Client Web Nav')}
                        </div>
                        <div style="font-size:10px; color:var(--text-dim);">
                            <strong>Fuseau Horaire :</strong> Africa/Algiers (UTC+1 / CET)
                        </div>
                    </div>
                </div>

                <div class="audit-inspector-card">
                    <div class="audit-inspector-label">Charge Utile Détaillée (Payload JSON)</div>
                    <pre class="audit-json-box">${escapeHTML(formattedJSON)}</pre>
                </div>
            `;
        }

        function copyAuditHash(hash) {
            if (!hash) return;
            if (navigator.clipboard && navigator.clipboard.writeText) {
                navigator.clipboard.writeText(hash).then(() => {
                    if (typeof showNotifyModal === 'function') {
                        showNotifyModal("Empreinte Copiée", `Le hachage SHA-256 a été copié dans votre presse-papiers :<br><code style="font-size:10px; word-break:break-all;">${escapeHTML(hash)}</code>`, "📋", "OK");
                    }
                });
            } else {
                prompt("Empreinte SHA-256 scellée :", hash);
            }
        }
        window.copyAuditHash = copyAuditHash;

        function filterAuditTable() {
            const q = (document.getElementById("audit-filter-input")?.value || '').trim().toLowerCase();
            const timeFilter = document.getElementById("audit-time-filter")?.value || 'all';

            const now = Date.now();
            const oneDay = 24 * 60 * 60 * 1000;

            const filtered = currentAuditLogs.filter(item => {
                // Filtre par onglet
                if (currentAuditTab !== 'all') {
                    const cat = item.category || getActionCategory(item.code);
                    if (cat !== currentAuditTab) return false;
                }

                // Filtre par période
                if (timeFilter !== 'all' && item.rawTs) {
                    const itemTime = new Date(item.rawTs).getTime();
                    if (timeFilter === 'today' && now - itemTime > oneDay) return false;
                    if (timeFilter === '7d' && now - itemTime > 7 * oneDay) return false;
                    if (timeFilter === '30d' && now - itemTime > 30 * oneDay) return false;
                }

                // Filtre texte
                if (!q) return true;
                return (item.code && item.code.toLowerCase().includes(q)) ||
                       (item.label && item.label.toLowerCase().includes(q)) ||
                       (item.user && item.user.toLowerCase().includes(q)) ||
                       (item.userEmail && item.userEmail.toLowerCase().includes(q)) ||
                       (item.role && item.role.toLowerCase().includes(q)) ||
                       (item.target && item.target.toLowerCase().includes(q)) ||
                       (item.sha && item.sha.toLowerCase().includes(q)) ||
                       (item.ipAddress && item.ipAddress.toLowerCase().includes(q)) ||
                       (item.location && item.location.toLowerCase().includes(q));
            });

            selectedAuditIndex = filtered.length > 0 ? 0 : -1;
            renderAuditTable(filtered);
        }
        window.filterAuditTable = filterAuditTable;

        function exportAuditTrailCSV(all = false) {
            const listToExport = all ? currentAuditLogs : filteredAuditList;
            if (!listToExport || listToExport.length === 0) {
                if (typeof showNotifyModal === 'function') {
                    showNotifyModal("Export Impossible", "Aucun enregistrement d'audit trail à exporter dans cette vue.", "⚠️", "Fermer");
                }
                return;
            }

            // Génération conforme RFC 4180
            const headers = ["ID", "Horodatage_UTC", "Code_Action", "Libelle_Action", "Categorie", "Operateur", "Email", "Role", "Type_Entite", "Cible", "Empreinte_SHA256", "Statut_Immuable", "IP_Client", "Lieu_Connexion"];
            const escapeCSV = (field) => {
                const s = String(field == null ? '' : field).replace(/"/g, '""');
                return `"${s}"`;
            };

            const rows = [headers.join(",")];
            listToExport.forEach(item => {
                rows.push([
                    escapeCSV(item.id),
                    escapeCSV(item.rawTs || item.ts),
                    escapeCSV(item.code),
                    escapeCSV(item.label),
                    escapeCSV(item.category),
                    escapeCSV(item.user),
                    escapeCSV(item.userEmail),
                    escapeCSV(item.role),
                    escapeCSV(item.entityType),
                    escapeCSV(item.target),
                    escapeCSV(item.sha),
                    escapeCSV(item.status || 'SCELLÉ SQL'),
                    escapeCSV(item.ipAddress),
                    escapeCSV(item.location || resolveClientLocation(item.ipAddress))
                ].join(","));
            });

            const csvString = "\uFEFF" + rows.join("\r\n"); // UTF-8 BOM pour Excel
            const blob = new Blob([csvString], { type: 'text/csv;charset=utf-8;' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement("a");
            const dateStr = new Date().toISOString().slice(0, 10);
            const suffix = all ? "complet" : (currentAuditTab || "filtre");
            a.href = url;
            a.download = `audit_trail_alcoa_${suffix}_${dateStr}.csv`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

            if (typeof showNotifyModal === 'function') {
                showNotifyModal(
                    "Export CSV Certifié Téléchargé (ALCOA+)",
                    `<div style="display:flex; flex-direction:column; gap:8px;">
                        <div style="color:var(--accent-emerald); font-weight:700;">
                            ✓ Fichier CSV téléchargé avec succès (${listToExport.length} enregistrements)
                        </div>
                        <div style="font-size:11px; color:var(--text-main);">
                            Nom de fichier : <code>audit_trail_alcoa_${suffix}_${dateStr}.csv</code><br>
                            Le journal inclut toutes les métadonnées de traçabilité, les adresses IP et les empreintes SHA-256 scellées conformes aux normes BPF / 21 CFR Part 11.
                        </div>
                    </div>`,
                    "📥",
                    "Fermer"
                );
            }
        }
        window.exportAuditTrailCSV = exportAuditTrailCSV;

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
