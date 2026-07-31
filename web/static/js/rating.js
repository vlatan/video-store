
// ==========================================================================
// Convert Review Dates into Local User Time
// ==========================================================================

document.querySelectorAll('.review-date').forEach(element => {
    const rawUtcString = element.getAttribute('data-utc-time');
    if (!rawUtcString) return;
    const dateObj = new Date(rawUtcString);
    element.textContent = dateObj.toLocaleDateString();
});


// ==========================================================================
// Sync All Checked Stars and Big Star Values
// ==========================================================================

const starRadios = document.querySelectorAll('input[name="rating"]');
const bigStarValues = document.querySelectorAll('.rating-big-star-value');

// Track currently checked value
let selectedValue = "?";
const checked = document.querySelector('input[name="rating"]:checked');
if (checked instanceof HTMLInputElement) {
    selectedValue = checked.value;
}

// Update the displayed rating value
const updateBigStar = (val = "?") => {
    bigStarValues.forEach(bsv => {
        bsv.textContent = val;
    });
};

starRadios.forEach(radio => {
    if (!(radio instanceof HTMLInputElement)) return;

    radio.addEventListener('change', () => {
        if (!radio.checked) return;
        selectedValue = radio.value;

        // Sync across all star sets (needed if they're in separate forms)
        starRadios.forEach(r => {
            if (r instanceof HTMLInputElement && r.value === selectedValue) {
                r.checked = true;
            }
            updateBigStar(selectedValue);
        });
    });

    // Target the visible label (or fall back to input)
    const hoverTarget = radio.closest('label') || radio;

    // Hover preview
    hoverTarget.addEventListener('mouseenter', () => {
        updateBigStar(radio.value);
    });

    // Reset to checked value on mouse leave
    hoverTarget.addEventListener('mouseleave', () => {
        updateBigStar(selectedValue);
    });
});


// ==========================================================================
// Update the Average Rating Display and the User Rating Button
// ==========================================================================

function updateRatingHTML(user_rating = 0, avg_rating = 0, rating_count = 0) {

    // Illegal values
    if (user_rating <= 0 || user_rating > 10) return;

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
    if (rateBtnOpen) {
        rateBtnOpen.innerHTML = `<span class="rating-user-star">&#9733;</span> ${user_rating}`;
    }
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
    let originalrateBtnSubmitText = rateBtnSubmit?.textContent.trim() || 'Rate';

    if (!(rateDialog instanceof HTMLDialogElement)) return;
    if (!(rateForm instanceof HTMLFormElement)) return;
    if (!(rateBtnSubmit instanceof HTMLButtonElement && rateBtnSubmit.type === 'submit')) return;

    rateBtnOpen?.addEventListener('click', () => rateDialog.showModal());
    rateBtnClose?.addEventListener('click', () => rateDialog.close());

    // Handle form submission
    rateForm?.addEventListener('submit', async (event) => {
        event.preventDefault();
        const form = event.currentTarget;
        if (!(form instanceof HTMLFormElement)) return;

        if (!form.checkValidity()) {
            form.reportValidity(); // shows the native browser bubble
            return;
        }

        const formData = new FormData(form);
        const payload = {
            ...Object.fromEntries(formData.entries()),
            rating: Number(formData.get('rating') || 0)
        };

        rateBtnSubmit.disabled = true;
        rateBtnSubmit.textContent = 'Posting...';
        rateDialog.close();

        try {
            const response = await postData(form.action, payload);
            if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
            const result = await response.json();

            // Update the rating HTML
            updateRatingHTML(payload.rating, result.avg_rating, result.rating_count)

            // The request went through, change the button text
            originalrateBtnSubmitText = "Update";
        } catch (error) {
            console.error("Failed to fetch or parse JSON:", error);
            setAlert("Something went wrong!");
        } finally {
            rateBtnSubmit.disabled = false;
            rateBtnSubmit.textContent = originalrateBtnSubmitText;
        }
    });
});


// ==========================================================================
// Reviews
// ==========================================================================

