<br />
Comments:<br />
<br />
{{$con := .content}}
{{if .content.comments}}
	{{range .content.comments}}
		{{.comment_content}} <a href="/b/content/delete_comment?type={{$con.type}}&content_id={{$con._id}}&comment_id={{.comment_id}}">Del</a><br />
	{{end}}
{{else}}
	No comments yet.<br />
{{end}}