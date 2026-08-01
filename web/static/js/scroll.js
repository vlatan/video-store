// Get references to the dom elements
const scroller = document.getElementById("scroller");
const htmlTemplate = document.getElementById("post_template");
const sentinel = document.getElementById("sentinel");
const spinner = sentinel?.querySelector('div');

let state = {
    nextCursor: scroller?.dataset.cursor,
    isLoading: false,
    hasMore: !!scroller?.dataset.cursor,
};

// Function to request new items and render to the dom
const loadItems = async (url = "", cursor = "") => {

    if (!(htmlTemplate instanceof HTMLTemplateElement)) return;
    if (!(sentinel instanceof HTMLElement)) return;

    // Prevent multiple simultaneous fetches
    if (state.isLoading || !state.hasMore) {
        return;
    }

    state.isLoading = true;
    spinner?.setAttribute("id", "spinner");

    try {

        const response = await getData(url, cursor);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        state.nextCursor = data.next_cursor || null;
        state.hasMore = !!data.next_cursor;

        // Iterate over the items in the response, create video cards
        // and append them as children to the scroller.
        for (const item of data.items) {
            const card = createVideoCard(item);
            if (card) scroller?.appendChild(card);
        }

        if (!state.hasMore) {
            sentinel.innerHTML = "No more videos";
        }

    } catch (error) {
        state.hasMore = false;
        sentinel.innerHTML = "Something went wrong";
        console.error("Failed to fetch items:", error);

    } finally {
        state.isLoading = false;
    }
};

if ('IntersectionObserver' in window) {
    // Create a new IntersectionObserver instance
    let intersectionObserver = new IntersectionObserver(([entry]) => {
        // If there is next page and the entry is intersecting
        if (state.hasMore && entry.isIntersecting) {

            const pathWithQueries = window.location.pathname + window.location.search + window.location.hash;

            // Call the loadItems function
            loadItems(`/api${pathWithQueries}`, state.nextCursor);
        }
        // add root margin for earlier intersection detecetion
    }, { rootMargin: "200px 0px" });

    // Instruct the IntersectionObserver to watch the sentinel
    if (sentinel) intersectionObserver.observe(sentinel);
}


/**
 * Builds a populated video card from the template for a single item.
 * Returns null if the template markup is missing required elements.
 *
 * @param {{video_id: string, thumbnail: {url: string}, title: string, srcset: string, id: string|number}} item
 * @returns {DocumentFragment | null}
 */
function createVideoCard(item) {

    if (!(htmlTemplate instanceof HTMLTemplateElement)) {
        throw new Error('Expected #post_template to be a <template> element');
    }

    // Clone the HTML template
    const templateClone = /** @type {DocumentFragment} */ (htmlTemplate.content.cloneNode(true));

    // Set the link source
    const videoLink = templateClone.querySelector('.video-link');
    if (!(videoLink instanceof HTMLAnchorElement)) {
        console.warn('Skipping item, .video-link missing or malformed:', item);
        return null;
    }
    videoLink.href = `/video/${item.video_id}/`;

    // Update the image
    const thumb = templateClone.querySelector('.video-img');
    if (!(thumb instanceof HTMLImageElement)) {
        console.warn('Skipping item, .video-img missing or malformed:', item);
        return null;
    }
    thumb.src = item.thumbnail.url;
    thumb.alt = item.title;
    thumb.srcset = item.srcset;

    // Set the title of the video
    const videoTitle = templateClone.querySelector('.video-title');
    if (videoTitle) videoTitle.textContent = item.title;

    // Set the data-id on the remove buttton if any
    const remove = templateClone.querySelector('.remove-option');
    remove?.setAttribute('data-id', `${item.id}`);

    return templateClone;
}
