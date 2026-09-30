let items = [];

document.addEventListener(
    "DOMContentLoaded",
    () => {
        initializeItemsPage();
    }
);

async function initializeItemsPage() {
    setupEventListeners();

    await loadCurrentUser();
    await loadItems();
}

function setupEventListeners() {
    const addButton =
        document.getElementById(
            "add-item-button"
        );

    const logoutButton =
        document.getElementById(
            "logout-button"
        );

    const closeButton =
        document.getElementById(
            "close-modal-button"
        );

    const cancelButton =
        document.getElementById(
            "cancel-button"
        );

    const form =
        document.getElementById(
            "item-form"
        );

    const searchInput =
        document.getElementById(
            "search-input"
        );

    const statusFilter =
        document.getElementById(
            "status-filter"
        );

    addButton.addEventListener(
        "click",
        () => {
            openCreateModal();
        }
    );

    logoutButton.addEventListener(
        "click",
        logout
    );

    closeButton.addEventListener(
        "click",
        closeModal
    );

    cancelButton.addEventListener(
        "click",
        closeModal
    );

    form.addEventListener(
        "submit",
        saveItem
    );

    searchInput.addEventListener(
        "input",
        renderItems
    );

    statusFilter.addEventListener(
        "change",
        renderItems
    );
}

// User

async function loadCurrentUser() {
    const { response, data } =
        await apiJson(
            "/api/v1/me"
        );

    if (!response.ok) {
        return;
    }

    const emailElement =
        document.getElementById(
            "user-email"
        );

    if (emailElement) {
        emailElement.textContent =
            data.email;
    }
}

// Items

async function loadItems() {
    clearError();

    const { response, data } =
        await apiJson(
            "/api/v1/items"
        );

    if (!response.ok) {
        showError(
            data?.error ||
            "Не удалось загрузить каталог."
        );

        return;
    }

    items = Array.isArray(data)
        ? data
        : data.items || [];

    renderItems();
}

function renderItems() {
    const list =
        document.getElementById(
            "items-list"
        );

    const emptyMessage =
        document.getElementById(
            "empty-message"
        );

    const searchInput =
        document.getElementById(
            "search-input"
        );

    const statusFilter =
        document.getElementById(
            "status-filter"
        );

    const search =
        searchInput.value
            .trim()
            .toLowerCase();

    const status =
        statusFilter.value;

    const filteredItems =
        items.filter((item) => {
            const matchesSearch =
                !search ||
                (item.title || "")
                    .toLowerCase()
                    .includes(search);

            const matchesStatus =
                !status ||
                item.status === status;

            return (
                matchesSearch &&
                matchesStatus
            );
        });

    list.innerHTML = "";

    if (filteredItems.length === 0) {
        emptyMessage.classList.remove(
            "hidden"
        );

        return;
    }

    emptyMessage.classList.add(
        "hidden"
    );

    for (const item of filteredItems) {
        list.appendChild(
            createItemCard(item)
        );
    }
}

function createItemCard(item) {
    const card =
        document.createElement("article");

    card.className = "item-card";

    const header =
        document.createElement("div");

    header.className = "item-header";

    const title =
        document.createElement("h2");

    title.className = "item-title";

    title.textContent =
        item.title || "Без названия";

    header.appendChild(title);

    card.appendChild(header);

    if (item.description) {
        const description =
            document.createElement("p");

        description.className =
            "item-description";

        description.textContent =
            item.description;

        card.appendChild(
            description
        );
    }

    const meta =
        document.createElement("div");

    meta.className = "item-meta";

    if (item.status) {
        const status =
            document.createElement("span");

        status.className = "status";

        status.textContent =
            getStatusLabel(
                item.status
            );

        meta.appendChild(status);
    }

    if (item.rating !== null &&
        item.rating !== undefined) {

        const rating =
            document.createElement("span");

        rating.className = "rating";

        rating.textContent =
            `★ ${item.rating}/10`;

        meta.appendChild(rating);
    }

    card.appendChild(meta);

    if (item.notes) {
        const notes =
            document.createElement("p");

        notes.className = "item-notes";

        notes.textContent =
            item.notes;

        card.appendChild(notes);
    }

    const actions =
        document.createElement("div");

    actions.className =
        "item-actions";

    const editButton =
        document.createElement("button");

    editButton.type = "button";
    editButton.textContent =
        "Изменить";

    editButton.addEventListener(
        "click",
        () => {
            openEditModal(item);
        }
    );

    const deleteButton =
        document.createElement("button");

    deleteButton.type = "button";
    deleteButton.className =
        "delete-button";

    deleteButton.textContent =
        "Удалить";

    deleteButton.addEventListener(
        "click",
        () => {
            deleteItem(item.id);
        }
    );

    actions.appendChild(editButton);
    actions.appendChild(deleteButton);

    card.appendChild(actions);

    return card;
}

