{{require admin/header.t}}

Installed modules:

{{if .admin.error}}
	{{.admin.error}}
{{else}}
	<ul>
		{{range .admin.menu}}
		<li>
			<a href="/admin/{{.}}">{{.}}</a>
		</li>
		{{end}}
	</ul>
{{end}}

{{require admin/footer.t}}