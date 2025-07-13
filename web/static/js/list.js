document.addEventListener('click', event => {
    const remove = event.target.closest('.remove-option');
    if (remove) {
        let action = 'unlike';
        let messageText = "Succesfully unliked.";
        if (window.location.pathname.includes('favorites')) {
            action = 'unfave';
            messageText = "Succesfully removed.";
        }
        const url = `/video/${remove.dataset.id}/${action}`;
        postData(url).then(response => {
            if (response.ok) {
                remove.parentElement.remove();
                setAlert(messageText);
            } else {
                setAlert("Something went wrong!");
            }
        });
    }
});