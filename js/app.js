$(function(){
  // support DELETE button in forms
  $('input.delete').click(function(){
    var el = $(this);
    var form = el.closest('form');
    var action = el.data('action');
    var consent = el.data('confirm');
    if (!consent || confirm(consent)) {
      form.prop('action', action).submit();
    }
  });
});
