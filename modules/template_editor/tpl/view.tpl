{{require admin/header.t}}

{{if .error}}
	An error occured: {{.error}}
{{else}}
	{{$raw_path := .raw_path}}
	{{if .dir}}
		Your dir ({{.filepath}}):<br />
		<br />
		{{range .dir}}
			<a href="/admin/template_editor/view?file={{$raw_path}}/{{.Name}}">{{.Name}}</a><br />
		{{end}}
	{{end}}
	
	{{if .file}}
		Your file ({{.filepath}}):<br />
		<br />
		<form>
			<textarea cols="90" rows="30">{{.file}}</textarea>
		</form>
	{{end}}
{{end}}

{{require admin/footer.t}}