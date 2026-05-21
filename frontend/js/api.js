const BASE = "http://localhost:8080/api";

async function apiCall(method, path, body = null) {
  const opts = {
    method,
    headers: { "Content-Type": "application/json" },
  };
  if (body) opts.body = JSON.stringify(body);
  const res = await fetch(BASE + path, opts);
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || "Ошибка сервера");
  return data;
}

const API = {
  register: (d) => apiCall("POST", "/auth/register", d),
  login: (d) => apiCall("POST", "/auth/login", d),
  logout: (d) => apiCall("POST", "/auth/logout", d),

  getProfile: (userId) => apiCall("GET", `/user/profile?user_id=${userId}`),
  updateProfile: (d) => apiCall("PUT", "/user/profile", d),

  listRestaurants: () => apiCall("GET", "/restaurants"),
  getRestaurant: (id) => apiCall("GET", `/restaurants/get?id=${id}`),
  rateRestaurant: (d) => apiCall("POST", "/restaurants/rate", d),

  searchItems: (q) => apiCall("GET", `/items/search?q=${encodeURIComponent(q)}`),
  getItem: (id) => apiCall("GET", `/items/get?id=${id}`),

  createOrder: (d) => apiCall("POST", "/orders/create", d),
  getOrder: (id) => apiCall("GET", `/orders/get?order_id=${id}`),
  listOrders: (userId, page = 1) => apiCall("GET", `/orders/list?user_id=${userId}&page=${page}&limit=10`),
  cancelOrder: (d) => apiCall("POST", "/orders/cancel", d),
  applyPromo: (d) => apiCall("POST", "/orders/promo", d),

  addToCart: (d) => apiCall("POST", "/cart/add", d),
  getCart: (userId) => apiCall("GET", `/cart?user_id=${userId}`),
  clearCart: (d) => apiCall("POST", "/cart/clear", d),
};