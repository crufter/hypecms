{{require admin/header.t}}
{{require display_editor/sidebar.t}}
<form action="/b/display_editor/save" method="post">
	Name:<br/>
	<input name="name" value="{{.point.name}}"><br/>
	<input type="hidden" name="prev_name" value="{{.point.name}}"><br/>
	<br/>
	Query:<br />
	<textarea name="queries" rows="30" cols="75">{{.point.queries}}</textarea><br />
	<br />
	<input type="submit">
</form>
{{require admin/footer.t}}