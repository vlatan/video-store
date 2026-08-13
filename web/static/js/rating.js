/**
 *  Get initial state from HTML accross all the rating/review dialogs, forms and buttons
 */
function getInitialState() {
    const checked = document.querySelector('input[name="rating"]:checked');
    const avgRatingDisplay = document.querySelector('.avg-rating-display');
    const btnSubmitReview = document.getElementById('btn-submit-review');

    return {
        // Data State
        userRating: checked instanceof HTMLInputElement ? checked.value : "?",
        avgRating: avgRatingDisplay?.querySelector('.rating-avg-val')?.textContent.trim() || "0.0",
        ratingCount: avgRatingDisplay?.querySelector('.rating-count-val')?.textContent.trim() || "0",
        userHasReview: btnSubmitReview?.dataset.hasReview === 'true',

        // Async Status Flags (Interim State)
        isSubmitting: false,
        isDeleting: false
    };
}

// Single rate/review source of truth
const ratingState = getInitialState();

/**
 *  Mutate state and immediately trigger UI updates
 * @param {Partial<typeof ratingState>} updates
 */
function setState(updates) {
    Object.assign(ratingState, updates);
    renderState();
}

/**
 *  Renders the HTML values of the rating/review UI state
 */
function renderState() {
    const isBusy = ratingState.isSubmitting || ratingState.isDeleting;

    // Disable/Enable all action buttons during async operations
    const buttonsToToggle = [
        document.getElementById('btn-open-rate'),
        document.querySelector('.btn-submit-rate'),
        document.getElementById('btn-open-review'),
        document.getElementById('btn-submit-review'),
        document.getElementById('btn-review-delete-init'),
        document.getElementById('btn-review-delete-cancel'),
        document.getElementById('btn-review-delete-confirm')
    ];

    buttonsToToggle.forEach(btn => {
        if (btn instanceof HTMLButtonElement) {
            btn.disabled = isBusy;
        }
    });

    // Update the average rating display
    if (!isBusy && ratingState.avgRating && ratingState.ratingCount) {
        upsertAvgRatingHTML(ratingState.avgRating, ratingState.ratingCount);
    }

    // Dynamic Open Rate Button Inner HTML
    const rateBtnOpen = document.getElementById('btn-open-rate');
    if (rateBtnOpen && !ratingState.isSubmitting) {
        let html = `<span class="rating-user-star">&#9734;</span><span>Rate</span>`;
        if (ratingState.userRating !== "?") {
            html = `<span class="rating-user-star">&#9733;</span><span>${ratingState.userRating}</span>`;
        }
        rateBtnOpen.innerHTML = html;
    }

    // Dynamic Submit Rating Button Text
    const rateBtnSubmit = document.querySelector('.btn-submit-rate');
    if (rateBtnSubmit) {
        if (ratingState.isSubmitting) {
            rateBtnSubmit.textContent = 'Posting...';
        } else {
            rateBtnSubmit.textContent = ratingState.userRating !== "?" ? 'Update' : 'Rate';
        }
    }

    // Dynamic Open Review Button Inner Text
    const reviewOpenBtnText = document.getElementById('btn-open-review-text');
    if (reviewOpenBtnText && !ratingState.isSubmitting) {
        let text = "Post Review"
        if (ratingState.userHasReview) text = "Update Review";
        reviewOpenBtnText.textContent = text;
    }

    // Dynamic Submit Review Button Inner Text
    const btnSubmitReview = document.getElementById('btn-submit-review');
    if (btnSubmitReview) {
        if (ratingState.isSubmitting) {
            btnSubmitReview.textContent = 'Posting...';
        } else {
            btnSubmitReview.textContent = ratingState.userHasReview ? 'Update' : 'Submit';
            btnSubmitReview.dataset.hasReview = String(ratingState.userHasReview);
        }
    }

    // Dynamic Init Delete Review Button
    const btnDeleteInit = document.getElementById('btn-review-delete-init');
    if (btnDeleteInit && !ratingState.isDeleting) btnDeleteInit.hidden = !ratingState.userHasReview;

    // Dynamic Delete Review Button Text
    const btnDeleteConfirm = document.getElementById('btn-review-delete-confirm');
    if (btnDeleteConfirm) {
        btnDeleteConfirm.textContent = ratingState.isDeleting ? 'Deleting...' : 'Delete';
    }
}

/**
 * Updates big star elements.
 * Expects "?" or an integer between 1 and 10.
 * @param {string|number} [val="?"]
 */
const updateBigStars = (val = "?") => {
    if (val !== "?") {
        const num = Number(val);
        if (!Number.isInteger(num) || num < 1 || num > 10) {
            throw new Error(
                `Invalid rating value "${val}". Expected "?" or an integer from 1 to 10.`
            );
        }
        val = String(num);
    }

    document.querySelectorAll('.rating-big-star-value').forEach(bsv => {
        bsv.textContent = val;
    });
};


/**
 * Listen rating stars radios change on check or hover and syncs rating state.
 */
