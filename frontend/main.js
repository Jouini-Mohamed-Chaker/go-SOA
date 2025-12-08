// Global state
const state = {
    token: localStorage.getItem('authToken') || '',
    username: localStorage.getItem('username') || '',
    baseUrl: 'http://localhost:8080'
};

// DOM Elements
const authStatus = document.getElementById('authStatus');
const authStatusText = document.getElementById('authStatusText');
const userInfo = document.getElementById('userInfo');
const userDetails = document.getElementById('userDetails');
const loginToken = document.getElementById('loginToken');
const tokenValue = document.getElementById('tokenValue');
const apiParams = document.getElementById('apiParams');
const apiResponse = document.getElementById('apiResponse');
const notification = document.getElementById('notification');

// Initialize
document.addEventListener('DOMContentLoaded', function () {
    initTabs();
    initAuth();
    initApiEndpointSelector();
    initEventListeners();

    // If we have a token, validate it on load
    if (state.token) {
        validateToken(state.token);
    }
});

// Tab functionality
function initTabs() {
    const tabs = document.querySelectorAll('.tab');
    tabs.forEach(tab => {
        tab.addEventListener('click', function () {
            const tabId = this.getAttribute('data-tab');

            // Update active tab
            tabs.forEach(t => t.classList.remove('active'));
            this.classList.add('active');

            // Show corresponding content
            document.querySelectorAll('.tab-content').forEach(content => {
                content.classList.remove('active');
            });
            document.getElementById(tabId).classList.add('active');
        });
    });
}

// Authentication state
function initAuth() {
    updateAuthStatus();
}

function updateAuthStatus() {
    if (state.token && state.username) {
        authStatus.classList.add('active');
        authStatusText.textContent = `Authenticated as ${state.username}`;
        userInfo.classList.remove('hidden');
        userDetails.innerHTML = `
                    <div>
                        <strong>${state.username}</strong>
                        <button id="logoutBtn" style="margin-left: 10px; padding: 5px 10px; font-size: 14px;">Logout</button>
                    </div>
                `;

        // Add logout event listener
        document.getElementById('logoutBtn').addEventListener('click', logout);
    } else {
        authStatus.classList.remove('active');
        authStatusText.textContent = 'Not authenticated';
        userInfo.classList.add('hidden');
    }
}

// API endpoint selector
function initApiEndpointSelector() {
    const endpointSelect = document.getElementById('apiEndpoint');
    endpointSelect.addEventListener('change', updateApiParams);
    updateApiParams(); // Initial call
}

function updateApiParams() {
    const endpoint = document.getElementById('apiEndpoint').value;
    let paramsHTML = '';

    switch (endpoint) {
        case 'books-search':
            paramsHTML = `
                        <div class="form-group">
                            <label for="searchTitle">Title</label>
                            <input type="text" id="searchTitle" placeholder="Enter book title">
                        </div>
                    `;
            break;

        case 'books-create':
            paramsHTML = `
                        <div class="form-group">
                            <label for="bookTitle">Title *</label>
                            <input type="text" id="bookTitle" placeholder="Enter book title" required>
                        </div>
                        <div class="form-group">
                            <label for="bookAuthor">Author</label>
                            <input type="text" id="bookAuthor" placeholder="Enter author name">
                        </div>
                        <div class="form-group">
                            <label for="bookIsbn">ISBN</label>
                            <input type="text" id="bookIsbn" placeholder="Enter ISBN">
                        </div>
                    `;
            break;

        case 'users-create':
            paramsHTML = `
                        <div class="form-group">
                            <label for="userUsername">Username *</label>
                            <input type="text" id="userUsername" placeholder="Enter username" required>
                        </div>
                        <div class="form-group">
                            <label for="userEmail">Email *</label>
                            <input type="email" id="userEmail" placeholder="Enter email" required>
                        </div>
                        <div class="form-group">
                            <label for="userFirstName">First Name</label>
                            <input type="text" id="userFirstName" placeholder="Enter first name">
                        </div>
                        <div class="form-group">
                            <label for="userLastName">Last Name</label>
                            <input type="text" id="userLastName" placeholder="Enter last name">
                        </div>
                    `;
            break;

        case 'loans-create':
            paramsHTML = `
                        <div class="form-group">
                            <label for="loanUserId">User ID *</label>
                            <input type="number" id="loanUserId" placeholder="Enter user ID" required>
                        </div>
                        <div class="form-group">
                            <label for="loanBookId">Book ID *</label>
                            <input type="number" id="loanBookId" placeholder="Enter book ID" required>
                        </div>
                    `;
            break;

        case 'loans-return':
            paramsHTML = `
                        <div class="form-group">
                            <label for="returnLoanId">Loan ID *</label>
                            <input type="number" id="returnLoanId" placeholder="Enter loan ID" required>
                        </div>
                    `;
            break;

        case 'loans-user':
            paramsHTML = `
                        <div class="form-group">
                            <label for="userLoansId">User ID *</label>
                            <input type="number" id="userLoansId" placeholder="Enter user ID" required>
                        </div>
                    `;
            break;

        default:
            paramsHTML = '<p>No parameters required for this endpoint.</p>';
    }

    apiParams.innerHTML = paramsHTML;
}

