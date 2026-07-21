// ==========================================================================
// Ratings
// ==========================================================================

document.querySelectorAll('.rating-section').forEach(widget => {
    const rateDialog = widget.querySelector('#rate-dialog');
    const rateForm = widget.querySelector('.rate-form');
    const rateBtnOpen = widget.querySelector('#btn-open-rate');
    const rateBtnClose = widget.querySelector('#btn-close-rate');
    const rateBtnSubmit = widget.querySelector('.btn-submit-rate');
    const bigStarValue = widget.querySelector('.rating-big-star-value');
    const bigStarOriginalTextContent = bigStarValue.textContent;

    rateBtnOpen.addEventListener('click', () => rateDialog.showModal());
    rateBtnClose.addEventListener('click', () => rateDialog.close());

    // Handle interactive star value in the big star
    widget.querySelectorAll('input[type="radio"]').forEach(input => {
        input.addEventListener('change', (event) => {
            currentRating = Number(event.target.value);
            bigStarValue.textContent = event.target.value;
        });
    });

    // Handle form submission
    rateForm.addEventListener('submit', async (event) => {
        event.preventDefault();

        if (!rateForm.checkValidity()) {
            rateForm.reportValidity(); // shows the native browser bubble
            return;
        }

        const formData = new FormData(event.currentTarget);
        const data = Object.fromEntries(formData.entries());
        if (data.rating) data.rating = Number(data.rating);

        rateBtnSubmit.disabled = true;
        rateBtnSubmit.textContent = 'Posting...';
        rateDialog.close();

        try {
            const response = await postData(event.currentTarget.action, data);
            if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
            const result = await response.json();

            let votesText = "votes";
            if (result.rating_count === 1) votesText = "vote";

            const avgRatingHTML = `
                <div class="btn-open-post-dialog avg-rating-display">
                    <span class="rating-global-star">&#9733;</span>
                    <div class="rating-meta" itemprop="aggregateRating" itemscope
                        itemtype="https://schema.org/AggregateRating">
                        <meta itemprop="worstRating" content="1">
                        <div class="rating-score">
                            <span class="rating-avg-val" itemprop="ratingValue">
                                ${result.avg_rating}
                            </span> / <span itemprop="bestRating">10</span>
                        </div>
                        <div class="rating-count">
                            <span class="rating-count-val" itemprop="ratingCount">
                                ${result.rating_count}
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
            rateBtnOpen.innerHTML = `<span class="rating-user-star">&#9733;</span> ${data.rating}`;

            // Reset the form and the big star
            bigStarValue.textContent = bigStarOriginalTextContent
            rateForm.reset()
        } catch (error) {
            console.error("Failed to fetch or parse JSON:", error);
            setAlert("Something went wrong!");
        } finally {
            rateBtnSubmit.disabled = false;
            rateBtnSubmit.textContent = 'Rate';
        }
    });
});


// ==========================================================================
// Reviews
// ==========================================================================

function getStarsHTML(rating) { return '★'.repeat(rating) + '☆'.repeat(10 - rating); }

document.querySelectorAll('.review-section').forEach(s => {
    const reviewDialog = s.querySelector('#review-dialog');
    const reviewForm = s.querySelector('.review-form');
    const reviewsList = s.querySelector('#reviews-list');
    const reviewOpenBtn = s.querySelector('#btn-open-review');
    const reviewCloseBtn = s.querySelector('#btn-close-review');
    const reviewSubmitBtn = s.querySelector('#submit-review');
    const reviewError = s.querySelector('#review-error');

    const showError = (msg) => {
        reviewError.textContent = msg;
        reviewError.hidden = false;
    };
    const clearError = () => {
        reviewError.textContent = '';
        reviewError.hidden = true;
    };

    reviewOpenBtn.addEventListener('click', () => reviewDialog.showModal());
    reviewCloseBtn.addEventListener('click', () => reviewDialog.close());

    reviewForm.addEventListener('submit', async (event) => {
        event.preventDefault();
        clearError();

        if (!reviewForm.checkValidity()) {
            reviewForm.reportValidity(); // shows the native browser bubble
            return;
        }

        const formData = new FormData(event.currentTarget);
        const headline = (formData.get('headline') || '').trim();
        const content = (formData.get('content') || '').trim();
        const rating = (formData.get('rating') || '').trim();

        if (!headline || !content || !rating) {
            showError('Please fill in the required fields');
            return;
        }

        const data = Object.fromEntries(formData.entries());
        if (data.rating) data.rating = Number(data.rating);

        reviewSubmitBtn.disabled = true;
        reviewSubmitBtn.textContent = 'Posting...';
        reviewDialog.close();

        try {
            const response = await postData(event.currentTarget.action, data);
            if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
            const result = await response.json();

            const card = document.createElement('div');
            card.className = 'review-card new-review';
            card.innerHTML = `
                <div class="review-header">
                    <h4>${result.author || 'Anonymous'} <span class="date-meta">Just now</span></h4>
                    <span class="stars-display">${getStarsHTML(parseInt(result.rating || formData.get('rating')))}</span>
                </div>
                <p class="review-headline"></p>
                <p class="review-content"></p>
            `;
            card.querySelector('.review-headline').textContent = result.headline || formData.get('headline');
            card.querySelector('.review-content').textContent = result.content || formData.get('content');
            reviewsList.prepend(card);

            // Reset the form
            reviewForm.reset()
        } catch (err) {
            console.error("Failed to fetch or parse JSON:", err);
            setAlert("Something went wrong!");
        } finally {
            reviewSubmitBtn.disabled = false;
            reviewSubmitBtn.textContent = 'Post Review';
        }
    });
});


// ==========================================================================
// Helpers
// ==========================================================================

function getStarsHTML(rating) { return '★'.repeat(rating) + '☆'.repeat(10 - rating); }