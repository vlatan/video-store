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

const listenForAction = async (event, action) => {
    const actionElement = event.target.closest(`.${action}`);
    if (!actionElement) return;
    actionElement.classList.toggle(`${action}-no`);
    actionElement.classList.toggle(`${action}-yes`);
    let currentAction = action;
    if (actionElement.classList.contains(`${action}-no`)) currentAction = `un${action}`;
    const url = `/api${window.location.pathname}${currentAction}`;
    try {
        const res = await postData(url);
        if (!res.ok) throw new Error(`HTTP error! Status: ${res.status}`);
        if (currentAction.includes('like')) { setLikeCounter(currentAction); return; }
        setFaveStatus(currentAction);
    } catch (error) {
        actionElement.classList.toggle(`${action}-no`);
        actionElement.classList.toggle(`${action}-yes`);
        console.error("Failed to fetch response:", error);
        setAlert("Sorry, could not record that action!")
    }
};

document.addEventListener('click', event => {
    listenForAction(event, 'like');
    listenForAction(event, 'fave');
});