/**
 *  Get initial state from HTML accross all the rating/review dialogs, forms and buttons
 */
function getInitialState() {
    const checked = document.querySelector('input[name="rating"]:checked');
    const avgRatingDisplay = document.querySelector('.avg-rating-display');
    const avgRating = avgRatingDisplay?.querySelector('.rating-avg-val');
    const ratingCount = avgRatingDisplay?.querySelector('.rating-count-val');
    const btnSubmitReview = document.getElementById('btn-submit-review');
    const reviewCount = document.getElementById('review-count');

    return {
        // Data State
        userRating: checked instanceof HTMLInputElement ? checked.value : "?",
        avgRating: parseFloat(avgRating?.textContent || "0.0"),
        ratingCount: parseInt(ratingCount?.textContent || "0", 10),
        reviewCount: parseInt(reviewCount?.textContent || "0", 10),
        userHasReview: btnSubmitReview?.dataset.hasReview === 'true',

        // Async Status Flags (Interim State)
        isSubmitting: false,
        isDeleting: false
    };
}

// Single rate/review source of truth
const opinionState = getInitialState();

/**
 *  Mutate state and immediately trigger UI updates
 * @param {Partial<typeof opinionState>} updates
 */
function setState(updates) {
    Object.assign(opinionState, updates);
    renderState();
}

/**
 *  Renders the HTML values of the rating/review UI state
 */
