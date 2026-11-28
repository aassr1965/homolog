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
            'home.badge': 'Industrie Pharmaceutique • Réglementation Algérienne & Internationale',
            'home.title': "Huit modules interconnectés pour piloter vos dossiers réglementaires et votre qualité.",
            'home.subtitle': "Préparez l'enregistrement de vos médicaments, l'homologation des dispositifs médicaux, vos agréments d'établissement et votre gestion documentaire sur une plateforme souveraine, en français, arabe et anglais.",
            'home.ctaLogin': 'Accéder à SANAD Pharma',
            'home.ctaPlatform': 'Découvrir la plateforme',
            'home.kpi1': 'Parcours officiels d’enregistrement et de renouvellement',
            'home.kpi2': 'Catégories de modifications post-autorisation gérées',
            'home.kpi3': 'Modèles de documents officiels prêts pour les autorités',
            'home.rolesTitle': 'Conçu pour chaque métier de votre laboratoire',
            'home.rolesSub': 'Chacun travaille dans son périmètre : rédaction, vérification qualité et signature finale sont strictement séparées.',
            'home.role1Title': 'Pharmacien Directeur Technique',
            'home.role1Desc': 'Espace centralisé de signature électronique sécurisée, libération des lots et contrôle de cohérence des notices français/arabe.',
            'home.role2Title': 'Affaires Réglementaires',
            'home.role2Desc': 'Montage pas-à-pas des dossiers, rétroplanning tenant compte des jours fériés, suivi des compléments et préparation des dépôts.',
            'home.role3Title': 'Assurance & Contrôle Qualité',
            'home.role3Desc': 'Relecture indépendante des pièces, coffre-fort documentaire structuré, suivi des écarts, actions correctives et audits.',
            'home.role4Title': 'Direction Générale & Informatique',
            'home.role4Desc': 'Installation locale sans internet ou Cloud souverain multi-sites, récupération complète des données à tout moment et traçabilité totale.',
            'plat.badge': 'Architecture Unifiée',
            'plat.title': 'Les principes partagés',
            'plat.intro': "Activez uniquement les modules dont vous avez besoin aujourd'hui. Tous reposent sur le même socle de fiabilité et de sécurité.",
            'plat.p1Title': 'Tout est tracé',
            'plat.p1Desc': 'Chaque consultation, téléchargement, refus ou signature est enregistré avec la date, l’heure et l’auteur.',
            'plat.p2Title': "Rien ne s'efface",
            'plat.p2Desc': 'Historique inviolable, suivi des versions avec motif obligatoire et conservation sécurisée des anciennes pièces.',
            'plat.p3Title': 'Chaque signature engage',
            'plat.p3Desc': 'Confirmation par double authentification et engagement formel du Pharmacien Directeur Technique.',
            'plat.p4Title': 'Une interface commune',
            'plat.p4Desc': 'Un espace de travail clair en 3 volets, identique de jour comme de nuit, en français, arabe et anglais.',
            'plat.ringTitle': 'Ce qui circule entre les modules',
            'plat.ringDesc': "Aucune ressaisie entre l'agrément de votre établissement, le montage de vos dossiers produits, les modifications et la libération des lots.",
            'plat.flow1': "Agréments ➔ Dossiers produits : Vos autorisations d'établissement et décisions du Directeur Technique sont reprises sans ressaisie.",
            'plat.flow2': "Documents ➔ Dossiers d'enregistrement : Certificats de bonnes pratiques, analyses et notices bilingues sont reliés automatiquement.",
            'plat.flow3': "Qualité ➔ Modifications : Tout changement usine qualifié génère le dossier de modification réglementaire adapté.",
            'plat.flow4': "Signature ➔ Dépôt officiel : Une fois visé par le Pharmacien Directeur Technique, le dossier prêt au dépôt est généré en un clic.",
            'plat.depTitle': 'Sur vos serveurs ou en Cloud souverain',
            'plat.depSub': 'Un logiciel autonome sans dépendance externe, conçu pour fonctionner en réseau interne isolé comme en Cloud multi-sites.',
            'plat.dep1Title': 'Installation sur site (hors-ligne)',
            'plat.dep1Desc': 'Fonctionnement 100 % autonome dans votre laboratoire. Vos données et clés de sécurité restent exclusivement chez vous.',
            'plat.dep2Title': 'Cloud Souverain multi-sites',
            'plat.dep2Desc': 'Cloisonnement strict entre établissements, supervision Groupe multi-usines et export complet de réversibilité à tout moment.',
            'plat.suiteTitle': 'Chaque module, en détail',
            'plat.suiteSub': 'Une suite modulaire complète couvrant tout le cycle réglementaire et qualité.',
            'mod.m1t': 'Référentiel Central',
            'mod.m1d': 'Annuaire unique de vos établissements, sites de fabrication, laboratoires de contrôle et catalogue de produits.',
            'mod.m2t': 'Agréments & Entreprise',
            'mod.m2d': 'Gestion des dossiers d’agrément ministériel (fabrication, importation, distribution, exploitation, exportation) et formulaires intelligents.',
            'mod.m3t': 'Dossiers Produits',
            'mod.m3d': 'Listes de contrôle automatiques pour les médicaments et dispositifs médicaux, alertes de complétude et calendrier intelligent.',
            'mod.m4t': 'Modifications Post-AMM',
            'mod.m4d': 'Qualification guidée des changements après autorisation, détection automatique de l’impact et dossier comparatif avant/après.',
            'mod.m5t': 'Coffre-Fort Documentaire',
            'mod.m5d': 'Classement pharmaceutique en 5 branches, contrôle des versions, verrouillage d’édition et libération des lots importés ou fabriqués.',
            'mod.m6t': 'Portail de Soumission',
            'mod.m6d': 'Calcul des redevances réglementaires, génération des bordereaux de paiement et préparation des dépôts électroniques et rendez-vous.',
            'mod.m7t': 'Système Qualité (Option)',
            'mod.m7d': 'Gestion des déviations, actions correctives et préventives, maîtrise des changements, audits internes et suivi des formations.',
            'mod.m8t': 'Assistant Rédactionnel IA',
            'mod.m8d': 'Aide à la rédaction des dossiers techniques et vérification de cohérence, toujours relue et validée par vos pharmaciens.',
            'conf.badge': 'Référentiels & Cadre Légal',
            'conf.title': 'La conformité, texte par texte',
            'conf.sub': 'Chaque contrôle bloquant et chaque document généré par SANAD Pharma s’appuie sur la réglementation algérienne et les standards internationaux.',
            'conf.c1t': 'Enregistrement des médicaments',
            'conf.c1d': 'Conforme au Décret exécutif n° 20-325 et au Guide de soumission de l’Agence Nationale des Produits Pharmaceutiques (ANPP).',
            'conf.c2t': 'Modifications après autorisation',
            'conf.c2d': 'Intègre la classification algérienne (mineure, modérée, majeure) sur 34 cas types et le regroupement des modifications.',
            'conf.c3t': 'Dispositifs médicaux',
            'conf.c3d': 'Conforme au Décret exécutif n° 20-324 : dossier en 6 sections, grille de recevabilité en 13 points et suivi des régularisations.',
            'conf.c4t': 'Agréments d’établissements',
            'conf.c4d': 'Conforme au Décret exécutif n° 21-82 du Ministère de l’Industrie Pharmaceutique : suivi des échéances et préparation des dépôts.',
            'conf.c5t': 'Signature électronique & Traçabilité',
            'conf.c5d': 'Aligné sur les exigences internationales (21 CFR Part 11 et Bonnes Pratiques de Fabrication) : double authentification et journal inviolable.',
            'conf.c6t': 'Validation du système (GAMP 5)',
            'conf.c6d': 'Cahier de recette et protocole de qualification opérationnelle et de performance maintenu à jour pour vos inspections.',
            'off.badge': 'Offres Modulaires',
            'off.title': 'Des formules adaptées à la taille de votre laboratoire',
            'off.sub': 'Démarrez avec le cœur réglementaire et ajoutez la gestion qualité ou la supervision multi-sites selon votre croissance.',
            'off.o1t': 'Socle Réglementaire',
            'off.o1tag': 'Enregistrement & Documents',
            'off.o1d': 'Dossiers médicaments et dispositifs médicaux, modifications, coffre-fort documentaire, signature électronique et préparation des dépôts.',
            'off.o2t': 'Suite Complète + Qualité',
            'off.o2tag': 'Réglementaire & Assurance Qualité',
            'off.o2d': 'Tout le socle réglementaire + gestion des écarts, actions correctives, maîtrise des changements, audits fournisseurs et formations.',
            'off.o3t': 'Groupe Multi-Sites',
            'off.o3tag': 'Holdings & Plusieurs Usines',
            'off.o3d': 'Pilotage centralisé pour les groupes exploitant plusieurs unités de production cloisonnées, avec export de sauvegarde complet.',
            'off.cta': 'Demander une étude',
            'demo.badge': 'Démonstration & Onboarding',
            'demo.title': 'Planifier une démonstration ou préparer votre onboarding',
            'demo.sub': "Nos spécialistes vous présentent l'application en direct et préparent l'ouverture de votre espace de travail.",
            'demo.alreadyTitle': 'Déjà client ou partenaire autorisé ?',
            'demo.alreadyDesc': "Accédez directement à l'écran de connexion sécurisé de l'application.",
            'demo.submit': 'Envoyer la demande',
            'contact.badge': 'Contact Direct',
            'contact.title': "Échanger avec l'équipe SANAD Pharma",
            'contact.sub': "Une question sur la validation, l'installation locale ou l'hébergement Cloud ? Écrivez-nous directement.",
            'contact.supportLabel': 'Accompagnement :',
            'contact.submit': 'Envoyer le message',
            'form.name': 'Nom et prénom *',
            'form.email': 'Email professionnel *',
            'form.org': 'Laboratoire / Entreprise',
            'form.role': 'Fonction',
            'form.phone': 'Téléphone',
            'form.mode': 'Mode souhaité',
            'form.msg': 'Vos besoins (enregistrement, agréments, qualité...)',
            'form.message': 'Message *',
            'footer.desc': "Plateforme souveraine pour l'enregistrement des médicaments, l'homologation des dispositifs médicaux, les agréments d'établissement et la gestion documentaire pharmaceutique.",
            'footer.col1': 'Plateforme',
            'footer.col2': 'Accès & Contact',
            'footer.disclaimer': "Un logiciel ne rend pas, à lui seul, une organisation conforme : la validation du système, vos procédures internes et votre infrastructure y concourent. SANAD Pharma fournit les mécanismes techniques et la documentation de qualification pour l'étayer.",
            'toast.sent': 'Votre demande a bien été enregistrée. Notre équipe vous contactera rapidement.',
            'toast.invalid': 'Veuillez renseigner votre nom et une adresse email valide.'
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
            'home.badge': 'الصناعة الصيدلانية • التنظيم الجزائري والمعايير الدولية',
            'home.title': 'ثماني وحدات مترابطة لإدارة ملفاتكم التنظيمية وضمان الجودة.',
            'home.subtitle': 'أعدوا ملفات تسجيل الأدوية، واعتماد المستلزمات الطبية، وتراخيص المؤسسات، وأرشفة الوثائق الصيدلانية على منصة سيادية بالعربية والفرنسية والإنجليزية، محلياً أو سحابياً.',
            'home.ctaLogin': 'الدخول إلى SANAD Pharma',
            'home.ctaPlatform': 'استكشف المنصة',
            'home.kpi1': 'مسارات رسمية للتسجيل والتجديد',
            'home.kpi2': 'فئة للتعديلات بعد التسجيل',
            'home.kpi3': 'نموذج وثائق رسمية جاهزة للهيئات',
            'home.rolesTitle': 'مصممة لكل مسؤولية داخل مؤسستكم',
            'home.rolesSub': 'يعمل كل طرف في مجاله: التحرير، المراجعة الجودة، والتوقيع النهائي مفصولة تماماً.',
            'home.role1Title': 'الصيدلي المدير التقني',
            'home.role1Desc': 'فضاء مركزي للتوقيع الإلكتروني المؤمّن، تحرير الدفعات ومطابقة النشرات بالعربية والفرنسية.',
            'home.role2Title': 'الشؤون التنظيمية',
            'home.role2Desc': 'تجهيز الملفات خطوة بخطوة، التخطيط الذكي بأيام العمل والعطل الرسمية، ومتابعة التحفظات.',
            'home.role3Title': 'تأكيد ومراقبة الجودة',
            'home.role3Desc': 'مراجعة مستقلة للوثائق، خزينة منظمة، متابعة الانحرافات والإجراءات التصحيحية والتدقيق.',
            'home.role4Title': 'الإدارة العامة وتقنية المعلومات',
            'home.role4Desc': 'تثبيت محلي بدون إنترنت أو سحابة سيادية متعددة المواقع مع إمكانية استرجاع كامل البيانات.',
            'plat.badge': 'بنية موحدة',
            'plat.title': 'المبادئ المشتركة',
            'plat.intro': 'فعّلوا الوحدات التي تحتاجونها اليوم فقط. جميعها تعمل على نفس قاعدة الأمان وسلامة البيانات.',
            'plat.p1Title': 'كل إجراء مسجل',
            'plat.p1Desc': 'كل اطلاع أو تحميل أو رفض أو توقيع يُحفظ بالتاريخ والساعة واسم المسؤول.',
            'plat.p2Title': 'لا شيء يُمحى',
            'plat.p2Desc': 'سجل غير قابل للتعديل، متابعة الإصدارات مع سبب التغيير وحفظ دائم للنسخ السابقة.',
            'plat.p3Title': 'كل توقيع مُلزم',
            'plat.p3Desc': 'تأكيد بالتحقق الثنائي والتزام رسمي من الصيدلي المدير التقني.',
            'plat.p4Title': 'واجهة موحدة',
            'plat.p4Desc': 'مساحة عمل واضحة بثلاثة أجزاء، متاحة ليلاً ونهاراً بالعربية والفرنسية والإنجليزية.',
            'plat.ringTitle': 'ما يتدفق بين الوحدات',
            'plat.ringDesc': 'ترابط تلقائي بين تراخيص المؤسسة، ملفات المنتجات، التعديلات وتحرير الدفعات دون إعادة إدخال.',
            'plat.flow1': 'التراخيص ➔ ملفات المنتجات: تُسترجع تراخيص المؤسسة وقرارات المدير التقني تلقائياً.',
            'plat.flow2': 'الوثائق ➔ ملفات التسجيل: تُربط الشهادات والتحاليل والنشرات المترجمة آلياً.',
            'plat.flow3': 'الجودة ➔ التعديلات: كل تغيير تقني معتمد يولّد ملف التعديل التنظيمي المناسب.',
            'plat.flow4': 'التوقيع ➔ الإيداع الرسمي: بعد تأشيرة المدير التقني يُجهز ملف الإيداع بنقرة واحدة.',
            'plat.depTitle': 'على خوادمكم أو في سحابة سيادية',
            'plat.depSub': 'برنامج مستقل دون أي اعتماد خارجي، يعمل في الشبكات المعزولة أو السحابة متعددة المواقع.',
            'plat.dep1Title': 'تثبيت محلي (بدون إنترنت)',
            'plat.dep1Desc': 'تشغيل مستقل 100% داخل مخبركم مع بقاء بياناتكم ومفاتيحكم حصراً لديكم.',
            'plat.dep2Title': 'سحابة سيادية متعددة المواقع',
            'plat.dep2Desc': 'عزل تام بين المؤسسات، إشراف مركزي للمجمعات وتصدير كامل للبيانات في أي وقت.',
            'plat.suiteTitle': 'كل وحدة بالتفصيل',
            'plat.suiteSub': 'منظومة متكاملة تغطي كامل الدورة التنظيمية والجودة.',
            'mod.m1t': 'المرجع المركزي',
            'mod.m1d': 'دليل موحد لمؤسساتكم، مواقع التصنيع، مخابر المراقبة وفهرس المنتجات.',
            'mod.m2t': 'التراخيص والمؤسسة',
            'mod.m2d': 'إدارة ملفات الاعتماد الوزاري (التصنيع، الاستيراد، التوزيع، الاستغلال، التصدير) والنماذج الذكية.',
            'mod.m3t': 'ملفات المنتجات',
            'mod.m3d': 'قوائم تحقق آلية للأدوية والمستلزمات الطبية مع متابعة الاكتمال والآجال.',
            'mod.m4t': 'التعديلات بعد التسجيل',
            'mod.m4d': 'تقييم موجه للتغييرات، كشف الأثر تلقائياً وإعداد ملف المقارنة قبل/بعد.',
            'mod.m5t': 'خزينة الوثائق',
            'mod.m5d': 'تصنيف صيدلاني في 5 فروع، إدارة الإصدارات، أقفال التحرير وتحرير الدفعات.',
            'mod.m6t': 'بوابة الإيداع',
            'mod.m6d': 'حساب الرسوم، توليد جداول الدفع وتحضير الإيداعات الإلكترونية والمواعيد.',
            'mod.m7t': 'نظام الجودة (اختياري)',
            'mod.m7d': 'إدارة الانحرافات، الإجراءات التصحيحية، إدارة التغيير، التدقيق ومتابعة التكوين.',
            'mod.m8t': 'مساعد التحرير الذكي',
            'mod.m8d': 'مساعدة في صياغة الملفات التقنية ومراجعة التناسق تحت الإشراف البشري للصيادلة.',
            'conf.badge': 'المرجعيات والإطار القانوني',
            'conf.title': 'المطابقة وفق كل نص تنظيمي',
            'conf.sub': 'كل قاعدة تحقق ومخرج رسمي في SANAD Pharma يستند إلى التنظيم الجزائري والمعايير الدولية.',
            'conf.c1t': 'تسجيل الأدوية',
            'conf.c1d': 'مطابق للمرسوم التنفيذي 20-325 ودليل الإيداع الصادر عن الوكالة الوطنية للمواد الصيدلانية.',
            'conf.c2t': 'التعديلات بعد التسجيل',
            'conf.c2d': 'يدمج التصنيف الجزائري (طفيفة، متوسطة، كبرى) عبر 34 حالة وتجميع التعديلات.',
            'conf.c3t': 'المستلزمات الطبية',
            'conf.c3d': 'مطابق للمرسوم التنفيذي 20-324: ملف من 6 أقسام وشبكة قبول من 13 نقطة.',
            'conf.c4t': 'تراخيص المؤسسات الصيدلانية',
            'conf.c4d': 'مطابق للمرسوم التنفيذي 21-82 لوزارة الصناعة الصيدلانية ومتابعة الآجال.',
            'conf.c5t': 'التوقيع الإلكتروني والتتبع',
            'conf.c5d': 'متوافق مع المعايير الدولية (21 CFR Part 11 وممارسات التصنيع الجيد): تحقق ثنائي وسجل غير قابل للتعديل.',
            'conf.c6t': 'تأهيل النظام (GAMP 5)',
            'conf.c6d': 'بروتوكول تأهيل تشغيلي وأدائي محيّن باستمرار لدعم عمليات التفتيش.',
            'off.badge': 'عروض مرنة',
            'off.title': 'باقات تتكيف مع حجم مؤسستكم',
            'off.sub': 'ابدؤوا بالأساس التنظيمي وأضيفوا نظام الجودة أو إدارة المواقع المتعددة حسب نموكم.',
            'off.o1t': 'الأساس التنظيمي',
            'off.o1tag': 'التسجيل والوثائق',
            'off.o1d': 'ملفات الأدوية والمستلزمات الطبية، التعديلات، خزينة الوثائق، التوقيع الإلكتروني وتحضير الإيداعات.',
            'off.o2t': 'المنظومة الكاملة + الجودة',
            'off.o2tag': 'التنظيم وتأكيد الجودة',
            'off.o2d': 'كامل الأساس التنظيمي + الانحرافات، الإجراءات التصحيحية، إدارة التغيير، تدقيق الموردين والتكوين.',
            'off.o3t': 'مجمع متعدد المواقع',
            'off.o3tag': 'المجمعات والمصانع المتعددة',
            'off.o3d': 'إشراف مركزي للمجمعات ذات الوحدات المعزولة مع تصدير كامل للنسخ الاحتياطي.',
            'off.cta': 'اطلب دراسة',
            'demo.badge': 'عرض تجريبي وإعداد الحساب',
            'demo.title': 'احجز عرضاً توضيحياً أو حضّر لإطلاق حسابك',
            'demo.sub': 'يعرض لكم خبراؤنا المنصة مباشرة ويجهزون فضاء العمل الخاص بمؤسستكم.',
            'demo.alreadyTitle': 'هل لديكم حساب أو دعوة مسبقة؟',
            'demo.alreadyDesc': 'انتقلوا مباشرة إلى شاشة تسجيل الدخول الآمنة للتطبيق.',
            'demo.submit': 'إرسال الطلب',
            'contact.badge': 'اتصال مباشر',
            'contact.title': 'تواصل مع فريق SANAD Pharma',
            'contact.sub': 'هل لديكم استفسار حول التأهيل أو التثبيت المحلي أو السحابة؟ راسلونا مباشرة.',
            'contact.supportLabel': 'المرافقة :',
            'contact.submit': 'إرسال الرسالة',
            'form.name': 'الاسم الكامل *',
            'form.email': 'البريد الإلكتروني المهني *',
            'form.org': 'المخبر / المؤسسة',
            'form.role': 'المنصب',
            'form.phone': 'الهاتف',
            'form.mode': 'نمط النشر المفضل',
            'form.msg': 'احتياجاتكم (التسجيل، التراخيص، الجودة...)',
            'form.message': 'الرسالة *',
            'footer.desc': 'منصة سيادية لتسجيل الأدوية، اعتماد المستلزمات الطبية، تراخيص المؤسسات وإدارة الوثائق الصيدلانية.',
            'footer.col1': 'المنصة',
            'footer.col2': 'الدخول والاتصال',
            'footer.disclaimer': 'البرمجية وحدها لا تجعل المؤسسة مطابقة: بل يشترك في ذلك تأهيل النظام وإجراءاتكم الداخلية وبنيتكم التحتية. توفر SANAD Pharma الآليات التقنية ووثائق التأهيل اللازمة لدعم ذلك.',
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
            'home.badge': 'Pharmaceutical Industry • Algerian & International Compliance',
            'home.title': 'Eight interconnected modules to manage your regulatory dossiers and quality.',
            'home.subtitle': 'Prepare medicine registrations, medical device approvals, facility licenses, and pharmaceutical document control on one sovereign platform in French, Arabic, and English — on-premises or in the cloud.',
            'home.ctaLogin': 'Launch SANAD Pharma',
            'home.ctaPlatform': 'Explore the platform',
            'home.kpi1': 'Official registration & renewal workflows',
            'home.kpi2': 'Post-approval variation categories covered',
            'home.kpi3': 'Official authority-ready report templates',
            'home.rolesTitle': 'Built for every team in your pharmaceutical organization',
            'home.rolesSub': 'Strict separation of responsibilities: drafting, quality review, and legal sign-off never overlap.',
            'home.role1Title': 'Qualified Person / Technical Director',
            'home.role1Desc': 'Centralized e-signature desk with two-factor authentication, batch release, and French/Arabic leaflet parity checks.',
            'home.role2Title': 'Regulatory Affairs',
            'home.role2Desc': 'Step-by-step dossier assembly, holiday-aware planning, authority query tracking, and submission packaging.',
            'home.role3Title': 'Quality Assurance & Control',
            'home.role3Desc': 'Independent document review, structured 5-branch document vault, deviations, corrective actions, and audits.',
            'home.role4Title': 'Executive & IT Governance',
            'home.role4Desc': 'Offline on-premises installation or multi-site sovereign cloud, full data portability anytime, and immutable audit logs.',
            'plat.badge': 'Unified Architecture',
            'plat.title': 'Shared principles',
            'plat.intro': 'Activate only the modules you need today. Every module runs on the same reliability and security foundation.',
            'plat.p1Title': 'Everything is traced',
            'plat.p1Desc': 'Every view, download, rejection, or approval is recorded with timestamp, user identity, and location.',
            'plat.p2Title': 'Nothing is erased',
            'plat.p2Desc': 'Tamper-proof audit history, version control with mandatory reason for change, and permanent archiving.',
            'plat.p3Title': 'Every signature binds',
            'plat.p3Desc': 'Two-factor re-authentication and formal legal attestation by the Technical Director.',
            'plat.p4Title': 'One unified interface',
            'plat.p4Desc': 'Clean 3-pane workspace, consistent day and night, in French, Arabic, and English.',
            'plat.ringTitle': 'What flows between modules',
            'plat.ringDesc': 'Zero re-entry between facility licensing, product registration dossiers, post-approval changes, and batch release.',
            'plat.flow1': 'Licensing ➔ Product Dossiers: Facility permits and Technical Director credentials are reused automatically.',
            'plat.flow2': 'Documents ➔ Registration: GMP certificates, lab analyses, and bilingual leaflets link directly.',
            'plat.flow3': 'Quality ➔ Variations: Approved factory changes automatically generate the matching regulatory variation file.',
            'plat.flow4': 'Signature ➔ Official Filing: Once signed off by the Technical Director, the submission package is built in one click.',
            'plat.depTitle': 'On your servers or in a sovereign Cloud',
            'plat.depSub': 'Self-contained software with zero external dependencies, built for offline intranets or multi-site cloud.',
            'plat.dep1Title': 'On-Premise (Offline)',
            'plat.dep1Desc': '100% offline operation inside your laboratory network. Your data and encryption keys stay strictly with you.',
            'plat.dep2Title': 'Sovereign Multi-Site Cloud',
            'plat.dep2Desc': 'Strict isolation between facilities, group-level oversight across plants, and full archive export anytime.',
            'plat.suiteTitle': 'Each module, in detail',
            'plat.suiteSub': 'A complete modular suite covering the full regulatory and quality lifecycle.',
            'mod.m1t': 'Central Directory',
            'mod.m1d': 'Single registry for your facilities, manufacturing and testing sites, and unified product catalog.',
            'mod.m2t': 'Facility Licensing',
            'mod.m2d': 'Ministry licensing workflows (manufacturing, import, distribution, exploitation, export) with smart forms.',
            'mod.m3t': 'Product Dossiers',
            'mod.m3d': 'Automated checklists for medicines and medical devices, completeness scoring, and smart deadlines.',
            'mod.m4t': 'Post-Approval Changes',
            'mod.m4d': 'Guided change classification, automatic impact detection, and before/after comparison dossiers.',
            'mod.m5t': 'Document Vault',
            'mod.m5d': '5-branch pharmaceutical filing, version control, edit locking, and batch release management.',
            'mod.m6t': 'Submission Gateway',
            'mod.m6d': 'Regulatory fee calculation, payment slip generation, and electronic filing & appointment preparation.',
            'mod.m7t': 'Quality System (Optional)',
            'mod.m7d': 'Deviations, corrective and preventive actions, change control, supplier audits, and SOP training.',
            'mod.m8t': 'AI Drafting Assistant',
            'mod.m8d': 'Drafting support and consistency checks for technical files, always reviewed and approved by your pharmacists.',
            'conf.badge': 'Regulatory Frameworks',
            'conf.title': 'Compliance, regulation by regulation',
            'conf.sub': 'Every blocking rule and official report in SANAD Pharma maps to Algerian regulations and international quality standards.',
            'conf.c1t': 'Medicine Registration',
            'conf.c1d': 'Aligned with Executive Decree 20-325 and the National Pharmaceutical Products Agency (ANPP) submission guide.',
            'conf.c2t': 'Post-Approval Variations',
            'conf.c2d': 'Implements Algerian minor, moderate, and major classifications across 34 cases plus variation grouping.',
            'conf.c3t': 'Medical Devices',
            'conf.c3d': 'Aligned with Executive Decree 20-324: 6-section technical file, 13-point admissibility grid, and transition tracking.',
            'conf.c4t': 'Facility Licensing',
            'conf.c4d': 'Aligned with Executive Decree 21-82 of the Ministry of Pharmaceutical Industry: deadline alerts and filing packages.',
            'conf.c5t': 'E-Signatures & Audit Trail',
            'conf.c5d': 'Aligned with international standards (21 CFR Part 11 and Good Manufacturing Practice): 2FA and tamper-proof logs.',
            'conf.c6t': 'System Validation (GAMP 5)',
            'conf.c6d': 'Operational and Performance Qualification protocol kept up to date for audits and authority inspections.',
            'off.badge': 'Modular Plans',
            'off.title': 'Plans tailored to your laboratory scale',
            'off.sub': 'Start with the regulatory core and add quality management or multi-site oversight as you grow.',
            'off.o1t': 'Regulatory Core',
            'off.o1tag': 'Registration & Documents',
            'off.o1d': 'Medicine and medical device dossiers, variations, document vault, e-signatures, and filing preparation.',
            'off.o2t': 'Full Suite + Quality',
            'off.o2tag': 'Regulatory & Quality Assurance',
            'off.o2d': 'Everything in Regulatory Core + deviations, corrective actions, change control, supplier audits, and training.',
            'off.o3t': 'Multi-Site Group',
            'off.o3tag': 'Holdings & Multiple Plants',
            'off.o3d': 'Centralized oversight across isolated production sites with full backup portability.',
            'off.cta': 'Request a proposal',
            'demo.badge': 'Demo & Onboarding',
            'demo.title': 'Schedule a demo or prepare your onboarding',
            'demo.sub': 'Our specialists walk you through the live application and prepare your workspace setup.',
            'demo.alreadyTitle': 'Already a customer or authorized partner?',
            'demo.alreadyDesc': 'Go directly to the secure application sign-in screen.',
            'demo.submit': 'Submit request',
            'contact.badge': 'Direct Contact',
            'contact.title': 'Talk with the SANAD Pharma team',
            'contact.sub': 'Questions about validation, offline installation, or cloud hosting? Reach out directly.',
            'contact.supportLabel': 'Support:',
            'contact.submit': 'Send message',
            'form.name': 'Full name *',
            'form.email': 'Work email *',
            'form.org': 'Laboratory / Company',
            'form.role': 'Role',
            'form.phone': 'Phone',
            'form.mode': 'Preferred deployment',
            'form.msg': 'Your needs (registration, licensing, quality...)',
            'form.message': 'Message *',
            'footer.desc': 'Sovereign platform for medicine registration, medical device approval, facility licensing, and pharmaceutical document control.',
            'footer.col1': 'Platform',
            'footer.col2': 'Access & Contact',
            'footer.disclaimer': 'Software alone does not make an organization compliant: system validation, your internal procedures, and your infrastructure all contribute. SANAD Pharma provides the technical controls and qualification documentation to support it.',
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
