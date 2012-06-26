{{require admin/header.t}}
<div></div>
{{if .error}}
There was an error getting the modules: {{.err}}
{{else}}
	<ul>
	Choose a module to uninstall:<br /><br />
	{{range .installed_modules}}
		<li><a href="/admin/b/install/{{.}}">{{.}}</a></li>
	{{end}}
	</ul>
{{end}}
{{require admin/footer.t}}