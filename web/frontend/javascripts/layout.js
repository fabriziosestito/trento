"use strict";

$(document).ready(function () {
  // enable bootstrap tooltips
  $('[data-toggle="tooltip"]').tooltip();

  let now = new Date();
  $("#last_update").html(now.toLocaleString());

  $(".tags-input").select2({
    tags: true,
    width: "300px"
  })


  $('.tags-input').each(
    function () {
      var tagsInput = $(this)
      var url = tagsInput.attr('data-url')
      $.ajax({
        type: 'GET',
        url: url
      }).then(function (data) {
        if (data == null) {
          return
        }

        data.forEach(tag => {
          var option = new Option(tag, tag, true, true);
          tagsInput.append(option).trigger('change');
        });
        // manually trigger the `select2:select` event
        tagsInput.trigger({
          type: 'select2:select',
          params: {
            data: data
          }
        });
      });
    }
  )
});