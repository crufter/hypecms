{{require header.t}}

<h1>Get:</h1>
{{if .main}}
	{{range .main}}
		<a href="/cars/{{._id}}">{{.name}} {{.content}}</a><br />
		<br />
	{{end}}
{{else}}
	No {{.main_noun}} yet.
{{end}}

{{require footer.t}}