document.addEventListener(
    "DOMContentLoaded",
    () => {
        const loginForm =
            document.getElementById("login-form");

        const registerForm =
            document.getElementById("register-form");

        if (loginForm) {
            setupLoginForm(loginForm);
        }

        if (registerForm) {
            setupRegisterForm(registerForm);
        }
    }
);

function showError(message) {
    const errorElement =
        document.getElementById(
            "error-message"
        );

    if (!errorElement) {
        return;
    }

    errorElement.textContent = message;
}

function clearError() {
    showError("");
}

function setupLoginForm(form) {
    form.addEventListener(
        "submit",
        async (event) => {
            event.preventDefault();

            clearError();

            const email =
                document.getElementById(
                    "email"
                ).value.trim();

            const password =
                document.getElementById(
                    "password"
                ).value;

            const response = await fetch(
                "/api/v1/auth/login",
                {
                    method: "POST",

                    headers: {
                        "Content-Type":
                            "application/json",
                    },

                    credentials: "include",

                    body: JSON.stringify({
                        email,
                        password,
                    }),
                }
            );

            let data = null;

            try {
                data = await response.json();
            } catch {

            }

            if (!response.ok) {
                if (response.status === 401) {
                    showError(
                        "Неверный email или пароль."
                    );
                } else {
                    showError(
                        data?.error ||
                        "Не удалось войти."
                    );
                }

                return;
            }

            accessToken =
                data.access_token;

            window.location.href =
                "/items";
        }
    );
}

function setupRegisterForm(form) {
    form.addEventListener(
        "submit",
        async (event) => {
            event.preventDefault();

            clearError();

            const email =
                document.getElementById(
                    "email"
                ).value.trim();

            const password =
                document.getElementById(
                    "password"
                ).value;

            const response = await fetch(
                "/api/v1/auth/register",
                {
                    method: "POST",

                    headers: {
                        "Content-Type":
                            "application/json",
                    },

                    credentials: "include",

                    body: JSON.stringify({
                        email,
                        password,
                    }),
                }
            );

            let data = null;

            try {
                data = await response.json();
            } catch {

            }

            if (!response.ok) {
                showError(
                    data?.error ||
                    "Не удалось создать аккаунт."
                );

                return;
            }

            window.location.href =
                "/";
        }
    );
}