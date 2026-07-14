document.querySelectorAll('.rate-widget').forEach(widget => {
    const rateDialog = widget.querySelector('.rate-dialog');
    const rateBtnOpen = widget.querySelector('.btn-open-rate');
    const rateBtnClose = widget.querySelector('.btn-close-rate');
    const rateBtnSubmit = widget.querySelector('.btn-submit-rate');
    const bigStarValue = widget.querySelector('.rating-big-star-value');
    const avgValDisplay = widget.querySelector('.rating-avg-val');
    const rateURL = widget.dataset.url || `/api${window.location.pathname}/rate`;

    let currentRating = null;
    let selectedStar = null;
    const bigStarOriginalTextContent = bigStarValue.textContent;

    rateBtnOpen.addEventListener('click', () => { rateDialog.showModal(); });
    rateBtnClose.addEventListener('click', () => {
        rateDialog.close()
        rateBtnSubmit.disabled = true;
        if (selectedStar) selectedStar.checked = false;
        bigStarValue.textContent = bigStarOriginalTextContent;
    });

    // Handle interactive star adjustments inside the modal
    widget.querySelectorAll('input[type="radio"]').forEach(input => {
        input.addEventListener('change', (e) => {
            selectedStar = e.target;
            currentRating = Number(selectedStar.value);
            bigStarValue.textContent = currentRating;
            rateBtnSubmit.disabled = false;
        });
    });

    // Handle submission workflow via the dedicated Rate CTA
    if (!rateBtnSubmit) return;
    rateBtnSubmit.addEventListener('click', async () => {
        if (!currentRating) return;
        rateDialog.close();
        rateBtnSubmit.disabled = true;
        if (selectedStar) selectedStar.checked = false;
        bigStarValue.textContent = bigStarOriginalTextContent;

        try {
            const res = await postData(rateURL, { 'rating': currentRating });
            if (!res.ok) throw new Error(`HTTP error! Status: ${res.status}`);
            const data = await res.json();

            // Update the distinct Video Rating value panel dynamically
            if (avgValDisplay && data.avg_rating) {
                avgValDisplay.textContent = data.avg_rating;
            }

            // Transition the User Rating trigger state to confirm submission visually
            rateBtnOpen.innerHTML = `<span class="rating-user-star">&#9733;</span> ${currentRating}`;
        } catch (error) {
            console.error("Failed to fetch or parse JSON:", error);
            setAlert("Something went wrong!");
        }
    });
});