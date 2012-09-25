<div id="left-sidebar">
	<ul>
		<li><a href="/admin/content">All contents</a></li>
		{{range .content_menu}}
			<li>
			<a href="/admin/content/type?type={{.}}">{{.}} contents</a>
				<ul>
				<li><a href="/admin/content/edit?type={{.}}">New {{.}} content</a></li>
				<li><a href="/admin/content/type-config?type={{.}}">Configure {{.}} content type</a></li>
				</ul>
			</li>
		{{end}}
		<li><a href="/admin/content/tags">Tags</a></li>
		<li><a href="/admin/content/comments">Comments</a></li>
	</ul>
</div>

<div id="inner-content">