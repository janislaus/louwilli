const pageNumber = "page-number";
const nameFilter = "name-filter";

$(document).ready(function () {
    localStorage.clear()

    setInterval(function () {
            $(".alert").fadeOut();
        }, 5000
    );
});

document.body.addEventListener("htmx:configRequest", function (configEvent) {
    // Paging Storage setup
    insertPageNumberIntoLocalStorage(configEvent);
    setPageNumberInRequestParameters(configEvent);

    // Filter Storage setup
    insertNameFilterIntoLocalStorage(configEvent);
    setNameFilterInRequestParameters(configEvent)
})

function insertNameFilterIntoLocalStorage(configEvent) {
    let nameFilterEvent = configEvent.detail.parameters[nameFilter];

    if (nameFilterEvent || nameFilterEvent === "") {
        localStorage[nameFilter] = nameFilterEvent;
    }
}

function insertPageNumberIntoLocalStorage(configEvent) {
    let hxTriggerHeader = configEvent.detail.headers['HX-Trigger'];

    if (hxTriggerHeader) {
        if (hxTriggerHeader === nameFilter) {
            localStorage.removeItem(pageNumber)
        } else {
            let splittedHeaderContent = hxTriggerHeader.split("-");

            if (splittedHeaderContent[0] && splittedHeaderContent[0] === "pageNumber") {
                localStorage[pageNumber] = splittedHeaderContent[1];
            }
        }
    }
}

function setPageNumberInRequestParameters(configEvent) {
    if (localStorage[pageNumber]) {
        configEvent.detail.parameters[pageNumber] = localStorage[pageNumber];
    }
}

function setNameFilterInRequestParameters(configEvent) {
    if (localStorage[nameFilter]) {
        configEvent.detail.parameters[nameFilter] = localStorage[nameFilter];
    }
}