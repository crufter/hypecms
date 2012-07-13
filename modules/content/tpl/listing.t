{{if .title}}
	<a href="/admin/content/edit/{{.type}}/{{._id}}">{{.title}}</a>
{{else}}
	<a href="/admin/content/edit/{{.type}}/{{._id}}">{{.name}}</a>
{{end}}&nbsp;&nbsp;&nbsp;
<a href="/b/content/delete?id={{._id}}&type={{.type}}" />Delete</a><br />