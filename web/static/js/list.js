document.addEventListener("click", async (event) => {
    if (!(event.target instanceof HTMLElement)) return;
    const remove = event.target.closest(".remove-option");
    if (!(remove instanceof HTMLElement)) return;

    let action = "unlike";
    let messageText = "Succesfully unliked.";
    if (window.location.pathname.includes("favorites")) {
        action = "unfave";
        messageText = "Succesfully removed.";
    }
    const url = `/api/video/${remove.dataset.id}/${action}`;
    try {
        const res = await deleteData(url);
        if (!res.ok) throw new Error(`HTTP error! Status: ${res.status}`);
        remove.parentElement?.remove();
        setAlert(messageText);

        // Decrease faved docs counter
        if (action === "unfave") {
            const countDisplay = document.getElementById("faved-counter");
            let count = parseInt(countDisplay.textContent.trim(), 10);
            if (Number.isNaN(count)) {
                console.error(
                    "Counter element does not contain a number:",
                    countDisplay.textContent,
                );
            } else if (count > 0) {
                countDisplay.textContent = String(count - 1);
            }
        }
    } catch (error) {
        console.error("Failed to fetch response:", error);
        setAlert("Something went wrong!");
    }
});
