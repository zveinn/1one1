"use strict";

let panels = document.querySelectorAll(".panels-container");

for (let panel of panels) {
    let buttons = panel.querySelectorAll(".btn");

    for (let button of buttons) {
        button.addEventListener("click", function() {
            let elemId = button.getAttribute("data-open");
            let elem = panel.querySelector("#" + elemId);

            elem.classList.toggle("slide");
            button.classList.toggle("open");
        });
    }
}
