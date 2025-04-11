let token = "";

function getUserDetails() {
  return {
    username: document.getElementById("username").value,
    password: document.getElementById("password").value,
  };
}

function showToast(message, isError = false) {
  const toast = document.getElementById("toast");
  toast.textContent = message;
  toast.style.backgroundColor = isError ? "#e74c3c" : "#2ecc71"; // Red for error, Green for success
  toast.classList.remove("hidden");
  toast.classList.add("show");

  setTimeout(() => {
    toast.classList.remove("show");
    toast.classList.add("hidden");
  }, 3000);
}

async function register() {
  const user = getUserDetails();
  try {
    const res = await fetch("http://localhost:8080/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(user),
    });

    const data = await res.json();
    if (res.ok) {
      showToast("✅ Registered successfully");
    } else {
      showToast("⚠️ " + (data.error || "Registration failed"), true);
    }
  } catch (err) {
    showToast("❌ Error: " + err.message, true);
  }
}

async function login() {
  const user = getUserDetails();
  try {
    const res = await fetch("http://localhost:8080/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(user),
    });

    const data = await res.json();
    if (data.token) {
      token = data.token;
      document.getElementById("authSection").classList.add("hidden");
      document.getElementById("actionsSection").classList.remove("hidden");
      document.getElementById("welcomeText").textContent = `Welcome, ${user.username}!`;
      showToast("✅ Logged in successfully");
    } else {
      showToast("⚠️ " + (data.error || "Login failed"), true);
    }
  } catch (err) {
    showToast("❌ Error: " + err.message, true);
  }
}

async function checkin() {
  try {
    const res = await fetch("http://localhost:8080/checkin", {
      method: "POST",
      headers: {
        Authorization: token,
      },
    });

    const data = await res.json();
    if (res.ok) {
      showToast("🟢 " + (data.message || "Checked in!"));
    } else {
      showToast("⚠️ Check-in failed", true);
    }
  } catch (err) {
    showToast("❌ Error: " + err.message, true);
  }
}

async function viewCheckins() {
    const res = await fetch("http://localhost:8080/checkins", {
      headers: {
        Authorization: token,
      },
    });
  
    const data = await res.json();
  
    const results = document.getElementById("checkinResults");
    results.innerHTML = ""; // Clear previous results
  
    if (data.checkins && data.checkins.length > 0) {
      const list = document.createElement("ul");
      data.checkins.forEach(item => {
        const li = document.createElement("li");
        li.textContent = `${item.username} checked in at ${new Date(item.time).toLocaleString()}`;
        list.appendChild(li);
      });
      results.appendChild(list);
    } else {
      results.textContent = "No check-ins found.";
    }
  }
  
  
function logout() {
  token = "";
  document.getElementById("authSection").classList.remove("hidden");
  document.getElementById("actionsSection").classList.add("hidden");
  showToast("🚪 Logged out");
}