document.querySelectorAll('.review-section').forEach(s => {
    const reviewDialog = s.querySelector('#review-dialog');
    const reviewForm = s.querySelector('.review-form');
    const reviewsList = s.querySelector('#reviews-list');
    const reviewOpenBtn = s.querySelector('#btn-open-review');
    const reviewCloseBtn = s.querySelector('#btn-close-review');
    const reviewSubmitBtn = s.querySelector('#submit-review');
    let originalreviewBtnSubmitText = reviewSubmitBtn?.textContent.trim() || 'Post Review';
    const reviewError = s.querySelector('#review-error');

    if (!(reviewDialog instanceof HTMLDialogElement)) return;
    if (!(reviewForm instanceof HTMLFormElement)) return;
    if (!(reviewSubmitBtn instanceof HTMLButtonElement && reviewSubmitBtn.type === 'submit')) return;

    const showError = (msg = "") => {
        if (!(reviewError instanceof HTMLElement)) return;
        reviewError.textContent = msg;
        reviewError.hidden = false;
    };
    const clearError = () => {
        if (!(reviewError instanceof HTMLElement)) return;
        reviewError.textContent = '';
        reviewError.hidden = true;
    };

    reviewOpenBtn?.addEventListener('click', () => reviewDialog.showModal());
    reviewCloseBtn?.addEventListener('click', () => reviewDialog.close());

    reviewForm.addEventListener('submit', async (event) => {
        event.preventDefault();
        const form = event.currentTarget;
        if (!(form instanceof HTMLFormElement)) return;
        clearError();

        if (!form.checkValidity()) {
            form.reportValidity(); // shows the native browser bubble
            return;
        }

        const formData = new FormData(form);
        const headline = String(formData.get('headline') || '').trim();
        const content = String(formData.get('content') || '').trim();
        const rating = String(formData.get('rating') || '').trim();

        if (!headline || !content || !rating) {
            showError('Please fill in the required fields');
            return;
        }

        const payload = {
            ...Object.fromEntries(formData.entries()),
            rating: Number(formData.get('rating') || 0)
        };

        reviewSubmitBtn.disabled = true;
        reviewSubmitBtn.textContent = 'Posting...';
        reviewDialog.close();

        try {
            const response = await postData(form.action, payload);
            if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
            const result = await response.json();

            // Get these from the header of the page
            const avatar = document.querySelector('.username-image')?.getAttribute('src') ?? "";
            const username = document.querySelector('.username-text')?.textContent.trim() ?? "";

            const now = new Date();
            const localDate = now.toLocaleDateString();

            const innerHTML = `
                <header class="review-header">
                    <img src="${avatar}" class=" review-user-avatar" width="20" height="20"
                        loading="lazy" alt="">
                    <span class="review-user-name">${username}</span>
                    <span class="review-user-rating">
                        <span class="rating-global-star">&#9733;</span>
                        <span>${payload.rating}</span>
                    </span>
                    <span class="review-date" data-utc-time="">${localDate}</span>
                </header>
                <h4 class="review-headline">${result.review.html_headline}</h4>
                <div class="review-content">${result.review.html_content}</div>
            `;

            // Look for this review in the DOM
            const reviewInDom = document.getElementById("current-user-review");

            // Look if the user has a review here at all
            const userHasReview = reviewSubmitBtn.dataset.hasReview === 'true';

            // Create review holder in case it is new review
            const card = document.createElement('div');

            if (reviewInDom) { // Review is in the DOM
                reviewInDom.innerHTML = innerHTML;
                reviewInDom.scrollIntoView({ behavior: 'smooth', block: 'center' });
                reviewInDom.classList.add('updated-review');
                setTimeout(() => reviewInDom.classList.remove('updated-review'), 2000);
                setAlert("Review updated!");
                // TODO: Change this check, not reliable
            } else if (userHasReview) {
                // Review isn't loaded in the DOM yet.
                setAlert("Review updated!");
            } else { // New review, prepend to the list
                card.className = 'review-card new-review';
                card.id = "current-user-review";
                card.innerHTML = innerHTML;
                reviewsList?.prepend(card);
                setAlert("Review posted!");
            }

            // Update the user rating and average rating HTML
            updateRatingHTML(payload.rating, result.stats.avg_rating, result.stats.rating_count)

            // The request went through, change the submit button text,
            // the data state and remove the new-review class.
            originalreviewBtnSubmitText = "Update Review"
            reviewSubmitBtn.dataset.hasReview = 'true';
            card.addEventListener('animationend', () => {
                card.classList.remove('new-review');
            }, { once: true }); // 'once: true' auto-removes the listener after it fires

        } catch (err) {
            console.error("Failed to fetch or parse JSON:", err);
            setAlert("Something went wrong!");
        } finally {
            reviewSubmitBtn.disabled = false;
            reviewSubmitBtn.textContent = originalreviewBtnSubmitText;
        }
    });
});


// ==========================================================================
// Load More Reviews
// ==========================================================================
document.getElementById('load-more-reviews-btn')?.addEventListener('click', async (event) => {
    const btn = event.currentTarget;
    if (!(btn instanceof HTMLButtonElement)) return;
    const videoId = btn.dataset.videoId;
    const cursor = btn.dataset.cursor;

    // Prevent fetch if no cursor
    if (!cursor) return;
    const originalBtnText = btn.innerHTML;

    try {
        btn.disabled = true;
        btn.innerHTML = '<span class="review-spinner"></span> Loading...';

        const response = await getData(`/api/video/${videoId}/reviews`, cursor);
        if (!response.ok) throw new Error('Failed to fetch reviews');
        const data = await response.json();

        /** @type {Record<string, any>[]} */
        const items = data.items || [];

        // Map the corresponding html to each review in the array
        const reviews = items.map(review => {

            // Convert review date to user's local date
            const dateObj = new Date(review.updated_at);
            const localDate = dateObj.toLocaleDateString();

            // Check if this is a current user review
            const reviewdId = review.is_current_user ? "id='current-user-review'" : ""

            return `
                <div class="review-card" ${reviewdId}>
                    <header class="review-header">
                        <img src="${review.user.local_avatar_url}" class=" review-user-avatar" width="20" height="20"
                            loading="lazy" alt="">
                        <span class="review-user-name">${review.user.name}</span>
                        <span class="review-user-rating">
                            <span class="rating-global-star">&#9733;</span>
                            <span>${review.rating}</span>
                        </span>
                        <span class="review-date" data-utc-time="${review.updated_at}">${localDate}</span>
                    </header>
                    <h4 class="review-headline">${review.html_headline}</h4>
                    <div class="review-content">${review.html_content}</div>
                </div>`;
        });

        // Append the reviews joined as string
        document.getElementById('reviews-list')?.insertAdjacentHTML('beforeend', reviews.join(''));

        if (data.next_cursor) {
            btn.dataset.cursor = data.next_cursor;
        } else {
            // The last page, reset the cursor, remove the load more button
            btn.dataset.cursor = "";
            btn.remove();
            return;
        }
    } catch (error) {
        console.error("Failed to fetch or parse JSON:", error);
        setAlert("Something went wrong!");
    } finally {
        // If the button is in the DOM (not removed),
        // enable it and restore the original text.
        if (btn.isConnected) {
            btn.disabled = false;
            btn.innerHTML = originalBtnText;
        }
    }
});
