document.querySelectorAll('.rate-widget').forEach(widget => {
    const rateDialog = widget.querySelector('.rate-dialog');
    const rateBtnOpen = widget.querySelector('.btn-open-rate');
    const rateBtnClose = widget.querySelector('.btn-close-rate');
    const rateURL = widget.dataset.url || `${window.location.pathname}rate`;

    rateBtnOpen.addEventListener('click', () => { rateDialog.showModal(); });
    rateBtnClose.addEventListener('click', () => rateDialog.close());

    widget.querySelectorAll('input[type="radio"]').forEach(input => {
        input.addEventListener('change', async (e) => {
            const rating = Number(e.target.value);
            rateDialog.close();
            try {
                const res = await postData(rateURL, { 'rating': rating });
                if (!res.ok) throw new Error(`HTTP error! Status: ${res.status}`);
                const data = await res.json();
                rateBtnOpen.innerHTML = `<span class="star-icon">★</span> <span style="color:#fff">${data.avg_rating}</span>/10`;
            } catch (error) {
                e.target.checked = false;
                console.error("Failed to fetch or parse JSON:", error);
                setAlert("Something went wrong!")
            }
        });
    });
});