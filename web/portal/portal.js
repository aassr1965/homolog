(() => {
    'use strict';

    const ROUTES = ['accueil', 'plateforme', 'conformite', 'offres', 'demo', 'contact'];

    const I18N = {
        fr: {
            'nav.home': 'Accueil',
            'nav.platform': 'Plateforme',
            'nav.compliance': 'Conformité',
            'nav.offers': 'Offres',
            'nav.demo': 'Démo / Onboarding',
            'nav.contact': 'Contact',
            'nav.login': 'Connexion',
            'nav.theme': 'Mode jour / nuit',
            'home.badge': 'ANPP • MIPH • 21 CFR Part 11 • ALCOA+',
            'home.title': "Huit modules interconnectés, un seul socle d'intégrité réglementaire.",
            'home.subtitle': "Pilotez l'enregistrement CTD des médicaments, l'homologation des dispositifs médicaux, les agréments MIPH Tabadol et votre GED BPF sur une plateforme souveraine, bilingue FR/AR et déployable sur site ou en Cloud.",
            'home.ctaLogin': 'Accéder à SANAD Pharma',
            'home.ctaPlatform': 'Découvrir la plateforme',
            'home.kpi1': 'Procédures CTD (NI/RI/MI/NF/RF/MF)',
            'home.kpi2': 'Rubriques Variations LD-DEPP07',
            'home.kpi3': 'États de sortie BPF scellés QR',
            'home.rolesTitle': 'Conçu pour chaque responsabilité de votre établissement',
            'home.rolesSub': 'Séparation stricte des fonctions (SoD Maker-Checker-Approver) et accès cloisonnés Zero-Trust.',
            'home.role1Title': 'Pharmacien Directeur Technique',
            'home.role1Desc': 'Parapheur électronique centralisé, visa légal 2FA TOTP, libération de lot DCE et contrôle de parité FR/AR.',
            'home.role2Title': 'Affaires Réglementaires',
            'home.role2Desc': 'Montage CTD Modules 1–5, Smart Planning en jours ouvrables algériens, carences ANPP et export E-TASJIL.',
            'home.role3Title': 'Assurance & Contrôle Qualité',
            'home.role3Desc': 'Revue indépendante Checker, gestion documentaire BPF en 5 branches, déviations, CAPA et audits fournisseurs.',
            'home.role4Title': 'Direction Générale & DSI',
            'home.role4Desc': 'Déploiement Air-Gapped ou SaaS multi-unités, réversibilité totale par archive .sanad et audit trail immuable.',
            'plat.badge': 'Architecture Unifiée',
            'plat.title': 'Les principes partagés',
            'plat.intro': "Chaque module peut être activé progressivement. Tous reposent sur le même moteur de conformité et d'intégrité des données.",
            'plat.p1Title': 'Tout est tracé',
            'plat.p1Desc': "Chaque lecture, téléchargement, refus ou visa laisse une trace horodatée et géolocalisée dans l'Audit Trail.",
            'plat.p2Title': "Rien ne s'efface",
            'plat.p2Desc': "Table d'audit append-only inviolable, versionnage mineur/majeur avec motif de révision et archivage scellé.",
            'plat.p3Title': 'Chaque signature engage',
            'plat.p3Desc': 'Re-authentification 2FA TOTP et déclaration solennelle du Pharmacien DT conformes au 21 CFR Part 11.',
            'plat.p4Title': 'Une interface commune',
            'plat.p4Desc': 'Ergonomie « Pharma Explorer » à 3 volets, identique de jour comme de nuit, en français, arabe RTL et anglais.',
            'plat.ringTitle': 'Ce qui circule entre les modules',
            'plat.ringDesc': "Aucune ressaisie entre l'agrément de votre établissement au MIPH, le montage de vos dossiers CTD/DM, la gestion des variations et la libération de lots.",
            'plat.depTitle': 'Sur vos serveurs ou en Cloud souverain',
            'plat.depSub': 'Binaire unique Go sans dépendance CDN externe, conçu pour fonctionner aussi bien en intranet isolé (Air-Gapped) qu’en architecture SaaS multi-unités.',
            'plat.dep1Title': 'On-Premise (Air-Gapped)',
            'plat.dep1Desc': 'Installation 100 % hors-ligne sur votre infrastructure laboratoire. Clés et secrets maîtrisés exclusivement par le Pharmacien DT et votre RSSI.',
            'plat.dep2Title': 'Cloud SaaS Souverain',
            'plat.dep2Desc': 'Cloisonnement strict par établissement, gouvernance Groupe multi-sites (DT Groupe) et réversibilité absolue via l’archive portative .sanad.',
            'plat.suiteTitle': 'Chaque module, en détail',
            'plat.suiteSub': 'Une suite modulaire complète couvrant tout le cycle réglementaire et qualité.',
            'mod.m1t': 'Sanad Registry',
            'mod.m1d': 'Référentiel des établissements MIPH, sites 1.H.8 (SA, PF, contrôle) et catalogue Médicaments & DM.',
            'mod.m2t': 'Sanad Corporate',
            'mod.m2d': 'Agréments des 5 catégories d’établissements (Décret 21-82), Smart e-Forms et export Tabadol.',
            'mod.m3t': 'Sanad CTD Studio',
            'mod.m3d': 'Checklists dynamiques NI/RI/MI/NF/RF/MF, gating DAG 5 phases, parité FR/AR et Smart Planning.',
            'mod.m4t': 'Sanad Variations',
            'mod.m4d': '34 rubriques LD-DEPP07/V01, reclassement automatique, Grouping Art. 09 et Change Impact Tree.',
            'mod.m5t': 'Sanad Vault',
            'mod.m5d': 'Coffre-fort GED BPF en 5 branches, baux d’édition 30 min, libération de lot DCE et CAS SHA-256.',
            'mod.m6t': 'Sanad Gateway',
            'mod.m6d': 'Bordereaux BV1/BV2, redevances 1.A.4/1.A.5, fiche E-TASJIL et veille créneaux RDV PP/DM.',
            'mod.m7t': 'Sanad eQMS',
            'mod.m7d': 'Option qualité BPF : déviations, CAPA, Change Control connecté aux variations, audits et formations SOP.',
            'mod.m8t': 'Sanad Copilot',
            'mod.m8d': 'Assistant IA souverain Human-in-the-Loop pour rédaction scientifique v0.1 et audit de cohérence.',
            'conf.badge': 'Référentiels & Textes Opposables',
            'conf.title': 'La conformité, référentiel par référentiel',
            'conf.sub': 'Chaque contrôle bloquant et chaque état de sortie de SANAD Pharma est rattaché à un texte réglementaire algérien ou standard international GxP.',
            'conf.c1t': 'Enregistrement ANPP (CTD 2025)',
            'conf.c1d': 'Décret exécutif n° 20-325, Guide CTD Juillet 2025 (Modules 1 à 5), distinction Princeps / Générique importé / Générique local et engagements 1.P.8 / 1.P.9.',
            'conf.c2t': 'Modifications LD-DEPP07/V01',
            'conf.c2d': 'Nomenclature algérienne Mmin, Mmod, Mmaj sur 34 rubriques, reclassement dynamique et Grouping selon l’arrêté du 7 novembre 2021.',
            'conf.c3t': 'Dispositifs Médicaux (DDM)',
            'conf.c3d': 'Décret exécutif n° 20-324 et arrêté du 10 mai 2021 : formulaire 6 sections, grille de recevabilité en 13 points et redevances BNA.',
            'conf.c4t': 'Établissements MIPH & Tabadol',
            'conf.c4d': 'Décret exécutif n° 21-82 et arrêtés 2025 : 5 catégories, caducité 2 ans, intérim PDT 92 jours et paquets Tabadol-J / Intaj.',
            'conf.c5t': '21 CFR Part 11 & Annexe 11 BPF',
            'conf.c5d': 'Argon2id, TOTP RFC 6238, signature électronique légale PDT, séparation stricte SoD et Audit Trail append-only.',
            'conf.c6t': 'GAMP 5 & Qualification OQ/PQ',
            'conf.c6d': 'Protocole de qualification opérationnelle et de performance maintenu à jour à chaque évolution pour vos audits et inspections.',
            'off.badge': 'Modèle Modulaire',
            'off.title': 'Des offres adaptées à la taille de votre laboratoire',
            'off.sub': 'Démarrez sur le socle réglementaire ANPP/MIPH et activez les extensions Qualité ou Multi-Sites selon votre feuille de route.',
            'off.o1t': 'Sanad Core RIM',
            'off.o1tag': 'Enregistrement & GED BPF',
            'off.o1d': 'CTD Studio, Variations LD-DEPP07, Homologation DM, Sanad Vault GED, Parapheur PDT, Gateway E-TASJIL/RDV et 28 États de sortie.',
            'off.o2t': 'Sanad Suite + eQMS',
            'off.o2tag': 'RIM + Assurance Qualité BPF',
            'off.o2d': 'Tout Sanad Core RIM + Déviations, CAPA, Change Control connecté au moteur de variations, Audits fournisseurs, Réclamations et Formations SOP.',
            'off.o3t': 'Sanad Groupe Multi-Unités',
            'off.o3tag': 'Holdings & Multi-Sites',
            'off.o3d': 'Supervision transverse DT Groupe sur plusieurs unités de production cloisonnées, gouvernance centralisée et archive de réversibilité .sanad.',
            'off.cta': 'Demander une étude',
            'demo.badge': 'Démonstration & Onboarding',
            'demo.title': 'Planifier une démonstration ou préparer votre onboarding',
            'demo.sub': "Nos ingénieurs réglementaires vous présentent l'environnement complet (CTD, Variations, Corporate Tabadol, Parapheur DT) et préparent l'ouverture de votre espace.",
            'demo.alreadyTitle': 'Déjà client ou partenaire autorisé ?',
            'demo.alreadyDesc': "Accédez directement à l'écran d'authentification sécurisé de l'application.",
            'demo.submit': 'Envoyer la demande',
            'contact.badge': 'Contact Direct',
            'contact.title': "Échanger avec l'équipe SANAD Pharma",
            'contact.sub': "Une question sur la qualification OQ/PQ, l'hébergement souverain ou le déploiement multi-sites ? Écrivez-nous directement.",
            'contact.supportLabel': 'Accompagnement :',
            'contact.submit': 'Envoyer le message',
            'form.name': 'Nom et prénom *',
            'form.email': 'Email professionnel *',
            'form.org': 'Laboratoire / Établissement',
            'form.role': 'Fonction',
            'form.phone': 'Téléphone',
            'form.mode': 'Mode souhaité',
            'form.msg': 'Vos objectifs (CTD, DM, MIPH, eQMS)',
            'form.message': 'Message *',
            'footer.desc': "Plateforme souveraine d'enregistrement ANPP, d'homologation DM, d'agréments MIPH et de GED pharmaceutique BPF.",
            'footer.col1': 'Plateforme',
            'footer.col2': 'Accès & Contact',
            'footer.disclaimer': "Un logiciel ne rend pas, à lui seul, une organisation conforme : la validation du système (OQ/PQ), vos procédures opératoires et votre infrastructure y concourent. SANAD Pharma fournit les mécanismes techniques et la documentation de qualification pour l'étayer.",
            'toast.sent': 'Votre demande a bien été enregistrée. Notre équipe vous contactera rapidement.',
            'toast.invalid': 'Veuillez renseigner un nom et une adresse email valide.'
        },
        ar: {
            'nav.home': 'الرئيسية',
            'nav.platform': 'المنصة',
            'nav.compliance': 'المطابقة',
            'nav.offers': 'العروض',
            'nav.demo': 'عرض تجريبي / تسجيل',
            'nav.contact': 'اتصل بنا',
            'nav.login': 'تسجيل الدخول',
            'nav.theme': 'الوضع النهاري / الليلي',
            'home.badge': 'ANPP • MIPH • 21 CFR Part 11 • ALCOA+',
            'home.title': 'ثماني وحدات مترابطة على قاعدة موحدة لنزاهة البيانات التنظيمية.',
            'home.subtitle': 'أدر ملفات تسجيل الأدوية CTD، واعتماد المستلزمات الطبية، وتراخيص وزارة الصناعة الصيدلانية Tabadol، والأرشفة الصيدلانية BPF على منصة سيادية ثنائية اللغة قابلة للتثبيت المحلي أو السحابي.',
            'home.ctaLogin': 'الدخول إلى SANAD Pharma',
            'home.ctaPlatform': 'استكشف المنصة',
            'home.kpi1': 'إجراءات CTD (NI/RI/MI/NF/RF/MF)',
            'home.kpi2': 'بنود التعديلات LD-DEPP07',
            'home.kpi3': 'تقارير رسمية مختومة برمز QR',
            'home.rolesTitle': 'مصممة لكل مسؤولية داخل مؤسستكم الصيدلانية',
            'home.rolesSub': 'فصل صارم للمهام (Maker-Checker-Approver) وعزل تام للصلاحيات.',
            'home.role1Title': 'الصيدلي المدير التقني',
            'home.role1Desc': 'توقيع إلكتروني مركزي بالتحقق الثنائي 2FA، تحرير الدفعات DCE ومطابقة النشرات فرنسي/عربي.',
            'home.role2Title': 'الشؤون التنظيمية',
            'home.role2Desc': 'إعداد ملفات CTD 1-5، التخطيط الذكي بأيام العمل الجزائرية، متابعة التحفظات وتصدير E-TASJIL.',
            'home.role3Title': 'تأكيد ومراقبة الجودة',
            'home.role3Desc': 'مراجعة مستقلة، نظام إدارة وثائق BPF في 5 فروع، الانحرافات، CAPA والتدقيق.',
            'home.role4Title': 'الإدارة العامة وتقنية المعلومات',
            'home.role4Desc': 'نشر محلي معزول Air-Gapped أو سحابي سيادي متعدد الوحدات مع أرشيف استرجاع .sanad.',
            'plat.badge': 'بنية موحدة',
            'plat.title': 'المبادئ المشتركة',
            'plat.intro': 'يمكن تفعيل كل وحدة تدريجياً، وتعمل جميعها على نفس محرك المطابقة وسلامة البيانات.',
            'plat.p1Title': 'كل إجراء مسجل',
            'plat.p1Desc': 'كل قراءة أو تحميل أو رفض أو تأشيرة يترك أثراً مؤرخاً وموثقاً في سجل التدقيق.',
            'plat.p2Title': 'لا شيء يُمحى',
            'plat.p2Desc': 'سجل تدقيق غير قابل للتعديل أو الحذف مع إدارة الإصدارات وسبب التغيير.',
            'plat.p3Title': 'كل توقيع مُلزم قانونياً',
            'plat.p3Desc': 'إعادة المصادقة الثنائية TOTP والتعهد الرسمي للصيدلي المدير التقني وفق 21 CFR Part 11.',
            'plat.p4Title': 'واجهة موحدة',
            'plat.p4Desc': 'تجربة مستخدم «Pharma Explorer» بثلاثة أجزاء، متاحة ليلاً ونهاراً بالعربية والفرنسية والإنجليزية.',
            'plat.ringTitle': 'ما يتدفق بين الوحدات',
            'plat.ringDesc': 'ترابط تلقائي بين تراخيص المؤسسة في الوزارة، ملفات CTD/DM، التعديلات وتحرير الدفعات دون إعادة إدخال.',
            'plat.depTitle': 'على خوادمكم أو في سحابة سيادية',
            'plat.depSub': 'ملف تنفيذي موحد بدون أي اعتماد خارجي، يعمل في الشبكات المعزولة أو السحابة متعددة المواقع.',
            'plat.dep1Title': 'تثبيت محلي (Air-Gapped)',
            'plat.dep1Desc': 'تشغيل كامل بدون إنترنت داخل شبكة المخبر مع تحكم حصري في مفاتيح التشفير.',
            'plat.dep2Title': 'سحابة سيادية SaaS',
            'plat.dep2Desc': 'عزل تام لكل وحدة إنتاجية، إشراف مركزي للمجمعات الصيدلانية وإمكانية الاسترجاع الكامل عبر .sanad.',
            'plat.suiteTitle': 'كل وحدة بالتفصيل',
            'plat.suiteSub': 'منظومة متكاملة تغطي كامل الدورة التنظيمية والجودة.',
            'mod.m1t': 'Sanad Registry',
            'mod.m1d': 'مرجع المؤسسات الصيدلانية MIPH، مواقع التصنيع 1.H.8 وفهرس الأدوية والمستلزمات الطبية.',
            'mod.m2t': 'Sanad Corporate',
            'mod.m2d': 'تراخيص الفئات الخمس للمؤسسات (المرسوم 21-82)، النماذج الذكية وتصدير Tabadol.',
            'mod.m3t': 'Sanad CTD Studio',
            'mod.m3d': 'قوائم التحقق الديناميكية NI/RI/MI/NF/RF/MF، ضوابط المراحل الخمس، مطابقة فرنسي/عربي والتخطيط الذكي.',
            'mod.m4t': 'Sanad Variations',
            'mod.m4d': '34 بنداً وفق LD-DEPP07/V01، إعادة التصنيف التلقائي، التجميع Art. 09 وشجرة الأثر.',
            'mod.m5t': 'Sanad Vault',
            'mod.m5d': 'خزينة الأرشفة الصيدلانية BPF في 5 فروع، أقفال التحرير، تحرير الدفعات DCE وبصمة SHA-256.',
            'mod.m6t': 'Sanad Gateway',
            'mod.m6d': 'جداول الدفع BV1/BV2، الرسوم 1.A.4/1.A.5، بطاقة E-TASJIL ومتابعة مواعيد RDV.',
            'mod.m7t': 'Sanad eQMS',
            'mod.m7d': 'خيار الجودة BPF: الانحرافات، CAPA، إدارة التغيير المرتبطة بالتعديلات، التدقيق والتكوين.',
            'mod.m8t': 'Sanad Copilot',
            'mod.m8d': 'مساعد ذكاء اصطناعي سيادي بإشراف بشري لصياغة المسودات العلمية وتدقيق التناسق.',
            'conf.badge': 'المرجعيات والنصوص القانونية',
            'conf.title': 'المطابقة وفق كل مرجع تنظيمي',
            'conf.sub': 'كل قاعدة تحقق ومخرج رسمي في SANAD Pharma مرتبط بنص تنظيمي جزائري أو معيار دولي GxP.',
            'conf.c1t': 'تسجيل الأدوية ANPP (CTD 2025)',
            'conf.c1d': 'المرسوم التنفيذي 20-325، دليل CTD يوليو 2025 (الوحدات 1 إلى 5) والتزامات التصنيع المحلي 1.P.8 / 1.P.9.',
            'conf.c2t': 'التعديلات LD-DEPP07/V01',
            'conf.c2d': 'التصنيف الجزائري Mmin و Mmod و Mmaj في 34 بنداً مع إعادة التصنيف التلقائي والتجميع.',
            'conf.c3t': 'المستلزمات الطبية (DDM)',
            'conf.c3d': 'المرسوم التنفيذي 20-324 وقرار 10 مايو 2021: استمارة 6 أقسام، شبكة القبول من 13 نقطة ورسوم BNA.',
            'conf.c4t': 'المؤسسات MIPH و Tabadol',
            'conf.c4d': 'المرسوم التنفيذي 21-82 وقرارات 2025: 5 فئات، مهلة سنتين، نيابة 92 يوماً وحزم Tabadol.',
            'conf.c5t': '21 CFR Part 11 و Annexe 11 BPF',
            'conf.c5d': 'تشفير Argon2id، تحقق ثنائي TOTP، توقيع إلكتروني قانوني، فصل الصلاحيات وسجل تدقيق غير قابل للتعديل.',
            'conf.c6t': 'GAMP 5 وتأهيل OQ/PQ',
            'conf.c6d': 'بروتوكول تأهيل تشغيلي وأدائي محيّن باستمرار لدعم عمليات التفتيش والتدقيق.',
            'off.badge': 'باقات مرنة',
            'off.title': 'عروض تتكيف مع حجم مؤسستكم',
            'off.sub': 'ابدأ بالأساس التنظيمي ANPP/MIPH وفعّل وحدات الجودة أو المجمعات حسب احتياجاتك.',
            'off.o1t': 'Sanad Core RIM',
            'off.o1tag': 'التسجيل والأرشفة BPF',
            'off.o1d': 'CTD Studio، التعديلات، اعتماد DM، خزينة GED، التوقيع الإلكتروني، بوابة E-TASJIL و28 مخرجاً رسمياً.',
            'off.o2t': 'Sanad Suite + eQMS',
            'off.o2tag': 'التنظيم + تأكيد الجودة BPF',
            'off.o2d': 'كامل Sanad Core RIM + الانحرافات، CAPA، إدارة التغيير، تدقيق الموردين، الشكاوى وتكوين SOP.',
            'off.o3t': 'Sanad Groupe متعدد الوحدات',
            'off.o3tag': 'المجمعات والمواقع المتعددة',
            'off.o3d': 'إشراف مركزي للمدير التقني للمجمع على عدة وحدات معزولة مع أرشيف الاسترجاع .sanad.',
            'off.cta': 'اطلب دراسة',
            'demo.badge': 'عرض تجريبي وإعداد الحساب',
            'demo.title': 'احجز عرضاً توضيحياً أو حضّر لإطلاق حسابك',
            'demo.sub': 'يعرض لكم خبراؤنا المنصة كاملة ويجهزون بيئة العمل الخاصة بمؤسستكم.',
            'demo.alreadyTitle': 'هل لديكم حساب أو دعوة مسبقة؟',
            'demo.alreadyDesc': 'انتقل مباشرة إلى شاشة تسجيل الدخول الآمنة للتطبيق.',
            'demo.submit': 'إرسال الطلب',
            'contact.badge': 'اتصال مباشر',
            'contact.title': 'تواصل مع فريق SANAD Pharma',
            'contact.sub': 'هل لديكم استفسار حول التأهيل OQ/PQ أو الاستضافة السيادية أو النشر متعدد المواقع؟ راسلونا مباشرة.',
            'contact.supportLabel': 'المرافقة :',
            'contact.submit': 'إرسال الرسالة',
            'form.name': 'الاسم الكامل *',
            'form.email': 'البريد الإلكتروني المهني *',
            'form.org': 'المخبر / المؤسسة',
            'form.role': 'المنصب',
            'form.phone': 'الهاتف',
            'form.mode': 'نمط النشر المفضل',
            'form.msg': 'أهدافكم (CTD, DM, MIPH, eQMS)',
            'form.message': 'الرسالة *',
            'footer.desc': 'منصة سيادية لتسجيل الأدوية ANPP، اعتماد المستلزمات الطبية، تراخيص MIPH والأرشفة الصيدلانية BPF.',
            'footer.col1': 'المنصة',
            'footer.col2': 'الدخول والاتصال',
            'footer.disclaimer': 'البرمجية وحدها لا تجعل المؤسسة مطابقة: بل يشترك في ذلك تأهيل النظام (OQ/PQ) وإجراءاتكم التشغيلية وبنيتكم التحتية. توفر SANAD Pharma الآليات التقنية ووثائق التأهيل اللازمة لدعم ذلك.',
            'toast.sent': 'تم تسجيل طلبكم بنجاح. سيتواصل معكم فريقنا قريباً.',
            'toast.invalid': 'يرجى إدخال الاسم وبريد إلكتروني صحيح.'
        },
        en: {
            'nav.home': 'Home',
            'nav.platform': 'Platform',
            'nav.compliance': 'Compliance',
            'nav.offers': 'Plans',
            'nav.demo': 'Demo / Onboarding',
            'nav.contact': 'Contact',
            'nav.login': 'Sign In',
            'nav.theme': 'Day / Night mode',
            'home.badge': 'ANPP • MIPH • 21 CFR Part 11 • ALCOA+',
            'home.title': 'Eight interconnected modules, one data integrity foundation.',
            'home.subtitle': 'Manage medicine CTD registrations, medical device approvals, MIPH Tabadol establishment licenses, and GMP document control on a sovereign, multilingual platform deployable on-premises or in the cloud.',
            'home.ctaLogin': 'Launch SANAD Pharma',
            'home.ctaPlatform': 'Explore the platform',
            'home.kpi1': 'CTD Procedures (NI/RI/MI/NF/RF/MF)',
            'home.kpi2': 'LD-DEPP07 Variation Categories',
            'home.kpi3': 'QR-Sealed GMP Output Reports',
            'home.rolesTitle': 'Built for every regulated role in your organization',
            'home.rolesSub': 'Strict Segregation of Duties (Maker-Checker-Approver) and Zero-Trust access control.',
            'home.role1Title': 'Qualified Person / Technical Director',
            'home.role1Desc': 'Centralized electronic signing desk, 2FA TOTP legal sign-off, DCE batch release, and FR/AR parity checks.',
            'home.role2Title': 'Regulatory Affairs',
            'home.role2Desc': 'CTD Modules 1–5 assembly, Algerian working-day Smart Planning, deficiency tracking, and E-TASJIL export.',
            'home.role3Title': 'Quality Assurance & Control',
            'home.role3Desc': 'Independent Checker review, 5-branch GMP vault, deviations, CAPA, and supplier audits.',
            'home.role4Title': 'Executive & IT Governance',
            'home.role4Desc': 'Air-gapped on-prem or multi-unit SaaS deployment, full .sanad archive reversibility, and immutable audit trail.',
            'plat.badge': 'Unified Architecture',
            'plat.title': 'Shared principles',
            'plat.intro': 'Enable modules at your own pace. Every module runs on the same compliance and data integrity engine.',
            'plat.p1Title': 'Everything is traced',
            'plat.p1Desc': 'Every view, download, rejection, or approval is timestamped and geolocated in the Audit Trail.',
            'plat.p2Title': 'Nothing is erased',
            'plat.p2Desc': 'Append-only immutable audit log, minor/major versioning with reason-for-change, and sealed archiving.',
            'plat.p3Title': 'Every signature binds',
            'plat.p3Desc': '2FA TOTP re-authentication and solemn declaration compliant with 21 CFR Part 11.',
            'plat.p4Title': 'One unified interface',
            'plat.p4Desc': 'High-density 3-pane "Pharma Explorer" UI, consistent day or night, in French, Arabic RTL, and English.',
            'plat.ringTitle': 'What flows between modules',
            'plat.ringDesc': 'Zero re-entry between MIPH establishment licensing, CTD/MD dossier assembly, post-approval variations, and batch release.',
            'plat.depTitle': 'On your servers or in a sovereign Cloud',
            'plat.depSub': 'Single Go binary with zero external CDN dependencies, ready for air-gapped intranets or multi-site SaaS.',
            'plat.dep1Title': 'On-Premise (Air-Gapped)',
            'plat.dep1Desc': '100% offline installation inside your laboratory infrastructure with full local key sovereignty.',
            'plat.dep2Title': 'Sovereign Cloud SaaS',
            'plat.dep2Desc': 'Strict tenant isolation, multi-unit Group Technical Director governance, and .sanad portability.',
            'plat.suiteTitle': 'Each module, in detail',
            'plat.suiteSub': 'A complete modular suite covering the full regulatory and quality lifecycle.',
            'mod.m1t': 'Sanad Registry',
            'mod.m1d': 'MIPH establishment directory, 1.H.8 manufacturing/QC sites, and unified Medicines & Medical Devices catalog.',
            'mod.m2t': 'Sanad Corporate',
            'mod.m2d': 'Licensing for all 5 establishment categories (Decree 21-82), Smart e-Forms, and Tabadol export packages.',
            'mod.m3t': 'Sanad CTD Studio',
            'mod.m3d': 'Dynamic NI/RI/MI/NF/RF/MF checklists, 5-phase DAG gating, FR/AR parity checks, and Smart Planning.',
            'mod.m4t': 'Sanad Variations',
            'mod.m4d': '34 LD-DEPP07/V01 categories, automatic escalation, Art. 09 Grouping, and Change Impact Tree.',
            'mod.m5t': 'Sanad Vault',
            'mod.m5d': '5-branch GMP document vault, 30-min edit leases, DCE batch release, and SHA-256 CAS storage.',
            'mod.m6t': 'Sanad Gateway',
            'mod.m6d': 'BV1/BV2 fee slips, 1.A.4/1.A.5 reconciliation, E-TASJIL summary sheet, and RDV slot alerts.',
            'mod.m7t': 'Sanad eQMS',
            'mod.m7d': 'Optional GMP quality suite: deviations, CAPA, variation-linked Change Control, audits, and SOP training.',
            'mod.m8t': 'Sanad Copilot',
            'mod.m8d': 'Sovereign Human-in-the-Loop AI assistant for v0.1 scientific drafting and consistency checks.',
            'conf.badge': 'Regulatory Frameworks',
            'conf.title': 'Compliance, framework by framework',
            'conf.sub': 'Every blocking rule and official output in SANAD Pharma maps to Algerian regulations and international GxP standards.',
            'conf.c1t': 'ANPP Registration (CTD 2025)',
            'conf.c1d': 'Executive Decree 20-325, July 2025 CTD Guide (Modules 1–5), Originator vs Imported/Local Generic rules, and 1.P.8 / 1.P.9 commitments.',
            'conf.c2t': 'Variations LD-DEPP07/V01',
            'conf.c2d': 'Algerian Mmin, Mmod, Mmaj classification across 34 categories, dynamic escalation, and Art. 09 Grouping.',
            'conf.c3t': 'Medical Devices (DDM)',
            'conf.c3d': 'Executive Decree 20-324 & May 10, 2021 Order: 6-section dossier, 13-point admissibility grid, and BNA fee slips.',
            'conf.c4t': 'MIPH Establishments & Tabadol',
            'conf.c4d': 'Executive Decree 21-82 & 2025 Orders: 5 categories, 2-year lapse tracking, 92-day interim cap, and Tabadol packages.',
            'conf.c5t': '21 CFR Part 11 & EU Annex 11',
            'conf.c5d': 'Argon2id, offline TOTP RFC 6238, legally binding PDT e-signature, strict SoD, and append-only Audit Trail.',
            'conf.c6t': 'GAMP 5 & OQ/PQ Qualification',
            'conf.c6d': 'Operational and Performance Qualification protocol continuously maintained for audits and inspections.',
            'off.badge': 'Modular Licensing',
            'off.title': 'Plans tailored to your laboratory scale',
            'off.sub': 'Start with the ANPP/MIPH regulatory core and activate Quality or Multi-Site extensions as you grow.',
            'off.o1t': 'Sanad Core RIM',
            'off.o1tag': 'Registration & GMP Vault',
            'off.o1d': 'CTD Studio, LD-DEPP07 Variations, MD Approval, Sanad Vault, PDT Signing Desk, E-TASJIL Gateway, and 28 official reports.',
            'off.o2t': 'Sanad Suite + eQMS',
            'off.o2tag': 'RIM + GMP Quality Assurance',
            'off.o2d': 'All of Sanad Core RIM + Deviations, CAPA, variation-linked Change Control, Supplier Audits, Complaints, and SOP Training.',
            'off.o3t': 'Sanad Multi-Unit Group',
            'off.o3tag': 'Holdings & Multi-Site',
            'off.o3d': 'Cross-unit Group Technical Director oversight across isolated production sites with .sanad portability.',
            'off.cta': 'Request a proposal',
            'demo.badge': 'Demo & Onboarding',
            'demo.title': 'Schedule a demo or prepare your onboarding',
            'demo.sub': 'Our regulatory engineers walk you through the full suite (CTD, Variations, Tabadol, Signing Desk) and prepare your workspace.',
            'demo.alreadyTitle': 'Already a customer or authorized partner?',
            'demo.alreadyDesc': 'Go directly to the secure application authentication screen.',
            'demo.submit': 'Submit request',
            'contact.badge': 'Direct Contact',
            'contact.title': 'Talk with the SANAD Pharma team',
            'contact.sub': 'Questions about OQ/PQ qualification, sovereign hosting, or multi-site rollout? Reach out directly.',
            'contact.supportLabel': 'Support:',
            'contact.submit': 'Send message',
            'form.name': 'Full name *',
            'form.email': 'Work email *',
            'form.org': 'Laboratory / Company',
            'form.role': 'Role',
            'form.phone': 'Phone',
            'form.mode': 'Preferred deployment',
            'form.msg': 'Your goals (CTD, MD, MIPH, eQMS)',
            'form.message': 'Message *',
            'footer.desc': 'Sovereign platform for ANPP medicine registration, medical device approval, MIPH licensing, and GMP document control.',
            'footer.col1': 'Platform',
            'footer.col2': 'Access & Contact',
            'footer.disclaimer': 'Software alone does not make an organization compliant: system qualification (OQ/PQ), your SOPs, and your infrastructure all contribute. SANAD Pharma provides the technical controls and qualification documentation to support it.',
            'toast.sent': 'Your request has been recorded. Our team will reach out shortly.',
            'toast.invalid': 'Please enter your name and a valid email address.'
        }
    };

    let currentLang = localStorage.getItem('sp_portal_lang') || 'fr';
    if (!I18N[currentLang]) currentLang = 'fr';

    function showToast(message, isError = false) {
        const el = document.getElementById('sp-toast');
        if (!el) return;
        el.textContent = message;
        el.className = 'sp-toast' + (isError ? ' error' : '');
        el.hidden = false;
        clearTimeout(showToast._t);
        showToast._t = setTimeout(() => { el.hidden = true; }, 4500);
    }

    function applyLanguage(lang) {
        if (!I18N[lang]) return;
        currentLang = lang;
        localStorage.setItem('sp_portal_lang', lang);
        document.documentElement.lang = lang;
        document.documentElement.dir = lang === 'ar' ? 'rtl' : 'ltr';

        document.querySelectorAll('[data-i18n]').forEach((node) => {
            const key = node.getAttribute('data-i18n');
            if (I18N[lang][key]) {
                node.textContent = I18N[lang][key];
            }
        });

        document.querySelectorAll('.sp-lang-btn').forEach((btn) => {
            btn.classList.toggle('active', btn.getAttribute('data-lang') === lang);
        });
    }

    function applyTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('sp_portal_theme', theme);
        const icon = document.getElementById('sp-theme-icon');
        if (icon) icon.textContent = theme === 'dark' ? '☀️' : '🌙';
    }

    function navigateTo(route) {
        const target = ROUTES.includes(route) ? route : 'accueil';
        document.querySelectorAll('.sp-view').forEach((sec) => {
            sec.hidden = sec.getAttribute('data-view') !== target;
        });
        document.querySelectorAll('[data-route]').forEach((link) => {
            link.classList.toggle('active', link.getAttribute('data-route') === target);
        });
        closeDrawer();
        window.scrollTo({ top: 0, behavior: 'smooth' });
    }

    function routeFromHash() {
        const raw = (window.location.hash || '#/accueil').replace(/^#\/?/, '').split('?')[0].toLowerCase();
        navigateTo(ROUTES.includes(raw) ? raw : 'accueil');
    }

    function openDrawer() {
        const drawer = document.getElementById('sp-mobile-drawer');
        const backdrop = document.getElementById('sp-drawer-backdrop');
        const burger = document.getElementById('sp-burger-btn');
        if (drawer) {
            drawer.classList.add('open');
            drawer.setAttribute('aria-hidden', 'false');
        }
        if (backdrop) backdrop.hidden = false;
        if (burger) burger.setAttribute('aria-expanded', 'true');
    }

    function closeDrawer() {
        const drawer = document.getElementById('sp-mobile-drawer');
        const backdrop = document.getElementById('sp-drawer-backdrop');
        const burger = document.getElementById('sp-burger-btn');
        if (drawer) {
            drawer.classList.remove('open');
            drawer.setAttribute('aria-hidden', 'true');
        }
        if (backdrop) backdrop.hidden = true;
        if (burger) burger.setAttribute('aria-expanded', 'false');
    }

    async function handleLeadSubmit(e) {
        e.preventDefault();
        const form = e.currentTarget;
        const data = Object.fromEntries(new FormData(form).entries());
        data.language = currentLang;

        if (!data.full_name || !data.email || !data.email.includes('@')) {
            showToast(I18N[currentLang]['toast.invalid'], true);
            return;
        }

        const btn = form.querySelector('button[type="submit"]');
        if (btn) btn.disabled = true;

        try {
            const resp = await fetch('/api/v1/public/lead', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            const payload = await resp.json().catch(() => ({}));
            if (!resp.ok) {
                showToast(payload.error || I18N[currentLang]['toast.invalid'], true);
                return;
            }
            form.reset();
            showToast(I18N[currentLang]['toast.sent']);
        } catch (_) {
            showToast(I18N[currentLang]['toast.invalid'], true);
        } finally {
            if (btn) btn.disabled = false;
        }
    }

    document.addEventListener('DOMContentLoaded', () => {
        applyTheme(localStorage.getItem('sp_portal_theme') || 'light');
        applyLanguage(currentLang);
        routeFromHash();

        window.addEventListener('hashchange', routeFromHash);

        document.querySelectorAll('.sp-lang-btn').forEach((btn) => {
            btn.addEventListener('click', () => applyLanguage(btn.getAttribute('data-lang')));
        });

        const toggleTheme = () => {
            const current = document.documentElement.getAttribute('data-theme') || 'light';
            applyTheme(current === 'dark' ? 'light' : 'dark');
        };
        document.getElementById('sp-theme-toggle')?.addEventListener('click', toggleTheme);
        document.getElementById('sp-theme-toggle-mobile')?.addEventListener('click', toggleTheme);

        document.getElementById('sp-burger-btn')?.addEventListener('click', openDrawer);
        document.getElementById('sp-drawer-close')?.addEventListener('click', closeDrawer);
        document.getElementById('sp-drawer-backdrop')?.addEventListener('click', closeDrawer);

        document.getElementById('sp-form-demo')?.addEventListener('submit', handleLeadSubmit);
        document.getElementById('sp-form-contact')?.addEventListener('submit', handleLeadSubmit);
    });
})();
