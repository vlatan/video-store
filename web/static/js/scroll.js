// Get references to the dom elements
const scroller = document.getElementById("scroller");
const sentinel = document.getElementById("sentinel");
const spinner = sentinel?.querySelector('div');

let state = {
    nextCursor: scroller?.dataset.cursor,
    isLoading: false,
    hasMore: !!scroller?.dataset.cursor,
};

// Function to request new items and render to the dom
const loadItems = async (url = "", cursor = "") => {

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
            scroller?.appendChild(card);
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
 * Builds a populated video card.
 *
 * @param {{video_id: string, thumbnail: {url: string}, title: string, srcset: string, id: string|number}} item
 * @returns {HTMLElement}
 */
function createVideoCard(item) {

    // Create the anchor element
    const a = document.createElement('a');
    a.className = 'video-link';
    a.href = `/video/${item.video_id}/`;

    // Create image wrapper
    const span = document.createElement('span');
    span.className = 'video-img-wrap';
    a.appendChild(span);

    // Create the image
    const img = document.createElement('img');
    img.className = 'video-img';
    img.src = item.thumbnail.url;
    img.alt = item.title;
    img.srcset = item.srcset;
    span.appendChild(img);

    // Create the title
    const title = document.createElement('h2');
    title.className = 'video-title';
    title.textContent = item.title;
    a.appendChild(title);

    // Needs slight modification if this is user favorites scroll.
    // Wrap the anchor in another block and offer remove button.
    if (window.location.pathname === '/user/favorites/') {

        const block = document.createElement('div');
        block.className = 'video-block';

        const remove = document.createElement('span');
        remove.className = 'remove-option';
        remove.setAttribute('data-id', `${item.video_id}`);
        remove.setAttribute('aria-label', 'Close');
        remove.textContent = '\u00D7';
        block.appendChild(remove);

        block.appendChild(a);
        return block;
    }

    return a;
}
