<div style="float: left; width: 20%;">
	<ul>
		<a href="/admin/content">All</a>
		{{range .content_menu}}
			<li>
			<a href="/admin/content/list/{{.}}">{{.}}</a>
				<ul>
				<li><a href="/admin/content/edit/{{.}}">New</a></li>
				<li><a href="/admin/content/type-config/{{.}}">Config</a></li>
				</ul>
			</li>
		{{end}}
	</ul>
</div>

<div style="float: right; width: 80%;">