(() => {

    const starRadios = document.querySelectorAll('input[name="rating"]');
    starRadios.forEach(radio => {
        if (!(radio instanceof HTMLInputElement)) return;

        radio.addEventListener('change', () => {
            if (!radio.checked) return;

            // Sync identical radios across multiple forms
            starRadios.forEach(r => {
                if (r instanceof HTMLInputElement && r.value === radio.value) {
                    r.checked = true;
                }
            });

            updateBigStars(radio.value);
        });

        const hoverTarget = radio.closest('label') || radio;

        hoverTarget.addEventListener('mouseenter', () => {
            updateBigStars(radio.value);
        });

        // Revert to currently checked radio or "?"
        hoverTarget.addEventListener('mouseleave', () => {
            const currentChecked = document.querySelector('input[name="rating"]:checked');
            updateBigStars(currentChecked instanceof HTMLInputElement ? currentChecked.value : "?");
        });
    });

})();


/**
 * Clear the rating or review form as well as clear the big stars values
 * @param {NodeListOf<Element>} inputs
 */
function clearForm(inputs) {
    inputs.forEach(field => {
        if (field instanceof HTMLTextAreaElement) {
            field.value = "";
        } else if (field instanceof HTMLInputElement) {
            if (field.type === 'radio' || field.type === 'checkbox') {
                field.checked = false;
            } else {
                field.value = "";
            }
        }
    });

    // Clear the big stars values
    updateBigStars();
}


/**
 * Update the Average Rating Display and the User Rating Button
 *
 * @param {number|string} avg_rating
 * @param {number|string} rating_count
 */
function upsertAvgRatingHTML(avg_rating, rating_count) {

    avg_rating = Number(avg_rating);
    rating_count = Number(rating_count);

    if (!Number.isFinite(avg_rating) || avg_rating < 0 || avg_rating > 10) {
        throw new Error("avg rating must be a number between 0 and 10");
    }

    if (!Number.isInteger(rating_count) || rating_count < 0) {
        throw new Error("rating count must be a non-negative integer");
    }

    // Remove the average rating display altogether if no rating at all
    const avgRatingDisplay = document.querySelector('.avg-rating-display');
    if (avg_rating === 0 && rating_count === 0) avgRatingDisplay?.remove();


    const votesText = rating_count === 1 ? "vote" : "votes";
    const avgRatingHTML = `
        <div class="btn-open-post-dialog avg-rating-display">
            <span class="rating-global-star">&#9733;</span>
            <div class="rating-meta" itemprop="aggregateRating" itemscope
                itemtype="https://schema.org/AggregateRating">
                <meta itemprop="worstRating" content="1">
                <div class="rating-score">
                    <span class="rating-avg-val" itemprop="ratingValue">${avg_rating}</span>
                    <span>/</span>
                    <span itemprop="bestRating">10</span>
                </div>
                <div class="rating-count">
                    <span class="rating-count-val" itemprop="ratingCount">${rating_count}</span>
                    <span>${votesText}</span>
                </div>
            </div>
        </div>
    `;


    // Upsert the average rating display
    const rateBtnOpen = document.getElementById('btn-open-rate');
    if (avgRatingDisplay) {
        avgRatingDisplay.outerHTML = avgRatingHTML;
    } else if (rateBtnOpen) {
        rateBtnOpen.insertAdjacentHTML('beforebegin', avgRatingHTML);
    }
}


// ==========================================================================
// Add Rating
// ==========================================================================

document.querySelectorAll('.rating-section').forEach(widget => {
    const rateDialog = widget.querySelector('#rate-dialog');
    const rateForm = widget.querySelector('.rate-form');
    const rateBtnOpen = widget.querySelector('#btn-open-rate');
    const rateBtnClose = widget.querySelector('#btn-close-rate');

    if (!(rateDialog instanceof HTMLDialogElement)) return;
    if (!(rateForm instanceof HTMLFormElement)) return;

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

        setState({ isSubmitting: true });
        rateDialog.close();

        try {
            const response = await postData(form.action, payload);
            if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
            const result = await response.json();

            // Set new state
            setState({
                userRating: String(payload.rating),
                avgRating: result.avg_rating,
                ratingCount: result.rating_count
            });

            // Show toast message
            setAlert("Post rated");
        } catch (error) {
            console.error("Failed to fetch or parse JSON:", error);
            setAlert("Something went wrong!");
        } finally {
            setState({ isSubmitting: false });
        }
    });
});


// ==========================================================================
// Add Review
// ==========================================================================

