$(function(){
  // foundation
  $(document).foundation();
  // buttons
  $('input[data-action],a[data-action]').click(function(e){
    e.preventDefault();
    var el = $(this);
    var form = el.closest('form');
    var action = el.data('action');
    var consent = el.data('confirm');
    if (!consent || confirm(consent)) {
      form.prop('action', action).submit();
    }
  });
  // datepicker
  $('.date').fdatepicker({
    language: language,
    format: dateFormat(),
    weekStart: 1
  });
  $('.date.changer').on('changeDate', function(e){
    var date = e.date.toISOString().substr(0,10);
    location.href = location.href.replace(/\?date=.*|$/, '?date='+date)
  });
  // select
  $.extend($.fn.select2.defaults, $.fn.select2.locales[language]);
  $('.select2').select2();
});

function dateFormat() {
  if (language == 'da') return 'dd/mm yyyy';
  return 'yyyy-mm-dd';
}