// Event listeners
function initEventListeners() {
    // Login
    document.getElementById('loginBtn').addEventListener('click', login);

    // Register
    document.getElementById('registerBtn').addEventListener('click', register);

    // Validate token
    document.getElementById('validateBtn').addEventListener('click', function () {
        const token = document.getElementById('validateToken').value;
        if (token) {
            validateToken(token);
        } else {
            showNotification('Please enter a token to validate', 'error');
        }
    });

    // Execute API
    document.getElementById('executeApi').addEventListener('click', executeApi);

    // Allow Enter key in login form
    document.getElementById('loginPassword').addEventListener('keypress', function (e) {
        if (e.key === 'Enter') {
            login();
        }
    });
}

// API Functions
async function makeRequest(url, method = 'GET', body = null, requiresAuth = false) {
    const headers = {
        'Content-Type': 'application/json'
    };

    if (requiresAuth && state.token) {
        headers['Authorization'] = `Bearer ${state.token}`;
    }

    const options = {
        method,
        headers
    };

    if (body) {
        options.body = JSON.stringify(body);
    }

    try {
        const response = await fetch(`${state.baseUrl}${url}`, options);
        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.message || `HTTP ${response.status}`);
        }

        return { success: true, data, status: response.status };
    } catch (error) {
        return {
            success: false,
            error: error.message,
            status: 0
        };
    }
}

// Auth Functions
async function login() {
    const username = document.getElementById('loginUsername').value;
    const password = document.getElementById('loginPassword').value;

    if (!username || !password) {
        showNotification('Please enter username and password', 'error');
        return;
    }

    const result = await makeRequest('/auth/login', 'POST', { username, password });

    if (result.success) {
        state.token = result.data.token;
        state.username = result.data.username;

        // Save to localStorage
        localStorage.setItem('authToken', state.token);
        localStorage.setItem('username', state.username);

        // Update UI
        tokenValue.textContent = state.token;
        loginToken.classList.remove('hidden');
        updateAuthStatus();

        showNotification('Login successful! Token saved.', 'success');

        // Switch to validate tab to show the token
        document.querySelector('.tab[data-tab="validate"]').click();
        document.getElementById('validateToken').value = state.token;
    } else {
        showNotification(`Login failed: ${result.error}`, 'error');
    }
}

async function register() {
    const username = document.getElementById('registerUsername').value;
    const email = document.getElementById('registerEmail').value;
    const firstName = document.getElementById('registerFirstName').value;
    const lastName = document.getElementById('registerLastName').value;
    const password = document.getElementById('registerPassword').value;

    if (!username || !email || !password) {
        showNotification('Please fill in all required fields', 'error');
        return;
    }

    const userData = {
        username,
        email,
        firstName,
        lastName,
        password
    };

    const result = await makeRequest('/auth/register', 'POST', userData);

    if (result.success) {
        showNotification(`User ${username} registered successfully!`, 'success');

        // Clear form
        document.getElementById('registerUsername').value = '';
        document.getElementById('registerEmail').value = '';
        document.getElementById('registerFirstName').value = '';
        document.getElementById('registerLastName').value = '';
        document.getElementById('registerPassword').value = '';

        // Switch to login tab with prefilled username
        document.querySelector('.tab[data-tab="login"]').click();
        document.getElementById('loginUsername').value = username;
    } else {
        showNotification(`Registration failed: ${result.error}`, 'error');
    }
}

