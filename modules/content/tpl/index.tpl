{{require admin/header.t}}
{{require content/sidebar.t}}

{{range .latest}}
	{{if .title}}
		{{.title}}
	{{else}}
		{{.name}}
	{{end}}<br />
{{end}}

{{require content/footer.t}}
{{require admin/footer.t}}