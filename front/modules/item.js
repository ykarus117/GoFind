import {api} from './api.js';
import {ui} from "./ui.js";

export class Item {
    constructor(id) {
        if (id < 0) {
            throw Error('id must be a positive integer');
        }
        this.id = id;
    }

    async init() {
        try {
            const resp = await api.getItem(this.id);
            this.name = resp.name;
            this.description = resp.description;
            this.tags = resp.tags;
            this.quantity = resp.quantity;
            this.container = resp.container;
            return this;
        } catch (e) {
            console.error('Failed to initialize item:', e);
            throw e;
        }
    }

    render(containerId, mode = 'view') {
        const container = document.getElementById(containerId);

        let templateId;
        if (mode === 'edit') {
            templateId = 'ItemTemplateEdit';
        } else {
            templateId = 'ItemTemplateView';
        }

        const template = document.getElementById(templateId);
        const clone = template.content.cloneNode(true);

        if (mode === 'view') {
            clone.querySelector('h2').textContent = this.name;
            clone.querySelector('button.edit').onclick = () => this.render(containerId, 'edit');
            clone.querySelector('.description').textContent = this.description !== "" ? this.description : "no description";
            clone.querySelector('.tags').textContent = this.tags.join(', ') !== "" ? this.tags : "[]";
            clone.querySelector('.quantity').textContent = this.quantity;
            clone.querySelector('.container').textContent = this.container !== "" ? this.container : "loose item";

        } else if (mode === 'edit') {
            const name = clone.querySelector('input[name="name"]');
            const description = clone.querySelector('input[name="description"]');
            const tags = clone.querySelector('input[name="tags"]');
            const quantity = clone.querySelector('input[name="quantity"]');
            const container = clone.querySelector('input[name="container"]');

            name.placeholder = this.name;
            description.placeholder = this.description !== "" ? this.description : 'no description available';
            tags.placeholder = this.tags.join(",").length > 1 ? this.tags.join(",") : '[]';
            quantity.placeholder = this.quantity
            container.placeholder = this.container !== "" ? this.container : 'loose item';

            // Save, update and switch back
            clone.querySelector('.save').onclick = () => {
                this.name = name.value !== "" ? name.value : this.name;

                this.description = description.value !== "" ? description.value : this.description;

                this.tags = tags.value !== "" ? tags.value.split(',').map(t => t.trim()).filter(Boolean) : this.tags

                this.quantity = parseInt(quantity.value) || this.quantity;

                this.container = container.value !== "" ? container.value : this.container;

                this.update().then(r =>
                    this.render(containerId, 'view')
                )
            };

            clone.querySelector('.delete').onclick = () => {
                try {
                    const resp = this.delete()
                }catch(e){
                    ui.showNotification("Error while deleting item" + e, 'error');
                }
            }
        }

        container.innerHTML = '';
        container.appendChild(clone);
    }

    async delete() {
        try {
            await api.deleteItem(this.id)
            ui.showNotification(`'${this.name}' deleted`, 'warning');
            ui.hideDetailsPanel();
        }catch (e) {
            ui.showNotification("Error while deleting item" + e, "error");
        }
    }

    async update() {
        try {
            await api.updateItem(this.id, {Item: this})
            ui.showNotification(`${this.name} updated successfully.`, 'success');
        }catch (e) {
            ui.showNotification("Error while updating item" + e, "error");
        }
    }
}