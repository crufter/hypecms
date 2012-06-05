<div>
<a href="/admin/edit-config">Edit config</a>
<a href="/admin/b/logout">Logout</a>
</div>
<br />

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