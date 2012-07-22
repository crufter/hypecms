{{require header.t}}

{{if .error}}
	{{.error}}
{{else}}
	{{range .content_list}}
		{{.title}}<br />
	{{end}}
{{end}}

{{require footer.t}}