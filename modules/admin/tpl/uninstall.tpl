{{require admin/header.t}}
<div></div>
{{if .error}}
There was an error getting the modules: {{.err}}
{{else}}
	<ul>
	<h3>Choose a module to uninstall:</h3><br />
	{{range .installed_modules}}
		<li><a href="/admin/b/uninstall/{{.}}">{{.}}</a></li>
	{{end}}
	</ul>
{{end}}
{{require admin/footer.t}}