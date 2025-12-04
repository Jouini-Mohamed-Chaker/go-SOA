// Configuration
const API_BASE_URL = 'http://localhost';
const BOOK_SERVICE = `${API_BASE_URL}:8081/api`;
const USER_SERVICE = `${API_BASE_URL}:8082/api`;
const LOAN_SERVICE = `${API_BASE_URL}:8083/loan`;

// State
let currentPage = {
    books: 1,
    users: 1
};

let isEditMode = {
    book: false,
    user: false
};

// Toast notification
function showToast(message, type = 'success') {
    const toast = document.getElementById('toast');
    const toastMessage = document.getElementById('toastMessage');
    
    toastMessage.textContent = message;
    toast.classList.remove('hidden', 'bg-gray-800', 'bg-red-600', 'bg-green-600');
    
    if (type === 'error') {
        toast.classList.add('bg-red-600');
    } else if (type === 'success') {
        toast.classList.add('bg-green-600');
    } else {
        toast.classList.add('bg-gray-800');
    }
    
    setTimeout(() => {
        toast.classList.add('hidden');
    }, 3000);
}

// Tab Navigation
function setupTabs() {
    const tabs = {
        booksTab: 'booksSection',
        usersTab: 'usersSection',
        loansTab: 'loansSection'
    };

    Object.keys(tabs).forEach(tabId => {
        document.getElementById(tabId).addEventListener('click', () => {
            // Remove active class from all tabs
            document.querySelectorAll('.tab-btn').forEach(btn => {
                btn.classList.remove('active', 'text-indigo-600', 'border-b-2', 'border-indigo-600');
                btn.classList.add('text-gray-600');
            });

            // Hide all sections
            document.querySelectorAll('.tab-content').forEach(section => {
                section.classList.add('hidden');
            });

            // Activate clicked tab
            const clickedTab = document.getElementById(tabId);
            clickedTab.classList.add('active', 'text-indigo-600', 'border-b-2', 'border-indigo-600');
            clickedTab.classList.remove('text-gray-600');

            // Show corresponding section
            document.getElementById(tabs[tabId]).classList.remove('hidden');

            // Load data
            if (tabId === 'booksTab') loadBooks();
            if (tabId === 'usersTab') loadUsers();
            if (tabId === 'loansTab') loadLoans();
        });
    });
}

// ==================== BOOKS ====================

async function loadBooks(page = 1, searchTitle = null) {
    try {
        let url = searchTitle 
            ? `${BOOK_SERVICE}/books/search?title=${encodeURIComponent(searchTitle)}&page=${page}&limit=10`
            : `${BOOK_SERVICE}/books?page=${page}&limit=10`;

        const response = await fetch(url);
        const data = await response.json();
        
        currentPage.books = data.page || page;
        renderBooks(data.data);
        
        // Update pagination info
        document.getElementById('booksPageInfo').textContent = `Page ${currentPage.books}`;
        
    } catch (error) {
        showToast('Error loading books: ' + error.message, 'error');
    }
}

function renderBooks(books) {
    const booksList = document.getElementById('booksList');
    
    if (!books || books.length === 0) {
        booksList.innerHTML = '<p class="text-gray-500 text-center py-4">No books found</p>';
        return;
    }
    
    booksList.innerHTML = books.map(book => `
        <div class="border rounded-lg p-4 hover:shadow-md transition">
            <div class="flex justify-between items-start">
                <div class="flex-1">
                    <h3 class="font-bold text-lg text-gray-800">${book.title}</h3>
                    <p class="text-gray-600 text-sm">by ${book.author}</p>
                    <div class="mt-2 space-y-1 text-sm">
                        <p><span class="font-semibold">ISBN:</span> ${book.isbn}</p>
                        <p><span class="font-semibold">Category:</span> ${book.category}</p>
                        <p><span class="font-semibold">Year:</span> ${book.publishYear}</p>
                        <p class="flex items-center">
                            <span class="font-semibold">Available:</span> 
                            <span class="ml-2 px-2 py-1 rounded ${book.availableQuantity > 0 ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
                                ${book.availableQuantity}
                            </span>
                        </p>
                    </div>
                </div>
                <div class="flex gap-2 ml-4">
                    <button onclick="editBook(${book.id})" class="text-blue-600 hover:text-blue-800">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button onclick="deleteBook(${book.id})" class="text-red-600 hover:text-red-800">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        </div>
    `).join('');
}

