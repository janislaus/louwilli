function finishGame(documentId) {
    $.ajax({
        url: "/finishGame",
        type: "POST",
        data: JSON.stringify({gameId: documentId}),
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            console.log("yeah! redirect after kafka event is send")
        },

    });
}


function coinDrop(playerName, oldCoinAmount) {
    const newCoinAmount = parseInt(oldCoinAmount) - 1;
    $.ajax({
        url: "/coinDrop",
        type: "POST",
        data: JSON.stringify({
            playerName: playerName,
            newCoinAmount: newCoinAmount
        }),
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function () {
            console.log("yeah! redirect after kafka event is send")
        },
    });
}

function resetGame(documentId) {
    $.ajax({
        url: "/resetGame",
        type: "POST",
        data: JSON.stringify({gameId: documentId}),
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            console.log("yeah! redirect after kafka event is send")
        },

    });
}

function restartGame(documentId) {
    $.ajax({
        url: "/restartGame",
        type: "POST",
        data: JSON.stringify({gameId: documentId}),
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            window.location = data.redirect
        },

    });
}

function confirmPlayer() {

    $.ajax({
        url: "/confirmPlayer",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            console.log("yeah! redirect after kafka event is send")
        },
    });
}

function confirmChangeSide() {

    $("#change-side-cofirm-button").addClass("invisible");

    $.ajax({
        url: "/confirmChangeSide",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            console.log("yeah! redirect after kafka event is send")
        },
    });
}


function startUp() {

    $("#startup_button").addClass("invisible");

    $.ajax({
        url: "/startUp",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            console.log("yeah! redirect after kafka event is send")
        },
    });
}

function setOffline() {

    $.ajax({
        url: "/setOffline",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            window.location = data.redirect
        },
    });
}

function setOnline() {

    $.ajax({
        url: "/setOnline",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            window.location = data.redirect
        },
    });
}

function startOfflineGame() {

    $.ajax({
        url: "/startOfflineGame",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            console.log("yeah! redirect after kafka event is send")
        },
    });
}

function playersCanBeReceived() {

    $.ajax({
        url: "/playersCanBeReceived",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            console.log("yeah! redirect after kafka event is send")
        },
    });
}
