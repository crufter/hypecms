{{require header.t}}

{{if .error}}
	{{.error}}
{{else}}
	{{range .tag_list}}
		{{.name}}<br />
	{{end}}
{{end}}

{{require footer.t}}