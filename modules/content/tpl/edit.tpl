{{require admin/header.t}}
{{require content/sidebar.t}}

You gotta add/edit your {{.type}} content here.<br />
<br />
<form action="/b/content/{{.op}}" method="post">
{{range .rules}}
{{.field}}<br />
<input name="{{.field}}" value="{{.value}}" /><br />
<br />
{{end}}
<input type="hidden" name="type" value="{{.type}}" />
<input type="submit">

{{require content/footer.t}}
{{require admin/footer.t}}