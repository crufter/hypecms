{{if .title}}
	<a href="/admin/content/edit/{{.type}}/{{._id}}">{{.title}}</a>
{{else}}
	<a href="/admin/content/edit/{{.type}}/{{._id}}">{{.name}}</a>
{{end}}
<form action="/b/content/delete" style="padding:0; margin:0;"><input type="hidden" name="id" value="{{._id}}" /><input type="submit"></form><br />