<a class="delete" href="/b/content/delete?id={{._id}}&type={{.type}}" />-</a>
{{if .title}}
	<a href="/admin/content/edit/{{.type}}/{{._id}}">{{.title}}</a>
{{else}}
	<a href="/admin/content/edit/{{.type}}/{{._id}}">{{.name}}</a>
{{end}}
<br />