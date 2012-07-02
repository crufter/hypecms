{{require admin/header.t}}
{{require content/sidebar.t}}

All enries: <br /><br />
{{range .latest}}
	{{require content/listing.t}}
{{end}}

{{require content/footer.t}}
{{require admin/footer.t}}