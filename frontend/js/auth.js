function saveUser(data) {
  localStorage.setItem("access_token", data.access_token || "");
  localStorage.setItem("refresh_token", data.refresh_token || "");
  localStorage.setItem("user", JSON.stringify(data.user || data));
}

function getUser() {
  const raw = localStorage.getItem("user");
  try { return raw ? JSON.parse(raw) : null; } catch { return null; }
}

function getToken() {
  return localStorage.getItem("access_token");
}

function clearUser() {
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
  localStorage.removeItem("user");
}

function requireAuth() {
  if (!getUser()) {
    window.location.href = "./auth.html";
    return false;
  }
  return true;
}

function toast(msg, type = "info") {
  let container = document.getElementById("toast-container");
  if (!container) {
    container = document.createElement("div");
    container.id = "toast-container";
    container.className = "toast-container";
    document.body.appendChild(container);
  }
  const t = document.createElement("div");
  t.className = `toast ${type}`;
  t.textContent = msg;
  container.appendChild(t);
  setTimeout(() => t.remove(), 3500);
}

function updateNavCart() {
  const user = getUser();
  if (!user) return;
  const countEl = document.getElementById("cart-count");
  if (!countEl) return;
  API.getCart(user.id).then((data) => {
    const items = data.items || [];
    countEl.textContent = items.length;
    countEl.style.display = items.length > 0 ? "flex" : "none";
  }).catch(() => {});
}

function renderNav() {
  const user = getUser();
  const navEl = document.getElementById("navbar");
  if (!navEl) return;

  navEl.innerHTML = `
    <a class="navbar-logo" href="./index.html">
      🍔 <span>AITUdel</span>
    </a>
    <div class="navbar-search">
      <input type="text" id="nav-search" placeholder="Поиск блюд..." />
      <button onclick="doNavSearch()">🔍</button>
    </div>
    <div class="navbar-actions">
      ${user ? `
        <a href="./orders.html">Заказы</a>
        <a href="./profile.html">${user.name || "Профиль"}</a>
        <a href="./cart.html" class="cart-btn">
          🛒 <span id="cart-count" class="cart-count" style="display:none">0</span>
        </a>
      ` : `
        <a href="./auth.html">Войти</a>
        <a href="./auth.html" class="cart-btn">Регистрация</a>
      `}
    </div>
  `;

  if (user) updateNavCart();
}

function doNavSearch() {
  const q = document.getElementById("nav-search").value.trim();
  if (q) window.location.href = `./index.html?search=${encodeURIComponent(q)}`;
}