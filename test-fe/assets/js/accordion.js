let accTriggers = document.querySelectorAll(".accordion");

for (let accTrigger of accTriggers) {
    accTrigger.addEventListener('click', function () {
        this.firstElementChild.classList.toggle("fa-plus-square");
        this.firstElementChild.classList.toggle("fa-minus-square");
        this.classList.toggle("active");
        let elementID = this.getAttribute("data-target");
 
        let subList = document.getElementById(elementID);
        subList.classList.toggle("active");
    });
}