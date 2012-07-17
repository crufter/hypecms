{{require admin/header.t}}
{{if .is_dir}}
	{{require template_editor/sidebar.t}}
{{end}}

<link rel="stylesheet" href="/shared/CodeMirror-2.3/lib/codemirror.css">
<script src="/shared/CodeMirror-2.3/lib/codemirror.js"></script>
<script src="/shared/CodeMirror-2.3/lib/util/overlay.js"></script>
<script src="/shared/CodeMirror-2.3/mode/xml/xml.js"></script>

<style type="text/css">
	.CodeMirror-scroll {height: 80%;}
	.CodeMirror {border: 1px solid black;}
	.cm-mustache {color: #004; font-weight: bold}
</style>

{{if .error}}
	An error occured: {{.error}}
{{else}}
	{{$can_mod := .can_modify}}
	{{$raw_path := .raw_path}}
	{{if $can_mod}}private{{else}}public{{end}} &nbsp;&nbsp;<a href="/admin/template_editor/view?file=">{{.template_name}}</a>/
	{{range .breadcrumb}}
		<a href="/admin/template_editor/view?file={{.Path}}">{{.Name}}</a>/
	{{end}}
	<br />
	<br />
	
	{{if .is_dir}}
		{{if $can_mod}}
			Create new file/dir: <form action="/b/template_editor/new_file"><input type="hidden" name="where" value="{{$raw_path}}"><input name="filepath"><input type="submit"></form>
		{{end}}
		{{range .dir}}
			{{if $can_mod}}
				<a href="/b/template_editor/delete_file?filepath={{$raw_path}}/{{.Name}}" title="Delete">-</a>&nbsp;&nbsp;
			{{end}}
			<a href="/admin/template_editor/view?file={{$raw_path}}/{{.Name}}">{{.Name}}</a>
			<br />
		{{end}}
	{{end}}
	
	{{if .file}}
		<form action="/b/template_editor/save_file">
			<textarea name="content" id="code" cols="90" rows="30">{{.file}}</textarea>
			{{if $can_mod}}
				<input type="hidden" name="filepath" value="{{$raw_path}}"><br />
				<input type="submit">
			{{else}}
				<br />
				You can not modify this file because it is part of a public template. <a href="/b/template_editor/fork_public">Make a private template out of this by forking.</a>
			{{end}}
		</form>
	{{end}}
{{end}}

<script src="/tpl/template_editor/codemirror_init.js"></script>

{{if .is_dir}}
	{{require template_editor/footer.t}}
{{end}}
{{require admin/footer.t}}