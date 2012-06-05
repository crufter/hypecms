<div></div>
{{if .error}}
There was an error getting the modules: {{.err}}
{{else}}
	<ul>
	Choose a module to install:<br /><br />
	{{range .admin.modules}}
		<li><a href="/admin/b/install/{{.}}">{{.}}</a></li>
	{{end}}
	</ul>
{{end}}