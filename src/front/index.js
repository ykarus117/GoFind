import {drawTree} from "./modules/app.js";

const API_BASE_URL = location.origin;
const responseArea = document.getElementById('response');
const detailsContainer = document.getElementById('detailsContainer');

let fullDataObject = {};
const fetchOptions = {credentials: 'same-origin'};

window.onload = () => {
    fetchAll();
}

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

async function fetchAll() {
    const response = await fetch(`${API_BASE_URL}/items`, fetchOptions);
    if (!response.ok) {
        responseArea.innerHTML = `<span style="color: darkred;">Error: HTTP error! status: ${response.status}</span>`;
    } else {
        const data = await response.json();
        displayResult(data);
        fullDataObject = data
        document.getElementById('treeContainer').append(drawTree(data, callback));
    }
}

function callback(selected) {
    if (selected["ref"] != null) {
        //TODO: get Item
    }else{
        //TODO: object
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
            ...fetchOptions,
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(itemData),
        });
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        displayResult(await response.json());
    } catch (error) {
        displayError(error);
    }
}

async function deleteItem(id, card) {
    if (!confirm(`Are you sure you want to delete item ${id}?`)) return;
    try {
        const response = await fetch(`${API_BASE_URL}/item/${id}`, {...fetchOptions, method: 'DELETE'});
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        displayResult({success: `Item ${id} deleted.`, status: response.status});
        fetchAll();
    } catch (error) {
        displayError(error);
    }
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
            ...fetchOptions,
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(objectData),
        });
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        displayResult(await response.json());
        fetchAll();
    } catch (error) {
        displayError(error);
    }
}

async function deleteObject(name) {
    if (!confirm(`Are you sure you want to delete object "${name}"? This will also detach its items.`)) return;
    try {
        const response = await fetch(`${API_BASE_URL}/object/${encodeURIComponent(name)}`, {
            ...fetchOptions,
            method: 'DELETE'
        });
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        displayResult({success: `Object ${name} deleted.`, status: response.status});
        fetchAll();
    } catch (error) {
        displayError(error);
    }
}

function populateObjectDetails (object){
    detailsContainer.getElementsByTagName('detail')[0].innerHTML = '';
}

function populateItemDetails(Item){

}


// --- Event Listeners ---
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
            ...fetchOptions,
            method: 'PUT',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(itemData),
        });
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        displayResult(await response.json());
        fetchAll();
    } catch (error) {
        displayError(error);
    }
});

document.getElementById('logoutBtn').addEventListener('click', () => {
    fetch(`${API_BASE_URL}/logout/${localStorage.getItem("username")}`, {...fetchOptions, method: 'POST'});
    localStorage.removeItem('username');
    window.location.replace("./login.html");
});

