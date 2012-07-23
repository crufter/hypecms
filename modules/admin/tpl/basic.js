$(function(){

$("a.delete").click(function() {
  return confirm("Are you sure want to delete this?");
});

$("a.delete").html("&#8212;")

})