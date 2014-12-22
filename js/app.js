$(function(){
  // support DELETE button in forms
  $('input.delete').click(function(){
    var form = $(this).closest('form');
    var action = form.prop('action').replace(/\w+$/, 'delete');
    form.prop('action', action).submit();
  });
});
