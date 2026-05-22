const STATUS_LABELS = {
  pending: "Ожидает",
  confirmed: "Подтверждён",
  cooking: "Готовится",
  on_the_way: "В пути",
  delivered: "Доставлен",
  cancelled: "Отменён",
};

function renderOrderCard(order) {
  const statusClass = `status-${order.status}`;
  const statusLabel = STATUS_LABELS[order.status] || order.status;
  const date = order.created_at ? new Date(order.created_at).toLocaleString("ru-RU") : "—";
  const items = order.items || [];

  return `
    <div class="order-card">
      <div class="order-card-header">
        <div>
          <div class="order-card-id">Заказ #${(order.id || "").slice(0, 8)}...</div>
          <div class="order-card-date">${date}</div>
        </div>
        <span class="order-status ${statusClass}">${statusLabel}</span>
      </div>

      ${order.delivery_address ? `
        <div style="font-size:13px;color:var(--gray-text);margin-bottom:8px">
          📍 ${order.delivery_address}
        </div>
      ` : ""}

      ${order.promo_code ? `
        <div style="font-size:13px;color:#2e7d32;margin-bottom:8px">
          🎟️ Промокод: ${order.promo_code} (−${Number(order.discount_amount || 0).toLocaleString()} ₸)
        </div>
      ` : ""}

      <div class="order-items-list">
        ${items.map(item => `
          <div class="order-item-row">
            <span>${item.name} × ${item.quantity}</span>
            <span>${(Number(item.price) * item.quantity).toLocaleString()} ₸</span>
          </div>
        `).join("")}
      </div>

      <div class="order-total-row">
        <span>Итого</span>
        <span>${Number(order.total_price || 0).toLocaleString()} ₸</span>
      </div>

      ${(order.status === "pending" || order.status === "confirmed") ? `
        <button class="btn btn-danger btn-sm" style="margin-top:12px"
          onclick="cancelOrder('${order.id}')">
          Отменить заказ
        </button>
      ` : ""}
    </div>
  `;
}

async function cancelOrder(orderId) {
  const user = getUser();
  if (!confirm("Отменить заказ?")) return;
  try {
    const data = await API.cancelOrder({ order_id: orderId, user_id: user.id });
    if (data.success) {
      toast("Заказ отменён", "info");
      loadOrders();
    } else {
      toast(data.message || "Не удалось отменить", "error");
    }
  } catch (e) {
    toast(e.message, "error");
  }
}

async function loadOrders() {
  if (!requireAuth()) return;
  renderNav();
  const user = getUser();

  try {
    const data = await API.listOrders(user.id);
    const orders = data.orders || [];

    if (orders.length === 0) {
      document.getElementById("orders-list").innerHTML = `
        <div class="empty-state">
          <div class="emoji">📦</div>
          <h3>Заказов пока нет</h3>
          <p>Сделайте первый заказ!</p>
          <a href="./index.html" class="btn btn-primary" style="margin-top:16px">Заказать</a>
        </div>
      `;
      return;
    }

    document.getElementById("orders-list").innerHTML = orders.map(renderOrderCard).join("");
  } catch (e) {
    document.getElementById("orders-list").innerHTML = `
      <div class="empty-state">
        <div class="emoji">📦</div>
        <h3>Заказов пока нет</h3>
        <a href="./index.html" class="btn btn-primary" style="margin-top:16px">Заказать</a>
      </div>
    `;
  }
}

document.addEventListener("DOMContentLoaded", loadOrders);