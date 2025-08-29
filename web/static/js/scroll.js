// Get references to the dom elements
const scroller = document.getElementById("scroller");
const template = document.getElementById("post_template");
const sentinel = document.getElementById("sentinel");
const spinner = sentinel.querySelector('div');

let state = {
    nextCursor: initialCursor,
    isLoading: false,
    hasMore: !!initialCursor,
};

// Function to request new items and render to the dom
const loadItems = (url, cursor) => {

    // Prevent multiple simultaneous fetches
    if (state.isLoading || !state.hasMore) {
        return;
    }

    state.isLoading = true;
    spinner.setAttribute("id", "spinner");

    try {
        getData(url, cursor).then(response => {

            // If bad response exit the function
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            // Unserialize the data
            response.json().then(data => {

                state.nextCursor = data.next_cursor || null;
                state.hasMore = !!data.next_cursor;

                // Iterate over the items in the response
                for (const item of data.items) {

                    // Clone the HTML template
                    const template_clone = template.content.cloneNode(true);

                    // Query & update the template content
                    template_clone.querySelector('.video-link').href = `/video/${item.video_id}/`;
                    const thumb = template_clone.querySelector('.video-img');
                    thumb.src = item.thumbnail.url;
                    thumb.alt = item.title;
                    thumb.srcset = item.srcset;
                    template_clone.querySelector('.video-title').innerHTML = item.title;
                    const remove = template_clone.querySelector('.remove-option');
                    if (remove) {
                        remove.setAttribute('data-id', `${item.id}`)
                    }

                    // Append template to dom
                    scroller.appendChild(template_clone);
                }
            })

            if (!state.hasMore) {
                sentinel.innerHTML = "No more videos";
            }
        })

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

            // Call the loadItems function
            loadItems(`${window.location.href}`, state.nextCursor);
        }
        // add root margin for earlier intersection detecetion
    }, { rootMargin: "100px 0px" });

    // Instruct the IntersectionObserver to watch the sentinel
    intersectionObserver.observe(sentinel);
}