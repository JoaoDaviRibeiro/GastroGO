(function () {
    const I18N_EVENT = 'gastro:locale-change';
    const LOCALE_STORAGE_KEY = 'gg_locale';
    const DEFAULT_LOCALE = 'en';
    const SUPPORTED_LOCALES = ['en', 'pt-BR'];

    const translations = {
        en: {
            meta: {
                title: {
                    auth: 'GastroGO Auth',
                    dashboard: 'GastroGO Dashboard',
                    profile: 'GastroGO Profile'
                }
            },
            common: {
                brand: 'GASTROGO',
                language: {
                    label: 'Language',
                    options: {
                        en: 'English',
                        'pt-BR': 'Português (BR)'
                    }
                },
                themeToggle: {
                    toLight: 'Switch to light mode',
                    toDark: 'Switch to dark mode'
                },
                session: {
                    expired: 'Your session has expired. Please login again.'
                },
                actions: {
                    logout: 'Logout',
                    refresh: 'Refresh',
                    openProfile: 'Open profile',
                    backToDashboard: 'Back to dashboard',
                    clearFilters: 'Clear filters',
                    reset: 'Reset',
                    saveChanges: 'Save changes',
                    saving: 'Saving...'
                },
                placeholders: {
                    phone: '+1 555 123 4567',
                    location: 'Lisbon, Portugal',
                    bio: 'Share your focus areas or kitchen specialties.',
                    profileName: 'Jamie Rivera',
                    profileTitle: 'Ops Lead',
                    searchRestaurants: 'Search by name, cuisine, or address'
                },
                select: {
                    allCuisines: 'All cuisines'
                },
                status: {
                    loading: 'Loading...',
                    genericError: 'Something went wrong.'
                }
            },
            auth: {
                hero: {
                    eyebrow: 'GASTROGO',
                    headline: 'Modern hospitality\nmeets precise access.',
                    body: 'Manage your kitchens, staff, and guest experiences from one secure portal. Sign in to continue or start fresh with a new workspace.'
                },
                stats: {
                    teamsValue: '200+',
                    teamsLabel: 'active teams today',
                    uptimeValue: '99.98%',
                    uptimeLabel: 'uptime last quarter'
                },
                modes: {
                    login: {
                        label: 'Log in',
                        headline: 'Sign in to GastroGO',
                        lede: 'Use your email and password to access your dashboard.',
                        cta: 'Let me in'
                    },
                    signup: {
                        label: 'Create account',
                        headline: 'Create your workspace',
                        lede: 'Use your email and password to spin up a secure account.',
                        cta: 'Join the chefs'
                    }
                },
                form: {
                    emailLabel: 'Email address',
                    emailPlaceholder: 'chef@gastrogosuite.com',
                    passwordLabel: 'Password',
                    passwordPlaceholder: '••••••••',
                    forgot: 'Forgot password?',
                    forgotWorking: 'Sending link...',
                    submitWorking: 'Working...'
                },
                status: {
                    processing: 'Processing...',
                    resetting: 'Checking your email...',
                    successRedirect: 'Login successful! Redirecting...',
                    signupSuccess: 'Sign up successful! Check your inbox for verification.',
                    credsMissing: 'Email and password are required.',
                    credsInvalid: 'Check your credentials.',
                    connectionFailed: 'Connection failed: {error}',
                    resetSuccess: 'If this email exists, a reset link is on its way.',
                    resetUnavailable: 'Unable to request a reset right now.',
                    resetMissingEmail: 'Please enter the email you registered with.',
                    resetFailed: 'Reset request failed: {error}'
                }
            },
            dashboard: {
                header: {
                    eyebrow: 'GASTROGO',
                    title: 'Welcome!',
                    body: 'Monitor every location, sync feedback, and keep your staff informed from a single pane.'
                },
                status: {
                    syncing: 'Syncing your properties...',
                    ready: 'Data synced. Ready to manage operations.',
                    error: 'Unable to load dashboard. Please try again.',
                    ratingSending: 'Submitting rating...',
                    ratingSaved: 'Thanks! {rating}-star rating saved.',
                    ratingDuplicate: 'Rating failed. You might have already rated this.',
                    ratingConnection: 'Connection error. Check your server logs.'
                },
                stats: {
                    active: {
                        label: 'Active restaurants',
                        sublabel: 'Total properties synced'
                    },
                    mapped: {
                        label: 'Mapped locations',
                        sublabel: 'GPS-ready venues'
                    },
                    missing: {
                        label: 'Missing addresses',
                        sublabel: 'Needs quick updates'
                    }
                },
                panel: {
                    title: 'Restaurant overview',
                    body: 'Search, filter, and rate performance to keep the signal flowing.',
                    loading: 'Pulling the latest data...',
                    empty: 'No restaurants match that filter.',
                    searchPlaceholder: 'Search by name, cuisine, or address',
                    resetFilters: 'Clear filters',
                    refresh: 'Refresh',
                    refreshCache: 'Refresh cache'
                },
                sort: {
                    ratingHighToLow: 'Google rating: high to low',
                    ratingLowToHigh: 'Google rating: low to high'
                },
                cards: {
                    cuisineTbd: 'Cuisine TBD',
                    delivery: 'delivery',
                    addressMissing: 'Address not listed yet',
                    gpsReady: 'GPS ready',
                    addCoordinates: 'Add coordinates',
                    viewOnMaps: 'View on Maps ↗',
                    ratingPrompt: 'How is this location performing?',
                    ratingUnavailable: 'Rating unavailable',
                    googleRating: '{rating} ★ · {reviews} reviews'
                },
                popularTimes: {
                    title: 'Popular times',
                    current: 'Now: {value}%',
                    weeklyPeak: 'Peak: {day} at {value}%',
                    todayHours: 'Today’s hourly pattern ({day})',
                    scale: 'Busy percentage',
                    unavailable: 'Mostly full on weekends.',
                    estimated: 'Estimated',
                    loading: 'Loading popular times...'
                },
                filter: {
                    allCuisines: 'All cuisines'
                },
                pagination: {
                    previous: 'Previous 15',
                    next: 'Next 15',
                    summary: 'Showing {start}-{end} of {total}',
                    empty: 'No restaurants to show.'
                },
                userPillFallback: 'Loading profile...',
                rating: {
                    ariaLabel: 'Rate {value} star{suffix}'
                }
            },
            profile: {
                header: {
                    eyebrow: 'GASTROGO',
                    title: 'Profile & Preferences',
                    body: 'Review your account details, update contact info, and control how GastroGO keeps you in the loop.'
                },
                status: {
                    loading: 'Loading your profile...',
                    fallback: 'Profile endpoint not found; showing dashboard info instead.',
                    ready: 'Profile loaded. Make your updates below.',
                    errorFetch: 'Unable to fetch profile. Please try again.',
                    reverted: 'Changes reverted.',
                    saving: 'Saving profile...',
                    saved: 'Profile updated successfully.',
                    saveError: 'Unable to save profile.'
                },
                popularTimes: {
                    title: 'Popular times',
                    current: 'Now: {value}%',
                    weeklyPeak: 'Peak: {day} at {value}%',
                    todayHours: 'Today’s hourly pattern ({day})',
                    scale: 'Busy percentage',
                    unavailable: 'Mostly full on weekends.',
                    estimated: 'Estimated',
                    loading: 'Loading popular times...'
                },
                section: {
                    personalTitle: 'Personal details',
                    personalLede: 'This information appears on workspace rosters and shared documents.'
                },
                form: {
                    fullName: 'Full name',
                    title: 'Role / Title',
                    phone: 'Phone number',
                    location: 'Location',
                    bio: 'Bio / Notes',
                    reset: 'Reset',
                    save: 'Save changes',
                    saving: 'Saving...'
                },
                panel: {
                    loading: 'Fetching profile...'
                }
            }
        },
        'pt-BR': {
            meta: {
                title: {
                    auth: 'GastroGO | Acesso',
                    dashboard: 'GastroGO | Painel',
                    profile: 'GastroGO | Perfil'
                }
            },
            common: {
                brand: 'GASTROGO',
                language: {
                    label: 'Idioma',
                    options: {
                        en: 'Inglês',
                        'pt-BR': 'Português (BR)'
                    }
                },
                themeToggle: {
                    toLight: 'Ativar modo claro',
                    toDark: 'Ativar modo escuro'
                },
                session: {
                    expired: 'Sua sessão expirou. Faça login novamente.'
                },
                actions: {
                    logout: 'Sair',
                    refresh: 'Atualizar',
                    openProfile: 'Abrir perfil',
                    backToDashboard: 'Voltar para o painel',
                    clearFilters: 'Limpar filtros',
                    reset: 'Reverter',
                    saveChanges: 'Salvar alterações',
                    saving: 'Salvando...'
                },
                placeholders: {
                    phone: '+55 11 99999-0000',
                    location: 'São Paulo, Brasil',
                    bio: 'Compartilhe áreas de foco ou especialidades da cozinha.',
                    profileName: 'Marina Souza',
                    profileTitle: 'Líder de Operações',
                    searchRestaurants: 'Busque por nome, cozinha ou endereço'
                },
                select: {
                    allCuisines: 'Todas as cozinhas'
                },
                status: {
                    loading: 'Carregando...',
                    genericError: 'Algo deu errado.'
                }
            },
            auth: {
                hero: {
                    eyebrow: 'GASTROGO',
                    headline: 'Hospitalidade moderna\ncom acesso preciso.',
                    body: 'Gerencie cozinhas, equipes e a experiência dos clientes em um portal seguro. Entre para continuar ou crie um novo workspace.'
                },
                stats: {
                    teamsValue: '200+',
                    teamsLabel: 'equipes ativas hoje',
                    uptimeValue: '99,98%',
                    uptimeLabel: 'uptime no último trimestre'
                },
                modes: {
                    login: {
                        label: 'Entrar',
                        headline: 'Entre no GastroGO',
                        lede: 'Use seu e-mail e senha para acessar o painel.',
                        cta: 'Deixe-me entrar'
                    },
                    signup: {
                        label: 'Criar conta',
                        headline: 'Crie seu workspace',
                        lede: 'Use seu e-mail e senha para abrir uma conta segura.',
                        cta: 'Juntar-se aos chefs'
                    }
                },
                form: {
                    emailLabel: 'Endereço de e-mail',
                    emailPlaceholder: 'chef@gastrogosuite.com',
                    passwordLabel: 'Senha',
                    passwordPlaceholder: '••••••••',
                    forgot: 'Esqueceu a senha?',
                    forgotWorking: 'Enviando link...',
                    submitWorking: 'Trabalhando...'
                },
                status: {
                    processing: 'Processando...',
                    resetting: 'Conferindo seu e-mail...',
                    successRedirect: 'Login realizado! Redirecionando...',
                    signupSuccess: 'Cadastro concluído! Confira seu e-mail para confirmar.',
                    credsMissing: 'E-mail e senha são obrigatórios.',
                    credsInvalid: 'Verifique suas credenciais.',
                    connectionFailed: 'Falha na conexão: {error}',
                    resetSuccess: 'Se este e-mail existir, enviaremos um link em instantes.',
                    resetUnavailable: 'Não foi possível solicitar a redefinição agora.',
                    resetMissingEmail: 'Informe o e-mail usado no cadastro.',
                    resetFailed: 'Falha ao solicitar redefinição: {error}'
                }
            },
            dashboard: {
                header: {
                    eyebrow: 'GASTROGO',
                    title: 'Bem Vindo!',
                    body: 'Monitore cada unidade, centralize feedbacks e mantenha a equipe alinhada em um único painel.'
                },
                status: {
                    syncing: 'Sincronizando seus estabelecimentos...',
                    ready: 'Dados atualizados. Operações prontas.',
                    error: 'Não foi possível carregar o painel. Tente novamente.',
                    ratingSending: 'Enviando avaliação...',
                    ratingSaved: 'Obrigado! Avaliação de {rating} estrela{suffix} salva.',
                    ratingDuplicate: 'Não foi possível registrar a avaliação. Talvez você já tenha enviado.',
                    ratingConnection: 'Erro de conexão. Verifique os logs do servidor.'
                },
                stats: {
                    active: {
                        label: 'Restaurantes ativos',
                        sublabel: 'Total sincronizado'
                    },
                    mapped: {
                        label: 'Locais mapeados',
                        sublabel: 'Unidades com GPS'
                    },
                    missing: {
                        label: 'Endereços pendentes',
                        sublabel: 'Precisa de atualização rápida'
                    }
                },
                panel: {
                    title: 'Visão geral dos restaurantes',
                    body: 'Pesquise, filtre e avalie o desempenho para manter o sinal vivo.',
                    loading: 'Buscando os dados mais recentes...',
                    empty: 'Nenhum restaurante corresponde ao filtro.',
                    searchPlaceholder: 'Busque por nome, cozinha ou endereço',
                    resetFilters: 'Limpar filtros',
                    refresh: 'Atualizar',
                    refreshCache: 'Atualizar cache'
                },
                sort: {
                    ratingHighToLow: 'Avaliação do Google: maior para menor',
                    ratingLowToHigh: 'Avaliação do Google: menor para maior'
                },
                cards: {
                    cuisineTbd: 'Cozinha indefinida',
                    delivery: 'delivery',
                    addressMissing: 'Endereço ainda não informado',
                    gpsReady: 'GPS pronto',
                    addCoordinates: 'Adicionar coordenadas',
                    viewOnMaps: 'Ver no Maps ↗',
                    ratingPrompt: 'Como está o desempenho desta unidade?',
                    ratingUnavailable: 'Sem avaliações públicas',
                    googleRating: '{rating} ★ · {reviews} avaliações'
                },
                popularTimes: {
                    title: 'Horários populares',
                    current: 'Agora: {value}%',
                    weeklyPeak: 'Pico: {day} às {value}%',
                    todayHours: 'Padrão horário de hoje ({day})',
                    scale: 'Percentual de movimento',
                    unavailable: 'Principalmente cheio nos finais de semana.',
                    estimated: 'Estimativa',
                    loading: 'Carregando horários populares...'
                },
                filter: {
                    allCuisines: 'Todas as cozinhas'
                },
                pagination: {
                    previous: '15 anteriores',
                    next: 'Próximos 15',
                    summary: 'Mostrando {start}-{end} de {total}',
                    empty: 'Nenhum restaurante para exibir.'
                },
                userPillFallback: 'Carregando perfil...',
                rating: {
                    ariaLabel: 'Avaliar com {value} estrela{suffix}'
                }
            },
            profile: {
                header: {
                    eyebrow: 'GASTROGO',
                    title: 'Perfil e preferências',
                    body: 'Revise seus dados, atualize contatos e defina como o GastroGO fala com você.'
                },
                status: {
                    loading: 'Carregando seu perfil...',
                    fallback: 'Endpoint de perfil ausente; exibindo dados do painel.',
                    ready: 'Perfil carregado. Ajuste os dados abaixo.',
                    errorFetch: 'Não foi possível buscar o perfil. Tente novamente.',
                    reverted: 'Alterações revertidas.',
                    saving: 'Salvando perfil...',
                    saved: 'Perfil atualizado com sucesso.',
                    saveError: 'Não foi possível salvar o perfil.'
                },
                section: {
                    personalTitle: 'Dados pessoais',
                    personalLede: 'Essas informações aparecem em listas do workspace e documentos compartilhados.'
                },
                form: {
                    fullName: 'Nome completo',
                    title: 'Cargo / Função',
                    phone: 'Telefone',
                    location: 'Localização',
                    bio: 'Bio / Notas',
                    reset: 'Reverter',
                    save: 'Salvar alterações',
                    saving: 'Salvando...'
                },
                panel: {
                    loading: 'Buscando perfil...'
                }
            }
        }
    };

    const deepGet = (locale, key) => {
        const source = translations[locale];
        if (!source) return undefined;
        return key.split('.').reduce((acc, segment) => {
            if (acc && Object.prototype.hasOwnProperty.call(acc, segment)) {
                return acc[segment];
            }
            return undefined;
        }, source);
    };

    const formatTemplate = (value, params) => {
        if (typeof value !== 'string') {
            return value;
        }
        if (!params) {
            return value;
        }
        return value.replace(/\{(\w+)\}/g, (_, token) => {
            if (Object.prototype.hasOwnProperty.call(params, token)) {
                return params[token];
            }
            return '';
        });
    };

    const resolveLocale = (candidate) => {
        if (!candidate) {
            return null;
        }
        if (SUPPORTED_LOCALES.includes(candidate)) {
            return candidate;
        }
        const lower = candidate.toLowerCase();
        if (lower.startsWith('pt')) {
            return 'pt-BR';
        }
        if (lower.startsWith('en')) {
            return 'en';
        }
        return null;
    };

    const getBrowserLocale = () => {
        if (typeof navigator === 'undefined') {
            return null;
        }
        return resolveLocale(navigator.language || navigator.userLanguage);
    };

    const storedLocale = (() => {
        try {
            return resolveLocale(localStorage.getItem(LOCALE_STORAGE_KEY));
        } catch (error) {
            return null;
        }
    })();

    let currentLocale = storedLocale || getBrowserLocale() || DEFAULT_LOCALE;

    const applyDocumentLanguage = (locale) => {
        if (typeof document === 'undefined') {
            return;
        }
        document.documentElement.setAttribute('lang', locale);
    };

    applyDocumentLanguage(currentLocale);

    const translate = (locale, key, params) => {
        const result = deepGet(locale, key);
        if (result === undefined) {
            return undefined;
        }
        if (typeof result === 'string') {
            return formatTemplate(result, params);
        }
        return result;
    };

    const t = (key, params) => {
        const primary = translate(currentLocale, key, params);
        if (primary !== undefined) {
            return primary;
        }
        const fallback = translate(DEFAULT_LOCALE, key, params);
        if (fallback !== undefined) {
            return fallback;
        }
        return key;
    };

    const setLocale = (nextLocale) => {
        const normalized = resolveLocale(nextLocale) || DEFAULT_LOCALE;
        if (normalized === currentLocale) {
            return currentLocale;
        }
        currentLocale = normalized;
        try {
            localStorage.setItem(LOCALE_STORAGE_KEY, currentLocale);
        } catch (error) {
            // Ignore storage errors
        }
        applyDocumentLanguage(currentLocale);
        window.dispatchEvent(new CustomEvent(I18N_EVENT, { detail: { locale: currentLocale } }));
        return currentLocale;
    };

    window.addEventListener('storage', (event) => {
        if (event.key !== LOCALE_STORAGE_KEY) {
            return;
        }
        if (!event.newValue) {
            return;
        }
        const next = resolveLocale(event.newValue);
        if (next && next !== currentLocale) {
            currentLocale = next;
            applyDocumentLanguage(currentLocale);
            window.dispatchEvent(new CustomEvent(I18N_EVENT, { detail: { locale: currentLocale } }));
        }
    });

    window.t = (key, params) => t(key, params);
    window.getLocale = () => currentLocale;
    window.setLocale = setLocale;
    window.getSupportedLocales = () => SUPPORTED_LOCALES.slice();

    window.useI18n = function useI18n() {
        if (!window.React || !React.useState) {
            throw new Error('React must be loaded before calling useI18n.');
        }

        const { useState, useEffect, useCallback } = React;
        const [locale, setLocaleState] = useState(currentLocale);
        const translateFn = useCallback((key, params) => window.t(key, params), [locale]);
        const setPreferredLocale = useCallback((value) => window.setLocale(value), []);

        useEffect(() => {
            const handleLocaleChange = (event) => {
                const next = event.detail?.locale || window.getLocale();
                setLocaleState((prev) => (prev === next ? prev : next));
            };

            window.addEventListener(I18N_EVENT, handleLocaleChange);
            return () => window.removeEventListener(I18N_EVENT, handleLocaleChange);
        }, []);

        return {
            locale,
            t: translateFn,
            setLocale: setPreferredLocale,
            options: SUPPORTED_LOCALES.slice()
        };
    };
})();
