{{require admin/header.t}}
{{require template_editor/sidebar.t}}

You are using the <b>{{if .can_modify}}private{{else}}public{{end}}</b> template named <b>{{.template_name}}</b>.<br />
<a href="/admin/template_editor/view?file=">Here</a> you can start viewing your template files.<br />
<br />
{{if .can_modify}}
	If you want to publish this private template of yours, so others can use and improve it, name the template and share it:
	<form action="/b/template_editor/publish_private" method="post">
		<input name="public_name">
		<input type="submit">
	</form>
{{else}}
	If you want to fork this private template so you can modify the create/modify/delete files and folders in it, click here.
{{end}}

{{require template_editor/footer.t}}
{{require admin/footer.t}}
