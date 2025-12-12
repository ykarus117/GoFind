    const API_BASE_URL = location.origin;
    const responseArea = document.getElementById('response');
    const mainContent = document.getElementById('mainContent');
    const objectGridContainer = document.getElementById('objectGridContainer');
    const detailsContainer = document.getElementById('detailsContainer');

    let fullDataObject = {};
    const fetchOptions = { credentials: 'same-origin' };

    function displayResult(data) {
    responseArea.innerHTML = '';
    try {
    const text = typeof data === 'string' ? data : JSON.stringify(data, null, 2);
    const formatted = text.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, (match) => {
    let cls = 'json-value-number';
    if (/^"/.test(match)) {
    cls = /:$/.test(match) ? 'json-key' : 'json-value-string';
} else if (/true|false/.test(match)) {
    cls = 'json-value-bool';
} else if (/null/.test(match)) {
    cls = 'json-value-null';
}
    return `<span class="${cls}">${match}</span>`;
});
    responseArea.innerHTML = `<pre>${formatted}</pre>`;
} catch (e) {
    responseArea.textContent = typeof data === 'string' ? data : JSON.stringify(data, null, 2);
}
}
    function displayError(error) {
    responseArea.innerHTML = `<span style="color:#ff5555;">Error: ${error.message}</span>`;
}

    function createCard(entity, type) {
    const card = document.createElement('div');
    card.className = 'card';
    const id = type === 'item' ? entity.id : entity.name;
    card.dataset.id = id;

    const isObject = type === 'object';
    const name = entity.name || (isObject ? '(Loose Items)' : '(Unnamed Item)');
    const headerText = isObject ? `Object: ${name}` : `Item: ${name} (ID: ${id})`;

    card.innerHTML = `
            <div class="card-header">
                <span><strong>${name}</strong></span>
                <span>&#9662;</span>
            </div>
            <div class="card-details"></div>
        `;

    const detailsDiv = card.querySelector('.card-details');
    if (isObject) {
    detailsDiv.innerHTML = `
                <div class="form-group"><label>Name:</label><input class="update-name" value="${entity.name || ''}"></div>
                <div class="form-group"><label>Description:</label><input class="update-description" value="${entity.description || ''}"></div>
                <div class="form-group"><label>Tags:</label><input class="update-tags" value="${(entity.tags || []).join(', ')}"></div>
                <div class="form-group"><label>Container:</label><input class="update-container" value="${entity.container || ''}"></div>
                <div class="actions">
                    <button class="update-btn">Update</button>
                    <button class="delete-btn">Delete</button>
                </div>
                <div class="sub-item-list"></div>`;
    card.querySelector('.update-btn').addEventListener('click', (e) => { e.stopPropagation(); updateObject(entity.name, card); });
    card.querySelector('.delete-btn').addEventListener('click', (e) => { e.stopPropagation(); deleteObject(entity.name); });
} else { // Item
    detailsDiv.innerHTML = `
                <div class="form-group"><label>Name:</label><input class="update-name" value="${entity.name || ''}"></div>
                <div class="form-group"><label>Quantity:</label><input class="update-quantity" type="number" value="${entity.quantity || 0}"></div>
                <div class="form-group"><label>Description:</label><input class="update-description" value="${entity.description || ''}"></div>
                <div class="form-group"><label>Tags:</label><input class="update-tags" value="${(entity.tags || []).join(', ')}"></div>
                <div class="form-group"><label>Container:</label><input class="update-container" value="${entity.container || ''}"></div>
                <div class="actions">
                    <button class="update-btn">Update</button>
                    <button class="delete-btn">Delete</button>
                </div>`;
    card.querySelector('.update-btn').addEventListener('click', (e) => { e.stopPropagation(); updateItem(entity.id, card); });
    card.querySelector('.delete-btn').addEventListener('click', (e) => { e.stopPropagation(); deleteItem(entity.id, card); });
}

    card.querySelector('.card-header').addEventListener('click', () => {
    detailsDiv.style.display = detailsDiv.style.display === 'block' ? 'none' : 'block';
});

    return card;
}

    function populateDetailsView(objectName) {
    detailsContainer.innerHTML = '';
    const objectGroup = fullDataObject[objectName];
    if (!objectGroup) {
    detailsContainer.textContent = `Object "${objectName}" not found.`;
    return;
}

    if (objectName !== "") {
    const ownCard = createCard(objectGroup[0], 'object');
    ownCard.querySelector('.card-details').style.display = 'block';
    detailsContainer.appendChild(ownCard);
}

    // Add sub-objects
    const subObjects = objectGroup.slice(1);

    if (subObjects.length > 0) {
    const subObjectsHeader = document.createElement('h3');
    subObjectsHeader.textContent = 'Sub-Objects';
    detailsContainer.appendChild(subObjectsHeader);
    subObjects.forEach(subObj => {
    const subObjectCard = createCard(subObj, 'object');
    detailsContainer.appendChild(subObjectCard);
});
}

    // Add items
    const mainObject = objectGroup[0];

    if (mainObject.items && mainObject.items.length > 0) {
    const itemsHeader = document.createElement('h3');
    itemsHeader.textContent = 'Items';
    detailsContainer.appendChild(itemsHeader);
    mainObject.items.forEach(item => {
    const itemCard = createCard(item, 'item');
    detailsContainer.appendChild(itemCard);
});
}
}

    function buildObjectGrid(data) {
    objectGridContainer.innerHTML = '';
    detailsContainer.innerHTML = 'Select an object to see its contents.';
    fullDataObject = data; // Store the full data

    for (const key in data) {
    const objectGroup = data[key];
    if (!objectGroup || objectGroup.length === 0) continue;
    const mainObject = objectGroup[0];

    // Skip the placeholder for loose items in the main grid, but show its items in details
    if (mainObject.name === "") {
    continue;
}

    const card = document.createElement('div');
    card.className = 'card';
    card.dataset.id = mainObject.name;
    card.innerHTML = `<div class="card-header"><strong>${mainObject.name}</strong></div>`;
    card.style.cursor = 'pointer';

    card.addEventListener('click', () => {
    document.querySelectorAll('#objectGridContainer .card').forEach(c => c.style.borderColor = 'var(--border-color)');
    card.style.borderColor = 'var(--primary-color)';
    populateDetailsView(mainObject.name);
});
    objectGridContainer.appendChild(card);
}
    // Add a card for loose items
    const looseItemsCard = document.createElement('div');
    looseItemsCard.className = 'card';
    looseItemsCard.innerHTML = `<div class="card-header"><strong>(Loose Items)</strong></div>`;
    looseItemsCard.style.cursor = 'pointer';
    looseItemsCard.addEventListener('click', () => {
    document.querySelectorAll('#objectGridContainer .card').forEach(c => c.style.borderColor = 'var(--border-color)');
    looseItemsCard.style.borderColor = 'var(--primary-color)';
    populateDetailsView("");
});
    objectGridContainer.appendChild(looseItemsCard);
}

    async function fetchAll() {
    const response = await fetch(`${API_BASE_URL}/items`, fetchOptions);
    if (!response.ok) {
        responseArea.innerHTML = `<span style="color: darkred;">Error: HTTP error! status: ${response.status}</span>`;
    }else {
        const data = await response.json();
        buildObjectGrid(data);
        displayResult(data);
    }
}

    // --- API Call Functions ---
    async function updateItem(id, card) {
    const itemData = {
    Item: {
    name: card.querySelector('.update-name').value,
    quantity: parseInt(card.querySelector('.update-quantity').value, 10) || 0,
    description: card.querySelector('.update-description').value,
    tags: card.querySelector('.update-tags').value.split(',').map(t => t.trim()).filter(Boolean),
    container: card.querySelector('.update-container').value,
}, Object: {},
};
    try {
    const response = await fetch(`${API_BASE_URL}/item/${id}`, {
    ...fetchOptions, method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(itemData),
});
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
    displayResult(await response.json());
    fetchAll();
} catch (error) { displayError(error); }
}

    async function deleteItem(id, card) {
    if (!confirm(`Are you sure you want to delete item ${id}?`)) return;
    try {
    const response = await fetch(`${API_BASE_URL}/item/${id}`, { ...fetchOptions, method: 'DELETE' });
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
    displayResult({ success: `Item ${id} deleted.`, status: response.status });
    fetchAll();
} catch (error) { displayError(error); }
}

    async function updateObject(name, card) {
    const objectData = {
    Item: {},
    Object: {
    name: card.querySelector('.update-name').value,
    description: card.querySelector('.update-description').value,
    tags: card.querySelector('.update-tags').value.split(',').map(t => t.trim()).filter(Boolean),
    container: card.querySelector('.update-container').value,
},
};
    try {
    const response = await fetch(`${API_BASE_URL}/object/${encodeURIComponent(name)}`, {
    ...fetchOptions, method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(objectData),
});
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
    displayResult(await response.json());
    fetchAll();
} catch (error) { displayError(error); }
}

    async function deleteObject(name) {
    if (!confirm(`Are you sure you want to delete object "${name}"? This will also detach its items.`)) return;
    try {
    const response = await fetch(`${API_BASE_URL}/object/${encodeURIComponent(name)}`, { ...fetchOptions, method: 'DELETE' });
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
    displayResult({ success: `Object ${name} deleted.`, status: response.status });
    fetchAll();
} catch (error) { displayError(error); }
}

    // --- Event Listeners ---
    document.getElementById('fetchAllBtn').addEventListener('click', fetchAll);

    document.getElementById('createBtn').addEventListener('click', async () => {
    const name = document.getElementById('create-name').value;
    const itemData = {
    Item: {
    name: name,
    quantity: parseInt(document.getElementById('create-quantity').value, 10) || 0,
    description: document.getElementById('create-description').value,
    tags: document.getElementById('create-tags').value.split(',').map(t => t.trim()).filter(Boolean),
    container: document.getElementById('create-object').value,
}, Object: {},
};
    try {
    const response = await fetch(`${API_BASE_URL}/item/${encodeURIComponent(name)}`, {
    ...fetchOptions, method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(itemData),
});
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
    displayResult(await response.json());
    fetchAll();
} catch (error) { displayError(error); }
});

    document.getElementById('createObjBtn').addEventListener('click', async () => {
    const name = document.getElementById('create-obj-name').value;
    const objectData = {
    Item: {},
    Object: {
    name: name,
    description: document.getElementById('create-obj-description').value,
    tags: document.getElementById('create-obj-tags').value.split(',').map(t => t.trim()).filter(Boolean),
    container: document.getElementById('create-obj-container').value,
},
};
    try {
    const response = await fetch(`${API_BASE_URL}/object/${encodeURIComponent(name)}`, {
    ...fetchOptions, method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(objectData),
});
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
    displayResult(await response.json());
    fetchAll();
} catch (error) { displayError(error); }
});

    document.getElementById('logoutBtn').addEventListener('click', () => {
        fetch(`${API_BASE_URL}/logout/${localStorage.getItem("username")}`, { ...fetchOptions, method: 'POST' });
        localStorage.removeItem('username');
        window.location.replace("./login.html");
    });
