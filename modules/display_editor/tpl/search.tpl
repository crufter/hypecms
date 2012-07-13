{{require admin/header.t}}
{{require display_editor/sidebar.t}}

{{if .has_points}}
	Search:<br />
	<form action="/admin/display_editor">
		<input name="point-name" value="{{.search}}">
		<input type="submit">
	</form>
	{{range .point_names}}
		<a href="/admin/display_editor/edit/{{.}}">{{.}}</a> <a href="/b/display_editor/delete?name={{.}}">Del</a><br />
	{{else}}
		Nothing matches your search criteria.
	{{end}}
{{else}}
	There are no display points yet.<br />
{{end}}

<br />
<br />
Create new:<br />
<form action="/b/display_editor/new">
	<input name="name">
	<input type="submit">
</form>

{{require admin/footer.t}}