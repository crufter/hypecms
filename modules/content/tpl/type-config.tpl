{{require admin/header.t}}
{{require content/sidebar.t}}
<h4>Configure {{.type}} content type.</h4>
(This page is not implemented yet.)<br />
<h5>These options modify behavior for every admin: </h5>

{{$top := .user_type_op}}
<h5>These options modify behavior for you only: </h5>
<form action="/b/content/save_type_config?type={{.type}}">
	<input name="content_safe_delete" type="checkbox" {{if $top.safe_delete_content}}CHECKED{{end}}><b>Safe delete contents</b><br />
	Check this in if you want the system to ask for a confirmation when deleting {{.type}} contents.<br />
	<br />
	
	<input name="tag_safe_delete" type="checkbox" {{if $top.safe_delete_tag_editor}}CHECKED{{end}}><b>Safe delete tags in editor</b><br />
	Check this in if you want the system to ask for a confirmation when deleting tags in the content editor.<br />
	<br />
	
	<input name="tag_safe_delete" type="checkbox" {{if $top.safe_delete_in_tag_list}}CHECKED{{end}}><b>Safe delete tags in listing</b><br />
	Check this in if you want the system to ask for a confirmation when deleting tags in the tag listing.<br />
	<br />
	
	<input type="submit">
</form>
{{require content/footer.t}}
{{require admin/footer.t}}