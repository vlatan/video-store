// ==========================================================================
// Ratings
// ==========================================================================

document.querySelectorAll('.rating-section').forEach(widget => {
    const rateDialog = widget.querySelector('#rate-dialog');
    const rateBtnOpen = widget.querySelector('#btn-open-rate');
    const rateBtnClose = widget.querySelector('#btn-close-rate');
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
        rateBtnSubmit.disabled = true;
        if (selectedStar) selectedStar.checked = false;
        bigStarValue.textContent = bigStarOriginalTextContent;
        rateDialog.close();

        try {
            const res = await postData(rateURL, { 'rating': currentRating });
            if (!res.ok) throw new Error(`HTTP error! Status: ${res.status}`);
            const data = await res.json();

            let votesText = "votes";
            if (data.rating_count === 1) votesText = "vote";

            const avgRatingHTML = `
                <div class="btn-open-post-dialog avg-rating-display">
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
            `;

            // Replace or insert average rating
            const avgRatingDisplay = widget.querySelector('.avg-rating-display');
            if (avgRatingDisplay) {
                avgRatingDisplay.outerHTML = avgRatingHTML;
            } else if (rateBtnOpen) {
                rateBtnOpen.insertAdjacentHTML('beforebegin', avgRatingHTML);
            }

            // Transform the user rating button
            rateBtnOpen.innerHTML = `<span class="rating-user-star">&#9733;</span> ${currentRating}`;
        } catch (error) {
            console.error("Failed to fetch or parse JSON:", error);
            setAlert("Something went wrong!");
        }
    });
});


// ==========================================================================
// Reviews
// ==========================================================================

document.querySelectorAll('.review-section').forEach(s => {
    const reviewDialog = s.querySelector('#review-dialog');
    const reviewForm = s.querySelector('#review-form');
    const reviewsList = s.querySelector('#reviews-list');
    const reviewOpenBtn = s.querySelector('#btn-open-review');
    const reviewCloseBtn = s.querySelector('#btn-close-review');
    const reviewSubmitBtn = s.querySelector('#submit-review');

    reviewOpenBtn.addEventListener('click', () => reviewDialog.showModal());
    reviewCloseBtn.addEventListener('click', () => reviewDialog.close());

    reviewForm.addEventListener('submit', async (event) => {
        event.preventDefault();
        reviewSubmitBtn.disabled = true;
        reviewSubmitBtn.textContent = 'Posting...';
        reviewDialog.close();

        const formData = new FormData(event.currentTarget);

        try {
            const response = await fetch(this.action, { method: 'POST', body: formData });
            if (!response.ok) throw new Error();
            const data = await response.json();

            const card = document.createElement('div');
            card.className = 'review-card new-review';
            card.innerHTML = `
                <div class="review-header">
                    <h4>${data.author || 'Anonymous'} <span class="date-meta">Just now</span></h4>
                    <span class="stars-display">${getStarsHTML(parseInt(data.rating || formData.get('rating')))}</span>
                </div>
                <p class="content"></p>
            `;
            card.querySelector('.content').textContent = data.text || formData.get('text');

            reviewsList.prepend(card);
            event.currentTarget.reset()
        } catch (err) {
            console.error("Failed to fetch or parse JSON:", err);
            setAlert("Something went wrong!");
        } finally {
            reviewSubmitBtn.disabled = false;
            reviewSubmitBtn.textContent = 'Post Review';
        }
    });
});
