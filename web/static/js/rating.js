document.querySelectorAll('.rate-widget').forEach(widget => {
    const userRatingColumn = widget.querySelector('#user-rating-column');
    const rateDialog = widget.querySelector('#rate-dialog');
    const rateBtnOpen = widget.querySelector('.btn-open-rate');
    const rateBtnClose = widget.querySelector('.btn-close-rate');
    const rateBtnSubmit = widget.querySelector('.btn-submit-rate');
    const bigStarValue = widget.querySelector('.rating-big-star-value');
    const rateURL = widget.dataset.url || `/api${window.location.pathname}/rate`;

    let currentRating = null;
    let selectedStar = null;
    const bigStarOriginalTextContent = bigStarValue.textContent;

    rateBtnOpen.addEventListener('click', () => rateDialog.showModal());
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

            let votesText = "votes";
            if (data.rating_count === 1) votesText = "vote";

            const avgRatingHTML = `
                <div class="rating-column" id="average-rating-column">
                    <span class="rating-column-label">AVG RATING</span>
                    <div class="rating-display">
                        <span class="rating-global-star">&#9733;</span>
                        <div class="rating-meta" itemprop="aggregateRating" itemscope
							itemtype="https://schema.org/AggregateRating">
                            <meta itemprop="worstRating" content="1">
                            <div class="rating-score">
                                <span class="rating-avg-val" itemprop="ratingValue">
                                    ${data.avg_rating}
                                </span> / <span itemprop="bestRating">10</span>
                            </div>
                            <div class="rating-count">
                                <span class="rating-count-val" itemprop="ratingCount">
                                    ${data.rating_count}
                                </span> ${votesText}
                            </div>
                        </div>
                    </div>
                </div>
            `;

            // Replace or insert average rating
            const avgRatingColumn = widget.querySelector('#average-rating-column');
            if (avgRatingColumn) {
                avgRatingColumn.outerHTML = avgRatingHTML;
            } else if (userRatingColumn) {
                userRatingColumn.insertAdjacentHTML('beforebegin', avgRatingHTML);
            }

            // Transform the user rating button
            rateBtnOpen.innerHTML = `<span class="rating-user-star">&#9733;</span> ${currentRating}`;
        } catch (error) {
            console.error("Failed to fetch or parse JSON:", error);
            setAlert("Something went wrong!");
        }
    });
});