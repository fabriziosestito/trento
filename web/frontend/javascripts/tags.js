$(function () {
  const tagsInputs = $(".tags-input");

  tagsInputs.select2({
    tags: true,
    width: '300px',
    templateSelection: (tag, container) => {
      const option = $('.tags-input option[value="' + tag.id + '"]')
      console.log(container)
      if (option.attr('parent')) {
        $(container).addClass('parent-tag')
        tag.locked = true
      }
      return tag.text
    },
  }).on('select2:unselecting', e => {
    if ($(e.params.args.data.element).attr('parent-tag')) {
      e.select2.pleaseStop();
    }
  }).on('select2:select', e => {
    if (e.params.data.id == null) {
      return
    }

    const url = $(e.target).attr('data-url')
    $.ajax({
      url: url,
      type: 'POST',
      data: JSON.stringify({ tag: e.params.data.id }),
      dataType: "json"
    })
  }).on('select2:unselect', e => {
    const url = $(e.target).attr('data-url') + "/" + e.params.data.id
    $.ajax({
      url: url,
      type: 'DELETE'
    })
  }).on('select2:open', function (e) {
    $('.select2-container--open .select2-dropdown--below').css('display', 'none');
  })

  // tagsInputs.each(
  //   (index, el) => {
  //     const url = $(el).attr('data-url');
  //     $.ajax({
  //       type: 'GET',
  //       url: url
  //     }).then(data => {
  //       if (data == null) {
  //         return
  //       }

  //       data.forEach(tag => {
  //         const option = new Option(tag, tag, true, true)
  //         $(el).append(option).trigger('change')
  //       })
  //     })

  //     const parentUrl = $(el).attr('data-parent-url');
  //     if (parentUrl) {
  //       $.ajax({
  //         type: 'GET',
  //         url: parentUrl
  //       }).then(data => {
  //         if (data == null) {
  //           return
  //         }

  //         data.forEach(tag => {
  //           const option = new Option(tag, tag, true, true)
  //           option.setAttribute('locked', 'locked')
  //           $(el).append(option).trigger('change')
  //         })
  //       })
  //     }
  //   }
  // )
})