async function editBook(id) {
    try {
        const response = await fetch(`${BOOK_SERVICE}/books/${id}`);
        const book = await response.json();
        
        document.getElementById('bookId').value = book.id;
        document.getElementById('bookIsbn').value = book.isbn;
        document.getElementById('bookTitle').value = book.title;
        document.getElementById('bookAuthor').value = book.author;
        document.getElementById('bookYear').value = book.publishYear;
        document.getElementById('bookCategory').value = book.category;
        document.getElementById('bookQuantity').value = book.availableQuantity;
        
        document.getElementById('bookFormTitle').textContent = 'Edit Book';
        isEditMode.book = true;
        
    } catch (error) {
        showToast('Error loading book: ' + error.message, 'error');
    }
}

async function deleteBook(id) {
    if (!confirm('Are you sure you want to delete this book?')) return;
    
    try {
        await fetch(`${BOOK_SERVICE}/books/${id}`, {
            method: 'DELETE'
        });
        showToast('Book deleted successfully');
        loadBooks(currentPage.books);
    } catch (error) {
        showToast('Error deleting book: ' + error.message, 'error');
    }
}

document.getElementById('bookForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const bookData = {
        isbn: document.getElementById('bookIsbn').value,
        title: document.getElementById('bookTitle').value,
        author: document.getElementById('bookAuthor').value,
        publishYear: parseInt(document.getElementById('bookYear').value),
        category: document.getElementById('bookCategory').value,
        availableQuantity: parseInt(document.getElementById('bookQuantity').value)
    };
    
    try {
        const bookId = document.getElementById('bookId').value;
        const url = bookId ? `${BOOK_SERVICE}/books/${bookId}` : `${BOOK_SERVICE}/books`;
        const method = bookId ? 'PUT' : 'POST';
        
        await fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(bookData)
        });
        
        showToast(bookId ? 'Book updated successfully' : 'Book created successfully');
        document.getElementById('bookForm').reset();
        document.getElementById('bookId').value = '';
        document.getElementById('bookFormTitle').textContent = 'Add New Book';
        isEditMode.book = false;
        loadBooks(currentPage.books);
        
    } catch (error) {
        showToast('Error saving book: ' + error.message, 'error');
    }
});

document.getElementById('cancelBookBtn').addEventListener('click', () => {
    document.getElementById('bookForm').reset();
    document.getElementById('bookId').value = '';
    document.getElementById('bookFormTitle').textContent = 'Add New Book';
    isEditMode.book = false;
});

document.getElementById('addBookBtn').addEventListener('click', () => {
    document.getElementById('bookForm').reset();
    document.getElementById('bookId').value = '';
    document.getElementById('bookFormTitle').textContent = 'Add New Book';
    isEditMode.book = false;
});

// Search functionality
document.getElementById('bookSearchBtn').addEventListener('click', () => {
    const searchTitle = document.getElementById('bookSearchInput').value.trim();
    if (searchTitle) {
        loadBooks(1, searchTitle);
    }
});

document.getElementById('bookSearchInput').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        const searchTitle = document.getElementById('bookSearchInput').value.trim();
        if (searchTitle) {
            loadBooks(1, searchTitle);
        }
    }
});

document.getElementById('bookClearBtn').addEventListener('click', () => {
    document.getElementById('bookSearchInput').value = '';
    loadBooks(1);
});

// Pagination
document.getElementById('booksPrevPage').addEventListener('click', () => {
    if (currentPage.books > 1) {
        const searchTitle = document.getElementById('bookSearchInput').value.trim();
        loadBooks(currentPage.books - 1, searchTitle || null);
    }
});

document.getElementById('booksNextPage').addEventListener('click', () => {
    const searchTitle = document.getElementById('bookSearchInput').value.trim();
    loadBooks(currentPage.books + 1, searchTitle || null);
});

// ==================== USERS ====================

async function loadUsers(page = 1) {
    try {
        const response = await fetch(`${USER_SERVICE}/users?page=${page}&limit=10`);
        const data = await response.json();
        
        currentPage.users = data.page || page;
        renderUsers(data.data);
        
        document.getElementById('usersPageInfo').textContent = `Page ${currentPage.users}`;
        
    } catch (error) {
        showToast('Error loading users: ' + error.message, 'error');
    }
}

