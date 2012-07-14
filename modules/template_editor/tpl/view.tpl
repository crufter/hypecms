{{require admin/header.t}}

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
			<textarea cols="90" rows="30">{{.file}}</textarea>
		</form>
	{{end}}
{{end}}

{{require admin/footer.t}}