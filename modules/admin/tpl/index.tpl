{{require admin/header.t}}

{{if .admin.error}}
	{{.admin.error}}
{{else}}
	<ul>
		<h3>Installed modules:</h3><br />
		{{range .admin.menu}}
		<li>
			<a href="/admin/{{.}}">{{.}}</a>
		</li>
		{{end}}
	</ul>
{{end}}

{{require admin/footer.t}}