document.addEventListener("change", function (event) {
  var target = event.target;
  if (target && target.matches("[data-ga-autosubmit]")) {
    target.form && target.form.requestSubmit();
  }
});