function renderState() {
    const isBusy = opinionState.isSubmitting || opinionState.isDeleting;

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

    // Update the average rating display and reviews header
    if (!isBusy) {
        upsertAvgRatingHTML(opinionState.avgRating, opinionState.ratingCount);
        updatetReviewsHeader(opinionState.reviewCount);
    }

    // Dynamic Open Rating Dialog Button
    const rateBtnOpen = document.getElementById('btn-open-rate');
    if (rateBtnOpen && !isBusy) {
        let html = `<span class="rating-user-star">&#9734;</span><span>Rate</span>`;
        if (opinionState.userRating !== "?") {
            html = `<span class="rating-user-star">&#9733;</span><span>${opinionState.userRating}</span>`;
        }
        rateBtnOpen.innerHTML = html;
    }

    // Dynamic Submit Rating Button
    const rateBtnSubmit = document.querySelector('.btn-submit-rate');
    if (rateBtnSubmit) {
        if (opinionState.isSubmitting) {
            rateBtnSubmit.textContent = 'Posting...';
        } else {
            rateBtnSubmit.textContent = opinionState.userRating !== "?" ? 'Update' : 'Rate';
        }
    }

    // Dynamic Open Review Dialog Button
    const reviewOpenBtnText = document.getElementById('btn-open-review-text');
    if (reviewOpenBtnText && !isBusy) {
        reviewOpenBtnText.textContent = opinionState.userHasReview ? "Update Review" : "Post Review";
    }

    // Dynamic Submit Review Button
    const btnSubmitReview = document.getElementById('btn-submit-review');
    if (btnSubmitReview) {
        if (opinionState.isSubmitting) {
            btnSubmitReview.textContent = 'Posting...';
        } else {
            btnSubmitReview.textContent = opinionState.userHasReview ? 'Update' : 'Submit';
            btnSubmitReview.dataset.hasReview = String(opinionState.userHasReview);
        }
    }

    // Dynamic Init Delete Rate Button
    const btnRateDeleteInit = document.getElementById('btn-rate-delete-init');
    if (btnRateDeleteInit && !isBusy) btnRateDeleteInit.hidden = opinionState.userRating === "?";

    // Dynamic Delete Rate Button
    const btnRateDeleteConfirm = document.getElementById('btn-rate-delete-confirm');
    if (btnRateDeleteConfirm) {
        btnRateDeleteConfirm.textContent = opinionState.isDeleting ? 'Deleting...' : 'Confirm';
    }

    // Dynamic Init Delete Review Button
    const btnReviewDeleteInit = document.getElementById('btn-review-delete-init');
    if (btnReviewDeleteInit && !isBusy) btnReviewDeleteInit.hidden = !opinionState.userHasReview;

    // Dynamic Delete Review Button
    const btnReviewDeleteConfirm = document.getElementById('btn-review-delete-confirm');
    if (btnReviewDeleteConfirm) {
        btnReviewDeleteConfirm.textContent = opinionState.isDeleting ? 'Deleting...' : 'Confirm';
    }

    // Clear the ratings - in all forms
    const ratingRadios = document.querySelectorAll('input[name="rating"]');
    if (!isBusy && opinionState.userRating === "?") {
        clearInputs(ratingRadios);
        updateBigStars();
    }

    // Clear the review headline and content - in all forms
    const reviewInputs = document.querySelectorAll('input[name="headline"], textarea[name="content"]');
    if (!isBusy && !opinionState.userHasReview) clearInputs(reviewInputs);

    // Look for current user review in the DOM and remove it if there
    const reviewInDom = document.getElementById("current-user-review");
    if (!isBusy && !opinionState.userHasReview) reviewInDom?.remove();
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
 * Listen rating stars radios change on checked/hover and syncs rating state.
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
function clearInputs(inputs) {
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
}


/**
 * Update or insert the average rating display
 *
 * @param {number} avg_rating
 * @param {number} rating_count
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
    if (avg_rating === 0 && rating_count === 0) {
        avgRatingDisplay?.remove();
        return;
    }

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

/**
 * Update the reviews header
 *
 * @param {number} review_count
 */
function updatetReviewsHeader(review_count) {

    review_count = Number(review_count);
    if (!Number.isInteger(review_count) || review_count < 0) {
        throw new Error("review count must be a non-negative integer");
    }

    const reviewCountWrapper = document.getElementById('review-count-wrapper')
    const reviewsHeaderTitle = document.querySelector('.reviews-title-wrapper h3');

    // Hide the review count and adjust the reviews list title
    if (review_count === 0) {
        if (reviewCountWrapper) reviewCountWrapper.style.display = 'none';
        if (reviewsHeaderTitle) reviewsHeaderTitle.textContent = "No Reviews Yet";
        return;
    }

    // Update the user reviews count and show it
    const countSpan = document.getElementById('review-count');
    if (countSpan) countSpan.textContent = String(review_count);
    if (reviewCountWrapper) reviewCountWrapper.style.display = 'inline';

    // Adjust the reviews list title
    if (reviewsHeaderTitle) {
        reviewsHeaderTitle.textContent = review_count > 1 ? "User Reviews" : "User Review";
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

        if (!rateForm.checkValidity()) {
            rateForm.reportValidity(); // shows the native browser bubble
            return;
        }

        const formData = new FormData(rateForm);
        const payload = {
            ...Object.fromEntries(formData.entries()),
            rating: Number(formData.get('rating') || 0)
        };

        setState({ isSubmitting: true });
        rateDialog.close();

        try {
            const response = await postData(rateForm.action, payload);
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
        clearError();

        if (!reviewForm.checkValidity()) {
            reviewForm.reportValidity(); // shows the native browser bubble
            return;
        }

        const formData = new FormData(reviewForm);
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
            const response = await postData(reviewForm.action, payload);
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

            if (reviewInDom) {
                // Review is in the DOM, update it
                reviewInDom.innerHTML = innerHTML;
                reviewInDom.scrollIntoView({ behavior: 'smooth', block: 'center' });
                reviewInDom.classList.add('updated-review');
                setTimeout(() => reviewInDom.classList.remove('updated-review'), 2000);
                setAlert("Review updated");
            } else if (userHasReview) {
                // Review isn't in the DOM, just inform the user it was updated
                setAlert("Review updated");
            } else {
                // New review, prepend it to the list
                const card = document.createElement('div');
                card.className = 'review-card load-review';
                card.id = "current-user-review";
                card.setAttribute("itemprop", "review");
                card.setAttribute("itemscope", "");
                card.setAttribute("itemtype", "https://schema.org/Review");
                card.innerHTML = innerHTML;
                reviewsList?.prepend(card);
                setAlert("Review posted");

                // Increase the review counter
                setState({ reviewCount: opinionState.reviewCount + 1 });

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
// Delete Rating/Review
// ==========================================================================

document.querySelectorAll('.opinion-section').forEach(s => {

    const opinionDialog = s.querySelector('.opinion-dialog');
    const defaultState = s.querySelector('.opinion-actions-default');
    const confirmState = s.querySelector('.opinion-actions-confirm');
    const btnDeleteInit = s.querySelector('.btn-opinion-delete-init');
    const btnDeleteCancel = s.querySelector('.btn-opinion-delete-cancel');
    const btnDeleteConfirm = s.querySelector('.btn-opinion-delete-confirm');

    if (!(opinionDialog instanceof HTMLDialogElement)) return;
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

        const videoId = btnDeleteConfirm.dataset.videoId;
        setState({ isDeleting: true });
        opinionDialog.close();

        try {
            // Send the request to the backend
            const response = await deleteData(`/api/video/${videoId}/unrate`);
            if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
            const result = await response.json();

            // Substract review count if user has review at all
            let reviewCount = opinionState.reviewCount;
            reviewCount = opinionState.userHasReview ? reviewCount - 1 : reviewCount;

            // Set new state
            setState({
                userRating: "?",
                userHasReview: false,
                avgRating: result.avg_rating,
                ratingCount: result.rating_count,
                reviewCount: reviewCount
            });

            // Show toast message
            setAlert("Rating/Review deleted");
        } catch (err) {
            console.error("Failed to fetch or parse JSON:", err);
            setAlert("Something went wrong!");
        } finally {
            confirmState.hidden = true;
            defaultState.hidden = false;
            setState({ isDeleting: false });
        }
    });
});
