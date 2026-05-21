function showAlert(msg, type) {
  document.getElementById("alert-box").innerHTML =
    `<div class="alert alert-${type}">${msg}</div>`;
}

async function doLogout() {
  const token = localStorage.getItem("refresh_token") || "";
  try { await API.logout({ refresh_token: token }); } catch (_) {}
  clearUser();
  window.location.href = "./auth.html";
}

async function saveProfile() {
  const user = getUser();
  const name = document.getElementById("edit-name").value.trim();
  const phone = document.getElementById("edit-phone").value.trim();
  if (!name) { showAlert("Введите имя", "error"); return; }
  try {
    const data = await API.updateProfile({ user_id: user.id, name, phone });
    const updated = data.user || data;
    user.name = updated.name || name;
    user.phone = updated.phone || phone;
    localStorage.setItem("user", JSON.stringify(user));
    renderProfile(user);
    showAlert("Профиль обновлён!", "success");
  } catch (e) {
    showAlert(e.message, "error");
  }
}

function renderProfile(user) {
  document.getElementById("profile-avatar").textContent =
    user.name ? user.name.charAt(0).toUpperCase() : "?";
  document.getElementById("profile-name").textContent = user.name || "—";
  document.getElementById("profile-email").textContent = user.email || "";
  document.getElementById("edit-name").value = user.name || "";
  document.getElementById("edit-phone").value = user.phone || "";

  const badgeEl = document.getElementById("student-badge");
  if (user.is_student) {
    badgeEl.innerHTML = `<div class="student-badge">🎓 Студент AITU — скидка 15% по AITU2025</div>`;
  } else if (user.role === "courier") {
    badgeEl.innerHTML = `<div class="student-badge" style="background:var(--yellow);color:var(--blue)">🛵 Курьер</div>`;
  } else {
    badgeEl.innerHTML = "";
  }
}

async function loadProfile() {
  if (!requireAuth()) return;
  renderNav();
  const user = getUser();
  renderProfile(user);

  try {
    const data = await API.getProfile(user.id);
    const serverUser = data.user || data;
    const merged = { ...user, ...serverUser };
    localStorage.setItem("user", JSON.stringify(merged));
    renderProfile(merged);
  } catch (_) {}
}

document.addEventListener("DOMContentLoaded", loadProfile);