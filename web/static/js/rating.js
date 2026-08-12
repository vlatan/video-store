/**
 * Updates big star elements.
 * Expects "?" or an integer between 1 and 10.
 * @param {string|number} [val="?"]
 */
const updateBigStar = (val = "?") => {
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


// ==========================================================================
// Sync All Checked Stars and Big Star Values on hover or checked
// ==========================================================================

const starRadios = document.querySelectorAll('input[name="rating"]');

// Track currently checked value
let selectedValue = "?";
const checked = document.querySelector('input[name="rating"]:checked');
if (checked instanceof HTMLInputElement) {
    selectedValue = checked.value;
}

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
        });

        updateBigStar(selectedValue);
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


/**
 * Clear the rating or review form as well as clear the big stars values
 * @param {NodeListOf<Element>} inputs
 */
function clearForm(inputs) {
    inputs.forEach(field => {
        if (field instanceof HTMLInputElement || field instanceof HTMLTextAreaElement) {
            field.value = "";
        }

        if (field instanceof HTMLInputElement && field.type === 'radio') {
            field.checked = false;
        }
    });

    // Reset the big star as well
    updateBigStar();
}


/**
 * Update the Average Rating Display and the User Rating Button
 *
 * @param {number} user_rating
 * @param {number} avg_rating
 * @param {number} rating_count
 */
