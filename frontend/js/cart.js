let cartItems = [];
let cartTotal = 0;
let appliedDiscount = 0;
let appliedPromo = "";

const ITEM_EMOJIS = ["🍔","🍕","🌮","🍜","🍣","🥗","🍗","🍱","🥙"];

function itemEmoji(i) {
  return ITEM_EMOJIS[i % ITEM_EMOJIS.length];
}

function renderCartItems() {
  if (cartItems.length === 0) {
    document.getElementById("cart-items-list").innerHTML = `
      <div class="empty-state">
        <div class="emoji">🛒</div>
        <h3>Корзина пуста</h3>
        <p>Добавьте что-нибудь вкусное!</p>
        <a href="./index.html" class="btn btn-primary" style="margin-top:16px">К ресторанам</a>
      </div>
    `;
    return;
  }

  document.getElementById("cart-items-list").innerHTML = cartItems.map((item, i) => `
    <div class="cart-item">
      <div class="cart-item-emoji">${itemEmoji(i)}</div>
      <div class="cart-item-info">
        <div class="cart-item-name">${item.name}</div>
        <div class="cart-item-price">${Number(item.price).toLocaleString()} ₸ × ${item.quantity}</div>
      </div>
      <div class="cart-item-subtotal">
        ${(Number(item.price) * item.quantity).toLocaleString()} ₸
      </div>
    </div>
  `).join("");
}

function updateSummary() {
  document.getElementById("sum-items").textContent = Number(cartTotal).toLocaleString() + " ₸";

  if (appliedDiscount > 0) {
    document.getElementById("sum-discount-row").style.display = "flex";
    document.getElementById("sum-discount").textContent = "−" + Number(appliedDiscount).toLocaleString() + " ₸";
  } else {
    document.getElementById("sum-discount-row").style.display = "none";
  }

  const final = cartTotal - appliedDiscount;
  document.getElementById("sum-total").textContent = Number(final).toLocaleString() + " ₸";
}

async function applyPromo() {
  const user = getUser();
  if (!user) { window.location.href = "./auth.html"; return; }
  const code = document.getElementById("promo-code").value.trim().toUpperCase();
  if (!code) { toast("Введите промокод", "error"); return; }

  try {
    const data = await API.applyPromo({
      order_id: "",
      user_id: user.id,
      promo_code: code,
      is_student: user.is_student || false,
      order_total: cartTotal,
    });
    appliedDiscount = data.discount_amount || 0;
    appliedPromo = code;
    updateSummary();
    toast(`Промокод применён! Скидка: ${Number(appliedDiscount).toLocaleString()} ₸ 🎉`, "success");
  } catch (e) {
    toast(e.message, "error");
  }
}

async function checkout() {
  const user = getUser();
  if (!user) { window.location.href = "./auth.html"; return; }
  if (cartItems.length === 0) { toast("Корзина пуста", "error"); return; }

  const address = document.getElementById("delivery-address").value.trim();
  if (!address) { toast("Введите адрес доставки", "error"); return; }

  const items = cartItems.map(item => ({
    item_id: item.item_id || String(item.id || ""),
    name: item.name,
    quantity: item.quantity,
    price: Number(item.price),
  }));

  try {
    await API.createOrder({
      user_id: user.id,
      restaurant_id: "1",
      delivery_address: address,
      items,
      promo_code: appliedPromo,
      is_student: user.is_student || false,
    });
    toast("Заказ оформлен! 🎉", "success");
    await API.clearCart({ user_id: user.id });
    setTimeout(() => { window.location.href = "./orders.html"; }, 1200);
  } catch (e) {
    toast("Ошибка оформления: " + e.message, "error");
  }
}

async function clearCart() {
  const user = getUser();
  if (!user) return;
  try {
    await API.clearCart({ user_id: user.id });
    cartItems = [];
    cartTotal = 0;
    appliedDiscount = 0;
    renderCartItems();
    updateSummary();
    updateNavCart();
    toast("Корзина очищена", "info");
  } catch (e) {
    toast(e.message, "error");
  }
}

async function loadCart() {
  if (!requireAuth()) return;
  renderNav();
  const user = getUser();

  try {
    const data = await API.getCart(user.id);
    cartItems = data.items || [];
    cartTotal = data.total || 0;
    renderCartItems();
    updateSummary();
  } catch (e) {
    cartItems = [];
    cartTotal = 0;
    renderCartItems();
    updateSummary();
  }
}

