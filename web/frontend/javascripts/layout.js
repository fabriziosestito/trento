$(document).ready(function () {
    // enable bootstrap tooltips
    $('[data-toggle="tooltip"]').tooltip();

    const now = new Date();
    $("#last_update").html(now.toLocaleString());
});
