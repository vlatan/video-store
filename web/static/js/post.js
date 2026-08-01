// ==========================================================================
// Video Player
// ==========================================================================

const playerContent = document.querySelector('.player-content');
playerContent?.addEventListener('click', () => {
    const videoId = playerContent.id;
    const iframe = document.createElement('iframe');
    iframe.src = `https://www.youtube-nocookie.com/embed/${videoId}?iv_load_policy=3&cc_load_policy=1&autoplay=1`;
    iframe.allow = "accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture";
    iframe.style.border = "0";
    iframe.allowFullscreen = true;
    playerContent.replaceWith(iframe);
});


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

        const reviewsList = document.getElementById('reviews-list');
        for (const review of data.items) {

            // Convert review date to user's local date
            const dateObj = new Date(review.updated_at);
            const localDate = dateObj.toLocaleDateString();

            // Build review card
            const card = document.createElement('div');
            card.className = 'review-card load-review';
            if (review.is_current_user) card.id = "current-user-review";
            card.innerHTML = buildReviewHTML(
                review.user.local_avatar_url,
                review.user.name,
                review.rating,
                localDate,
                review.html_headline,
                review.html_content,
            );

            reviewsList?.append(card);
        }

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


// ==========================================================================
// Review Helpers
// ==========================================================================

/**
 * @param {string} str
 * @returns {string}
 */
function escapeHtml(str) {
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

/**
 * Only allow http/https URLs through — rejects javascript:, data:,
 * and malformed strings that could break out of the src attribute.
 * @param {string} url
 * @returns {string}
 */
function sanitizeImageUrl(url) {
    try {
        const parsed = new URL(url, window.location.href);
        if (parsed.protocol === 'http:' || parsed.protocol === 'https:') {
            return escapeHtml(parsed.href);
        }
    } catch {
        // invalid URL — fall through
    }
    return '';
}

/**
 * Builds the HTML markup for a single review card.
 *
 * @param {string} avatar - URL of the reviewer's avatar image
 * @param {string} username - Display name of the reviewer
 * @param {number} rating - The user rating
 * @param {string} date - Pre-formatted, human-readable date string
 * @param {string} html_headline - Assumes escaped html headline
 * @param {string} html_content - Assumes escaped html content
 * @returns {string} Sanitized HTML markup for the review card's contents
 */
function buildReviewHTML(avatar, username, rating, date, html_headline, html_content) {

    const safeAvatar = sanitizeImageUrl(avatar);
    const safeUsername = escapeHtml(username);
    const safeDate = escapeHtml(date);

    const safeRating = Number(rating);
    if (!Number.isFinite(rating) || rating < 0 || rating > 10) {
        throw new Error('rating must be a number between 0 and 10');
    }

    return `
        <header class="review-header">
            <div class="review-meta">
                <img src="${safeAvatar}" class=" review-user-avatar" width="20" height="20"
                    loading="lazy" alt="${safeUsername}">
                <span class="review-user-name">${safeUsername}</span>
                <span class="review-user-rating">
                    <span class="rating-global-star">&#9733;</span>
                    <span>${safeRating}</span>
                </span>
                <span class="review-date" data-utc-time="">${safeDate}</span>
            </div>
            <h4 class="review-headline">${html_headline}</h4>
        </header>
        <div class="review-content">${html_content}</div>
    `;
}