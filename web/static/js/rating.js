document.querySelectorAll('.rate-widget').forEach(widget => {
    const rateDialog = widget.querySelector('.rate-dialog');
    const rateBtnOpen = widget.querySelector('.btn-open-rate');
    const rateBtnClose = widget.querySelector('.btn-close-rate');
    const rateBtnSubmit = widget.querySelector('.btn-submit-rate');
    const bigStarValue = widget.querySelector('.rating-big-star-value');
    const avgValDisplay = widget.querySelector('.avg-val');
    const rateURL = widget.dataset.url || `${window.location.pathname}/rate`;

    let currentRating = null;

    rateBtnOpen.addEventListener('click', () => { rateDialog?.showModal(); });
    rateBtnClose.addEventListener('click', () => rateDialog?.close());

    // Handle interactive star adjustments inside the modal
    widget.querySelectorAll('input[type="radio"]').forEach(input => {
        input.addEventListener('change', (e) => {
            currentRating = Number(e.target.value);
            bigStarValue.textContent = currentRating;
            if (rateBtnSubmit) rateBtnSubmit.disabled = false;
        });
    });

    // Handle submission workflow via the dedicated Rate CTA
    if (rateBtnSubmit) {
        rateBtnSubmit.addEventListener('click', async () => {
            if (!currentRating) return;

            rateDialog.close();

            try {
                const res = await postData(rateURL, { 'rating': currentRating });
                if (!res.ok) throw new Error(`HTTP error! Status: ${res.status}`);

                const data = await res.json();

                // Update the distinct Video Rating value panel dynamically
                if (avgValDisplay && data.avg_rating) {
                    avgValDisplay.textContent = data.avg_rating;
                }

                // Transition the User Rating trigger state to confirm submission visually
                rateBtnOpen.innerHTML = `<span class="rating-user-star " style="color: #5799ef;">★</span> ${currentRating}`;
            } catch (error) {
                console.error("Failed to fetch or parse JSON:", error);
                setAlert("Something went wrong!");
            }
        });
    }
});