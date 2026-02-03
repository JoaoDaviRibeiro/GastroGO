async function auth(action) {
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const messageDiv = document.getElementById('message');

    messageDiv.innerText = "Processing...";

    try {
        const response = await fetch(`/api/${action}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password }),
        });

        const data = await response.json();

        if (response.ok) {
            if (action === 'signup') {
                messageDiv.style.color = "green";
                messageDiv.innerText = "Sign up successful! Check your email.";
            } else {
                // SUCCESSFUL LOGIN
                messageDiv.style.color = "blue";
                messageDiv.innerText = "Login successful! Redirecting...";
                
                // Store token for later use
                localStorage.setItem('sb_token', data.access_token);
                
                // 2. Redirect to the dashboard
                setTimeout(() => {
                    window.location.href = "/dashboard.html";
                }, 1000);
            }
        } else {
            messageDiv.style.color = "red";
            messageDiv.innerText = "Error: " + (data.error_description || data.error || "Check credentials");
        }
    } catch (err) {
        messageDiv.style.color = "red";
        messageDiv.innerText = "Connection failed: " + err.message;
    }
}

// Function to call the PROTECTED route
async function loadDashboard() {
    const token = localStorage.getItem('sb_token');
    const display = document.getElementById('dashboard-data');

    if (!token) {
        window.location.href = "/";
        return;
    }

    try {
        // 1. Fetch User Info (Dashboard Info)
        const userResponse = await fetch('/api/dashboard', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (userResponse.status === 401) return handleSessionExpiry();

        // 2. Fetch Restaurant Data
        const resResponse = await fetch('/api/restaurants', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (userResponse.ok && resResponse.ok) {
            const userData = await userResponse.json();
            const restaurants = await resResponse.json();

            let restaurantHTML = '<h3>🍽️ GastroGO Restaurants</h3><ul>';
            if (restaurants && restaurants.length > 0) {
                restaurants.forEach(res => {
                    restaurantHTML += `
                        <li>
                            <strong>${res.name}</strong> - ${res.cuisine}<br>
                            <small>${res.address} | Rating: ${res.rating} ⭐</small>
                        </li><br>`;
                });
            } else {
                restaurantHTML += '<li>No restaurants found. Add your first one!</li>';
            }
            restaurantHTML += '</ul>';

            display.innerHTML = `
                <p>Welcome back, <strong>${userData.email}</strong>!</p>
                <hr>
                ${restaurantHTML}
            `;
        } else {
            // Any other error (500, etc)
            handleSessionExpiry();
        }
    } catch (err) {
        console.error("Dashboard load failed:", err);
    }
}

// Clean up logic to prevent repetitive code
function handleSessionExpiry() {
    localStorage.removeItem('sb_token');
    alert("Your session has expired. Please login again.");
    window.location.href = "/";
}

function logout() {
    localStorage.removeItem('sb_token');
    window.location.href = "/";
}