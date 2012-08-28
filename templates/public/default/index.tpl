{{require header.t}}
Index<br />
{{if .queries.blog}}
	{{range .queries.blog}}
		<a href="{{._id}}">{{.title}}</a><br />
	{{end}}
	{{$navi := .queries.blog_navi}}
	{{require admin/navi.t}}
{{else}}
	No blog post query.
{{end}}
<br />
{{require footer.t}}