function updateRatingHTML(user_rating, avg_rating, rating_count) {

    user_rating = Number(user_rating);
    avg_rating = Number(avg_rating);
    rating_count = Number(rating_count);

    if (!Number.isFinite(user_rating) || user_rating <= 0 || user_rating > 10) {
        throw new Error("user rating can't be <= 0 or > 10");
    }
    if (!Number.isFinite(avg_rating) || avg_rating < 0 || avg_rating > 10) {
        throw new Error("avg rating must be a number between 0 and 10");
    }
    if (!Number.isInteger(rating_count) || rating_count < 0) {
        throw new Error("rating count must be a non-negative integer");
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
// Add Rating
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
            setAlert("Post rated");
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
// Add Review
// ==========================================================================

document.querySelectorAll('.review-section').forEach(s => {
    const reviewDialog = s.querySelector('#review-dialog');
    const reviewForm = s.querySelector('.review-form');
    const reviewsList = s.querySelector('#reviews-list');
    const reviewOpenBtn = s.querySelector('#btn-open-review');
    const reviewOpenBtnText = s.querySelector('#btn-open-review-text');
    let originalreviewOpenBtnText = reviewOpenBtnText?.textContent.trim() || 'Post Review';
    const reviewCloseBtn = s.querySelector('#btn-close-review');
    const reviewSubmitBtn = s.querySelector('#btn-submit-review');
    let originalreviewBtnSubmitText = reviewSubmitBtn?.textContent.trim() || 'Submit';
    const btnDeleteInit = s.querySelector('#btn-review-delete-init');
    const reviewError = s.querySelector('#review-error');

    if (!(reviewDialog instanceof HTMLDialogElement)) return;
    if (!(reviewForm instanceof HTMLFormElement)) return;
    if (!(reviewOpenBtnText instanceof HTMLElement)) return;
    if (!(reviewSubmitBtn instanceof HTMLButtonElement && reviewSubmitBtn.type === 'submit')) return;
    if (!(btnDeleteInit instanceof HTMLButtonElement && btnDeleteInit.type === 'button')) return;

    // We'll need the rate button submit value too
    const rateBtnSubmit = document.querySelector('.btn-submit-rate');
    if (!(rateBtnSubmit instanceof HTMLButtonElement && rateBtnSubmit.type === 'submit')) return;
    let originalrateBtnSubmitText = rateBtnSubmit?.textContent.trim() || 'Rate';

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

        // Disable the rate button
        rateBtnSubmit.disabled = true;
        rateBtnSubmit.textContent = 'Posting...';

        // Disable the review button
        reviewSubmitBtn.disabled = true;
        reviewSubmitBtn.textContent = 'Posting...';

        // Close the dialog
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

            // Update the user rating and average rating HTML
            updateRatingHTML(payload.rating, result.stats.avg_rating, result.stats.rating_count)

            // The request went through, change buttons
            originalrateBtnSubmitText = "Update";
            originalreviewOpenBtnText = "Update Review";
            originalreviewBtnSubmitText = "Update";
            reviewSubmitBtn.dataset.hasReview = 'true';
            btnDeleteInit.hidden = false;
        } catch (err) {
            console.error("Failed to fetch or parse JSON:", err);
            setAlert("Something went wrong!");
        } finally {
            rateBtnSubmit.disabled = false;
            rateBtnSubmit.textContent = originalrateBtnSubmitText;
            reviewSubmitBtn.disabled = false;
            reviewSubmitBtn.textContent = originalreviewBtnSubmitText;
            reviewOpenBtnText.textContent = originalreviewOpenBtnText;
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

    const reviewOpenBtnText = s.querySelector('#btn-open-review-text');
    if (!(reviewOpenBtnText instanceof HTMLElement)) return;
    let originalreviewOpenBtnText = reviewOpenBtnText.textContent.trim() || 'Post Review';

    const reviewSubmitBtn = s.querySelector('#btn-submit-review');
    if (!(reviewSubmitBtn instanceof HTMLButtonElement && reviewSubmitBtn.type === 'submit')) return;
    let originalreviewBtnSubmitText = reviewSubmitBtn.textContent.trim() || 'Submit';

    // We'll need the rate button submit value too
    const rateBtnSubmit = document.querySelector('.btn-submit-rate');
    if (!(rateBtnSubmit instanceof HTMLButtonElement && rateBtnSubmit.type === 'submit')) return;
    let originalrateBtnSubmitText = rateBtnSubmit?.textContent.trim() || 'Rate';

    btnDeleteInit?.addEventListener('click', () => {
        defaultState.hidden = true;
        confirmState.hidden = false;
    });

    btnDeleteCancel?.addEventListener('click', () => {
        confirmState.hidden = true;
        defaultState.hidden = false;
    });

    btnDeleteConfirm?.addEventListener('click', async () => {

        // Disable the delete and submit buttons and close the review dialog
        btnDeleteConfirm.disabled = true;
        btnDeleteConfirm.textContent = 'Deleting...';
        reviewSubmitBtn.disabled = true;
        reviewDialog.close();

        try {
            // Send the request to the backend
            // const response = await postData(reviewForm.action);
            // if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

            // Look for this review in the DOM and remove it if there
            const reviewInDom = document.getElementById("current-user-review");
            reviewInDom?.remove();

            // TODO: Update the average rating with updateRatingHTML(), 
            // meaning the request neeeds to return the new average and count here too.

            // TODO: Update the review count in the reviews list header, decrease by one
            // TODO: Reset the big star value in the review form as well as in the rating form.

            // Clear the rating form
            const rateForm = document.querySelector('.rate-form');
            const rateFormInputs = rateForm?.querySelectorAll("input, select, textarea");
            if (rateFormInputs) clearForm(rateFormInputs);

            // Clear the review form
            const reviewFormImputs = reviewForm.querySelectorAll("input, select, textarea");
            clearForm(reviewFormImputs);

            // The request went through, change buttons
            originalrateBtnSubmitText = "Rate";
            originalreviewOpenBtnText = "Post Review";
            originalreviewBtnSubmitText = "Post";
            reviewSubmitBtn.dataset.hasReview = 'false';
            btnDeleteInit.hidden = true;
        } catch (error) {
            console.error('Review deletion failed', error);
            setAlert("Something went wrong!");
        } finally {
            // Reset the state of the buttons and texts
            confirmState.hidden = true;
            defaultState.hidden = false;

            rateBtnSubmit.disabled = false;
            rateBtnSubmit.textContent = originalrateBtnSubmitText;

            reviewSubmitBtn.disabled = false;
            reviewSubmitBtn.textContent = originalreviewBtnSubmitText;

            btnDeleteConfirm.disabled = false;
            reviewOpenBtnText.textContent = originalreviewOpenBtnText;
        }
    });
});
