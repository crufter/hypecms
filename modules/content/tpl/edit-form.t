<form action="/b/content/{{.op}}" method="post">
{{range .fields}}
	{{.key}}<br />
	<input name="{{.key}}" value="{{.value}}" /><br />
	<br />
{{end}}
{{if .tags_on}}
	{{if .content._tags}}
		{{range .content._tags}}
			{{.name}}<br />
		{{end}}
		<br />
	{{else}}
		No tags yet.<br />
		<br />
	{{end}}
{{end}}
<input type="hidden" name="type" value="{{.type}}" />
<input type="hidden" name="id" value="{{.content._id}}" />
<input type="submit">