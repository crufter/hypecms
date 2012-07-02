<form action="/b/content/{{.op}}" method="post">
{{range .fields}}
	{{.key}}<br />
	<input name="{{.key}}" value="{{.value}}" /><br />
	<br />
{{end}}
<input type="hidden" name="type" value="{{.type}}" />
<input type="hidden" name="id" value="{{.content._id}}" />
<input type="submit">