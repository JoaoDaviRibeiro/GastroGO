const THEME_STORAGE_KEY = 'gg_theme';
const THEME_EVENT = 'gastro:theme-change';

const prefersDarkMode = () => window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;

const applyThemeToBody = (theme) => {
    const normalized = theme === 'light' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', normalized);
    document.documentElement.style.setProperty('color-scheme', normalized);

    const annotatePageNodes = () => {
        if (!document.body) return;
        document.querySelectorAll('.page').forEach((node) => {
            node.setAttribute('data-theme', normalized);
        });
    };

    const setBodyAttr = () => {
        if (document.body) {
            document.body.setAttribute('data-theme', normalized);
            annotatePageNodes();
        }
    };

    if (document.body) {
        setBodyAttr();
    } else {
        document.addEventListener('DOMContentLoaded', () => {
            setBodyAttr();
            annotatePageNodes();
        }, { once: true });
    }

    window.currentTheme = normalized;
    window.dispatchEvent(new CustomEvent(THEME_EVENT, { detail: { theme: normalized } }));
};

const initializeTheme = () => {
    const stored = localStorage.getItem(THEME_STORAGE_KEY);
    if (stored === 'light' || stored === 'dark') {
        applyThemeToBody(stored);
        return stored;
    }

    const fallback = prefersDarkMode() ? 'dark' : 'light';
    applyThemeToBody(fallback);
    return fallback;
};

const setPreferredTheme = (theme) => {
    const normalized = theme === 'light' ? 'light' : 'dark';
    localStorage.setItem(THEME_STORAGE_KEY, normalized);
    applyThemeToBody(normalized);
    return normalized;
};

const togglePreferredTheme = () => {
    const nextTheme = window.currentTheme === 'light' ? 'dark' : 'light';
    return setPreferredTheme(nextTheme);
};

const initialTheme = initializeTheme();
window.currentTheme = initialTheme;
window.setPreferredTheme = setPreferredTheme;
window.togglePreferredTheme = togglePreferredTheme;
window.getPreferredTheme = () => window.currentTheme;

const translate = (key, params) => {
    if (typeof window !== 'undefined' && typeof window.t === 'function') {
        return window.t(key, params);
    }
    return key;
};

if (window.matchMedia) {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleSchemeChange = (event) => {
        const stored = localStorage.getItem(THEME_STORAGE_KEY);
        if (stored === 'light' || stored === 'dark') return;
        applyThemeToBody(event.matches ? 'dark' : 'light');
    };
    mediaQuery.addEventListener('change', handleSchemeChange);
}

window.addEventListener('storage', (event) => {
    if (event.key === THEME_STORAGE_KEY && event.newValue) {
        applyThemeToBody(event.newValue);
    }
});

