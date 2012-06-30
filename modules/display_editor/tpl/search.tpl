{{require admin/header.t}}
{{range .point_names}}
	<a href="/admin/display_editor/{{.}}">{{.}}</a>
{{end}}
{{require admin/footer.t}}