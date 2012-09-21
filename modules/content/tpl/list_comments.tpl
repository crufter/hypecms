{{require admin/header.t}}
{{require content/sidebar.t}}

<h4>Comments</h4>
{{if .comment_list}}
	{{range .comment_list}}
		<div>
			{{if .in_moderation}}
				<span>Awaiting moderation</span>
				{{.content}}
				{{.created_by.guest_name}}
			{{else}}
				Comment:<br />
				{{if is_map ._contents_parent}}
					<a href="/{{._contents_parent.slug}}">{{._contents_parent.title}}</a>
				{{else}}
					Unresolved content.
				{{end}}
			{{end}}
			<br />
		</div>
		<br />
	{{end}}
{{else}}
	No comments yet.
{{end}}


{{require content/footer.t}}
{{require admin/footer.t}}