async function auth(action, credentials = {}) {
    const { email, password } = credentials;

    if (!email || !password) {
        return { success: false, tone: 'error', message: translate('auth.status.credsMissing') };
    }

    try {
        const response = await fetch(`/api/${action}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        const data = await response.json();

        if (response.ok) {
            if (action === 'signup') {
                return {
                    success: true,
                    tone: 'success',
                    message: translate('auth.status.signupSuccess')
                };
            }

            localStorage.setItem('sb_token', data.access_token);

            return {
                success: true,
                tone: 'info',
                message: translate('auth.status.successRedirect')
            };
        }

        return {
            success: false,
            tone: 'error',
            message: data.error_description || data.error || translate('auth.status.credsInvalid')
        };
    } catch (err) {
        return {
            success: false,
            tone: 'error',
            message: translate('auth.status.connectionFailed', { error: err.message })
        };
    }
}

async function requestPasswordReset(email) {
    if (!email) {
        return { success: false, tone: 'error', message: translate('auth.status.resetMissingEmail') };
    }

    try {
        const response = await fetch('/api/password-reset', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email })
        });

        if (response.ok) {
            return {
                success: true,
                tone: 'success',
                message: translate('auth.status.resetSuccess')
            };
        }

        const data = await response.json().catch(() => ({}));
        return {
            success: false,
            tone: 'error',
            message: data.error_description || data.error || translate('auth.status.resetUnavailable')
        };
    } catch (err) {
        return {
            success: false,
            tone: 'error',
            message: translate('auth.status.resetFailed', { error: err.message })
        };
    }
}

window.auth = auth;
window.requestPasswordReset = requestPasswordReset;

async function loadDashboard() {
    const token = localStorage.getItem('sb_token');
    const display = document.getElementById('dashboard-data');

    if (!token) {
        window.location.href = "/";
        return;
    }

    try {
        // Fetch User Info and Restaurants in parallel for better performance
        const [userRes, resRes] = await Promise.all([
            fetch('/api/dashboard', { headers: { 'Authorization': `Bearer ${token}` } }),
            fetch('/api/restaurants', { headers: { 'Authorization': `Bearer ${token}` } })
        ]);

        if (userRes.status === 401 || resRes.status === 401) return handleSessionExpiry();

        if (userRes.ok && resRes.ok) {
            const userData = await userRes.json();
            const restaurants = await resRes.json();

            let restaurantHTML = '<h3>🍽️ GastroGO Restaurants</h3>';
            
            if (restaurants && restaurants.length > 0) {
                restaurants.forEach(res => {
                    const mapsUrl = `https://www.google.com/maps?q=${res.lat},${res.lng}`;
                    const hasCoords = res.lat !== 0 && res.lng !== 0;

                    restaurantHTML += `
                        <div class="restaurant-card">
                            <strong>${res.name}</strong> (${res.cuisine})<br>
                            <small>📍 ${res.address || 'Address not listed'}</small><br>
                            
                            <div class="star-rating" data-rid="${res.id}">
                                <span class="star" onclick="submitRating(${res.id}, 1)">★</span>
                                <span class="star" onclick="submitRating(${res.id}, 2)">★</span>
                                <span class="star" onclick="submitRating(${res.id}, 3)">★</span>
                                <span class="star" onclick="submitRating(${res.id}, 4)">★</span>
                                <span class="star" onclick="submitRating(${res.id}, 5)">★</span>
                            </div>

                            <br>
                            <small>
                                ${hasCoords ? `<a href="${mapsUrl}" target="_blank">View on Map</a>` : 'No GPS data'}
                            </small>
                        </div><hr style="border: 0.5px solid #eee;">`;
                });
            } else {
                restaurantHTML += '<p>No restaurants found. Add your first one!</p>';
            }

            display.innerHTML = `
                <p>Welcome back, <strong>${userData.email}</strong>!</p>
                <hr>
                ${restaurantHTML}
            `;
        }
    } catch (err) {
        console.error("Dashboard load failed:", err);
    }
}

// Logic to send the rating to the Go Backend
async function submitRating(restaurantId, ratingValue) {
    const token = localStorage.getItem('sb_token');
    
    try {
        const response = await fetch('/api/rate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            // CRITICAL: Ensure we send pure numbers, not strings
            body: JSON.stringify({ 
                restaurant_id: parseInt(restaurantId), 
                rating: parseFloat(ratingValue) 
            })
        });

        if (response.ok) {
            alert(`You rated this ${ratingValue} stars!`);
        } else {
            const errorData = await response.json().catch(() => ({}));
            // Show the actual error from the backend (409 Conflict)
            alert(errorData.error || "Rating failed. You might have already rated this.");
        }
    } catch (err) {
        console.error("Rating failed:", err);
        alert("Connection error. Check your server logs.");
    }
}

function handleSessionExpiry() {
    localStorage.removeItem('sb_token');
    alert(translate('common.session.expired'));
    window.location.href = "/";
}

function logout() {
    localStorage.removeItem('sb_token');
    window.location.href = "/";
}