async function validateToken(token) {
    const result = await makeRequest('/auth/validate', 'POST', { token });

    if (result.success) {
        if (result.data.valid) {
            showNotification(`Token is valid for user: ${result.data.username}`, 'success');
            apiResponse.textContent = JSON.stringify(result.data, null, 2);
        } else {
            showNotification('Token is invalid', 'error');
            apiResponse.textContent = JSON.stringify(result.data, null, 2);
        }
    } else {
        showNotification(`Validation failed: ${result.error}`, 'error');
        apiResponse.textContent = result.error;
    }
}

function logout() {
    state.token = '';
    state.username = '';

    localStorage.removeItem('authToken');
    localStorage.removeItem('username');

    updateAuthStatus();
    loginToken.classList.add('hidden');

    showNotification('Logged out successfully', 'success');
}

// API Execution
async function executeApi() {
    if (!state.token) {
        showNotification('You need to login first to use protected endpoints', 'error');
        return;
    }

    const endpoint = document.getElementById('apiEndpoint').value;
    let url = '';
    let method = 'GET';
    let body = null;

    // Build request based on selected endpoint
    switch (endpoint) {
        case 'books':
            url = '/api/books';
            break;

        case 'books-search':
            const title = document.getElementById('searchTitle').value;
            url = `/api/books/search?title=${encodeURIComponent(title || '')}`;
            break;

        case 'books-create':
            url = '/api/books';
            method = 'POST';
            body = {
                title: document.getElementById('bookTitle').value || 'Sample Book',
                author: document.getElementById('bookAuthor').value || 'Unknown Author',
                isbn: document.getElementById('bookIsbn').value || '1234567890'
            };
            break;

        case 'users':
            url = '/api/users';
            break;

        case 'users-create':
            url = '/api/users';
            method = 'POST';
            body = {
                username: document.getElementById('userUsername').value || `user_${Date.now()}`,
                email: document.getElementById('userEmail').value || `user_${Date.now()}@example.com`,
                firstName: document.getElementById('userFirstName').value || 'John',
                lastName: document.getElementById('userLastName').value || 'Doe'
            };
            break;

        case 'loans':
            url = '/api/loans';
            break;

        case 'loans-create':
            url = '/api/loans';
            method = 'POST';
            body = {
                userId: parseInt(document.getElementById('loanUserId').value) || 1,
                bookId: parseInt(document.getElementById('loanBookId').value) || 1
            };
            break;

        case 'loans-return':
            const loanId = document.getElementById('returnLoanId').value;
            url = `/api/loans/${loanId}/return`;
            method = 'PUT';
            break;

        case 'loans-user':
            const userId = document.getElementById('userLoansId').value;
            url = `/api/loans/user/${userId}`;
            break;
    }

    apiResponse.textContent = 'Loading...';

    const result = await makeRequest(url, method, body, true);

    if (result.success) {
        apiResponse.textContent = JSON.stringify(result.data, null, 2);
        showNotification(`API call successful (Status: ${result.status})`, 'success');
    } else {
        apiResponse.textContent = `Error: ${result.error}\n\nMake sure you are authenticated and the API server is running at ${state.baseUrl}`;
        showNotification(`API call failed: ${result.error}`, 'error');
    }
}

// Utility Functions
function showNotification(message, type = 'success') {
    notification.textContent = message;
    notification.className = `notification ${type}`;
    notification.classList.remove('hidden');

    setTimeout(() => {
        notification.classList.add('hidden');
    }, 3000);
}

// Format JSON for display
function formatJson(json) {
    try {
        return JSON.stringify(JSON.parse(json), null, 2);
    } catch {
        return json;
    }
}
