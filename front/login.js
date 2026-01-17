import {ui} from "./modules/ui.js";

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
        const response = await fetch(`${API_BASE_URL}/login`, { ...fetchOptions, method: 'POST', body: formData });
        if (!response.ok) {
            if (response.status === 401) {
                ui.showNotification('Username or password not valid', 'warning');
            }else{
                ui.showNotification('Error: ' + response.status, 'error');
            }
        }else {

            window.location.replace("./main.html");
        }
}

async function register() {
    const username = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;
    const formData = new FormData();
    formData.append('username', username);
    formData.append('password', password);
        const response = await fetch(`${API_BASE_URL}/register`, {
            ...fetchOptions,
            method: 'POST',
            body: formData,
        })
        if (!response.ok) {
            if (response.status === 400) {
                ui.showNotification('Username or password not valid', 'warning');
            }
        }else{
            ui.showNotification(`${username} registered`, 'success');
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