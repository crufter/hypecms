$(function(){

$("a.delete").click(function() {
  return confirm("Are you sure want to delete this?");
});

$("a.delete").html("&#8212;");

var pathname = document.location.pathname;
$("a").each(
function(index, elem) {
	if ($(elem).attr("href") == pathname) {
		$(elem).html("<img class=\"you-are-here\" src=\"/tpl/admin/play_9x12.png\" />" + $(elem).html());
	}
})

if (pathname.split("/").length > 2) {
	var modname = pathname.split("/")[2];
	modname = modname.replace("_", " ")
	$("#left-sidebar").append("<div id=\"module_text\">" + modname + "</div>");
}

})