{{require admin/header.t}}
{{require template_editor/sidebar.t}}

You are using the <b>{{if .can_modify}}private{{else}}public{{end}}</b> template named <b>{{.template_name}}</b>.<br />
<br />
<h3>Start browsing</h3>
<a href="/admin/template_editor/view?file=">Here</a> you can start viewing your template files.<br />
<br />
{{if .can_modify}}
	<h3>Publish this private template</h3>
	If you want to publish this private template of yours, so others can use and improve it, name the template and share it:
	<form action="/b/template_editor/publish_private" method="post">
		<input name="public_name" value="{{.template_name}}">
		<input type="submit">
	</form>
	<br />
	<h3>Fork this private template</h3>
	If you want to fork this private template into an other private template (thus creating a backup), you can do it here:
	<form action="/b/template_editor/fork_private" method="post">
		<input name="new_template_name">
		<input type="submit">
	</form>
{{else}}
	<h3>Fork this public template</h3>
	<a href="/b/template_editor/fork_public">Click here</a> if you want to fork this public template (thus creating a private one out of it) so you can create/modify/delete files and folders in it.
{{end}}

{{require template_editor/footer.t}}
{{require admin/footer.t}}
