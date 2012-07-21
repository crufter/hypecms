<form action="/b/content/{{.op}}" method="post">
{{range .fields}}
	{{.key}}<br />
	<input name="{{.key}}" value="{{.value}}" /><br />
	<br />
{{end}}
{{if .tags_on}}
	{{$content_id := .content._id}}
	{{if .content._tags}}
		{{range .content._tags}}
			{{.name}} ({{.count}}) <a href="/b/content/pull_tags?content_id={{$content_id}}&tag_id={{._id}}">x</a> <br /> 
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