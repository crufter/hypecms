<form action="/b/content/{{.op}}" method="post">
{{range .fields}}
	{{.fieldname}}<br />
	<input name="{{.fieldname}}" value="{{.value}}" /><br />
	<br />
{{end}}
<input type="hidden" name="type" value="{{.type}}" />
<input type="hidden" name="type" value="{{.content._id}}" />
<input type="submit">