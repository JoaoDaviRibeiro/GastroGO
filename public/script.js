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
                
                // 1. Store the token in localStorage
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
        window.location.href = "/"; // Kick back to login if no token
        return;
    }

    try {
        const response = await fetch('/api/dashboard', {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}` // 3. Send the token in the header
            }
        });

        if (response.ok) {
            const data = await response.json();
            display.innerHTML = `
                <p><strong>Status:</strong> ${data.message}</p>
                <p><strong>Logged in as:</strong> ${data.email}</p>
            `;
        } else {
            // Token likely expired or invalid
            localStorage.removeItem('sb_token');
            window.location.href = "/";
        }
    } catch (err) {
        console.error("Dashboard load failed:", err);
    }
}

function logout() {
    localStorage.removeItem('sb_token');
    window.location.href = "/";
}