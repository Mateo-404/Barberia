import { ENDPOINTS } from "../js/config.js";
const administradores = ENDPOINTS.administradores;

// Toggle password visibility
document
  .getElementById("togglePassword")
  .addEventListener("click", function () {
    const password = document.getElementById("password");
    const type =
      password.getAttribute("type") === "password" ? "text" : "password";
    password.setAttribute("type", type);
    this.textContent = type === "password" ? "👁️" : "🙈";
  });

// Form validation and submission
document.getElementById("loginForm").addEventListener("submit", function (e) {
  e.preventDefault();

  const form = this;
  const submitBtn = form.querySelector('button[type="submit"]');
  const loginText = submitBtn.querySelector(".login-text");
  const spinner = submitBtn.querySelector(".spinner-border");

  // Validate form
  if (!form.checkValidity()) {
    e.stopPropagation();
    form.classList.add("was-validated");
    return;
  }

  // Show loading state
  submitBtn.disabled = true;
  loginText.classList.add("d-none");
  spinner.classList.remove("d-none");

  // Get form data
  const formData = {
    email: document.getElementById("email").value.trim(),
    contrasenia: document.getElementById("password").value,
    //rememberMe: document.getElementById("rememberMe").checked,
  };

  // Simulate API call
  setTimeout(() => {
    // Replace with actual API call
    fetch(`${administradores}/login`, {
      method: "POST",
      headers: {
      "Content-Type": "application/json",
      },
      body: JSON.stringify(formData),
    })
      .then((response) => {
        if (!response.ok) {
          console.log("Response not OK:", response);
          throw new Error("Error en la solicitud de inicio de sesión");
        }
        return response.json();
      })
      .then((data) => {
        // Success - redirect to dashboard
        window.location.href = "/dashboard";
      })
      .catch((error) => {
        // Error handling
        console.error("Error:", error);
        alert("Error al iniciar sesión. Verifica tus credenciales.");
      })
      .finally(() => {
        // Reset button state
        submitBtn.disabled = false;
        loginText.classList.remove("d-none");
        spinner.classList.add("d-none");
      });
  }, 1000);
});

// Clear validation on input
document.querySelectorAll(".form-control").forEach((input) => {
  input.addEventListener("input", function () {
    if (this.classList.contains("is-invalid")) {
      this.classList.remove("is-invalid");
    }
  });
});

// Auto-focus email field
document.addEventListener("DOMContentLoaded", function () {
  document.getElementById("email").focus();
});
