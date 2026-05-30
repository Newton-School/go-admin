document.addEventListener("change", function (event) {
  var target = event.target;
  if (target && target.matches("[data-ga-autosubmit]")) {
    target.form && target.form.requestSubmit();
  }
});

document.addEventListener("submit", function (event) {
  var form = event.target;
  if (!form || !form.matches("[data-ga-action-form]")) {
    return;
  }
  var select = form.querySelector("[data-ga-action-select]");
  if (select && select.value) {
    form.action = select.value;
  }
});
