{{require header.t}}
SZeretlek Andiiiiiiiiyiiii!!!<br />
{{if .queries.blog}}
	{{range .queries.blog}}
		<a href="{{._id}}">{{.title}}</a><br />
	{{end}}
{{else}}
	No blog post query.
{{end}}
<br />
{{require footer.t}}