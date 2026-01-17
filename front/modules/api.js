const API_BASE_URL = location.origin
const fetchOptions = {credentials: 'same-origin'}

async function callApi (endpoint, method = 'GET', body = null) {
    const options = {...fetchOptions, method}
    if (body){
        options.headers = {'Content-Type': 'application/json', ...options.headers}
        options.body = JSON.stringify(body)
    }
    const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
    if (!response.ok) {
        const errorDetail = await response.text()
        throw new Error(`${response.status}: ${response.statusText}`, {cause: response.status});
    }
    if (response.headers.get('content-type') === 'application/json') {
        return response.json();
    }
    return response.text();
}

export const api = {
    getItems: ()=> callApi('/items'),

    getItem: (id) => callApi(`/item/${id}`),
    createItem: (itemData) => callApi('/item/0', 'PUT', itemData),
    updateItem: (id, itemData) => callApi(`/item/${id}`, 'POST', itemData),
    deleteItem: (id) => callApi(`/item/${id}`, 'DELETE'),

    getObject: (id) => callApi(`/object/${encodeURIComponent(id)}`),
    createObject: (objectData) => callApi(`/object/0`, 'PUT', objectData),
    updateObject: (id, objectData) => callApi(`/object/${encodeURIComponent(id)}`, 'POST', objectData),
    deleteObject: (id) => callApi(`/object/${encodeURIComponent(id)}`, 'DELETE'),
    
    logout: (username) => callApi(`/logout/${username}`, 'POST'),

    search: (keyword) => callApi(`/search/${keyword}`),
    autocomplete: (keyword) => callApi(`/autocomplete/${keyword}`),
}