function renderUsers(users) {
    const usersList = document.getElementById('usersList');
    
    if (!users || users.length === 0) {
        usersList.innerHTML = '<p class="text-gray-500 text-center py-4">No users found</p>';
        return;
    }
    
    usersList.innerHTML = users.map(user => `
        <div class="border rounded-lg p-4 hover:shadow-md transition">
            <div class="flex justify-between items-start">
                <div class="flex-1">
                    <h3 class="font-bold text-lg text-gray-800">${user.firstName} ${user.lastName}</h3>
                    <p class="text-gray-600 text-sm">@${user.username}</p>
                    <div class="mt-2 text-sm">
                        <p><i class="fas fa-envelope mr-2 text-gray-500"></i>${user.email}</p>
                    </div>
                </div>
                <div class="flex gap-2 ml-4">
                    <button onclick="editUser(${user.id})" class="text-blue-600 hover:text-blue-800">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button onclick="deleteUser(${user.id})" class="text-red-600 hover:text-red-800">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        </div>
    `).join('');
}

async function editUser(id) {
    try {
        const response = await fetch(`${USER_SERVICE}/users/${id}`);
        const user = await response.json();
        
        document.getElementById('userId').value = user.id;
        document.getElementById('userUsername').value = user.username;
        document.getElementById('userEmail').value = user.email;
        document.getElementById('userFirstName').value = user.firstName;
        document.getElementById('userLastName').value = user.lastName;
        
        document.getElementById('userFormTitle').textContent = 'Edit User';
        isEditMode.user = true;
        
    } catch (error) {
        showToast('Error loading user: ' + error.message, 'error');
    }
}

async function deleteUser(id) {
    if (!confirm('Are you sure you want to delete this user?')) return;
    
    try {
        await fetch(`${USER_SERVICE}/users/${id}`, {
            method: 'DELETE'
        });
        showToast('User deleted successfully');
        loadUsers(currentPage.users);
    } catch (error) {
        showToast('Error deleting user: ' + error.message, 'error');
    }
}

document.getElementById('userForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const userData = {
        username: document.getElementById('userUsername').value,
        email: document.getElementById('userEmail').value,
        firstName: document.getElementById('userFirstName').value,
        lastName: document.getElementById('userLastName').value
    };
    
    try {
        const userId = document.getElementById('userId').value;
        const url = userId ? `${USER_SERVICE}/users/${userId}` : `${USER_SERVICE}/users`;
        const method = userId ? 'PUT' : 'POST';
        
        await fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(userData)
        });
        
        showToast(userId ? 'User updated successfully' : 'User created successfully');
        document.getElementById('userForm').reset();
        document.getElementById('userId').value = '';
        document.getElementById('userFormTitle').textContent = 'Add New User';
        isEditMode.user = false;
        loadUsers(currentPage.users);
        
    } catch (error) {
        showToast('Error saving user: ' + error.message, 'error');
    }
});

document.getElementById('cancelUserBtn').addEventListener('click', () => {
    document.getElementById('userForm').reset();
    document.getElementById('userId').value = '';
    document.getElementById('userFormTitle').textContent = 'Add New User';
    isEditMode.user = false;
});

document.getElementById('addUserBtn').addEventListener('click', () => {
    document.getElementById('userForm').reset();
    document.getElementById('userId').value = '';
    document.getElementById('userFormTitle').textContent = 'Add New User';
    isEditMode.user = false;
});

// Pagination
document.getElementById('usersPrevPage').addEventListener('click', () => {
    if (currentPage.users > 1) {
        loadUsers(currentPage.users - 1);
    }
});

document.getElementById('usersNextPage').addEventListener('click', () => {
    loadUsers(currentPage.users + 1);
});

// ==================== LOANS (SOAP) ====================

async function loadLoans() {
    try {
        const soapRequest = `
            <?xml version="1.0" encoding="utf-8"?>
            <soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
                <soap:Body>
                    <getAllLoans xmlns="http://loan.service"/>
                </soap:Body>
            </soap:Envelope>
        `;
        
        const response = await fetch(LOAN_SERVICE, {
            method: 'POST',
            headers: {
                'Content-Type': 'text/xml',
                'SOAPAction': 'getAllLoans'
            },
            body: soapRequest
        });
        
        const text = await response.text();
        const parser = new DOMParser();
        const xml = parser.parseFromString(text, 'text/xml');
        
        // Parse SOAP response
        const loans = [];
        const loanNodes = xml.getElementsByTagName('loan');
        
        for (let i = 0; i < loanNodes.length; i++) {
            const loan = {
                id: loanNodes[i].getElementsByTagName('id')[0]?.textContent,
                userId: loanNodes[i].getElementsByTagName('userId')[0]?.textContent,
                bookId: loanNodes[i].getElementsByTagName('bookId')[0]?.textContent,
                loanDate: loanNodes[i].getElementsByTagName('loanDate')[0]?.textContent,
                dueDate: loanNodes[i].getElementsByTagName('dueDate')[0]?.textContent,
                returnDate: loanNodes[i].getElementsByTagName('returnDate')[0]?.textContent,
                status: loanNodes[i].getElementsByTagName('status')[0]?.textContent
            };
            loans.push(loan);
        }
        
        renderLoans(loans);
        
    } catch (error) {
        showToast('Error loading loans: ' + error.message, 'error');
        console.error('SOAP Error:', error);
    }
}

