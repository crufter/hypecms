{{require admin/header.t}}
{{require content/sidebar.t}}

<h4>All enries: </h4>
{{range .latest}}
	{{require content/listing.t}}
{{end}}

{{require content/footer.t}}
{{require admin/footer.t}}