{{require admin/adminheader.t}}

Installed modules:

{{if .admin.error}}
	{{.admin.error}}
{{else}}
	<ul>
		<li>
		{{range .admin.menu}}
			<a href="/admin/{{.}}">{{.}}</a>
		{{end}}
		</li>
	</ul>
{{end}}