function getStatusLabel(status) {
    const labels = {
        planned: "Запланировано",
        watching: "В процессе",
        completed: "Завершено",
        dropped: "Брошено",
    };

    return labels[status] || status;
}

// Create / Edit

function openCreateModal() {
    document.getElementById(
        "modal-title"
    ).textContent = "Добавить в каталог";

    document.getElementById(
        "item-id"
    ).value = "";

    document.getElementById(
        "item-title"
    ).value = "";

    document.getElementById(
        "item-description"
    ).value = "";

    document.getElementById(
        "item-status"
    ).value = "planned";

    document.getElementById(
        "item-rating"
    ).value = "";

    document.getElementById(
        "item-notes"
    ).value = "";

    openModal();
}

function openEditModal(item) {
    document.getElementById(
        "modal-title"
    ).textContent = "Редактировать item";

    document.getElementById(
        "item-id"
    ).value = item.id;

    document.getElementById(
        "item-title"
    ).value =
        item.title || "";

    document.getElementById(
        "item-description"
    ).value =
        item.description || "";

    document.getElementById(
        "item-status"
    ).value =
        item.status || "planned";

    document.getElementById(
        "item-rating"
    ).value =
        item.rating ?? "";

    document.getElementById(
        "item-notes"
    ).value =
        item.notes || "";

    openModal();
}

function openModal() {
    document
        .getElementById("item-modal")
        .classList.remove("hidden");
}

function closeModal() {
    document
        .getElementById("item-modal")
        .classList.add("hidden");
}

async function saveItem(event) {
    event.preventDefault();

    clearError();

    const id =
        document.getElementById(
            "item-id"
        ).value;

    const title =
        document.getElementById(
            "item-title"
        ).value.trim();

    const description =
        document.getElementById(
            "item-description"
        ).value.trim();

    const status =
        document.getElementById(
            "item-status"
        ).value;

    const ratingValue =
        document.getElementById(
            "item-rating"
        ).value;

    const notes =
        document.getElementById(
            "item-notes"
        ).value.trim();

    const payload = {
        title: title || null,
        description:
            description || null,
        status: status || null,
        rating:
            ratingValue === ""
                ? null
                : Number(ratingValue),
        notes: notes || null,
    };

    let response;
    let data;

    if (id) {
        ({ response, data } =
            await apiJson(
                `/api/v1/items/${id}`,
                {
                    method: "PUT",
                    body: JSON.stringify(
                        payload
                    ),
                }
            ));
    } else {
        ({ response, data } =
            await apiJson(
                "/api/v1/items",
                {
                    method: "POST",
                    body: JSON.stringify(
                        payload
                    ),
                }
            ));
    }

    if (!response.ok) {
        showError(
            data?.error ||
            "Не удалось сохранить item."
        );

        return;
    }

    closeModal();

    await loadItems();
}

// Delete

async function deleteItem(id) {
    const confirmed =
        window.confirm(
            "Удалить этот item?"
        );

    if (!confirmed) {
        return;
    }

    clearError();

    const response =
        await apiFetch(
            `/api/v1/items/${id}`,
            {
                method: "DELETE",
            }
        );

    if (!response.ok) {
        let data = null;

        try {
            data =
                await response.json();
        } catch {

        }

        showError(
            data?.error ||
            "Не удалось удалить item."
        );

        return;
    }

    await loadItems();
}

// Logout

async function logout() {
    await fetch(
        "/api/v1/auth/logout",
        {
            method: "POST",
            credentials: "include",
        }
    );

    accessToken = null;

    window.location.href =
        "/";
}

// Errors

function showError(message) {
    const errorElement =
        document.getElementById(
            "error-message"
        );

    if (!errorElement) {
        return;
    }

    errorElement.textContent =
        message;
}

function clearError() {
    showError("");
}