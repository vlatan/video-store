// sleep time expects milliseconds
const sleep = time => {
    return new Promise(resolve => setTimeout(resolve, time));
};

// Send POST request to backend
const postData = async (url = '', data = {}) => {
    const response = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });
    return response;
};

// Send GET request to backend
const getData = async (url = "", page = 2) => {
    // set page query param to url
    // https://developer.mozilla.org/en-US/docs/Web/API/URLSearchParams/set
    const currenURL = new URL(url);
    const params = new URLSearchParams(currenURL.search);
    params.set("page", page);
    currenURL.search = params.toString();
    return await fetch(currenURL.toString());
};

// Set alert message
const setAlert = message => {
    const alert = document.createElement('div');
    alert.classList.add('alert');
    alert.innerText = message;
    document.getElementById('footer').prepend(alert);
    sleep(2000).then(() => {
        alert.remove();
    });
};

document.addEventListener('click', event => {

    // Modals
    document.querySelectorAll('[data-modal]').forEach(element => {
        const modalName = element.dataset.modal;
        const modalBody = document.querySelector(`[data-body="${modalName}"]`);
        const closeModal = event.target.closest(`[data-close="${modalName}"]`);
        if (event.target.closest(`[data-modal="${modalName}"]`)) {
            modalBody.style.display = 'flex';
        } else if (event.target === modalBody || closeModal) {
            modalBody.removeAttribute('style');
        }
    });

    // User Profile Dropdown menu
    const dropContent = document.querySelector('.dropdown-content');
    if (dropContent) {
        const notDropped = !dropContent.classList.contains('show-dropdown');
        const usernameClicked = event.target.closest('.username');
        const deleteAccountClicked = event.target.closest('.delete-account');
        const menuNotClicked = !event.target.closest('.show-dropdown');
        if (notDropped && usernameClicked) {
            dropContent.classList.add('show-dropdown');
        } else if (deleteAccountClicked || menuNotClicked) {
            dropContent.classList.remove('show-dropdown');
        }
    }

    // Categories Dropdown menu
    const catDropContent = document.querySelector('.category-dropdown-content');
    const hamburgerIcon = document.querySelector('.hamburger-icon');
    if (catDropContent) {
        const catNotDropped = !catDropContent.classList.contains('category-show-dropdown');
        const categoriesClicked = event.target.closest('.categories');
        const catMenuNotClicked = !event.target.closest('.category-show-dropdown');
        if (catNotDropped && categoriesClicked) {
            catDropContent.classList.add('category-show-dropdown');
            hamburgerIcon.classList.add('hamburger-icon-change');
        } else if (catMenuNotClicked) {
            catDropContent.classList.remove('category-show-dropdown');
            hamburgerIcon.classList.remove('hamburger-icon-change');
        }
    }

    // Mobile search form
    const searchForm = document.getElementById('searchForm');
    const logo = document.querySelector('a.logo');
    const searchIcon = document.querySelector('.search-button-mobile');
    const dropdowns = document.querySelectorAll('.dropdown');
    const arrow = document.querySelector('button.search-arrow')
    const arrowClicked = event.target.closest('button.search-arrow');
    const outsideFormClicked = !event.target.closest('#searchForm');
    if (event.target.closest('.search-button-mobile')) {
        arrow.style.display = "block";
        searchForm.style.display = 'flex'
        logo.style.display = "none";
        searchIcon.style.display = "none";
        for (const dropdown of dropdowns) {
            dropdown.style.display = "none";
        }
    } else if (arrowClicked || outsideFormClicked) {
        arrow.removeAttribute('style');
        searchForm.removeAttribute('style');
        logo.removeAttribute('style');
        searchIcon.removeAttribute('style');
        for (const dropdown of dropdowns) {
            dropdown.removeAttribute('style');
        }
    }
});

// Cookies disclaimer
const acceptCookies = localStorage.getItem('acceptCookies');
const privacyPath = "/page/privacy/";
const currentPath = window.location.pathname;
if (currentPath !== privacyPath && acceptCookies !== 'true') {
    const snackbar = document.createElement('div');
    snackbar.classList.add('snackbar');
    document.getElementById('footer').after(snackbar);

    const snackbarLabel = document.createElement('div');
    snackbarLabel.classList.add('snackbar-label');
    snackbarLabel.innerText = "We serve cookies on this site to analyze traffic, \
    remember your preferences, and optimize your experience.";
    snackbar.appendChild(snackbarLabel);

    const snackbarActions = document.createElement('div');
    snackbarActions.classList.add('snackbar-actions');
    snackbar.appendChild(snackbarActions);

    const detailsLink = document.createElement('a');
    detailsLink.classList.add('cookies-button');
    detailsLink.href = privacyPath;
    detailsLink.target = '_blank';
    detailsLink.innerText = "More details";
    snackbarActions.appendChild(detailsLink);

    const buttonOK = document.createElement('button');
    buttonOK.classList.add('cookies-button');
    buttonOK.innerText = "OK";
    snackbarActions.appendChild(buttonOK);

    buttonOK.addEventListener('click', () => {
        localStorage.setItem('acceptCookies', true);
        snackbar.remove();
    });
}