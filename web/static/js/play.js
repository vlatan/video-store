const v = document.querySelector('.player-content');

const playerIframe = function () {
    const e = document.createElement("iframe");
    e.setAttribute("src", "https://www.youtube-nocookie.com/embed/" + v.id + "?iv_load_policy=3&cc_load_policy=1&autoplay=1");
    e.setAttribute("frameborder", "0");
    e.setAttribute("allow", "accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture");
    e.setAttribute("allowfullscreen", "");
    this.parentNode.replaceChild(e, this);
};

v.onclick = playerIframe;