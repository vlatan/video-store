// Sleep time expects milliseconds
const sleep = time => {
    return new Promise(resolve => setTimeout(resolve, time));
};

// Send POST request to backend
const postData = async (url = '', data = {}) => {
    // Create headers object
    const headers = new Headers();
    headers.append("Content-Type", "application/json");

    // If CSRF Token send with the POST request
    let csrfToken = document.getElementsByName("gorilla.csrf.Token");
    if (csrfToken) {
        headers.append("X-CSRF-Token", csrfToken[0].value);
    }

    const response = await fetch(url, {
        method: 'POST',
        headers: headers,
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
    const catDropContent = document.querySelector('.categories-dropdown-content');
    if (catDropContent) {
        const catDropped = catDropContent.classList.contains('categories-show-dropdown');
        const hamburgrIconClicked = event.target.closest('.hamburger-icon');
        const closeIconClicked = event.target.closest('.categories-close-icon');
        const catMenuClicked = event.target.closest('.categories-show-dropdown');
        if (!catDropped && hamburgrIconClicked) {
            catDropContent.classList.add('categories-show-dropdown');
        } else if (!catMenuClicked || closeIconClicked) {
            catDropContent.classList.remove('categories-show-dropdown');
        }
    }

    // Mobile search form
    const searchForm = document.getElementById('searchForm');
    const logo = document.querySelector('a.logo');
    const searchIcon = document.querySelector('.search-button-mobile');
    const hamburgerIcon = document.querySelector('.hamburger-icon');
    const dropdowns = document.querySelectorAll('.dropdown');
    const arrow = document.querySelector('button.search-arrow')
    const arrowClicked = event.target.closest('button.search-arrow');
    const outsideFormClicked = !event.target.closest('#searchForm');
    if (event.target.closest('.search-button-mobile')) {
        arrow.style.display = "block";
        searchForm.style.display = 'flex'
        logo.style.display = "none";
        searchIcon.style.display = "none";
        hamburgerIcon.style.display = "none";
        for (const dropdown of dropdowns) {
            dropdown.style.display = "none";
        }
    } else if (arrowClicked || outsideFormClicked) {
        arrow.removeAttribute('style');
        searchForm.removeAttribute('style');
        logo.removeAttribute('style');
        searchIcon.removeAttribute('style');
        hamburgerIcon.removeAttribute('style');
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


// ==========================================================================
// Form
// ==========================================================================

const formInputs = document.querySelectorAll('.form-input');
const formSubmit = document.querySelector('.form-button');
const formSpinner = document.querySelector('.submit-spinner');

if (formSubmit) {
    formSubmit.addEventListener('click', () => {

        // Check if all required inputs have values
        let ok = true
        for (const inputElement of formInputs) {
            if (inputElement.required && inputElement.value.trim() === '') {
                ok = false
            }
        }

        if (ok) {
            if (formSpinner) {
                formSpinner.classList.add('show')
            }
        }
    });
}
