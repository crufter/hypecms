{{require admin/header.t}}
{{if .is_dir}}
	{{require template_editor/sidebar.t}}
{{end}}

<!-- Including codemirror -->
<link rel="stylesheet" href="/shared/CodeMirror-2.3/lib/codemirror.css">
<script src="/shared/CodeMirror-2.3/lib/codemirror.js"></script>
<script src="/shared/CodeMirror-2.3/lib/util/overlay.js"></script>
<script src="/shared/CodeMirror-2.3/mode/xml/xml.js"></script>

<style type="text/css">
	.CodeMirror-scroll {height: 60%;}
	.CodeMirror {border: 1px solid black;}
	.cm-mustache {color: #004; font-weight: bold}
</style>
<!-- /Including codemirror -->

{{$current := .current}}
{{$typ := .typ}}
{{$name := .template_name}}
{{$included := .included}}

{{if .error}}
	An error occured: {{.error}}
{{else}}
	{{$can_mod := .can_modify}}
	{{$raw_path := .raw_path}}
	{{$typ}} &nbsp;&nbsp;
	
	<!-- Breadcrumb. -->
	{{if $current}}
		<a href="/admin/template_editor/view?file=">
	{{else}}
		<a href="/admin/template_editor/view/{{$typ}}/{{$name}}?file=">
	{{end}}
	{{.template_name}}</a>/
	
	{{range .breadcrumb}}
		{{if $current}}
			<a href="/admin/template_editor/view?file={{.Path}}">{{.Name}}</a>/
		{{else}}
			<a href="/admin/template_editor/view/{{$typ}}/{{$name}}?file={{.Path}}">{{.Name}}</a>/
		{{end}}
	{{end}}
	<br />
	<br />
	<!-- /Breadcrumb. -->
	
	<!-- Directory listing. -->
	{{if .is_dir}}
		{{if $can_mod}}
			{{if $current}}
				Create new file/dir: <form action="/b/template_editor/new_file"><input type="hidden" name="where" value="{{$raw_path}}"><input name="filepath"><input type="submit"></form>
			{{else}}
				<!-- Soon you will be able to create files in noncurrent templates too. -> <!-- TODO -->
			{{end}}
		{{end}}
		{{range .dir}}
			{{if $can_mod}}
				{{if $current}}
					<a class="delete" href="/b/template_editor/delete_file?filepath={{$raw_path}}/{{.Name}}" title="Delete">-</a>&nbsp;&nbsp;
				{{else}}
					<!-- Soon you will be able to delete files in noncurrent templates too. -> <!-- TODO -->
				{{end}}
			{{end}}
			{{if $current}}
				<a href="/admin/template_editor/view?file={{$raw_path}}/{{.Name}}">{{.Name}}</a>
			{{else}}
				<a href="/admin/template_editor/view/{{$typ}}/{{$name}}?file={{$raw_path}}/{{.Name}}">{{.Name}}</a>
			{{end}}
			<br />
		{{end}}
	{{end}}
	<!-- /Directory listing. -->
	
	<!-- File editing. -->
	{{if .file}}
		<form action="/b/template_editor/save_file">
			<textarea name="content" id="code" cols="90" rows="30">{{.file}}</textarea>
			{{if $can_mod}}
				{{if $current}}
				{{else}}
					<!--TODO: include additional parameters here too. -->
				{{end}}
				<input type="hidden" name="filepath" value="{{$raw_path}}"><br />
				<input type="submit">
			{{else}}
				<br />
				You can not modify this file because it is part of a public template. <a href="/b/template_editor/fork_public">Make a private template out of this by forking.</a>
			{{end}}
		</form>
		<h3>Included files:</h3>
		{{if .included}}
			{{range .included}}
				<a href="/admin/template_editor/view/{{.Typ}}/{{.Tempname}}?file=/{{.Filepath}}">{{.Filepath}}</a><br />
			{{end}}
		{{else}}
			None.<br />
		{{end}}
	{{end}}
	<!-- /File editing. -->
{{end}}

<script src="/tpl/template_editor/codemirror_init.js"></script>

{{if .is_dir}}
	{{require template_editor/footer.t}}
{{end}}
{{require admin/footer.t}}