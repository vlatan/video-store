const setFaveStatus = action => {
    let text = 'Save';
    if (action === 'fave') {
        text = '&#10003; Saved';
    }
    document.querySelector('[data-status]').innerHTML = text;
};

const setLikeCounter = action => {
    let likes = document.querySelector('[data-likes]');
    let text = likes.textContent.trim();
    let counter = parseInt(text.charAt(0));
    if (isNaN(counter)) {
        counter = 0;
    }
    if (action === 'like') {
        let liked = document.createElement('span');
        liked.innerHTML = '&#10003;';
        liked.setAttribute('data-liked', '');
        likes.before(liked);
        counter += 1;
    } else {
        counter -= 1;
        document.querySelector('[data-liked]').remove();
    }
    if (counter === 0) {
        text = 'Like';
    } else if (counter === 1) {
        text = '1 Like';
    } else {
        text = `${counter} Likes`;
    }
    document.querySelector('[data-likes]').textContent = text;
};

const performAction = async action => {
    const url = `${window.location.pathname}${action}`;
    return await fetch(url, { method: 'POST' });
};

const listenForAction = (event, action) => {
    const actionElement = event.target.closest(`.${action}`);
    if (actionElement) {
        actionElement.classList.toggle(`${action}-no`);
        actionElement.classList.toggle(`${action}-yes`);
        let currentAction = action;
        if (actionElement.classList.contains(`${action}-no`)) {
            currentAction = `un${action}`;
        }
        performAction(currentAction)
            .then(response => {
                if (response.ok) {
                    if (currentAction.includes('like')) {
                        setLikeCounter(currentAction);
                    } else {
                        setFaveStatus(currentAction);
                    }
                }
            });
    }
};

document.addEventListener('click', event => {
    listenForAction(event, 'like');
    listenForAction(event, 'fave');
});