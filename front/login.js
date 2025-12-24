const fetchOptions = { credentials: 'same-origin' };

document.getElementById('loginSection').addEventListener('keydown', enterLogin);
document.getElementById('registerBtn').addEventListener('click', register);
document.getElementById('loginBtn').addEventListener('click', login);
const API_BASE_URL = location.origin;

async function enterLogin(event) {
    if (event.code === 'Enter') {
        await login()
    }
}

async function login()  {
    const username = document.getElementById('login-username').value;
    localStorage.setItem('username', username);
    const password = document.getElementById('login-password').value;
    const formData = new FormData();
    formData.append('username', username);
    formData.append('password', password);
    try {
        const response = await fetch(`${API_BASE_URL}/login`, { ...fetchOptions, method: 'POST', body: formData });
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        window.location.replace("./main.html");
    } catch (error) {
        document.getElementById('loginError').innerHTML = error.message;
    }
}

async function register() {
    const username = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;
    const formData = new FormData();
    formData.append('username', username);
    formData.append('password', password);
    try {
        const response = await fetch(`${API_BASE_URL}/register`, {
            ...fetchOptions,
            method: 'POST',
            body: formData,
        })
        if (!(response.ok)) throw new Error(`Error: ${await response.text()}`);
        document.getElementById('loginError').style.color = 'green';
        document.getElementById('loginError').innerHTML = 'User Registered'
    }catch (error) {
        document.getElementById('loginError').innerHTML = error.message;
    }
}

window.onload = () => {
    fetch(`${location.origin}/login`, {
        ...fetchOptions,
        method: 'GET',
    }).then(response => {
        if (response.ok) {
            window.location.replace("./main.html");
        }
    })
}