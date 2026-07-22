
// ==========================================================================
// Sync All Checked Stars and Big Star Values
// ==========================================================================

const starEadios = document.querySelectorAll('input[name="rating"]');
const bigStarValues = document.querySelectorAll('.rating-big-star-value');

starEadios.forEach(radio => {
    radio.addEventListener('change', (event) => {
        if (!event.target.checked) return;

        const value = event.target.value;

        // Sync across all star sets (needed if they're in separate forms)
        starEadios.forEach(r => {
            if (r.value === value) r.checked = true;
        });

        // Update the displayed rating value
        bigStarValues.forEach(bsv => {
            bsv.textContent = value;
        });
    });
});


// ==========================================================================
// Update the Average Rating Display and the User Rating Button
// ==========================================================================

function updateRatingHTML(user_rating, avg_rating, rating_count) {

    const votesText = rating_count === 1 ? "vote" : "votes";
    const avgRatingHTML = `
        <div class="btn-open-post-dialog avg-rating-display">
            <span class="rating-global-star">&#9733;</span>
            <div class="rating-meta" itemprop="aggregateRating" itemscope
                itemtype="https://schema.org/AggregateRating">
                <meta itemprop="worstRating" content="1">
                <div class="rating-score">
                    <span class="rating-avg-val" itemprop="ratingValue">
                        ${avg_rating}
                    </span> / <span itemprop="bestRating">10</span>
                </div>
                <div class="rating-count">
                    <span class="rating-count-val" itemprop="ratingCount">
                        ${rating_count}
                    </span> ${votesText}
                </div>
            </div>
        </div>
    `;

    const avgRatingDisplay = document.querySelector('.avg-rating-display');
    const rateBtnOpen = document.querySelector('#btn-open-rate');

    // Transform the average display
    if (avgRatingDisplay) {
        avgRatingDisplay.outerHTML = avgRatingHTML;
    } else if (rateBtnOpen) {
        rateBtnOpen.insertAdjacentHTML('beforebegin', avgRatingHTML);
    }

    // Transform the user rating button
    rateBtnOpen.innerHTML = `<span class="rating-user-star">&#9733;</span> ${user_rating}`;
}



// ==========================================================================
// Ratings
// ==========================================================================

document.querySelectorAll('.rating-section').forEach(widget => {
    const rateDialog = widget.querySelector('#rate-dialog');
    const rateForm = widget.querySelector('.rate-form');
    const rateBtnOpen = widget.querySelector('#btn-open-rate');
    const rateBtnClose = widget.querySelector('#btn-close-rate');
    const rateBtnSubmit = widget.querySelector('.btn-submit-rate');

    rateBtnOpen.addEventListener('click', () => rateDialog.showModal());
    rateBtnClose.addEventListener('click', () => rateDialog.close());

    // Handle form submission
    rateForm.addEventListener('submit', async (event) => {
        event.preventDefault();
        const form = event.currentTarget;

        if (!form.checkValidity()) {
            form.reportValidity(); // shows the native browser bubble
            return;
        }

        const formData = new FormData(form);
        const data = Object.fromEntries(formData.entries());
        if (data.rating) data.rating = Number(data.rating);

        rateBtnSubmit.disabled = true;
        rateBtnSubmit.textContent = 'Posting...';
        rateDialog.close();

        try {
            const response = await postData(form.action, data);
            if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
            const result = await response.json();

            // Update the rating HTML
            updateRatingHTML(data.rating, result.avg_rating, result.rating_count)
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
        const form = event.currentTarget;
        clearError();

        if (!form.checkValidity()) {
            form.reportValidity(); // shows the native browser bubble
            return;
        }

        const formData = new FormData(form);
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
            const response = await postData(form.action, data);
            if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
            const result = await response.json();

            const card = document.createElement('div');
            card.className = 'review-card new-review';
            card.innerHTML = `
                <div class="review-header">
                    <h4>${result.author || 'Anonymous'} <span class="date-meta">Just now</span></h4>
                    <span class="stars-display">${getStarsHTML(data.rating)}</span>
                </div>
                <p class="review-headline"></p>
                <p class="review-content"></p>
            `;
            card.querySelector('.review-headline').textContent = data.headline;
            card.querySelector('.review-content').textContent = data.content;
            reviewsList.prepend(card);

            // Update the rating HTML
            updateRatingHTML(data.rating, result.avg_rating, result.rating_count)
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