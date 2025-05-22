const popupWindow = url => {
    const w = window.top.innerWidth * 0.4;
    const h = window.top.innerHeight * 0.9;
    const y = window.top.outerHeight / 2 + window.top.screenY - (h / 2);
    const x = window.top.outerWidth / 2 + window.top.screenX - (w / 2);
    return window.open(url, '_blank', `width=${w}, height=${h}, top=${y}, left=${x}`);
};