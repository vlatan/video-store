document.addEventListener('click', async (event) => {
    const remove = event.target.closest('.remove-option');
    if (!remove) return;
    let action = 'unlike';
    let messageText = "Succesfully unliked.";
    if (window.location.pathname.includes('favorites')) {
        action = 'unfave';
        messageText = "Succesfully removed.";
    }
    const url = `/video/${remove.dataset.id}/${action}`;
    try {
        const res = await postData(url);
        if (!res.ok) throw new Error(`HTTP error! Status: ${res.status}`);
        remove.parentElement.remove();
        setAlert(messageText);
    } catch (error) {
        console.error("Failed to fetch response:", error);
        setAlert("Something went wrong!");
    }
});