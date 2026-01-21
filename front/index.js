import {drawTree} from "./modules/app.js";
import {ui} from "./modules/ui.js";
import {api} from "./modules/api.js";
import {Item} from "./modules/item.js";

const API_BASE_URL = location.origin;
const responseArea = document.getElementById('response');
const detailsContainer = document.getElementById('detailsPanel');
const createPanel = document.getElementById('createPanel');
const itemPanel = document.getElementById('itemsPanel');

let fullDataObject = {};
const fetchOptions = {credentials: 'same-origin'};

window.onload = async () => {
    try {
        fullDataObject = await api.getItems()
        document.getElementById('treeContainer').innerHTML = null
        document.getElementById('treeContainer').append(drawTree(fullDataObject, callback));
    }catch(err) {
        if (err.cause === 401) {
            window.location.replace("./login.html");
        }else{
            ui.showNotification('Error: '+ err.message,'error');
        }
    }
}

async function Update() {
    try {
        fullDataObject = await api.getItems()
        document.getElementById('treeContainer').innerHTML = null
        document.getElementById('treeContainer').append(drawTree(fullDataObject, callback));
    }catch(err) {
        ui.showNotification('Error: '+ err.message,'error');
    }
}

async function callback(selected) {
    if (selected["ref"] != null) {
        const item = await new Item(selected["ref"]).init();
        ui.showDetailsPanel()
        item.render("details",'view');
    } else {
        const response = await fetch(`${API_BASE_URL}/object/${selected["name"]}`, {
            ...fetchOptions,
            method: 'GET',
            headers: {'Accept': 'application/json'},
        });
        if (response.ok){
            ui.populateDetails(await response.json());
        }else{
            ui.showNotification('Error: ' + response.statusText, 'error');
        }
    }
}

function buildItem(form){
    if(!form) return;
    const inputs = form.elements
    const formName = form.id
    return {
        Item: {
            name: inputs[`${formName}-name`].value,
            quantity: parseInt(inputs[`${formName}-quantity`].value, 10) || 1,
            description: inputs[`${formName}-description`].value,
            tags: inputs[`${formName}-tags`].value.split(',').map(t => t.trim()).filter(Boolean),
            container: inputs[`${formName}-container`].value,
        }, Object: {},
    };
}

function buildObject(form) {
    if(!form) return;
    const inputs = form.elements
    const formName = form.id
    return {
        Object: {
            name: inputs[`${formName}-name`].value,
            description: inputs[`${formName}-description`].value,
            tags: inputs[`${formName}-tags`].value.split(',').map(t => t.trim()).filter(Boolean),
            container: inputs[`${formName}-container`].value,
        },
    }
}

// --- Event Listeners ---

document.getElementById('logoutBtn').addEventListener('click', async () => {
    try {
        const response = await api.logout();
        localStorage.removeItem('username');
        window.location.replace('./login.html');
    }catch(err) {
        console.log(err);
        ui.showNotification('Error: '+ err.message, 'error');
    }
});

document.getElementById('detailSectionBtn').addEventListener('click', () => {
    ui.hideDetailsPanel()
})

document.getElementById('newBtn').addEventListener('click', () => {
    ui.showCreatePanel();
})

document.getElementById('createItemBtn').addEventListener('click', async () => {
    const data = buildItem(document.getElementById('createForm'));
    if (!data) {
        return;
    }
    try {
        const response = api.createItem(data);
        ui.showNotification('Item created', 'success');
    }catch(err) {
        ui.showNotification('Error: ' + err.message, 'error');
    }

})

document.getElementById('createObjectBtn').addEventListener('click', async () => {
    const data = buildObject(document.getElementById('createForm'));
    if (!data) {
        return;
    }
    try {
        const response = api.createObject();
        ui.showNotification( `New item created`, 'success');
    }catch(err) {
        ui.showNotification('Error: ' + err.message, 'error');
    }
})

document.getElementById('closePanel').addEventListener('click', () => {
    console.log("HERE")
    ui.hideCreatePanel();
})

document.getElementById('deleteBtn').addEventListener('click', () => {
    const name = document.getElementById('D-name').placeholder;
    try {
        if (!document.getElementById('D-quantity')) {
            const response =  api.deleteObject(name);
            ui.showNotification(`Object ${name} deleted`, 'warning');
        }else{
            const response = api.deleteItem(document.getElementById('D-id').placeholder);
            ui.showNotification(`Item ${name} deleted`, 'warning');
        }

    }catch(err) {
        ui.showNotification('Error: ' + err.message, 'error');
    }

})

document.getElementById('searchBar').addEventListener('input', () => {
    ui.showSearchResults();
})