function renderLoans(loans) {
    const loansList = document.getElementById('loansList');
    
    if (!loans || loans.length === 0) {
        loansList.innerHTML = '<p class="text-gray-500 text-center py-4">No loans found</p>';
        return;
    }
    
    loansList.innerHTML = loans.map(loan => {
        const isActive = loan.status === 'ACTIVE';
        const loanDate = new Date(loan.loanDate).toLocaleDateString();
        const dueDate = new Date(loan.dueDate).toLocaleDateString();
        const returnDate = loan.returnDate ? new Date(loan.returnDate).toLocaleDateString() : 'Not returned';
        
        return `
            <div class="border rounded-lg p-4 hover:shadow-md transition">
                <div class="flex justify-between items-start">
                    <div class="flex-1">
                        <div class="flex items-center gap-2 mb-2">
                            <h3 class="font-bold text-lg text-gray-800">Loan #${loan.id}</h3>
                            <span class="px-2 py-1 text-xs rounded ${isActive ? 'bg-yellow-100 text-yellow-800' : 'bg-green-100 text-green-800'}">
                                ${loan.status}
                            </span>
                        </div>
                        <div class="space-y-1 text-sm">
                            <p><span class="font-semibold">User ID:</span> ${loan.userId}</p>
                            <p><span class="font-semibold">Book ID:</span> ${loan.bookId}</p>
                            <p><span class="font-semibold">Loan Date:</span> ${loanDate}</p>
                            <p><span class="font-semibold">Due Date:</span> ${dueDate}</p>
                            <p><span class="font-semibold">Return Date:</span> ${returnDate}</p>
                        </div>
                    </div>
                    ${isActive ? `
                        <button onclick="returnLoan(${loan.id})" class="ml-4 bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
                            Return
                        </button>
                    ` : ''}
                </div>
            </div>
        `;
    }).join('');
}

async function returnLoan(loanId) {
    if (!confirm('Mark this loan as returned?')) return;
    
    try {
        const soapRequest = `
            <?xml version="1.0" encoding="utf-8"?>
            <soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
                <soap:Body>
                    <returnLoan xmlns="http://loan.service">
                        <loanId>${loanId}</loanId>
                    </returnLoan>
                </soap:Body>
            </soap:Envelope>
        `;
        
        await fetch(LOAN_SERVICE, {
            method: 'POST',
            headers: {
                'Content-Type': 'text/xml',
                'SOAPAction': 'returnLoan'
            },
            body: soapRequest
        });
        
        showToast('Loan returned successfully');
        loadLoans();
        
    } catch (error) {
        showToast('Error returning loan: ' + error.message, 'error');
    }
}

document.getElementById('loanForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const userId = document.getElementById('loanUserId').value;
    const bookId = document.getElementById('loanBookId').value;
    
    try {
        const soapRequest = `
            <?xml version="1.0" encoding="utf-8"?>
            <soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
                <soap:Body>
                    <createLoan xmlns="http://loan.service">
                        <userId>${userId}</userId>
                        <bookId>${bookId}</bookId>
                    </createLoan>
                </soap:Body>
            </soap:Envelope>
        `;
        
        const response = await fetch(LOAN_SERVICE, {
            method: 'POST',
            headers: {
                'Content-Type': 'text/xml',
                'SOAPAction': 'createLoan'
            },
            body: soapRequest
        });
        
        const text = await response.text();
        const parser = new DOMParser();
        const xml = parser.parseFromString(text, 'text/xml');
        
        const success = xml.getElementsByTagName('success')[0]?.textContent === 'true';
        const message = xml.getElementsByTagName('message')[0]?.textContent;
        
        if (success) {
            showToast('Loan created successfully');
            document.getElementById('loanForm').reset();
            loadLoans();
        } else {
            showToast(message || 'Error creating loan', 'error');
        }
        
    } catch (error) {
        showToast('Error creating loan: ' + error.message, 'error');
    }
});

document.getElementById('addLoanBtn').addEventListener('click', () => {
    document.getElementById('loanForm').reset();
});

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    setupTabs();
    loadBooks();
});