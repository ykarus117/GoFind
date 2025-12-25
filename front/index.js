import {drawTree} from "./modules/app.js";

const API_BASE_URL = location.origin;
const responseArea = document.getElementById('response');
const detailsContainer = document.getElementById('detailsPanel');
const createPanel = document.getElementById('createPanel');
const itemPanel = document.getElementById('itemsPanel');

let fullDataObject = {};
const fetchOptions = {credentials: 'same-origin'};

window.onload = () => {
    if (!localStorage.getItem("username")) {
        window.location.replace("./login.html");
        return;
    }
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

async function callback(selected) {
    if (selected["ref"] != null) {
        const response = await fetch(`${API_BASE_URL}/item/${selected["ref"]}`, {
            ...fetchOptions,
            method: 'GET',
            headers: {'Accept': 'application/json'},
        })
        if (response.ok) {
            populateDetails(await response.json());
        }else{
            displayError(response.status);
        }
    } else {
        const response = await fetch(`${API_BASE_URL}/object/${selected["name"]}`, {
            ...fetchOptions,
            method: 'GET',
            headers: {'Accept': 'application/json'},
        });
        if (response.ok){
            populateDetails(await response.json());
        }else{
            displayError(response.status);
        }
    }
}

function populateDetails (object){
    if (!object) return;

    if (detailsContainer.classList.contains('fade-out')){
        detailsContainer.classList.remove('fade-out');
        detailsContainer.classList.add('fade-in');
    }
    const details = document.getElementById('details');
    document.getElementById('detailHeaderName').innerText = object["name"];

    itemPanel.innerHTML = '';
    details.innerHTML = '';

    for (const key in object) {
        if (Array.isArray(object[key])) {
            for (const key2 in object[key]) {
                const p = document.createElement('details')
                const summary = document.createElement('summary');
                summary.append(object[key][key2]["name"]);

                for (const element in object[key][key2]) {
                    p.innerHTML += `<label for="D-${element}">${element}:</label><input id="D-${element}" type="text" placeholder="${object[key][key2][element]}">`
                }

                p.appendChild(summary);
                itemPanel.appendChild(p);
            }
        }else{
            const div = document.createElement('div');
            div.classList.add('form-group');
            div.innerHTML = `<label for="D-${key}">${key}:</label><input id="D-${key}" type="text" placeholder="${object[key]}">`
            details.appendChild(div)
        }
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

// --- Event Listeners ---
document.getElementById('updateBtn').addEventListener('click', async () => {
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

document.getElementById('detailSectionBtn').addEventListener('click', () => {
    detailsContainer.classList.remove('fade-in');
    detailsContainer.classList.add('fade-out');
})

document.getElementById('newBtn').addEventListener('click', () => {
    createPanel.classList.remove('fade-out');
    createPanel.classList.add('fade-in');
})

document.getElementById('closePanel').addEventListener('click', () => {
    createPanel.classList.remove('fade-in');
    createPanel.classList.add('fade-out');

})
