const needLogin = document.querySelector('.need-login');

document.addEventListener('click', event => {
    if (event.target.closest('[data-auth]')) {
        needLogin.style.display = 'block';
    }

    else if (needLogin.style.display === 'block' && event.target != needLogin) {
        needLogin.removeAttribute('style');
    }
});