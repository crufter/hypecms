{{require admin/header.t}}

<link rel="stylesheet" href="/shared/CodeMirror-2.3/lib/codemirror.css">
<script src="/shared/CodeMirror-2.3/lib/codemirror.js"></script>
<script src="/shared/CodeMirror-2.3/mode/clike/clike.js"></script>
<script src="/shared/CodeMirror-2.3/keymap/emacs.js"></script>

<style type="text/css">
	.CodeMirror {border-top: 1px solid #eee; border: 1px solid black;}
	.CodeMirror-scroll {height: 85%;}
</style>

<div>
</div>

<form action="/admin/b/save-config">
<input type="submit">
<textarea id="code" name="option" style="display: block; width: 100%; height: 92%">
{{.admin.options_json}}
</textarea>
</form>

<script src="/tpl/admin/codemirror_init.js"></script>

{{require admin/footer.t}}