document.addEventListener("DOMContentLoaded", loadCart);let cartItems = [];
let cartTotal = 0;
let appliedDiscount = 0;
let appliedPromo = "";

const ITEM_EMOJIS = ["🍔","🍕","🌮","🍜","🍣","🥗","🍗","🍱","🥙"];

function itemEmoji(i) {
  return ITEM_EMOJIS[i % ITEM_EMOJIS.length];
}

function renderCartItems() {
  if (cartItems.length === 0) {
    document.getElementById("cart-items-list").innerHTML = `
      <div class="empty-state">
        <div class="emoji">🛒</div>
        <h3>Корзина пуста</h3>
        <p>Добавьте что-нибудь вкусное!</p>
        <a href="./index.html" class="btn btn-primary" style="margin-top:16px">К ресторанам</a>
      </div>
    `;
    return;
  }

  document.getElementById("cart-items-list").innerHTML = cartItems.map((item, i) => `
    <div class="cart-item">
      <div class="cart-item-emoji">${itemEmoji(i)}</div>
      <div class="cart-item-info">
        <div class="cart-item-name">${item.name}</div>
        <div class="cart-item-price">${Number(item.price).toLocaleString()} ₸ × ${item.quantity}</div>
      </div>
      <div class="cart-item-subtotal">
        ${(Number(item.price) * item.quantity).toLocaleString()} ₸
      </div>
    </div>
  `).join("");
}

function updateSummary() {
  document.getElementById("sum-items").textContent = Number(cartTotal).toLocaleString() + " ₸";

  if (appliedDiscount > 0) {
    document.getElementById("sum-discount-row").style.display = "flex";
    document.getElementById("sum-discount").textContent = "−" + Number(appliedDiscount).toLocaleString() + " ₸";
  } else {
    document.getElementById("sum-discount-row").style.display = "none";
  }

  const final = cartTotal - appliedDiscount;
  document.getElementById("sum-total").textContent = Number(final).toLocaleString() + " ₸";
}

async function applyPromo() {
  const user = getUser();
  if (!user) { window.location.href = "./auth.html"; return; }
  const code = document.getElementById("promo-code").value.trim().toUpperCase();
  if (!code) { toast("Введите промокод", "error"); return; }

  try {
    const data = await API.applyPromo({
      order_id: "",
      user_id: user.id,
      promo_code: code,
      is_student: user.is_student || false,
      order_total: cartTotal,
    });
    appliedDiscount = data.discount_amount || 0;
    appliedPromo = code;
    updateSummary();
    toast(`Промокод применён! Скидка: ${Number(appliedDiscount).toLocaleString()} ₸ 🎉`, "success");
  } catch (e) {
    toast(e.message, "error");
  }
}

async function checkout() {
  const user = getUser();
  if (!user) { window.location.href = "./auth.html"; return; }
  if (cartItems.length === 0) { toast("Корзина пуста", "error"); return; }

  const address = document.getElementById("delivery-address").value.trim();
  if (!address) { toast("Введите адрес доставки", "error"); return; }

  const items = cartItems.map(item => ({
    item_id: item.item_id || String(item.id || ""),
    name: item.name,
    quantity: item.quantity,
    price: Number(item.price),
  }));

  try {
    await API.createOrder({
      user_id: user.id,
      restaurant_id: "1",
      delivery_address: address,
      items,
      promo_code: appliedPromo,
      is_student: user.is_student || false,
    });
    toast("Заказ оформлен! 🎉", "success");
    await API.clearCart({ user_id: user.id });
    setTimeout(() => { window.location.href = "./orders.html"; }, 1200);
  } catch (e) {
    toast("Ошибка оформления: " + e.message, "error");
  }
}

async function clearCart() {
  const user = getUser();
  if (!user) return;
  try {
    await API.clearCart({ user_id: user.id });
    cartItems = [];
    cartTotal = 0;
    appliedDiscount = 0;
    renderCartItems();
    updateSummary();
    updateNavCart();
    toast("Корзина очищена", "info");
  } catch (e) {
    toast(e.message, "error");
  }
}

async function loadCart() {
  if (!requireAuth()) return;
  renderNav();
  const user = getUser();

  try {
    const data = await API.getCart(user.id);
    cartItems = data.items || [];
    cartTotal = data.total || 0;
    renderCartItems();
    updateSummary();
  } catch (e) {
    cartItems = [];
    cartTotal = 0;
    renderCartItems();
    updateSummary();
  }
}

document.addEventListener("DOMContentLoaded", loadCart);