{{require admin/header.t}}
<div></div>
{{if .error}}
There was an error getting the modules: {{.err}}
{{else}}
	<ul>
	<h3>Choose a module to install:</h3><br />
	{{range .admin.modules}}
		<li><a href="/admin/b/install?module={{.}}">{{.}}</a></li>
	{{end}}
	</ul>
{{end}}
{{require admin/footer.t}}