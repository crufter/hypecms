{{require admin/header.t}}
{{require template_editor/sidebar.t}}

{{if .error}}
	There was an error: {{.error}}
{{else}}
	{{range .dir}}
		{{.}}<br />
	{{end}}
{{end}}

{{require template_editor/footer.t}}
{{require admin/footer.t}}