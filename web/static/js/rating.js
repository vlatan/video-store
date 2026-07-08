document.querySelectorAll('.rate-widget').forEach(widget => {
    const rateDialog = widget.querySelector('.rate-dialog');
    const rateBtnOpen = widget.querySelector('.btn-open-rate');
    const rateBtnClose = widget.querySelector('.btn-close-rate');
    const rateErrBox = widget.querySelector('.rate-error');
    const rateURL = widget.dataset.url || `${window.location.pathname}rate`;

    rateBtnOpen.addEventListener('click', () => { rateErrBox.hidden = true; rateDialog.showModal(); });
    rateBtnClose.addEventListener('click', () => rateDialog.close());

    widget.querySelectorAll('input[type="radio"]').forEach(input => {
        input.addEventListener('change', async (e) => {
            const rating = Number(e.target.value);
            rateDialog.close();

            try {
                const res = await postData(rateURL, { 'rating': rating });
                if (!res.ok) throw new Error();
                rateBtnOpen.innerHTML = `<span class="star-icon">★</span> <span style="color:#fff">${rating}</span>/10`;
            } catch {
                rateErrBox.hidden = false;
                e.target.checked = false;
            }
        });
    });
});