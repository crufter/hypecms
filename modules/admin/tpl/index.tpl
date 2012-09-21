{{require admin/header.t}}

{{if .admin.error}}
	{{.admin.error}}
{{else}}
	<ul>
		<h3>Installed modules:</h3><br />
		{{if .admin.menu}}
			{{range .admin.menu}}
			<li>
				<a href="/admin/{{.}}">{{.}}</a>
			</li>
			{{end}}
		{{else}}
			<li>
				No installed modules yet. <a href="/admin/install">Click here to install one</a>.
			</li>
		{{end}}
	</ul>
{{end}}

{{require admin/footer.t}}