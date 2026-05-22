let cartItems = [];
let cartTotal = 0;
let appliedDiscount = 0;
let appliedPromo = "";

const ITEM_EMOJIS = ["🍔", "🍕", "🌮", "🍜", "🍣", "🥗", "🍗", "🍱", "🥙"];

function itemEmoji(i) {
    return ITEM_EMOJIS[i % ITEM_EMOJIS.length];
}

function renderCartItems() {
    const container = document.getElementById("cart-items-list");
    if (!container) return;

    if (cartItems.length === 0) {
        container.innerHTML = `
      <div class="empty-state">
        <div class="emoji">🛒</div>
        <h3>Корзина пуста</h3>
        <p>Добавьте что-нибудь вкусное!</p>
        <a href="./index.html" class="btn btn-primary" style="margin-top:16px">К ресторанам</a>
      </div>
    `;
        return;
    }

    container.innerHTML = cartItems.map((item, i) => `
    <div class="cart-item">
      <div class="cart-item-emoji">${itemEmoji(i)}</div>
      <div class="cart-item-info">
        <div class="cart-item-name">${item.name || 'Без названия'}</div>
        <div class="cart-item-price">${Number(item.price || 0).toLocaleString()} ₸ × ${item.quantity || 1}</div>
      </div>
      <div class="cart-item-subtotal">
        ${(Number(item.price || 0) * (item.quantity || 1)).toLocaleString()} ₸
      </div>
    </div>
  `).join("");
}

function updateSummary() {
    const sumItemsEl = document.getElementById("sum-items");
    const sumDiscountRowEl = document.getElementById("sum-discount-row");
    const sumDiscountEl = document.getElementById("sum-discount");
    const sumTotalEl = document.getElementById("sum-total");

    if (sumItemsEl) {
        sumItemsEl.textContent = Number(cartTotal).toLocaleString() + " ₸";
    }

    if (sumDiscountRowEl && sumDiscountEl) {
        if (appliedDiscount > 0) {
            sumDiscountRowEl.style.display = "flex";
            sumDiscountEl.textContent = "−" + Number(appliedDiscount).toLocaleString() + " ₸";
        } else {
            sumDiscountRowEl.style.display = "none";
        }
    }

    if (sumTotalEl) {
        const final = cartTotal - appliedDiscount;
        sumTotalEl.textContent = (final < 0 ? 0 : final).toLocaleString() + " ₸";
    }
}

async function applyPromo() {
    const user = getUser();
    if (!user) { window.location.href = "./auth.html"; return; }

    const promoInput = document.getElementById("promo-code");
    if (!promoInput) return;

    const code = promoInput.value.trim().toUpperCase();
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

    const addressInput = document.getElementById("delivery-address");
    const address = addressInput ? addressInput.value.trim() : "";
    if (!address) { toast("Введите адрес доставки", "error"); return; }

    const items = cartItems.map(item => {
        const rawId = String(item.item_id || item.id || "");

        const isUUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(rawId);

        const safeItemId = isUUID ? rawId : "00000000-0000-0000-0000-000000000002";

        return {
            item_id: safeItemId,
            name: item.name,
            quantity: item.quantity,
            price: Number(item.price),
        };
    });

    try {
        const restaurantId = cartItems[0]?.restaurant_id || "00000000-0000-0001-0000-000000000001";

        await API.createOrder({
            user_id: user.id,
            restaurant_id: restaurantId,
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
        appliedPromo = "";

        renderCartItems();
        updateSummary();

        if (typeof updateNavCart === "function") updateNavCart();

        toast("Корзина очищена", "info");
    } catch (e) {
        toast(e.message, "error");
    }
}

async function loadCart() {
    if (typeof requireAuth === "function" && !requireAuth()) return;

    if (typeof renderNav === "function") renderNav();

    const user = getUser();
    if (!user) return;

    try {
        const data = await API.getCart(user.id);
        cartItems = data.items || [];
        cartTotal = data.total || 0;
    } catch (e) {
        cartItems = [];
        cartTotal = 0;
    } finally {
        renderCartItems();
        updateSummary();
    }
}

document.addEventListener("DOMContentLoaded", loadCart);