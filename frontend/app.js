const API_URL = 'http://localhost:8080';


const sections = {
    auth: document.getElementById('auth-section'),
    menu: document.getElementById('menu-section'),
    cart: document.getElementById('cart-section'),
    orders: document.getElementById('orders-section')
};

const navBtns = {
    menu: document.getElementById('nav-menu'),
    orders: document.getElementById('nav-orders'),
    cart: document.getElementById('nav-cart'),
    login: document.getElementById('nav-login')
};


let token = localStorage.getItem('token');
let cart = JSON.parse(localStorage.getItem('cart') || '[]');
let currentRestaurantId = '1'; 


function init() {
    updateNav();
    updateCartBadge();
    if (token) {
        showSection('menu');
        loadMenu();
    } else {
        showSection('auth');
    }
}


function showSection(sectionName) {
    Object.values(sections).forEach(s => s.classList.add('hidden'));
    sections[sectionName].classList.remove('hidden');
    
    if (sectionName === 'menu') loadMenu();
    if (sectionName === 'orders') loadOrders();
    if (sectionName === 'cart') renderCart();
}

function updateNav() {
    if (token) {
        navBtns.login.textContent = 'Logout';
        navBtns.orders.classList.remove('hidden');
    } else {
        navBtns.login.textContent = 'Login';
        navBtns.orders.classList.add('hidden');
    }
}


function showToast(msg, type = 'success') {
    const toast = document.getElementById('toast');
    toast.textContent = msg;
    toast.className = `toast ${type}`;
    toast.classList.remove('hidden');
    setTimeout(() => toast.classList.add('hidden'), 3000);
}


document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', (e) => {
        document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.auth-form').forEach(f => f.classList.remove('active'));
        e.target.classList.add('active');
        document.getElementById(`${e.target.dataset.target}-form`).classList.add('active');
    });
});

document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    try {
        const res = await fetch(`${API_URL}/auth/login`, {
            method: 'POST',
            body: JSON.stringify({
                email: document.getElementById('l-email').value,
                password: document.getElementById('l-password').value
            })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.message || 'Login failed');
        
        token = data.access_token;
        localStorage.setItem('token', token);
        showToast('Login successful!');
        init();
    } catch (err) {
        showToast(err.message, 'error');
    }
});

document.getElementById('register-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    try {
        const res = await fetch(`${API_URL}/auth/register`, {
            method: 'POST',
            body: JSON.stringify({
                name: document.getElementById('r-name').value,
                email: document.getElementById('r-email').value,
                password: document.getElementById('r-password').value,
                phone: document.getElementById('r-phone').value,
                role: 'user'
            })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.message || 'Registration failed');
        
        token = data.access_token;
        localStorage.setItem('token', token);
        showToast('Welcome to AITUdel!');
        init();
    } catch (err) {
        showToast(err.message, 'error');
    }
});

navBtns.login.addEventListener('click', () => {
    if (token) {
        localStorage.removeItem('token');
        token = null;
        init();
        showToast('Logged out');
    } else {
        showSection('auth');
    }
});


navBtns.menu.addEventListener('click', () => showSection('menu'));
document.getElementById('btn-search').addEventListener('click', loadMenu);

async function loadMenu() {
    const q = document.getElementById('search-input').value;
    try {
        const res = await fetch(`${API_URL}/items/search?query=${q}`);
        const data = await res.json();
        const grid = document.getElementById('menu-grid');
        grid.innerHTML = '';
        
        if (data.items) {
            data.items.forEach(item => {
                grid.innerHTML += `
                    <div class="item-card">
                        <div class="item-name">${item.name}</div>
                        <div class="item-desc">${item.description}</div>
                        <div class="item-footer">
                            <div class="item-price">${item.price} ₸</div>
                            <button class="btn-secondary" onclick="addToCart('${item.id}', '${item.name}', ${item.price})">Add</button>
                        </div>
                    </div>
                `;
            });
        }
    } catch (err) {
        console.error('Failed to load menu', err);
    }
}


navBtns.cart.addEventListener('click', () => showSection('cart'));

function addToCart(id, name, price) {
    const existing = cart.find(i => i.item_id === id);
    if (existing) {
        existing.quantity += 1;
    } else {
        cart.push({ item_id: id, name, price, quantity: 1 });
    }
    saveCart();
    showToast(`${name} added to cart`);
}

function saveCart() {
    localStorage.setItem('cart', JSON.stringify(cart));
    updateCartBadge();
    if (!sections.cart.classList.contains('hidden')) renderCart();
}

function updateCartBadge() {
    document.getElementById('cart-count').textContent = cart.reduce((sum, i) => sum + i.quantity, 0);
}

function renderCart() {
    const list = document.getElementById('cart-items');
    list.innerHTML = '';
    let total = 0;
    
    cart.forEach((item, index) => {
        total += item.price * item.quantity;
        list.innerHTML += `
            <div class="cart-item">
                <div>
                    <strong>${item.name}</strong> x${item.quantity}
                </div>
                <div>
                    ${item.price * item.quantity} ₸
                    <button class="btn-secondary" style="padding: 0.2rem 0.5rem; margin-left:1rem" onclick="removeFromCart(${index})">X</button>
                </div>
            </div>
        `;
    });
    
    document.getElementById('cart-total-price').textContent = `${total} ₸`;
}

window.removeFromCart = (index) => {
    cart.splice(index, 1);
    saveCart();
};

document.getElementById('btn-checkout').addEventListener('click', async () => {
    if (!token) return showSection('auth');
    if (cart.length === 0) return showToast('Cart is empty', 'error');
    
    const address = document.getElementById('delivery-address').value;
    const promo = document.getElementById('promo-input').value;
    
    if (!address) return showToast('Please enter delivery address', 'error');

    try {
        const res = await fetch(`${API_URL}/orders`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                restaurant_id: currentRestaurantId,
                delivery_address: address,
                items: cart,
                promo_code: promo
            })
        });
        
        const data = await res.json();
        if (!res.ok) throw new Error(data.message || 'Checkout failed');
        
        cart = [];
        saveCart();
        showToast('Order placed successfully!');
        showSection('orders');
    } catch (err) {
        showToast(err.message, 'error');
    }
});


navBtns.orders.addEventListener('click', () => showSection('orders'));

async function loadOrders() {
    if (!token) return;
    try {
        const res = await fetch(`${API_URL}/orders`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        const data = await res.json();
        const grid = document.getElementById('orders-grid');
        grid.innerHTML = '';
        
        if (data.orders) {
            data.orders.forEach(o => {
                grid.innerHTML += `
                    <div class="item-card">
                        <div class="item-name">Order #${o.id.substring(0,8)}</div>
                        <div class="item-desc">Status: <strong style="color: var(--primary)">${o.status}</strong></div>
                        <div class="item-footer">
                            <div class="item-price">${o.total_price} ₸</div>
                        </div>
                    </div>
                `;
            });
        }
    } catch (err) {
        console.error(err);
    }
}

init();
