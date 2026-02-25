async function auth(action, credentials = {}) {
    const { email, password } = credentials;

    if (!email || !password) {
        return { success: false, tone: 'error', message: 'Email and password are required.' };
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
                    message: 'Sign up successful! Check your inbox for verification.'
                };
            }

            localStorage.setItem('sb_token', data.access_token);

            return {
                success: true,
                tone: 'info',
                message: 'Login successful! Redirecting...'
            };
        }

        return {
            success: false,
            tone: 'error',
            message: data.error_description || data.error || 'Check your credentials.'
        };
    } catch (err) {
        return {
            success: false,
            tone: 'error',
            message: `Connection failed: ${err.message}`
        };
    }
}

async function requestPasswordReset(email) {
    if (!email) {
        return { success: false, tone: 'error', message: 'Please enter the email you registered with.' };
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
                message: 'If this email exists, a reset link is on its way.'
            };
        }

        const data = await response.json().catch(() => ({}));
        return {
            success: false,
            tone: 'error',
            message: data.error_description || data.error || 'Unable to request a reset right now.'
        };
    } catch (err) {
        return {
            success: false,
            tone: 'error',
            message: `Reset request failed: ${err.message}`
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
    alert("Your session has expired. Please login again.");
    window.location.href = "/";
}

function logout() {
    localStorage.removeItem('sb_token');
    window.location.href = "/";
}