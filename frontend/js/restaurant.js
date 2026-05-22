const ITEM_EMOJIS = ["🍔","🍕","🌮","🍜","🍣","🥗","🍗","🍱","🥙","🫕","🍛","🥘","🍲","🌯","🥪"];

function itemEmoji(index) {
  return ITEM_EMOJIS[index % ITEM_EMOJIS.length];
}

let currentRestId = null;

function renderMenuItem(item, index) {
  const inStock = item.stock > 0;
  return `
    <div class="menu-item-card">
      <div class="menu-item-emoji">${itemEmoji(index)}</div>
      <div class="menu-item-name">${item.name}</div>
      <div class="menu-item-desc">${item.description || "Вкусное блюдо"}</div>
      <div class="menu-item-footer">
        <div class="menu-item-price">${Number(item.price).toLocaleString()} ₸</div>
        <span class="stock-label">${inStock ? `В наличии: ${item.stock}` : "Нет в наличии"}</span>
      </div>
      ${inStock
        ? `<button class="btn btn-primary" style="margin-top:10px;width:100%"
             onclick="addToCart('${item.id}','${(item.name||'').replace(/'/g,"\\'")}',${item.price})">
             + В корзину
           </button>`
        : `<button class="btn btn-outline" style="margin-top:10px;width:100%" disabled>Нет в наличии</button>`
      }
    </div>
  `;
}

async function addToCart(itemId, name, price) {
  const user = getUser();
  if (!user) { window.location.href = "./auth.html"; return; }
  try {
    await API.addToCart({
      user_id: user.id,
      item: { item_id: String(itemId), name, quantity: 1, price: Number(price) }
    });
    toast("Добавлено в корзину! 🛒", "success");
    updateNavCart();
  } catch (e) {
    toast(e.message, "error");
  }
}

async function doRate() {
  const user = getUser();
  if (!user) { window.location.href = "./auth.html"; return; }
  const rating = parseInt(document.getElementById("rate-value").value);
  const comment = document.getElementById("rate-comment").value.trim();
  try {
    await API.rateRestaurant({
      restaurant_id: currentRestId,
      user_id: 0,
      rating,
      comment,
    });
    toast("Спасибо за отзыв! ⭐", "success");
    document.getElementById("rate-comment").value = "";
  } catch (e) {
    toast(e.message, "error");
  }
}

async function loadPage() {
  renderNav();
  const params = new URLSearchParams(window.location.search);
  const id = params.get("id");
  if (!id) { window.location.href = "./index.html"; return; }
  currentRestId = parseInt(id);

  try {
    const data = await API.getRestaurant(id);
    const rest = data.restaurant || {};
    document.title = `AITUdel — ${rest.name || "Меню"}`;
    document.getElementById("rest-header").innerHTML = `
      <div class="menu-header-icon">🏪</div>
      <div class="menu-header-info">
        <h2>${rest.name || "Ресторан"}</h2>
        <p>${rest.description || "Вкусная еда с доставкой"}</p>
        <div class="rating" style="margin-top:8px">
          <span class="rating-star">★</span>
          ${rest.rating ? Number(rest.rating).toFixed(1) : "Нет оценок"}
        </div>
      </div>
    `;
  } catch (e) {
    document.getElementById("rest-header").innerHTML = `
      <div class="menu-header-icon">🏪</div>
      <div class="menu-header-info"><h2>Ресторан</h2><p>Меню загружается...</p></div>
    `;
  }

  try {
    const data = await API.searchItems("");
    const items = data.items || [];
    if (items.length === 0) {
      document.getElementById("menu-grid").innerHTML = `
        <div class="empty-state"><div class="emoji">🍽️</div><h3>Меню пока пусто</h3></div>
      `;
      return;
    }
    document.getElementById("menu-grid").innerHTML =
      items.map((item, i) => renderMenuItem(item, i)).join("");
    if (getUser()) document.getElementById("rate-section").style.display = "block";
  } catch (e) {
    document.getElementById("menu-grid").innerHTML = `
      <div class="empty-state"><div class="emoji">🍽️</div><h3>Меню загрузится после запуска сервера</h3></div>
    `;
  }
}

document.addEventListener("DOMContentLoaded", loadPage);