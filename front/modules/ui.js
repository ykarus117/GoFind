const API_BASE_URL = location.origin;
const responseArea = document.getElementById('response');
const createPanel = document.getElementById('createPanel');
const itemPanel = document.getElementById('itemsPanel');
const detailsPanel = document.getElementById('detailsPanel');
const details = document.getElementById('details');
const notificationPanel = document.getElementById('notificationPanel');

let detailsPanelVisible = true

function removeNotificationCard(node) {
    node.classList.remove('show');
    node.classList.add('hide');
    setTimeout(() => {notificationPanel.removeChild(node)}, 1000);
}

export const ui = {
    populateDetails: (object) => {
        if (!object) return;

        if (!detailsPanelVisible) {
            ui.showDetailsPanel();
        }

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
    },

    showNotification(content, options) {
        let div = document.createElement('div');
        div.classList.add('notification');
        div.classList.add(options);
        div.classList.add('show');
        let text = document.createElement('p')

        text.innerText = content;
        div.append(text);
        notificationPanel.append(div);
        let timer = setTimeout(()=>{removeNotificationCard(div)}, 3500);
        div.addEventListener('mouseover', () => {
            clearTimeout(timer);
            timer = setTimeout(() => {removeNotificationCard(div)}, 2000);
        })
    },

    showCreatePanel: () => {
        createPanel.classList.remove('fade-out');
        createPanel.classList.add('fade-in');
    },

    hideCreatePanel: () => {
        createPanel.classList.remove('fade-in');
        createPanel.classList.add('fade-out');
    },

    showDetailsPanel: () => {
        detailsPanelVisible = true;
        detailsPanel.classList.remove('contract-left');
        detailsPanel.classList.add('expand-right');
    },

    hideDetailsPanel: () => {
        detailsPanelVisible = false;
        detailsPanel.classList.remove('expand-right');
        detailsPanel.classList.add('contract-left');
    },

    toggleDetails: () => {
        if (!detailsPanelVisible) {
            ui.hideDetailsPanel();
        }else{
            ui.showDetailsPanel();
        }
        detailsPanelVisible = !detailsPanelVisible
    },
}