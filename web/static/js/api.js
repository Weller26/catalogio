let accessToken = null;

async function refreshAccessToken() {
    const response = await fetch(
        "/api/v1/auth/refresh",
        {
            method: "POST",
            credentials: "include",
        }
    );

    if (!response.ok) {
        accessToken = null;
        return false;
    }

    const data = await response.json();

    accessToken = data.access_token;

    return true;
}

async function apiFetch(
    url,
    options = {},
    retry = true
) {
    const headers = new Headers(
        options.headers || {}
    );

    if (accessToken) {
        headers.set(
            "Authorization",
            `Bearer ${accessToken}`
        );
    }

    if (
        options.body &&
        !headers.has("Content-Type")
    ) {
        headers.set(
            "Content-Type",
            "application/json"
        );
    }

    let response = await fetch(url, {
        ...options,
        headers,
        credentials: "include",
    });

    if (
        response.status === 401 &&
        retry
    ) {
        const refreshed = await refreshAccessToken();

        if (!refreshed) {
            window.location.href = "/";
            return response;
        }

        return apiFetch(
            url,
            options,
            false
        );
    }

    return response;
}

async function apiJson(
    url,
    options = {}
) {
    const response = await apiFetch(
        url,
        options
    );

    let data = null;

    try {
        data = await response.json();
    } catch {

    }

    return {
        response,
        data,
    };
}