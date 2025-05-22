// Edit Title
const editTitleButton = document.querySelector('.edit-title');
const videoTitle = document.querySelector('.video-title');
let originalTitle = videoTitle.innerText;


editTitleButton.addEventListener('click', () => {
    videoTitle.classList.toggle('editing');
    if (editTitleButton.innerText === "Edit Title") {
        videoTitle.contentEditable = true;
        videoTitle.focus();
        editTitleButton.innerText = "Save Title";
    } else {
        const url = `${window.location.pathname}edit`;
        postData(url, { title: videoTitle.innerText })
            .then(response => {
                if (response.ok) {
                    setAlert("Title succesfully edited!");
                    originalTitle = videoTitle.innerText;
                } else {
                    videoTitle.innerText = originalTitle;
                    setAlert("Sorry, something went wrong!");
                }
            });
        videoTitle.removeAttribute('contenteditable');
        videoTitle.blur();
        editTitleButton.innerText = "Edit Title";
    }
});


const resetTitle = (title, button) => {
    title.removeAttribute('contenteditable');
    button.innerText = "Edit Title";
    title.classList.remove('editing');
    title.innerText = originalTitle;
};


videoTitle.addEventListener('keydown', (event) => {
    if (event.code === 'Escape') {
        resetTitle(videoTitle, editTitleButton);
    }
});


document.addEventListener('click', (event) => {
    if (event.target !== editTitleButton &&
        event.target !== videoTitle &&
        videoTitle.classList.contains('editing')) {
        resetTitle(videoTitle, editTitleButton);
    }
});