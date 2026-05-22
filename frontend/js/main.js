const EMOJI_MAP = ["🍔", "🍕", "🌮", "🍜", "🍣", "🥗", "🍗", "🍱", "🥙", "🫕"];

function restEmoji(index) {
  return EMOJI_MAP[index % EMOJI_MAP.length];
}

function renderRestaurantCard(rest, index) {
  return `
    <div class="rest-card" onclick="window.location.href='./restaurant.html?id=${rest.id}'">
      <div class="rest-card-img">${restEmoji(index)}</div>
      <div class="rest-card-body">
        <div class="rest-card-name">${rest.name}</div>
        <div class="rest-card-desc">${rest.description || "Вкусная еда с доставкой"}</div>
        <div class="rest-card-footer">
          <div class="rating">
            <span class="rating-star">★</span>
            ${rest.rating ? Number(rest.rating).toFixed(1) : "Новый"}
          </div>
          <span class="badge badge-open">Открыт</span>
        </div>
      </div>
    </div>
  `;
}

function renderSearchItem(item) {
  return `
    <div class="menu-item-card">
      <div class="menu-item-emoji">🍽️</div>
      <div class="menu-item-name">${item.name}</div>
      <div class="menu-item-desc">${item.description || ""}</div>
      <div class="menu-item-footer">
        <div class="menu-item-price">${Number(item.price).toLocaleString()} ₸</div>
        <span class="stock-label">В наличии: ${item.stock}</span>
      </div>
      <button class="btn btn-primary btn-sm" style="margin-top:8px"
        onclick="addSearchItemToCart('${item.id}', '${(item.name||'').replace(/'/g,"\\'")}', ${item.price})">
        + В корзину
      </button>
    </div>
  `;
}

async function addSearchItemToCart(itemId, name, price) {
  const user = getUser();
  if (!user) { window.location.href = "./auth.html"; return; }
  try {
    await API.addToCart({
      user_id: user.id,
      item: { item_id: String(itemId), name, quantity: 1, price: Number(price) }
    });
    toast("Добавлено в корзину!", "success");
    updateNavCart();
  } catch (e) {
    toast(e.message, "error");
  }
}

async function doHeroSearch() {
  const q = document.getElementById("hero-search").value.trim();
  if (!q) return;

  document.getElementById("main-section").style.display = "none";
  document.getElementById("search-results-section").style.display = "block";
  document.getElementById("search-results").innerHTML =
    '<div class="loader"><div class="spinner"></div></div>';

  try {
    const data = await API.searchItems(q);
    const items = data.items || [];
    if (items.length === 0) {
      document.getElementById("search-results").innerHTML = `
        <div class="empty-state">
          <div class="emoji">🔍</div>
          <h3>Ничего не найдено</h3>
          <p>Попробуйте другой запрос</p>
        </div>
      `;
      return;
    }
    document.getElementById("search-results").innerHTML = items.map(renderSearchItem).join("");
  } catch (e) {
    document.getElementById("search-results").innerHTML =
      `<div class="alert alert-error">${e.message}</div>`;
  }
}

async function loadRestaurants() {
  try {
    const data = await API.listRestaurants();
    const restaurants = data.restaurants || [];
    if (restaurants.length === 0) {
      document.getElementById("restaurants-grid").innerHTML = `
        <div class="empty-state">
          <div class="emoji">🏪</div>
          <h3>Рестораны скоро появятся</h3>
        </div>
      `;
      return;
    }
    document.getElementById("restaurants-grid").innerHTML =
      restaurants.map((r, i) => renderRestaurantCard(r, i)).join("");
  } catch (e) {
    document.getElementById("restaurants-grid").innerHTML = `
      <div class="empty-state">
        <div class="emoji">🏪</div>
        <h3>Рестораны скоро появятся</h3>
        <p style="font-size:13px;margin-top:6px">Запустите API Gateway для загрузки данных</p>
      </div>
    `;
  }
}

document.addEventListener("DOMContentLoaded", () => {
  renderNav();
  loadRestaurants();

  const params = new URLSearchParams(window.location.search);
  const searchQ = params.get("search");
  if (searchQ) {
    document.getElementById("hero-search").value = searchQ;
    doHeroSearch();
  }
});