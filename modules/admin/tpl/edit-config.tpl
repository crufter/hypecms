{{require admin/header.t}}
<div>
</div>

<form action="/admin/b/save-config">
<input type="submit">
<textarea name="option" style="display: block; width: 100%; height: 92%">
{{.admin.options_json}}
</textarea>
</form>
{{require admin/footer.t}}