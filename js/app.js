$(function(){
  // foundation
  $(document).foundation();
  // actions
  $('input[data-action],a[data-action]').click(function(e){
    e.preventDefault();
    performAction($(this));
  });
  $('select[data-action]').change(function(){
    performAction($(this));
  });
  // datepicker
  $('.date').fdatepicker({
    language: language,
    format: dateFormat(),
    weekStart: 1
  });
  $('.date.changer').on('changeDate', function(e){
    var date = e.date.toISOString().substr(0,10);
    Qurl().query('date', date);
    location.reload();
  });
  // select
  $.extend($.fn.select2.defaults, $.fn.select2.locales[language]);
  $('.select2').select2();
  // misc
  activateTab();
  activateLinks();
});

function dateFormat() {
  if (language == 'da') return 'dd/mm yyyy';
  return 'yyyy-mm-dd';
}

function performAction(el) {
  var form = el.closest('form');
  var action = el.data('action') || el.attr('href');
  var consent = el.data('confirm');
  if (consent && !confirm(consent)) {
    return;
  }
  if (action) {
    form.prop('action', action);
  }
  form.submit();
}

// a hack for buggy foundation tabs
function activateTab() {
  if (location.hash) {
    $('.tabs li a').each(function(){
      var hash = '#' + $(this).attr('href').split('#')[1];
      if (hash == location.hash) {
        $(this).click();
      }
    });
  }
}

function activateLinks() {
  var exp = /(\b[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\b)/ig;
  var t = $('.form-description').text().replace(exp, '<a href="mailto:$1">$1</a>');
  $('.form-description').html(t);
}
