<div id="left-sidebar">
	<ul>
		<li><a href="/admin/content">List all contents</a></li>
		{{range .content_menu}}
			<li>
			<a href="/admin/content/list/{{.}}">List {{.}} contents</a>
				<ul>
				<li><a href="/admin/content/edit/{{.}}">New {{.}} content</a></li>
				<li><a href="/admin/content/type-config/{{.}}">Configure {{.}} content type</a></li>
				</ul>
			</li>
		{{end}}
		<li><a href="/admin/content/tags">List tags</a></li>
		<li><a href="/admin/content/list-comments">Comments</a></li>
	</ul>
</div>

<div id="inner-content">