document.querySelectorAll('.review-section').forEach(s => {
    const reviewDialog = s.querySelector('#review-dialog');
    const reviewForm = s.querySelector('.review-form');
    const reviewsList = s.querySelector('#reviews-list');
    const reviewOpenBtn = s.querySelector('#btn-open-review');
    const reviewCloseBtn = s.querySelector('#btn-close-review');
    const reviewSubmitBtn = s.querySelector('#btn-submit-review');
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

        // Check for input errors
        if (!headline || !content || !rating) {
            showError('Please fill in the required fields');
            return;
        }

        const payload = {
            ...Object.fromEntries(formData.entries()),
            rating: Number(formData.get('rating') || 0)
        };

        setState({ isSubmitting: true });
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

            const innerHTML = buildReviewHTML(
                avatar,
                username,
                payload.rating,
                localDate,
                result.review.html_headline,
                result.review.html_content,
            );

            // Look for this review in the DOM
            const reviewInDom = document.getElementById("current-user-review");

            // Look if the user has a review here at all
            const userHasReview = reviewSubmitBtn.dataset.hasReview === 'true';

            if (reviewInDom) { // Review is in the DOM
                reviewInDom.innerHTML = innerHTML;
                reviewInDom.scrollIntoView({ behavior: 'smooth', block: 'center' });
                reviewInDom.classList.add('updated-review');
                setTimeout(() => reviewInDom.classList.remove('updated-review'), 2000);
                setAlert("Review updated");
            } else if (userHasReview) {
                // Review isn't loaded in the DOM yet.
                setAlert("Review updated");
            } else { // New review, prepend to the list
                const card = document.createElement('div');
                card.className = 'review-card load-review';
                card.id = "current-user-review";
                card.setAttribute("itemprop", "review");
                card.setAttribute("itemscope", "");
                card.setAttribute("itemtype", "https://schema.org/Review");
                card.innerHTML = innerHTML;
                reviewsList?.prepend(card);
                setAlert("Review posted");

                // Update the user reviews counter
                const countSpan = document.getElementById('review-count');
                const numReviews = parseInt(countSpan?.textContent || "0", 10) + 1

                // Adjust the reviews list header
                const reviewsHeaderTitle = document.querySelector('.reviews-title-wrapper h3');
                if (reviewsHeaderTitle) {
                    if (numReviews === 1) {
                        reviewsHeaderTitle.textContent = "User Review";
                    } else if (numReviews > 1) {
                        reviewsHeaderTitle.textContent = "User Reviews";
                    }
                }

                if (countSpan) {
                    countSpan.textContent = String(numReviews);
                }

                const reviewCountWrapper = document.getElementById('review-count-wrapper')
                if (reviewCountWrapper) {
                    reviewCountWrapper.style.display = 'inline';
                }

                // Remove the class load-review after the animation.
                // 'once: true' auto-removes the listener after it fires.
                card.addEventListener('animationend', () => {
                    card.classList.remove('load-review');
                }, { once: true });
            }

            // Set new state
            setState({
                userRating: String(payload.rating),
                avgRating: result.stats.avg_rating,
                ratingCount: result.stats.rating_count,
                userHasReview: true
            });
        } catch (err) {
            console.error("Failed to fetch or parse JSON:", err);
            setAlert("Something went wrong!");
        } finally {
            setState({ isSubmitting: false });
        }
    });
});


// ==========================================================================
// Delete Review
// ==========================================================================

document.querySelectorAll('.review-section').forEach(s => {

    const reviewDialog = s.querySelector('#review-dialog');
    const reviewForm = s.querySelector('.review-form');
    const defaultState = s.querySelector('#review-actions-default');
    const confirmState = s.querySelector('#review-actions-confirm');
    const btnDeleteInit = s.querySelector('#btn-review-delete-init');
    const btnDeleteCancel = s.querySelector('#btn-review-delete-cancel');
    const btnDeleteConfirm = s.querySelector('#btn-review-delete-confirm');

    if (!(reviewDialog instanceof HTMLDialogElement)) return;
    if (!(reviewForm instanceof HTMLFormElement)) return;
    if (!(defaultState instanceof HTMLElement)) return;
    if (!(confirmState instanceof HTMLElement)) return;
    if (!(btnDeleteInit instanceof HTMLButtonElement && btnDeleteInit.type === 'button')) return;
    if (!(btnDeleteConfirm instanceof HTMLButtonElement && btnDeleteInit.type === 'button')) return;

    btnDeleteInit?.addEventListener('click', () => {
        defaultState.hidden = true;
        confirmState.hidden = false;
    });

    btnDeleteCancel?.addEventListener('click', () => {
        confirmState.hidden = true;
        defaultState.hidden = false;
    });

    btnDeleteConfirm?.addEventListener('click', async () => {

        setState({ isDeleting: true });
        reviewDialog.close();

        try {
            // Send the request to the backend
            // const response = await postData(reviewForm.action);
            // if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

            // Look for this review in the DOM and remove it if there
            const reviewInDom = document.getElementById("current-user-review");
            reviewInDom?.remove();

            // Clear the rating form
            const rateForm = document.querySelector('.rate-form');
            const rateFormInputs = rateForm?.querySelectorAll("input, select, textarea");
            if (rateFormInputs) clearForm(rateFormInputs);

            // Clear the review form
            const reviewFormImputs = reviewForm.querySelectorAll("input, select, textarea");
            clearForm(reviewFormImputs);

            // Set new state
            setState({ userRating: "?", userHasReview: false });
        } catch (error) {
            console.error('Review deletion failed', error);
            setAlert("Something went wrong!");
        } finally {
            confirmState.hidden = true;
            defaultState.hidden = false;
            setState({ isDeleting: false });
        }
    });
});
