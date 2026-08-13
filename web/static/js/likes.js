const setFaveStatus = (action = "") => {
    let saves = document.querySelector('[data-saves]');
    let text = saves?.textContent.trim();
    if (action === 'fave') {
        text = 'Saved';
        let saved = document.createElement('span');
        saved.innerHTML = '&#10003;';
        saved.setAttribute('data-saved', '');
        saves?.before(saved);
    } else {
        text = 'Save';
        saves = document.querySelector('[data-saved]');
        saves?.remove();
    }
    if (saves) saves.textContent = text;
}

const setLikeCounter = (action = "") => {
    let likes = document.querySelector('[data-likes]');
    let text = String(likes?.textContent.trim());
    let counter = parseInt(text.charAt(0));
    if (isNaN(counter)) {
        counter = 0;
    }
    if (action === 'like') {
        let liked = document.createElement('span');
        liked.innerHTML = '&#10003;';
        liked.setAttribute('data-liked', '');
        likes?.before(liked);
        counter += 1;
    } else {
        counter -= 1;
        document.querySelector('[data-liked]')?.remove();
    }
    if (counter === 0) {
        text = 'Like';
    } else if (counter === 1) {
        text = '1 Like';
    } else {
        text = `${counter} Likes`;
    }

    likes = document.querySelector('[data-likes]')
    if (likes) likes.textContent = text;
};

const listenForAction = async (action = "", videoId = "") => {
    document.addEventListener('click', async event => {
        if (!(event.target instanceof HTMLElement)) return;
        const actionElement = event.target.closest(`.${action}`);
        if (!actionElement) return;
        actionElement.classList.toggle(`${action}-no`);
        actionElement.classList.toggle(`${action}-yes`);
        const unAction = actionElement.classList.contains(`${action}-no`);
        const currentAction = unAction ? `un${action}` : action;
        const url = `/api/video/${videoId}/${currentAction}`;
        try {
            const res = unAction ? await deleteData(url) : await postData(url);
            if (!res.ok) throw new Error(`HTTP error! Status: ${res.status}`);
            if (currentAction.includes('like')) { setLikeCounter(currentAction); return; }
            setFaveStatus(currentAction);
        } catch (error) {
            actionElement.classList.toggle(`${action}-no`);
            actionElement.classList.toggle(`${action}-yes`);
            console.error("Failed to fetch response:", error);
            setAlert("Sorry, could not record that action!")
        }
    });
};

(() => {
    const socialButtons = document.getElementById('social-buttons');
    const videoId = socialButtons?.dataset.videoId;
    listenForAction('like', videoId);
    listenForAction('fave', videoId);
})();