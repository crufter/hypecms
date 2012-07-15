{{require admin/header.t}}

<link rel="stylesheet" href="/shared/CodeMirror-2.3/lib/codemirror.css">
<script src="/shared/CodeMirror-2.3/lib/codemirror.js"></script>
<script src="/shared/CodeMirror-2.3/lib/util/overlay.js"></script>
<script src="/shared/CodeMirror-2.3/mode/xml/xml.js"></script>

{{if .error}}
	An error occured: {{.error}}
{{else}}
	{{$raw_path := .raw_path}}
	<a href="/admin/template_editor/view?file=">root</a>/
	{{range .breadcrumb}}
		<a href="/admin/template_editor/view?file={{.Path}}">{{.Name}}</a>/
	{{end}}
	<br />
	<br />
	{{if .dir}}
		{{range .dir}}
			<a href="/admin/template_editor/view?file={{$raw_path}}/{{.Name}}">{{.Name}}</a><br />
		{{end}}
	{{end}}
	
	{{if .file}}
		<form>
			<textarea id="code" cols="90" rows="30">{{.file}}</textarea>
		</form>
	{{end}}
{{end}}

<script src="/tpl/template_editor/codemirror_init.js"></script>

{{require admin/footer.t}}