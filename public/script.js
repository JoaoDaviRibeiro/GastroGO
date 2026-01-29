async function auth(action) {
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const messageDiv = document.getElementById('message');

    messageDiv.innerText = "Processing...";

    try {
        const response = await fetch(`/api/${action}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });

        const data = await response.json();

        if (response.ok) {
            if (action === 'signup') {
                messageDiv.style.color = "green";
                messageDiv.innerText = "Sign up successful! Check your email for a confirmation link.";
            } else {
                messageDiv.style.color = "blue";
                messageDiv.innerText = "Login successful! Token received.";
                console.log("Supabase Token:", data.access_token);
                // Pro-tip: In a real app, you'd save this token in localStorage or a Cookie
            }
        } else {
            messageDiv.style.color = "red";
            messageDiv.innerText = "Error: " + (data.error || "Something went wrong");
        }
    } catch (err) {
        messageDiv.style.color = "red";
        messageDiv.innerText = "Connection failed: " + err.message;
    }
}