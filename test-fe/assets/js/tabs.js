let widgets = document.querySelectorAll(".tabs-container");

for (let widget of widgets) {
    let tabTriggers = widget.querySelectorAll(".tab-btn");

    for (let tabTrigger of tabTriggers) {
        tabTrigger.addEventListener("click", function() {
            // remove all active classes
            widget.querySelector(".tab-btn.active").classList.remove("active");

            // add active class to clicked button
            this.classList.add("active");

            // remove all visible classes
            widget.querySelector(".tab.visible").classList.remove("visible");
            
            // add visible to button pointer data attr
            let elemId = this.getAttribute("data-target");
            let targetElem = widget.querySelector("#" + elemId);
            targetElem.classList.add("visible");
        });
    }
}