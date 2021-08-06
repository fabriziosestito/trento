$(document).ready(function() {
  // enable bootstrap tooltips
  $('[data-toggle="tooltip"]').tooltip();

  let now = new Date();
  $("#last_update").html(now.toLocaleString());

  $(".tags-input").select2({
    tags: true,
    width: "300px"
  })
});
