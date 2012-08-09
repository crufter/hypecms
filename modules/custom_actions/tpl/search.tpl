{{require admin/header.t}}
{{require custom_action/sidebar.t}}

{{if .has_points}}
	Search:<br />
	<form action="/admin/custom_action">
		<input name="point-name" value="{{.search}}">
		<input type="submit">
	</form>
	{{range .point_names}}
		<a class="delete" href="/b/custom_action/delete?name={{.}}">-</a> <a href="/admin/custom_action/edit/{{.}}">{{.}}</a><br />
	{{else}}
		Nothing matches your search criteria.
	{{end}}
{{else}}
	There are no display points yet.<br />
{{end}}

<br />
<br />
Create new:<br />
<form action="/b/custom_action/new">
	<input name="name">
	<input type="submit">
</form>

{{require admin/footer.t}}