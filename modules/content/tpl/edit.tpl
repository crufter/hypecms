{{require admin/header.t}}
{{require content/sidebar.t}}

<h4>{{if .content._id}}Edit{{else}}Insert{{end}} {{.type}} content</h4>
<br />
{{require content/edit-form.t}}

{{require content/footer.t}}
{{require admin/footer.t}}