{{require header.t}}
Index<br />
{{if .queries.blog}}
	{{range .queries.blog}}
		<a href="{{._id}}">{{.title}}</a><br />
	{{end}}
	{{range .queries.blog_navi}}
		<a href="{{.Url}}">{{.Page}}</a> 
	{{end}}
{{else}}
	No blog post query.
{{end}}
<br />
{{require footer.t}}