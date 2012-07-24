<div id="left-sidebar">
	<ul>
		<a href="/admin/content">List all contents</a>
		{{range .content_menu}}
			<li>
			<a href="/admin/content/list/{{.}}">List {{.}}</a>
				<ul>
				<li><a href="/admin/content/edit/{{.}}">New {{.}} contents</a></li>
				<li><a href="/admin/content/type-config/{{.}}">Edit {{.}} config</a></li>
				</ul>
			</li>
		{{end}}
		<a href="/admin/content/tags">List tags</a>
	</ul>
</div>

<div id="inner-content">