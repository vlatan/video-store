// Edit Description
const editDescriptionButton = document.querySelector('.edit-description');
const unpublishedWrap = document.querySelector('.unpublished');
const videoDescription = document.querySelector('.video-description');
let originalDescription = videoDescription.innerText;


editDescriptionButton.addEventListener('click', () => {
    videoDescription.classList.toggle('editing');
    if (editDescriptionButton.innerText === "Edit Desc") {
        videoDescription.contentEditable = true;
        videoDescription.focus();
        editDescriptionButton.innerText = "Save Desc";
    } else {
        const url = `${window.location.pathname}edit`;
        postData(url, { description: videoDescription.innerText })
            .then(response => {
                if (response.ok) {
                    originalDescription = videoDescription.innerText;
                    if (unpublishedWrap) {
                        unpublishedWrap.replaceWith(videoDescription);
                    }
                    setAlert("Description succesfully updated!");
                } else {
                    videoDescription.innerText = originalDescription;
                    setAlert("Sorry, something went wrong!");
                }
            });
        videoDescription.removeAttribute('contenteditable');
        videoDescription.blur();
        editDescriptionButton.innerText = "Edit Desc";
    }
});


const resetDescription = (description, button) => {
    description.removeAttribute('contenteditable');
    button.innerText = "Edit Desc";
    description.classList.remove('editing');
    description.innerText = originalDescription;
};


videoDescription.addEventListener('keydown', event => {
    if (event.code === 'Escape') {
        resetDescription(videoDescription, editDescriptionButton);
    }
});


document.addEventListener('click', event => {
    if (event.target !== editDescriptionButton &&
        event.target !== videoDescription &&
        videoDescription.classList.contains('editing')) {
        resetDescription(videoDescription, editDescriptionButton);
    }
});